package handlers

import "errors"

var ErrInvalidParams = errors.New("invalid parameters")

type SearchVideoParams struct {
	Query string
}

type DownloadVideoParams struct {
	VideoID *string
	Query   *string
}
