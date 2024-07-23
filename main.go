package main

import (
	discordbot "github.com/Zach51920/discord-bot/bot"
	"github.com/Zach51920/discord-bot/config"
	"log/slog"
	"os"
)

func main() {
	initLogger()

	// read config
	cfg, err := config.Load("config.yaml")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// create the bot
	bot := discordbot.New(cfg)
	if err = bot.Run(); err != nil {
		slog.Error("failed to run bot", "error", err)
		os.Exit(1)
	}
}

func initLogger() {
	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelDebug)

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: lvl,
	}))
	slog.SetDefault(logger)

}
