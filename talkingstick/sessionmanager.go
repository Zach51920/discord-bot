package talkingstick

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"sync"
	"time"
)

type Action string

var (
	ActionTogglePlayPause Action = "ts_toggle_play_pause"
	ActionSkipUser        Action = "ts_skip_user"
	ActionQuitSession     Action = "ts_quit_session"
)

var ErrSessionExists = errors.New("a talking stick session already exists")
var ErrSessionNotFound = errors.New("channel has no active talking stick session")
var ErrUnknownAction = errors.New("unknown action")

type SessionManager interface {
	// Create a new talking stick session
	Create(guildID, channelID string, duration time.Duration) error
	// Handle a request action
	Handle(channelID string, action Action) error
	// Close all running sessions. Blocks until all sessions are finished closing
	Close() error
}

type SessManager struct {
	mu         *sync.Mutex
	wg         *sync.WaitGroup
	sess       *discordgo.Session
	tsSessions map[string]*tsSession
}

func NewSessionManager(s *discordgo.Session) SessionManager {
	return &SessManager{
		sess:       s,
		mu:         &sync.Mutex{},
		wg:         &sync.WaitGroup{},
		tsSessions: make(map[string]*tsSession),
	}
}

func (s *SessManager) Create(guildID, channelID string, duration time.Duration) error {
	// check if a session already exists
	if tss := s.getSession(channelID); tss != nil {
		return ErrSessionExists
	}

	// load the voice channels members
	members := loadVoiceMembers(s.sess, guildID, channelID)
	shuffleDGMembers(members)
	head := newMemberList(members)

	// create a new session
	tss := newTSSession(s.sess, channelID, duration, head)
	if err := tss.CreateControlPanel(); err != nil {
		return fmt.Errorf("failed to create control panel: %w", err)
	}
	s.register(tss)
	go s.launch(tss)
	return nil
}

func (s *SessManager) Handle(channelID string, action Action) error {
	slog.Debug("received handle action request", "channel_id", channelID, "action", action)

	tss := s.getSession(channelID)
	if tss == nil {
		return ErrSessionNotFound
	}
	defer tss.resetTimer()

	actions := map[Action]func(){
		ActionQuitSession:     tss.Quit,
		ActionSkipUser:        func() { tss.Pass(nil) },
		ActionTogglePlayPause: s.togglePlayPauseHandler(tss),
	}

	handler, ok := actions[action]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownAction, action)
	}

	handler()
	return nil
}

func (s *SessManager) Close() error {
	for _, tss := range s.tsSessions {
		tss.Quit()
	}
	s.wg.Wait()
	return nil
}

func (s *SessManager) togglePlayPauseHandler(tss *tsSession) func() {
	return func() {
		if tss.Running() {
			tss.Pause()
		} else {
			tss.Play()
		}
		tss.RefreshControlPanel()
	}
}
func (s *SessManager) getSession(channelID string) *tsSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.tsSessions[channelID]
}

func (s *SessManager) launch(tss *tsSession) {
	defer s.unregister(tss.channelID)
	tss.Start()
}

func (s *SessManager) register(tss *tsSession) {
	slog.Debug("registering TS session", "channel_id", tss.channelID)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.tsSessions[tss.channelID] = tss
	s.wg.Add(1)
}

func (s *SessManager) unregister(channelID string) {
	slog.Debug("unregistering TS session", "channel_id", channelID)

	s.mu.Lock()
	defer s.mu.Unlock()
	tss, ok := s.tsSessions[channelID]
	if ok {
		tss.Quit()
	}
	delete(s.tsSessions, channelID)
	s.wg.Done()
}
