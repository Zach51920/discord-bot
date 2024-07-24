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

	talkingStickChannels map[string]*TalkingStickSess
	shutdownCh           chan struct{}
}

func New() *Handlers {
	return &Handlers{
		ytClient:             youtube.New(),
		wg:                   sync.WaitGroup{},
		shutdownCh:           make(chan struct{}),
		talkingStickChannels: make(map[string]*TalkingStickSess)}
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

	// execute the command
	commands := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"yt-download":         h.Download,
		"yt-search":           h.Search,
		"bedtime-ban":         h.Bedtime,
		"talking-stick-start": h.TalkingStick,
		"coinflip":            h.CoinFlip,
	}
	data := i.ApplicationCommandData()
	cmd, ok := commands[data.Name]
	if !ok {
		writeMessage(s, i, "Unknown request command")
		return
	}
	cmd(s, i)
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
	turnDuration := opts.GetIntDefault("duration", 15)
	duration := time.Duration(turnDuration) * time.Second

	// get the channels active voice members
	vs, err := getVoiceState(s, i.GuildID, i.Member.User.ID)
	if err != nil {
		writeMessage(s, i, "Failed to get voice state. Are you in a voice channel?")
		return
	}

	// check if talking stick is currently being run in the requested channel
	if h.getTalkingStickSess(vs.ChannelID) != nil {
		writeMessage(s, i, "The talking stick is already being passed around your voice channel")
		return
	}

	slog.Debug("creating talking stick session for channel", "channel_id", vs.ChannelID)
	tss := NewStickSession(s, vs.GuildID, vs.ChannelID, duration)
	go func() {
		h.addTalkingStickSess(tss)
		defer h.removeTalkingStickSess(tss.channelID)
		tss.Start(s)
	}()
	writeMessage(s, i, "Talking stick initiated")
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
