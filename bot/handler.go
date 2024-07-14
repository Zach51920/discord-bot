package bot

import (
	"github.com/bwmarrin/discordgo"
	"log/slog"
)

func (b *Bot) handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	b.wg.Add(1)
	defer b.wg.Done()
	logRequest(i)

	// defer the message response and let the user know we got the request
	if err := acknowledgeRequest(s, i); err != nil {
		slog.Error("failed to acknowledge request: " + err.Error())
		return
	}
	defer followupRequest(s, i)

	data := i.ApplicationCommandData()
	handleFn, ok := b.handlers[data.Name]
	if !ok {
		s.Lock()
		defer s.Unlock()
		if _, err := s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
			Content: "Unknown request command"}); err != nil {
			slog.Error("failed to write response", "error", err)
		}
		slog.Error("unknown request command", "command", data.Name)
		return
	}
	handleFn(s, i)
}

func logRequest(i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	args := make([]any, 4, 4+(len(data.Options)*2))
	args[0], args[1] = "user", i.Member.User.Username
	args[2], args[3] = "command", data.Name
	for _, opt := range data.Options {
		args = append(args, opt.Name, opt.Value)
	}
	slog.Info("received request", args...)
}

func acknowledgeRequest(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	s.Lock()
	defer s.Unlock()

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
}

func followupRequest(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.Lock()
	defer s.Unlock()

	// check if we already responded
	resp, err := s.InteractionResponse(i.Interaction)
	if err != nil {
		slog.Error("failed to get response message", "error", err)
		return
	}
	if len(resp.Embeds) != 0 || len(resp.Attachments) != 0 || resp.Content != "" {
		// we have already responded, return
		return
	}

	// if we haven't made a response, something went wrong
	if _, err = s.FollowupMessageCreate(i.Interaction, false,
		&discordgo.WebhookParams{
			Content: "An unexpected error has occurred.",
		}); err != nil {
		slog.Error("failed to write response", "error", err)
	}
}
