package interactions

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"net/http"
)

type responseParam func(param *discordgo.WebhookParams)

func writeResponse(s *discordgo.Session, i *discordgo.InteractionCreate, params ...responseParam) {
	webhookParams := new(discordgo.WebhookParams)
	for _, param := range params {
		param(webhookParams)
	}

	slog.Info("writing response",
		"interaction", i.ID,
		"content", webhookParams.Content,
		"files", len(webhookParams.Files),
		"embeds", len(webhookParams.Embeds))

	s.Lock()
	defer s.Unlock()
	if _, err := s.FollowupMessageCreate(i.Interaction, false, webhookParams); err != nil {
		msg, _ := getRESTErrorMessage(err)
		slog.Error("failed to write response", "error", msg)
	}
}

func writeMessage(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	writeResponse(s, i, withMessage(msg))
}

func withEmbeds(embeds []*discordgo.MessageEmbed) responseParam {
	return func(p *discordgo.WebhookParams) {
		p.Embeds = embeds
	}
}

func withFiles(files []*discordgo.File) responseParam {
	return func(p *discordgo.WebhookParams) {
		p.Files = files
	}
}

func withMessage(format string, a ...any) responseParam {
	return func(p *discordgo.WebhookParams) {
		p.Content = fmt.Sprintf(format, a...)
	}
}

func getRESTErrorMessage(err error) (string, bool) {
	var restErr *discordgo.RESTError
	if ok := errors.As(err, &restErr); !ok || restErr == nil {
		return err.Error(), false
	}
	switch restErr.Response.StatusCode {
	case http.StatusRequestEntityTooLarge:
		return "Request entity is too large.", true
	}
	if restErr.Message != nil {
		return restErr.Message.Message, true
	}
	return "An unexpected error has occurred", true
}

func acknowledgeRequest(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	s.Lock()
	defer s.Unlock()

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
}

// ensureFollowup checks if we already responded to the message. If we haven't, send an error response
func ensureFollowup(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.Lock()
	// check if we already responded
	resp, err := s.InteractionResponse(i.Interaction)
	if err != nil {
		s.Unlock()
		slog.Error("failed to get response message", "error", err)
		return
	}
	s.Unlock()
	if len(resp.Embeds) != 0 || len(resp.Attachments) != 0 || resp.Content != "" {
		// we have already responded, return
		return
	}

	// if we haven't made a response, something went wrong
	writeMessage(s, i, "An unexpected error has occurred")
}
