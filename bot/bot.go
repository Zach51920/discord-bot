package bot

import (
	"context"
	"fmt"
	"github.com/Zach51920/discord-bot/config"
	"github.com/Zach51920/discord-bot/postgres"
	"github.com/bwmarrin/discordgo"
	ranna "github.com/ranna-go/ranna/pkg/client"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Bot struct {
	config config.Config

	sess       *discordgo.Session
	dbProvider *postgres.Provider
	rClient    ranna.Client

	closers []io.Closer
	wg      sync.WaitGroup
}

func New(cfg config.Config) *Bot {
	return &Bot{
		config:  cfg,
		wg:      sync.WaitGroup{},
		closers: make([]io.Closer, 0),
	}
}

func (b *Bot) Run() error {
	if err := b.init(); err != nil {
		return err
	}

	if err := b.sess.Open(); err != nil {
		return err
	}
	slog.Info("bot started...")
	defer b.Shutdown()

	// wait for a shutdown signal
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)
	<-shutdownChan

	return nil
}

// Shutdown triggers a graceful shutdown. If the shutdown takes more than 5 seconds pull the plug.
func (b *Bot) Shutdown() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)

	slog.Info("received shutdown signal, waiting for processes to finish...")
	go func() {
		b.wg.Wait() // wait for processes to finish
		slog.Info("process finished successfully")
		ctxCancel()
	}()

	<-ctx.Done()
	if err := b.sess.Close(); err != nil {
		slog.Error("failed to close session", "error", err)
	}
}

func (b *Bot) init() error {
	// init ranna client
	rClient, err := ranna.New(ranna.Options{
		Endpoint:  b.config.Ranna.Endpoint,
		Version:   b.config.Ranna.Version,
		UserAgent: b.config.Ranna.UserAgent,
	})
	if err != nil {
		return fmt.Errorf("create ranna: %w", err)
	}
	b.rClient = rClient

	// init postgres provider
	pqConfig := postgres.Config{PostgresURL: os.Getenv("OVERLORD_DB_URL")}
	if b.dbProvider, err = postgres.NewProvider(pqConfig); err != nil {
		return fmt.Errorf("create postgres provider")
	}

	// init bot
	token := config.GetString("BOT_TOKEN")
	b.sess, err = discordgo.New("Bot " + token)
	if err != nil {
		slog.Error("failed to create session", "error", err)
		os.Exit(1)
	}
	b.RegisterCommands()
	b.RegisterHandlers()
	b.RegisterIntents()
	return nil
}

func (b *Bot) sendAlert(format string, a ...any) {
	if !b.config.Bot.Alerts {
		return
	}

	content := fmt.Sprintf(format, a...)
	channelID := config.GetString("ALERT_CHANNEL_ID")
	slog.Info("sending alert", "content", content)

	b.sess.Lock()
	defer b.sess.Unlock()
	if _, err := b.sess.ChannelMessageSend(channelID, content); err != nil {
		slog.Error("failed to send alert", "error", err)
	}
}
