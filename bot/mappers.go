package bot

import (
	"fmt"
	"github.com/Zach51920/discord-bot/google"
	"github.com/bwmarrin/discordgo"
)

func mapYTSearchResults(results google.YTSearchResults) []*discordgo.MessageEmbed {
	embeds := make([]*discordgo.MessageEmbed, len(results.Items))
	for i, item := range results.Items {
		embeds[i] = mapYTSearchItem(item)
	}
	return embeds
}

func mapYTSearchItem(item google.YTSearchItem) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       item.Snippet.Title,
		Description: item.Snippet.Description,
		URL:         fmt.Sprintf("https://www.youtube.com/watch?v=%s", item.ID.VideoID),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL:    item.Snippet.Thumbnails.Default.URL,
			Width:  item.Snippet.Thumbnails.Default.Width,
			Height: item.Snippet.Thumbnails.Default.Height,
		},
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Video ID", Value: item.ID.VideoID},
		},
	}
}