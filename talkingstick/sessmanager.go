package talkingstick

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"sync"
	"time"
)

var ErrNoSession = errors.New("no session exists for channel")
var ErrSessionExists = errors.New("a session already exists for channel")
var ErrSessionInactive = errors.New("the session is inactive")
var ErrDuplicateMember = errors.New("user is already a member of the session")
var ErrMemberNotInChannel = errors.New("user is not a member of the current voice channel")
var ErrMemberNotInSession = errors.New("user is not a member of the session")

type SessionManager interface {
	// Create a new talking stick session. This does NOT automatically start the session
	Create(guildID, channelID string, duration time.Duration) error

	// Start the talking stick session for the specified channel. If no session exists, an error is returned
	Start(channelID string) error

	// Pass the talking stick to a specified user
	Pass(channelID, userID string) error

	// Skip the remainder of the current members turn
	Skip(channelID string) error

	// Pause the talking stick session if it's ongoing. If no session exists, an error is returned

	// AddMember adds a member to the talking stick session in a random location

	// RemoveMember removes a member from the talking stick session

	// End the session for the specified channel
	End(channelID string) error
}

type SessManager struct {
	mu         *sync.Mutex
	sess       *discordgo.Session
	tsSessions map[string]session
}

func NewSessionManager(s *discordgo.Session) SessionManager {
	return &SessManager{sess: s, mu: &sync.Mutex{}, tsSessions: make(map[string]session)}
}

func (s *SessManager) Create(guildID, channelID string, duration time.Duration) error {
	slog.Debug("creating talking stick session", "channel_id", channelID)

	// check if a session is already ongoing
	if tss := s.getSession(channelID); tss != nil {
		return ErrSessionExists
	}

	// get the channel members
	mem := loadVoiceMembers(s.sess, guildID, channelID)
	shuffleDGMembers(mem)
	memList := newMemberList(mem)

	// create the session
	tss := newTSSession(s.sess, channelID, duration, memList)
	s.addSession(tss)
	return nil
}

func (s *SessManager) Start(channelID string) error {
	tss := s.getSession(channelID)
	if tss == nil {
		return ErrNoSession
	}
	// start the session if it's not currently running
	if !tss.Running() {
		go tss.Start()
	}
	return nil
}

func (s *SessManager) End(channelID string) error {
	tss := s.getSession(channelID)
	if tss == nil {
		return ErrNoSession
	}
	tss.Stop()
	tss.Close()
	s.delSession(channelID)
	return nil
}

func (s *SessManager) Skip(channelID string) error {
	tss := s.getSession(channelID)
	if tss == nil {
		return ErrNoSession
	}
	if !tss.Running() {
		return ErrSessionInactive
	}
	tss.Pass(nil)
	return nil
}

func (s *SessManager) Pass(channelID, userID string) error {
	tss := s.getSession(channelID)
	if tss == nil {
		return ErrNoSession
	}
	if !tss.Running() {
		return ErrSessionInactive
	}
	member, ok := getMember(tss.Active(), userID)
	if !ok {
		return ErrMemberNotInSession
	}
	tss.Pass(member)
	return nil
}

func (s *SessManager) getSession(channelID string) session {
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
