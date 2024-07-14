package bot

import (
	"context"
	"github.com/Zach51920/discord-bot/config"
	"github.com/Zach51920/discord-bot/handlers"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Bot struct {
	handlers map[string]handlers.HandlerFn
	sess     *discordgo.Session
	wg       sync.WaitGroup
}

func New() *Bot {
	token := config.GetString("BOT_TOKEN")
	sess, err := discordgo.New("Bot " + token)
	if err != nil {
		slog.Error("failed to create session", "error", err)
		os.Exit(1)
	}
	if err = RegisterCommands(sess, config.GetString("GUILD_ID")); err != nil {
		slog.Error("failed to register commands", "error", err)
		os.Exit(1)
	}
	return &Bot{
		handlers: getHandlers(),
		sess:     sess,
		wg:       sync.WaitGroup{},
	}
}

func (b *Bot) Run() error {
	b.sess.AddHandler(b.handler)
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

func getHandlers() map[string]handlers.HandlerFn {
	handle := handlers.New()
	return map[string]handlers.HandlerFn{
		"yt-download": handle.Download,
		"yt-search":   handle.Search,
		"bedtime-ban": handle.Bedtime,
	}
}
