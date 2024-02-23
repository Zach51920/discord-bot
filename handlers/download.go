package handlers

import (
	"errors"
	"fmt"
	"github.com/kkdai/youtube/v2"
	"io"
	"strings"
)

type File struct {
	Name         string
	ContentType  string
	ReaderCloser io.ReadCloser
}

func (h *Handlers) Download(params DownloadVideoParams) (Response, error) {
	video, err := h.ytClient.GetVideo(params.VideoID)
	if err != nil {
		if errors.Is(err, youtube.ErrInvalidCharactersInVideoID) || errors.Is(err, youtube.ErrVideoIDMinLength) {
			return ErrorResponse("Invalid Video ID"), nil
		}
		var playbackErr *youtube.ErrPlayabiltyStatus
		if errors.As(err, &playbackErr) {
			return ErrorResponse(playbackErr.Reason), nil
		}
		return Response{}, fmt.Errorf("video error: %w", err)
	}

	formats := video.Formats.WithAudioChannels() // only get videos with audio
	if len(formats) == 0 {
		return Response{}, fmt.Errorf("no video formats contain audio channels")
	}

	stream, _, err := h.ytClient.GetStream(video, &formats[0])
	if err != nil {
		return Response{}, fmt.Errorf("failed to get stream: %w", err)
	}

	return Response{Files: []*File{{
		Name:         strings.Join(strings.Split(video.Title, " "), "-") + ".mp4",
		ContentType:  video.Formats[0].MimeType,
		ReaderCloser: stream,
	}}}, nil
}
