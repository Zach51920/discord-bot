package handlers

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
