package main

import (
	discordbot "github.com/Zach51920/discord-bot/bot"
	"log/slog"
	"os"
)

func main() {
	// init logger
	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelDebug)

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: lvl,
	}))
	slog.SetDefault(logger)

	// create the bot
	bot := discordbot.New()
	if err := bot.Run(); err != nil {
		slog.Error("failed to run bot", "error", err)
		os.Exit(1)
	}
}
