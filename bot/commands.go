package bot

import (
	"github.com/Zach51920/discord-bot/config"
	"github.com/bwmarrin/discordgo"
	"log/slog"
)

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "talking-stick-start",
		Description: "Initiate a talking stick session in your current voice channel",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionNumber,
				Name:        "duration",
				Description: "How long does each user have the talking stick (seconds)",
				Required:    false,
			},
		},
	},
	{
		Name:        "yt-download",
		Description: "Download a YouTube video.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "url",
				Description: "VideoID of the YouTube video to download.",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "query",
				Description: "Search query for the YouTube video to download.",
				Required:    false,
			},
		},
	},
	{
		Name:        "yt-search",
		Description: "BETA | Search for YouTube videos to play and/or download.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "query",
				Description: "Search query to find YouTube videos.",
				Required:    true,
			},
		},
	},
	{
		Name:        "bedtime-ban",
		Description: "Handle a \"baby rager\" by putting them to bed.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "User to take a nap",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "Why is the user going to bed?",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionNumber,
				Name:        "duration",
				Description: "How long (hours) until the ban expires",
				Required:    false,
			},
		},
	},
}

func (b *Bot) RegisterCommands() {
	guilds, err := b.sess.UserGuilds(0, "", "")
	if err != nil {
		slog.Error("failed to get guilds", "error", err)
		return
	}

	for _, guild := range guilds {
		b.overwriteCommands(guild.ID, guild.Name)
	}
}

func (b *Bot) overwriteCommands(guildID, guildName string) {
	applicationID := config.GetString("APPLICATION_ID")
	if _, err := b.sess.ApplicationCommandBulkOverwrite(applicationID, guildID, commands); err != nil {
		slog.Error("failed to overwrite commands", "error", err, "guild", guildName)
	}
}
