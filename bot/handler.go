package bot

import (
	"fmt"
	"github.com/Zach51920/discord-bot/handlers"
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
)

func (b *Bot) handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	b.wg.Add(1)
	defer b.wg.Done()
	logRequest(s, i)

	// defer the message response and let the user know we got the request
	if err := acknowledgeRequest(s, i); err != nil {
		log.Println("[ERROR] failed to acknowledge request: " + err.Error())
		return
	}

	data := i.ApplicationCommandData()
	resp, err := b.handleCommand(data.Name, handlers.NewRequestOptions(data.Options))
	defer resp.Close()
	if err == nil {
		err = writeResponse(s, i, resp)
	}
	if err != nil {
		log.Println("[ERROR] " + err.Error())
		writeError(s, i, "An unexpected error has occurred")
		return
	}
}

func (b *Bot) handleCommand(command string, options handlers.RequestOptions) (handlers.Response, error) {
	switch command {
	case "download":
		return b.handlers.Download(options)
	case "watch":
		return b.handlers.NotImplemented()
	case "listen":
		return b.handlers.NotImplemented()
	case "search":
		return b.handlers.SearchVideos(options)
	default:
		return b.handlers.UnknownCommand()
	}
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

func derefStr(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
