package interactions

import (
	"fmt"
	"github.com/Zach51920/discord-bot/youtube"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"math/rand"
	"sync"
	"time"
)

type Handlers struct {
	wg       sync.WaitGroup
	ytClient *youtube.Client
	mu       sync.Mutex

	talkingStickChannels map[string]bool
	shutdownCh           chan struct{}
}

func New() *Handlers {
	return &Handlers{
		ytClient:             youtube.New(),
		wg:                   sync.WaitGroup{},
		shutdownCh:           make(chan struct{}),
		talkingStickChannels: make(map[string]bool)}
}

func (h *Handlers) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	h.wg.Add(1)
	defer h.wg.Done()
	logRequest(i)

	// defer the message response and let the user know we got the request
	if err := acknowledgeRequest(s, i); err != nil {
		slog.Error("failed to acknowledge request: " + err.Error())
		return
	}
	defer ensureFollowup(s, i)

	// handle the request
	data := i.ApplicationCommandData()
	switch data.Name {
	case "yt-download":
		h.Download(s, i)
	case "yt-search":
		h.Search(s, i)
	case "bedtime-ban":
		h.Bedtime(s, i)
	case "talking-stick-start":
		h.TalkingStick(s, i)
	case "coinflip":
		h.CoinFlip(s, i)
	default:
		writeMessage(s, i, "Unknown request command")
	}
}

func (h *Handlers) Search(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := NewRequestOptions(i.ApplicationCommandData().Options)
	query, _ := opts.GetString("query")

	results, err := h.ytClient.Search(query)
	if err != nil {
		slog.Error("failed to search youtube", "error", err)
		return
	}
	embeds := mapYTSearchResults(results)
	writeResponse(s, i, withEmbeds(embeds))
}

func (h *Handlers) Download(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := NewRequestOptions(i.ApplicationCommandData().Options)
	videoID, err := h.getVideoIDFromRequest(opts)
	if err != nil {
		slog.Error("failed to get videoID", "error", err)
		return
	}

	video, err := h.ytClient.Download(videoID)
	if err != nil {
		slog.Error("failed to get video", "error", err)
		return
	}

	files := []*discordgo.File{mapFile(video)}
	writeResponse(s, i, withFiles(files))
}

func (h *Handlers) Bedtime(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := NewRequestOptions(i.ApplicationCommandData().Options)
	user, _ := opts.GetUser(s)
	role, _ := opts.GetRole(s, i)
	reason, ok := opts.GetString("reason")
	if !ok {
		reason = "baby raging"
	}
	if user == nil && role == nil {
		writeMessage(s, i, "Invalid request, all optional parameters cannot be null.")
		return
	}

	s.Lock()
	defer s.Unlock()

	timeout := time.Now().Add(16 * time.Hour)
	if err := s.GuildMemberTimeout(i.GuildID, user.ID, &timeout); err != nil {
		slog.Error("failed to timeout user", "error", err)
		return
	}

	content := fmt.Sprintf("The user %s has been bedtime banned until %v for \"%s\".", user.Username, timeout.String(), reason)
	if _, err := s.ChannelMessageSend(i.ChannelID, content); err != nil {
		slog.Error("failed to send message", "error", err)
		return
	}
}

func (h *Handlers) CoinFlip(s *discordgo.Session, i *discordgo.InteractionCreate) {
	isHeads := rand.Intn(2) == 0
	result := map[bool]struct {
		prefix       string
		defaultTitle string
		optionKey    string
		imageURL     string
	}{
		true: {
			prefix:       "Heads: ",
			defaultTitle: "The coin landed on heads",
			optionKey:    "landed-heads",
			imageURL:     "https://media1.tenor.com/m/9RsE4H_eUAEAAAAd/coinflip-heads.gif",
		},
		false: {
			prefix:       "Tails: ",
			defaultTitle: "The coin landed on tails",
			optionKey:    "landed-tails",
			imageURL:     "https://media1.tenor.com/m/C_cJS3GKhwcAAAAd/coinflip-tails.gif",
		},
	}[isHeads]

	opts := NewRequestOptions(i.ApplicationCommandData().Options)
	title, ok := opts.GetString(result.optionKey)
	if ok {
		title = result.prefix + title
	} else {
		title = result.defaultTitle
	}

	embed := &discordgo.MessageEmbed{
		Type:  discordgo.EmbedTypeGifv,
		Title: title,
		Image: &discordgo.MessageEmbedImage{
			URL:    result.imageURL,
			Width:  240,
			Height: 240,
		},
	}
	writeResponse(s, i, withEmbeds([]*discordgo.MessageEmbed{embed}))
}

func (h *Handlers) TalkingStick(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := NewRequestOptions(i.ApplicationCommandData().Options)
	turnDuration, ok := opts.GetInt("duration")
	if !ok {
		turnDuration = 15
	}
	duration := time.Duration(turnDuration) * time.Second

	// get the guilds active voice members
	userID := i.Member.User.ID
	vs, err := s.State.VoiceState(i.GuildID, userID)
	if err != nil {
		slog.Error("failed to get voice state", "user_id", userID)
		return
	}
	if vs.ChannelID == "" || vs == nil {
		writeMessage(s, i, "You need to be in a voice channel to use this command")
		return
	}

	// check if talking stick is currently being run in the requested channel
	if h.isRunningTalkingStick(vs.ChannelID) {
		writeMessage(s, i, "The talking stick is already being passed around your voice channel")
		return
	}

	slog.Debug("getting members for voice channel", "channel_id", vs.ChannelID)
	members, err := getVoiceChannelMembers(s, i.GuildID, vs.ChannelID)
	if err != nil {
		slog.Error("failed to get voice channel members", "channel_id", vs.ChannelID, "error", err)
		return
	}
	if len(members) == 0 {
		writeMessage(s, i, "The channel has no active voice members")
		return
	}

	writeMessage(s, i, "Talking stick initiated")
	go h.startTalkingStick(s, members, vs.ChannelID, duration)
}

