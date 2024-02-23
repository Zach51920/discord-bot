package handlers

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

type Response struct {
	IsError bool // send the response with an error prefix
	Message *string
	Files   []*File
	Embeds  []*discordgo.MessageEmbed
}

func ErrorResponse(message string) Response {
	return Response{
		IsError: true,
		Message: &message,
	}
}

func (r *Response) Close() error {
	if len(r.Files) == 0 {
		return nil
	}

	var err error
	errorCount := 0
	for _, f := range r.Files {
		if e := f.ReaderCloser.Close(); e != nil {
			errorCount++
			err = e
		}
	}
	return fmt.Errorf("failed to close %d files: %w", errorCount, err)
}
