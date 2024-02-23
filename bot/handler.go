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

	resp, err := b.handleCommand(s, i)
	defer resp.Close()
	if err != nil {
		log.Println("[ERROR] " + err.Error())
		writeError(s, i, "An unexpected error has occurred")
		return
	}
	if err = writeResponse(s, i, resp); err != nil {
		log.Println("[ERROR] " + err.Error())
		writeError(s, i, "An unexpected error has occurred")
		return
	}
}

func (b *Bot) handleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) (handlers.Response, error) {
	switch i.ApplicationCommandData().Name {
	case "download":
		videoID := getPtrStrOption(i, "url")
		query := getPtrStrOption(i, "query")
		return b.handlers.Download(handlers.DownloadVideoParams{VideoID: videoID, Query: query})
	case "watch":
		return b.handlers.NotImplemented()
	case "listen":
		return b.handlers.NotImplemented()
	case "search":
		query := getStrOption(i, "query")
		return b.handlers.SearchVideos(handlers.SearchVideoParams{Query: query})
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

func getStrOption(i *discordgo.InteractionCreate, name string) string {
	for _, data := range i.ApplicationCommandData().Options {
		if name == data.Name && data.Type == discordgo.ApplicationCommandOptionString {
			return data.Value.(string)
		}
	}
	return ""
}

func getPtrStrOption(i *discordgo.InteractionCreate, name string) *string {
	for _, data := range i.ApplicationCommandData().Options {
		if name == data.Name && data.Type == discordgo.ApplicationCommandOptionString {
			val := data.Value.(string)
			return &val
		}
	}
	return nil
}

func derefStr(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
