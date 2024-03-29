package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

func (b *Bot) RegisterCommands() error {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "download",
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
			Name:        "search",
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
			Name:        "watch",
			Description: "COMING SOON | Play a YouTube video directly into the voice channel.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "url",
					Description: "VideoID of the YouTube video to play",
					Required:    true,
				},
			},
		},
		{
			Name:        "listen",
			Description: "COMING SOON | Stream audio from a YouTube video into the voice channel.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "url",
					Description: "VideoID of the YouTube video to play",
					Required:    true,
				},
			},
		},
	}
	if _, err := b.session.ApplicationCommandBulkOverwrite(b.config.ApplicationID, b.config.GuildID, commands); err != nil {
		return fmt.Errorf("bulk overwrite failed: %w", err)
	}

	b.session.AddHandler(b.handler)
	return nil
}
