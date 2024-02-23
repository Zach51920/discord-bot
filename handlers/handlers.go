package handlers

import (
	"github.com/Zach51920/discord-bot/google"
	"github.com/kkdai/youtube/v2"
	"log"
)

type Handlers struct {
	logger   *log.Logger
	gClient  google.Client
	ytClient youtube.Client
}

func New() *Handlers {
	return &Handlers{
		logger:   log.Default(),
		gClient:  google.Client{},
		ytClient: youtube.Client{},
	}
}

func (h *Handlers) NotImplemented() (Response, error) {
	msg := "Command is not implemented"
	return Response{
		IsError: true,
		Message: &msg,
	}, nil
}

func (h *Handlers) UnknownCommand() (Response, error) {
	msg := "Unknown command"
	return Response{
		IsError: true,
		Message: &msg,
	}, nil
}
