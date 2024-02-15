package bot

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"log"
)

type request struct {
	session     *discordgo.Session
	interaction *discordgo.InteractionCreate
}

func (r *request) action() string {
	return r.interaction.ApplicationCommandData().Name
}

func (r *request) getStrOption(name string) string {
	for _, data := range r.interaction.ApplicationCommandData().Options {
		if name == data.Name && data.Type == discordgo.ApplicationCommandOptionString {
			return data.Value.(string)
		}
	}
	return ""
}

func (r *request) writeResponse(content string, file *discordgo.File) error {
	r.session.Lock()
	defer r.session.Unlock()

	files := make([]*discordgo.File, 0)
	if file != nil {
		files = append(files, file)
	}
	return r.session.InteractionRespond(
		r.interaction.Interaction,
		&discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
				Files:   files,
			},
		},
	)
}

func (r *request) overwriteResponse(content string, file *discordgo.File) error {
	r.session.Lock()
	defer r.session.Unlock()

	files := make([]*discordgo.File, 0)
	if file != nil {
		files = append(files, file)
	}
	resp := &discordgo.WebhookEdit{
		Content: &content,
		Files:   files,
	}
	if _, err := r.session.InteractionResponseEdit(r.interaction.Interaction, resp); err != nil {
		return err
	}
	return nil
}

func (r *request) handleWriteError(writeErr error) error {
	var restErr *discordgo.RESTError
	if errors.As(writeErr, &restErr) {
		if restErr.Message != nil {
			return r.overwriteResponse(restErr.Message.Message, nil)
		}
	}
	log.Println("[ERROR] an unknown error has occurred:", writeErr)
	return r.overwriteResponse("an unknown error has occurred", nil)
}
