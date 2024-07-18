package interactions

import (
	"github.com/bwmarrin/discordgo"
	"log/slog"
)

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
