package talkingstick

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"sync"
	"time"
)

type session interface {
	Start()
	Stop()
	Close()
	Pass(target *tsMember) bool
	Running() bool
	Active() *tsMember
}

type tsSession struct {
	channelID string
	active    *tsMember
	prev      *tsMember

	sess     *discordgo.Session
	duration time.Duration
	ticker   *time.Ticker
	messages []*discordgo.Message

	mu     sync.Mutex
	stopCh chan struct{}

	isRunning bool
}

func newTSSession(s *discordgo.Session, channelID string, duration time.Duration, members *tsMember) *tsSession {
	return &tsSession{
		sess:      s,
		channelID: channelID,
		active:    members,
		duration:  duration,
		ticker:    time.NewTicker(duration),
		messages:  make([]*discordgo.Message, 0),
		mu:        sync.Mutex{},
		stopCh:    make(chan struct{}),
	}
}

func (tss *tsSession) Start() {
	tss.SetRunning(true)
	defer tss.SetRunning(false)

	// recreate the stopCh
	tss.stopCh = make(chan struct{})

	tss.Pass(nil)
	for {
		select {
		case <-tss.stopCh:
			return
		case <-tss.ticker.C:
			if !tss.Pass(nil) {
				return
			}
		}
	}
}

func (tss *tsSession) Pass(target *tsMember) bool {
	tss.prev = tss.active
	if target != nil {
		tss.active = target
	} else {
		tss.active = tss.active.next
	}
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

	tss.ticker.Reset(tss.duration)
	return true
}

func (tss *tsSession) Close() {
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

func (tss *tsSession) Stop() {
	if tss.Running() {
		close(tss.stopCh)
	}
}

func (tss *tsSession) Running() bool {
	tss.mu.Lock()
	defer tss.mu.Unlock()
	return tss.isRunning
}

func (tss *tsSession) SetRunning(isRunning bool) {
	tss.mu.Lock()
	defer tss.mu.Unlock()
	tss.isRunning = isRunning
}

func (tss *tsSession) Active() *tsMember {
	return tss.active
}