func (h *Handlers) startTalkingStick(s *discordgo.Session, members []*discordgo.Member, channelID string, duration time.Duration) {
	h.wg.Add(1)
	defer h.wg.Done()

	h.addTalkingStickChannel(channelID)
	defer h.removeTalkingStickChannel(channelID)

	// create a channel for messages, so we can delete them all later
	messageCh := make(chan *discordgo.Message, len(members)+5)
	defer h.cleanupMessages(s, messageCh)

	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	shuffleMembers(members)
	h.passTalkingStick(s, channelID, members[0], true, messageCh)
	for idx := 1; idx <= len(members); idx++ {
		select {
		case <-h.shutdownCh:
			h.removePrioritySpeaker(s, channelID, members[idx].User.ID)
			return
		case <-ticker.C:
			// have to do it like this so the last member gets a turn
			if idx == len(members) {
				break
			}
			currentMember := members[idx]
			h.passTalkingStick(s, channelID, currentMember, false, messageCh)
		}
	}
	if len(members) > 0 {
		s.Lock()
		ttsMessage, err := s.ChannelMessageSendTTS(channelID, "The talking stick session has been terminated")
		if err != nil {
			s.Unlock()
			slog.Error("failed to send tts message", "channel_id", channelID, "error", err)
		}
		s.Unlock()
		messageCh <- ttsMessage
		h.removePrioritySpeaker(s, channelID, members[len(members)-1].User.ID)
	}
}

func (h *Handlers) passTalkingStick(s *discordgo.Session, channelID string, member *discordgo.Member, isFirst bool, ch chan *discordgo.Message) {
	slog.Debug("passing the talking stick", "user_id", member.User.ID)
	msg := fmt.Sprintf("Passing the talking stick to %s", member.User.Mention())
	if isFirst {
		msg = fmt.Sprintf("%s has the talking stick", member.User.Mention())
	}

	s.Lock()
	defer s.Unlock()
	// announce the passing of the talking stick
	ttsMessage, err := s.ChannelMessageSendTTS(channelID, msg)
	if err != nil {
		slog.Error("failed to send tts message", "channel_id", channelID, "error", err)
	}
	ch <- ttsMessage

	if err = s.ChannelPermissionSet(channelID, member.User.ID,
		discordgo.PermissionOverwriteTypeMember, discordgo.PermissionVoicePrioritySpeaker, 0); err != nil {
		slog.Error("Failed to set priority speaker", "error", err, "user", member.User.Username)
	}
}

func (h *Handlers) removePrioritySpeaker(s *discordgo.Session, channelID, userID string) {
	if err := s.ChannelPermissionSet(channelID, userID,
		discordgo.PermissionOverwriteTypeMember, 0, discordgo.PermissionVoicePrioritySpeaker); err != nil {
		slog.Error("Failed to remove priority speaker", "error", err, "user_id", userID)
	}
}

func (h *Handlers) cleanupMessages(s *discordgo.Session, ch chan *discordgo.Message) {
	for message := range ch {
		// if the message is a text to speech message, make sure we waited a few seconds for it to finish announcing it
		if message.TTS && time.Since(message.Timestamp) < 5*time.Second {
			duration := 5*time.Second - time.Since(message.Timestamp)
			slog.Debug("waiting for TTS message to finish announcing", "message_id", message.ID, "duration", duration.String())
			timer := time.NewTimer(duration)
			<-timer.C
		}
		slog.Debug("deleting message", "message_id", message.ID)
		s.Lock()
		if err := s.ChannelMessageDelete(message.ChannelID, message.ID); err != nil {
			slog.Error("failed to delete message", "message_id", err)
		}
		s.Unlock()
	}
}

func (h *Handlers) addTalkingStickChannel(channelID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.talkingStickChannels[channelID] = true
}

func (h *Handlers) removeTalkingStickChannel(channelID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.talkingStickChannels, channelID)
}

func (h *Handlers) isRunningTalkingStick(channelID string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.talkingStickChannels[channelID]
}

func getVoiceChannelMembers(s *discordgo.Session, guildID, channelID string) ([]*discordgo.Member, error) {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		return nil, fmt.Errorf("error accessing guild state: %w", err)
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
	return members, nil
}

func shuffleMembers(members []*discordgo.Member) {
	for i := len(members) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		members[i], members[j] = members[j], members[i]
	}
}

func (h *Handlers) getVideoIDFromRequest(opts RequestOptions) (string, error) {
	videoID, ok := opts.GetString("url")
	if ok {
		return videoID, nil
	}

	query, _ := opts.GetString("query")
	result, err := h.ytClient.Search(query)
	if err != nil {
		return "", fmt.Errorf("search error: %w", err)
	}
	if len(result.Items) == 0 {
		return "", fmt.Errorf("no search results found for query: %s", query)
	}
	return result.Items[0].ID.VideoID, nil
}

func (h *Handlers) Close() error {
	h.wg.Wait()
	close(h.shutdownCh)
	return nil
}
