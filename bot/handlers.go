package bot

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/kkdai/youtube/v2"
	"io"
	"log"
	"strings"
)

var errRuntime = errors.New("Uh oh, something went wrong :(")
var errInvalidVideID = errors.New("Invalid Video ID")

func (b *Bot) RegisterHandlers() error {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "youpirate",
			Description: "A tool for do",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "url",
					Description: "url of the youtube video to download",
					Required:    true,
				},
			},
		},
	}
	if _, err := b.session.ApplicationCommandBulkOverwrite(b.config.ApplicationID, b.config.GuildID, commands); err != nil {
		return fmt.Errorf("bulk overwrite failed: %w", err)
	}

	b.session.AddHandler(b.handleDownload)
	return nil
}

func (b *Bot) handleDownload(s *discordgo.Session, i *discordgo.InteractionCreate) {
	b.wg.Add(1)
	defer b.wg.Done()

	req := request{s, i}
	if req.action() != "youpirate" {
		_ = req.writeResponse("I dont know that command... What are you trying to do?", nil)
		return
	}

	videoID := req.getStrOption("url")
	log.Printf("[INFO] %s requested to download %s\n", i.Member.User.Username, videoID)

	// respond immediately to prevent premature timeout
	if err := req.writeResponse("downloading...", nil); err != nil {
		log.Println("[ERROR] failed to send response:", err.Error())
		return
	}

	file, stream, err := b.getFile(videoID)
	if err != nil {
		_ = req.overwriteResponse(err.Error(), nil)
		return
	}
	defer func() {
		if err = stream.Close(); err != nil {
			b.sendAlert("failed to close stream: " + err.Error()) // potential memory leaks are worth a pesky notification
		}
	}()

	if err = req.overwriteResponse("", file); err != nil {
		_ = req.handleWriteError(err)
		return
	}
}

// getFile gets a YouTube video and returns it as a *discordgo.File. An io.Closer is also returned to close the video stream
func (b *Bot) getFile(videoID string) (*discordgo.File, io.Closer, error) {
	video, err := b.youtube.GetVideo(videoID)
	if err != nil {
		if errors.Is(err, youtube.ErrInvalidCharactersInVideoID) || errors.Is(err, youtube.ErrVideoIDMinLength) {
			return nil, nil, errInvalidVideID
		}
		log.Println("[ERROR] failed to get video:", err)
		return nil, nil, errRuntime
	}

	// only get videos with audio
	formats := video.Formats.WithAudioChannels()
	if len(formats) == 0 {
		b.sendAlert("Unable to get formats with audio channels: " + videoID)
		return nil, nil, errRuntime
	}
	stream, _, err := b.youtube.GetStream(video, &formats[0])
	if err != nil {
		log.Println("[ERROR] failed to get stream:", err.Error())
		return nil, nil, errRuntime
	}

	return &discordgo.File{
		Name:        strings.Join(strings.Split(video.Title, " "), "-") + ".mp4",
		ContentType: video.Formats[0].MimeType,
		Reader:      stream,
	}, stream, nil
}
