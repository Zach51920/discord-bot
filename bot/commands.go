package bot

import (
	"fmt"
	"github.com/Zach51920/discord-bot/config"
	"github.com/bwmarrin/discordgo"
)

func RegisterCommands(s *discordgo.Session, guildID string) error {
	commands := []*discordgo.ApplicationCommand{
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

	if _, err := s.ApplicationCommandBulkOverwrite(
		config.GetString("APPLICATION_ID"),
		guildID,
		commands); err != nil {
		return fmt.Errorf("bulk overwrite failed: %w", err)
	}
	return nil
}
