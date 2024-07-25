package talkingstick

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"sync"
	"time"
)

type Session interface {
	Start()
	End()
	Pass() bool
	Close()
}

type tsSession struct {
	channelID string
	active    *tsMember
	prev      *tsMember

	sess     *discordgo.Session
	duration time.Duration
	ticker   *time.Ticker
	messages []*discordgo.Message

	mu         sync.Mutex
	shutdownCh chan struct{}

	// selfDestruct removes the session from the session manager
	selfDestruct func()
}

func newTSSession(s *discordgo.Session, channelID string, duration time.Duration, members *tsMember, selfDestructFn func()) *tsSession {
	return &tsSession{
		sess:         s,
		channelID:    channelID,
		active:       members,
		prev:         nil,
		duration:     duration,
		ticker:       time.NewTicker(duration),
		messages:     make([]*discordgo.Message, 0),
		mu:           sync.Mutex{},
		shutdownCh:   make(chan struct{}),
		selfDestruct: selfDestructFn,
	}
}

func (tss *tsSession) Start() {
	tss.Pass()
	for {
		select {
		case <-tss.shutdownCh:
			return
		case <-tss.ticker.C:
			if !tss.Pass() {
				return
			}
		}
	}
}

func (tss *tsSession) Pass() bool {
	if tss.active == nil {
		return false
	}

	tss.mu.Lock()
	defer tss.mu.Unlock()

	msg := fmt.Sprintf("Passing the talking stick to %s", tss.active.data.User.Mention())
	ttsMessage, err := tss.sess.ChannelMessageSendTTS(tss.channelID, msg)
	if err != nil {
		slog.Error("failed to send tts message", "channel_id", tss.channelID, "error", err)
	} else {
		tss.messages = append(tss.messages, ttsMessage)
	}

	if err = tss.sess.ChannelPermissionSet(tss.channelID, tss.active.data.User.ID,
		discordgo.PermissionOverwriteTypeMember, discordgo.PermissionVoicePrioritySpeaker, 0); err != nil {
		slog.Error("Failed to set priority speaker", "error", err, "user", tss.active.data.User.Username)
	}

	tss.prev = tss.active
	tss.active = tss.active.next
	tss.ticker.Reset(tss.duration)
	return true
}

func (tss *tsSession) End() {
	close(tss.shutdownCh)
}

func (tss *tsSession) Close() {
	defer tss.selfDestruct()

	tss.mu.Lock()
	defer tss.mu.Unlock()

	tss.ticker.Stop()

	if tss.prev != nil {
		if err := tss.sess.ChannelPermissionSet(tss.channelID, tss.prev.data.User.ID,
			discordgo.PermissionOverwriteTypeMember, 0, discordgo.PermissionVoicePrioritySpeaker); err != nil {
			slog.Error("Failed to remove priority speaker", "error", err, "user", tss.active.data.User.Username)
		}
	}

	closingMsg, err := tss.sess.ChannelMessageSendTTS(tss.channelID, "The talking stick session has ended")
	if err != nil {
		slog.Error("failed to send closing tts message", "channel_id", tss.channelID, "error", err)
	} else {
		tss.messages = append(tss.messages, closingMsg)
	}

	for _, message := range tss.messages {
		if message.TTS && time.Since(message.Timestamp) < 5*time.Second {
			time.Sleep(5*time.Second - time.Since(message.Timestamp))
		}
		if err = tss.sess.ChannelMessageDelete(message.ChannelID, message.ID); err != nil {
			slog.Error("failed to delete message", "message_id", err)
		}
	}
}
