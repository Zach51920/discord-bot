package interactions

import (
	"github.com/Zach51920/discord-bot/talkingstick"
	"github.com/Zach51920/discord-bot/youtube"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"sync"
)

type Handlers struct {
	wg sync.WaitGroup
	mu sync.Mutex

	ytClient   *youtube.Client
	tsManager  talkingstick.SessionManager
	shutdownCh chan struct{}
}

func New(s *discordgo.Session) *Handlers {
	return &Handlers{
		ytClient:   youtube.New(),
		wg:         sync.WaitGroup{},
		shutdownCh: make(chan struct{}),
		tsManager:  talkingstick.NewSessionManager(s),
	}
}

func (h *Handlers) HandleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

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
		"yt-download":   h.Download,
		"yt-search":     h.Search,
		"bedtime-ban":   h.Bedtime,
		"talking-stick": h.TalkingStick,
		"coinflip":      h.CoinFlip,
	}
	data := i.ApplicationCommandData()
	handler, ok := commands[data.Name]
	if !ok {
		writeMessage(s, i, "Unknown request command")
		return
	}
	handler(s, i)
}

func (h *Handlers) HandleButtons(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	h.wg.Add(1)
	defer h.wg.Done()
	customID := i.MessageComponentData().CustomID
	slog.Info("received button press event", "custom_id", customID)

	// acknowledge the request
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate}); err != nil {
		slog.Error("failed to respond to interaction", "error", err)
	}

	// get voice channel
	vs, err := getVoiceState(s, i.GuildID, i.Member.User.ID)
	if err != nil {
		writeMessage(s, i, "Failed to get voice state. Are you in a voice channel?")
		return
	}

	// perform the action
	if err = h.tsManager.Handle(vs.ChannelID, i.MessageComponentData()); err != nil {
		slog.Error("failed to perform action", "channel_id", vs.ChannelID, "error", err)
		return
	}
}

func (h *Handlers) Close() error {
	h.wg.Wait()
	close(h.shutdownCh)
	return h.tsManager.Close()
}
