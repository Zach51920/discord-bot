package main

import discordbot "github.com/Zach51920/discord-bot/bot"

func main() {
	bot, err := discordbot.New()
	if err != nil {
		panic("failed to create bot: " + err.Error())
	}

	if err = bot.Run(); err != nil {
		panic("failed to run bot: " + err.Error())
	}
}
