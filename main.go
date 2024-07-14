package main

import (
	discordbot "github.com/Zach51920/discord-bot/bot"
	"log/slog"
	"os"
)

func main() {
	bot := discordbot.New()
	if err := bot.Run(); err != nil {
		slog.Error("failed to run bot", "error", err)
		os.Exit(1)
	}
}
