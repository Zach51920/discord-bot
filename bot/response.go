package bot

import (
	"fmt"
	"github.com/Zach51920/discord-bot/handlers"
	"github.com/bwmarrin/discordgo"
	"log"
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
		return fmt.Errorf("followup message: %w", err)
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
