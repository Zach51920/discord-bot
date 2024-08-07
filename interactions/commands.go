package interactions

import (
	"errors"
	"fmt"
	"github.com/Zach51920/discord-bot/talkingstick"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"math/rand"
	"time"
)

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
	turnDuration := opts.GetIntDefault("duration", 15)
	duration := time.Duration(turnDuration) * time.Second

	// get the users active voice channel
	vs, err := getVoiceState(s, i.GuildID, i.Member.User.ID)
	if err != nil {
		writeMessage(s, i, "Failed to get voice state. Are you in a voice channel?")
		return
	}

	if err = h.tsManager.Create(vs.GuildID, vs.ChannelID, duration); err != nil {
		if errors.Is(err, talkingstick.ErrSessionExists) {
			writeMessage(s, i, "A talking stick already exists in your current voice channel")
			return
		}
		slog.Error("failed to create talking stick session", "channel_id", vs.ChannelID, "error", err)
		return
	}
	writeMessage(s, i, "Initiating talking stick session")
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
