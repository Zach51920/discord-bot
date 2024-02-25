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

func (h *Handlers) Download(options RequestOptions) (Response, error) {
	params := GetDownloadVideoParams(options)
	videoID, err := h.getVideoIDFromParams(params)
	if err != nil {
		if errors.Is(err, ErrInvalidParams) {
			return ErrorResponse("Invalid Parameters"), nil
		}
		return Response{}, fmt.Errorf("video ID error: %w", err)
	}

	video, err := h.ytClient.GetVideo(videoID)
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

func (h *Handlers) getVideoIDFromParams(params DownloadVideoParams) (string, error) {
	if params.VideoID != nil {
		return *params.VideoID, nil
	} else if params.Query == nil {
		return "", ErrInvalidParams
	}

	result, err := h.gClient.SearchYT(*params.Query)
	if err != nil {
		return "", fmt.Errorf("search error: %w", err)
	}
	if len(result.Items) == 0 {
		return "", fmt.Errorf("no search results found for query: %s", *params.Query)
	}
	return result.Items[0].ID.VideoID, nil
}
