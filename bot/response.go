package bot

import (
	"errors"
	"fmt"
	"github.com/Zach51920/discord-bot/handlers"
	"github.com/bwmarrin/discordgo"
	"log"
	"net/http"
)

func writeResponse(s *discordgo.Session, i *discordgo.InteractionCreate, resp handlers.Response) error {
	if resp.IsError {
		writeError(s, i, derefStr(resp.Message))
		return nil
	}

	s.Lock()
	defer s.Unlock()
	if _, err := s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Content: derefStr(resp.Message),
		Files:   mapFiles(resp.Files),
		Embeds:  resp.Embeds,
	}); err != nil {
		msg, _ := getRESTErrorMessage(err)
		return fmt.Errorf("followup message: %s", msg)
	}
	return nil
}

func writeMessage(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) error {
	s.Lock()
	defer s.Unlock()

	_, err := s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{Content: msg})
	return err
}

func writeError(s *discordgo.Session, i *discordgo.InteractionCreate, errMsg string) {
	if err := writeMessage(s, i, "Error: "+errMsg); err != nil {
		log.Println("[ERROR] failed to write error: " + err.Error())
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
