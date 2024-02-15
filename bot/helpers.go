package bot

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/kkdai/youtube/v2"
	"log"
	"net/http"
	"strings"
)

func getStrOption(i *discordgo.InteractionCreate, name string) string {
	for _, data := range i.ApplicationCommandData().Options {
		if name == data.Name && data.Type == discordgo.ApplicationCommandOptionString {
			return data.Value.(string)
		}
	}
	return ""
}

func logRequest(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	args := make([]string, len(data.Options))
	for j, opt := range data.Options {
		args[j] = fmt.Sprintf("%s:%v ", opt.Name, opt.Value)
	}
	log.Printf("[INFO] %s made a %s request: %s", i.Member.User.Username, data.Name, strings.Join(args, ","))
}

func acknowledgeRequest(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	s.Lock()
	defer s.Unlock()

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
}

func writeMessage(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) error {
	s.Lock()
	defer s.Unlock()

	_, err := s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{Content: msg})
	return err
}

func writeFile(s *discordgo.Session, i *discordgo.InteractionCreate, file *discordgo.File) error {
	s.Lock()
	defer s.Unlock()

	_, err := s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{Files: []*discordgo.File{file}})
	return err
}

func writeError(s *discordgo.Session, i *discordgo.InteractionCreate, errMsg string) {
	msg := "Error: " + errMsg
	if err := writeMessage(s, i, msg); err != nil {
		log.Println("[ERROR] failed to write error: " + err.Error())
	}
}

func handleStreamError(s *discordgo.Session, i *discordgo.InteractionCreate, err error) {
	if errors.Is(err, youtube.ErrInvalidCharactersInVideoID) || errors.Is(err, youtube.ErrVideoIDMinLength) {
		writeError(s, i, "Invalid video ID")
		return
	}
	var playbackErr *youtube.ErrPlayabiltyStatus
	if errors.As(err, &playbackErr) {
		writeError(s, i, playbackErr.Reason)
		return
	}
	log.Println("[ERROR] failed to download video: " + err.Error())
	writeError(s, i, "An unknown error has occurred")
	return

}

func handleWriteError(s *discordgo.Session, i *discordgo.InteractionCreate, discordWriteErr error) {
	var restErr *discordgo.RESTError
	if errors.As(discordWriteErr, &restErr) {
		if restErr.Message != nil {
			writeError(s, i, restErr.Message.Message)
			return
		}
		// message is nil with a 413 error code :/
		if restErr.Response.StatusCode == http.StatusRequestEntityTooLarge {
			writeError(s, i, "Request entity is too large")
			return
		}
	}
	log.Println("[ERROR] an unknown error has occurred:", discordWriteErr)
	writeError(s, i, "An unknown error has occurred")
}
