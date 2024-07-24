package main

import (
	"fmt"
	discordbot "github.com/Zach51920/discord-bot/bot"
	"github.com/Zach51920/discord-bot/config"
	"github.com/Zach51920/discord-bot/logger"
	"github.com/Zach51920/discord-bot/postgres"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}
	logger.Init(cfg.Logger)
	postgres.RunMigrations()

	bot := discordbot.New(cfg)
	if err = bot.Run(); err != nil {
		panic(fmt.Errorf("failed to run bot: %w", err))
	}
}
