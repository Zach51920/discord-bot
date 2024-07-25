package talkingstick

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"sync"
	"time"
)

type SessionManager interface {
	Run(channelID string) error
	Get(channelID string) Session
	Create(guildID, channelID string, duration time.Duration) error
}

type SessManager struct {
	mu         *sync.Mutex
	sess       *discordgo.Session
	tsSessions map[string]*tsSession
}

func NewSessionManager(s *discordgo.Session) SessionManager {
	return &SessManager{sess: s, mu: &sync.Mutex{}, tsSessions: make(map[string]*tsSession)}
}

func (s *SessManager) Get(channelID string) Session {
	return s.getSession(channelID)
}

func (s *SessManager) Create(guildID, channelID string, duration time.Duration) error {
	slog.Debug("creating talking stick session", "channel_id", channelID)

	// check if a session is already ongoing
	if tss := s.getSession(channelID); tss != nil {
		return errors.New("There is already an active talking stick session in your voice channel")
	}

	// get the channel members
	mem := loadVoiceMembers(s.sess, guildID, channelID)
	shuffleDGMembers(mem)
	memList := newMemberList(mem)

	// create the session
	tss := newTSSession(s.sess, channelID, duration, memList, s.getSelfDestructFn(channelID))
	s.addSession(tss)
	return nil
}

func (s *SessManager) Run(channelID string) error {
	slog.Debug("running talking stick session", "channel_id", channelID)

	tss := s.getSession(channelID)
	if tss == nil {
		return errors.New("There are no active talking stick sessions in your current channel")
	}
	go func() {
		defer tss.Close()
		tss.Start()
	}()
	return nil
}

func (s *SessManager) getSession(channelID string) *tsSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.tsSessions[channelID]
}

func (s *SessManager) addSession(tss *tsSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tsSessions[tss.channelID] = tss
}

func (s *SessManager) delSession(channelID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tsSessions, channelID)
}

func (s *SessManager) getSelfDestructFn(channelID string) func() {
	return func() { s.delSession(channelID) }
}
