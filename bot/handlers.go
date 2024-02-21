package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
)

// commandHandler is the entrypoint for all requests
func (b *Bot) commandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	b.wg.Add(1)
	defer b.wg.Done()

	logRequest(s, i)

	// defer the message response since we might take a few seconds
	if err := acknowledgeRequest(s, i); err != nil {
		log.Println("failed to acknowledge request: " + err.Error())
	}

	switch i.ApplicationCommandData().Name {
	case "download":
		b.handleDownload(s, i)
	case "watch":
		b.handleNotImplemented(s, i)
	case "listen":
		b.handleNotImplemented(s, i)
	case "search":
		b.handleSearch(s, i)
	default:
		b.handleUnknownCommand(s, i)
	}
}

func (b *Bot) handleDownload(s *discordgo.Session, i *discordgo.InteractionCreate) {
	videoID := getStrOption(i, "url")
	video, stream, err := b.getStream(videoID)
	if err != nil {
		handleStreamError(s, i, err)
		return
	}
	defer func() {
		if err = stream.Close(); err != nil {
			log.Println("[WARNING] failed to close stream: " + err.Error())
		}
	}()

	if err = writeFile(s, i, &discordgo.File{
		Name:        strings.Join(strings.Split(video.Title, " "), "-") + ".mp4",
		ContentType: video.Formats[0].MimeType,
		Reader:      stream,
	}); err != nil {
		handleWriteError(s, i, err)
		return
	}
}

func (b *Bot) handleSearch(s *discordgo.Session, i *discordgo.InteractionCreate) {
	results, err := b.gClient.SearchYT(getStrOption(i, "query"))
	if err != nil {
		log.Println("[ERROR] failed to search youtube: " + err.Error())
		writeError(s, i, "An unexpected error has occurred")
		return
	}

	message := &discordgo.WebhookParams{Embeds: mapYTSearchResults(results)}
	if _, err = s.FollowupMessageCreate(i.Interaction, false, message); err != nil {
		log.Println("[ERROR] failed to write response: " + err.Error())
		return
	}
}

func (b *Bot) handleNotImplemented(s *discordgo.Session, i *discordgo.InteractionCreate) {
	msg := fmt.Sprintf("Error: `/%s` is not implemented", i.ApplicationCommandData().Name)
	if err := writeMessage(s, i, msg); err != nil {
		log.Printf("[ERROR] failed to write response: " + err.Error())
	}
}

func (b *Bot) handleUnknownCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	msg := fmt.Sprintf("Error: Unknown command `/%s`", i.ApplicationCommandData().Name)
	if err := writeMessage(s, i, msg); err != nil {
		log.Printf("[ERROR] failed to write response: " + err.Error())
	}
}
