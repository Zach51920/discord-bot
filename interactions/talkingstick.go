package interactions

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"math/rand"
	"sync"
	"time"
)

type linkedMember struct {
	data *discordgo.Member
	next *linkedMember
}

type TalkingStickSess struct {
	active    *linkedMember
	prev      *linkedMember
	duration  time.Duration
	channelID string
	messages  []*discordgo.Message
	ticker    *time.Ticker
	mu        sync.Mutex
	done      chan struct{}
}

func NewStickSession(s *discordgo.Session, guildID, channelID string, duration time.Duration) *TalkingStickSess {
	members := loadMembers(s, guildID, channelID)
	shuffledMembers := newLinkedMembers(members)
	return &TalkingStickSess{
		active:    shuffledMembers,
		channelID: channelID,
		duration:  duration,
		ticker:    time.NewTicker(duration),
		messages:  make([]*discordgo.Message, 0),
		done:      make(chan struct{}),
	}
}

func (tss *TalkingStickSess) PassStick(s *discordgo.Session) bool {
	if tss.active == nil {
		return false
	}

	tss.mu.Lock()
	defer tss.mu.Unlock()

	msg := fmt.Sprintf("Passing the talking stick to %s", tss.active.data.User.Mention())
	ttsMessage, err := s.ChannelMessageSendTTS(tss.channelID, msg)
	if err != nil {
		slog.Error("failed to send tts message", "channel_id", tss.channelID, "error", err)
	} else {
		tss.messages = append(tss.messages, ttsMessage)
	}

	if err = s.ChannelPermissionSet(tss.channelID, tss.active.data.User.ID,
		discordgo.PermissionOverwriteTypeMember, discordgo.PermissionVoicePrioritySpeaker, 0); err != nil {
		slog.Error("Failed to set priority speaker", "error", err, "user", tss.active.data.User.Username)
	}

	tss.prev = tss.active
	tss.active = tss.active.next
	tss.ticker.Reset(tss.duration)
	return true
}

func (tss *TalkingStickSess) Start(s *discordgo.Session) {
	defer tss.Close(s)

	tss.PassStick(s)
	for {
		select {
		case <-tss.done:
			return
		case <-tss.ticker.C:
			if !tss.PassStick(s) {
				return
			}
		}
	}
}

func (tss *TalkingStickSess) Close(s *discordgo.Session) {
	tss.mu.Lock()
	defer tss.mu.Unlock()

	close(tss.done)
	tss.ticker.Stop()

	if tss.prev != nil {
		if err := s.ChannelPermissionSet(tss.channelID, tss.prev.data.User.ID,
			discordgo.PermissionOverwriteTypeMember, 0, discordgo.PermissionVoicePrioritySpeaker); err != nil {
			slog.Error("Failed to remove priority speaker", "error", err, "user", tss.active.data.User.Username)
		}
	}

	closingMsg, err := s.ChannelMessageSendTTS(tss.channelID, "The talking stick session has ended")
	if err != nil {
		slog.Error("failed to send closing tts message", "channel_id", tss.channelID, "error", err)
	} else {
		tss.messages = append(tss.messages, closingMsg)
	}

	for _, message := range tss.messages {
		if message.TTS && time.Since(message.Timestamp) < 5*time.Second {
			time.Sleep(5*time.Second - time.Since(message.Timestamp))
		}
		if err = s.ChannelMessageDelete(message.ChannelID, message.ID); err != nil {
			slog.Error("failed to delete message", "message_id", err)
		}
	}
}

func loadMembers(s *discordgo.Session, guildID, channelID string) []*discordgo.Member {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		slog.Error("failed to access guild state", "error", err)
		return nil
	}
	var members []*discordgo.Member
	for _, vs := range guild.VoiceStates {
		if vs.ChannelID == channelID {
			member, err := s.GuildMember(guildID, vs.UserID)
			if err != nil {
				slog.Error("failed to fetch member", "user_id", vs.UserID, "error", err)
				continue
			}
			members = append(members, member)
		}
	}
	return members
}

func newLinkedMembers(members []*discordgo.Member) *linkedMember {
	if len(members) == 0 {
		return nil
	}
	// shuffle the members
	for i := len(members) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		members[i], members[j] = members[j], members[i]
	}

	head := &linkedMember{data: members[0]}
	current := head
	for i := 1; i < len(members); i++ {
		newNode := &linkedMember{data: members[i]}
		current.next = newNode
		current = newNode
	}
	//current.next = head //  make the list circular
	return head
}

func (h *Handlers) addTalkingStickSess(tss *TalkingStickSess) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.talkingStickChannels[tss.channelID] = tss
}

func (h *Handlers) removeTalkingStickSess(channelID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.talkingStickChannels, channelID)
}

func (h *Handlers) getTalkingStickSess(channelID string) *TalkingStickSess {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.talkingStickChannels[channelID]
}
