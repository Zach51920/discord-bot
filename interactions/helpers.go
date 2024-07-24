package interactions

import (
	"fmt"
	"github.com/Zach51920/discord-bot/youtube"
	"github.com/bwmarrin/discordgo"
	"log/slog"
)

func logRequest(i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	args := make([]any, 4, 4+(len(data.Options)*2))
	args[0], args[1] = "user", i.Member.User.Username
	args[2], args[3] = "command", data.Name
	for _, opt := range data.Options {
		args = append(args, opt.Name, opt.Value)
	}
	slog.Info("received request", args...)
}

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

func getVoiceState(s *discordgo.Session, guildID, userID string) (*discordgo.VoiceState, error) {
	vs, err := s.State.VoiceState(guildID, userID)
	if err != nil {
		slog.Error("Failed to get voice state", "user_id", userID, "error", err)
		return nil, err
	}
	if vs == nil || vs.ChannelID == "" {
		return nil, fmt.Errorf("user not in a voice channel")
	}
	return vs, nil
}
