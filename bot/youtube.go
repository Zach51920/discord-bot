package bot

import (
	"fmt"
	"github.com/kkdai/youtube/v2"
	"io"
)

func (b *Bot) getStream(videoID string) (*youtube.Video, io.ReadCloser, error) {
	video, err := b.ytClient.GetVideo(videoID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get video: %w", err)
	}

	// only get videos with audio
	formats := video.Formats.WithAudioChannels()
	if len(formats) == 0 {
		return nil, nil, fmt.Errorf("no video formats contain audio channels")
	}
	stream, _, err := b.ytClient.GetStream(video, &formats[0])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get stream: %w", err)
	}
	return video, stream, nil
}
