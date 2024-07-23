package events

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
	ranna "github.com/ranna-go/ranna/pkg/client"
	"sync"
)

type Handler struct {
	db      *sqlx.DB
	rClient ranna.Client

	sess *discordgo.Session
	wg   sync.WaitGroup

	messageCh  chan *discordgo.Message
	shutdownCh chan struct{}
}

func New(sess *discordgo.Session, rClient ranna.Client, db *sqlx.DB) *Handler {
	handler := &Handler{
		db:         db,
		rClient:    rClient,
		sess:       sess,
		wg:         sync.WaitGroup{},
		messageCh:  make(chan *discordgo.Message),
		shutdownCh: make(chan struct{}),
	}
	go handler.Start()
	return handler
}

func (h *Handler) Start() {
	for {
		select {
		case m := <-h.messageCh:
			h.handleMessage(m)
		case <-h.shutdownCh:
			return
		}
	}
}

func (h *Handler) Close() error {
	h.wg.Wait()
	close(h.messageCh)
	close(h.shutdownCh)
	return nil
}
