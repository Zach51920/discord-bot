package interactions

import (
	"fmt"
	"github.com/Zach51920/discord-bot/youtube"
	"github.com/bwmarrin/discordgo"
)

func mapYTSearchResults(results youtube.YTSearchResults) []*discordgo.MessageEmbed {
	embeds := make([]*discordgo.MessageEmbed, len(results.Items))
	for i, item := range results.Items {
		embeds[i] = &discordgo.MessageEmbed{
			Title:       item.Snippet.Title,
			Description: item.Snippet.Description,
			URL:         getYouTubeURL(item.ID.VideoID),
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
	return embeds
}

func getYouTubeURL(videoID string) string {
	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
}

func mapFile(f *youtube.File) *discordgo.File {
	if f == nil {
		return nil
	}
	return &discordgo.File{
		Name:        f.Name,
		ContentType: f.ContentType,
		Reader:      f.ReaderCloser,
	}
}
