package interactions

import (
	"fmt"
	"github.com/Zach51920/discord-bot/youtube"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"sync"
	"time"
)

type Handlers struct {
	wg       sync.WaitGroup
	ytClient *youtube.Client
}

func New() *Handlers {
	return &Handlers{ytClient: youtube.New()}
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
	return nil
}
