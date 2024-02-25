package handlers

import (
	"errors"
)

var ErrInvalidParams = errors.New("invalid parameters")

type SearchVideoParams struct {
	Query string
}

type DownloadVideoParams struct {
	VideoID *string
	Query   *string
}

func GetDownloadVideoParams(options RequestOptions) DownloadVideoParams {
	return DownloadVideoParams{
		VideoID: options.GetStringPtr("url"),
		Query:   options.GetStringPtr("query"),
	}
}

func GetSearchVideoParams(options RequestOptions) SearchVideoParams {
	query, _ := options.GetString("query")
	return SearchVideoParams{
		Query: query,
	}
}
