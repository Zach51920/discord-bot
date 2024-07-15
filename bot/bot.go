package bot

import (
	"context"
	"fmt"
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
	return &Bot{
		sess: sess,
		wg:   sync.WaitGroup{},
	}
}

func (b *Bot) Run() error {
	if err := b.sess.Open(); err != nil {
		return err
	}
	slog.Info("bot started...")
	defer b.Shutdown()
	b.RegisterCommands()
	b.RegisterHandlers()

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

func (b *Bot) sendAlert(format string, a ...any) {
	content := fmt.Sprintf(format, a...)
	channelID := config.GetString("ALERT_CHANNEL_ID")
	slog.Info("sending alert", "content", content)

	b.sess.Lock()
	defer b.sess.Unlock()
	if _, err := b.sess.ChannelMessageSend(channelID, content); err != nil {
		slog.Error("failed to send alert", "error", err)
	}
}
