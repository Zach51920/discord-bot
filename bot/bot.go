package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Zach51920/discord-bot/handlers"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Bot struct {
	handlers *handlers.Handlers
	session  *discordgo.Session

	wg     *sync.WaitGroup
	config *config
}

type config struct {
	Token          string `json:"botToken"`
	ApplicationID  string `json:"applicationID"`
	GuildID        string `json:"guildID"`
	AlertChannelID string `json:"alertChannelID"`
}

func New() (Bot, error) {
	cfg, err := loadConfig()
	if err != nil {
		return Bot{}, fmt.Errorf("failed to read config: %w", err)
	}

	sess, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		return Bot{}, fmt.Errorf("failed to create bot: %w", err)
	}
	return Bot{
		handlers: handlers.New(),
		session:  sess,
		config:   &cfg,
		wg:       &sync.WaitGroup{},
	}, nil
}

// Run starts the bot and keeps it running until it receives a shutdown signal.
func (b *Bot) Run() error {
	if err := b.Start(); err != nil {
		return err
	}
	defer b.Stop()

	// wait for a shutdown signal
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)
	<-shutdownChan

	return nil
}

// Start opens a new bot session.
func (b *Bot) Start() error {
	if err := b.RegisterCommands(); err != nil {
		return fmt.Errorf("failed to register commands: %w", err)
	}

	if err := b.session.Open(); err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	log.Println("bot started...")
	return nil
}

// Stop triggers a graceful shutdown. If the shutdown takes more than 5 seconds pull the plug.
func (b *Bot) Stop() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)

	b.SendAlert("shutting down bot")
	log.Println("[INFO] received shutdown signal, waiting for processes to finish...")
	go func() {
		b.wg.Wait() // wait for processes to finish
		log.Println("[INFO] processes finished successfully")
		ctxCancel()
	}()

	<-ctx.Done()
	if err := b.session.Close(); err != nil {
		b.SendAlert("failed to close session: " + err.Error())
	}
}

func loadConfig() (config, error) {
	_ = godotenv.Load(".env")
	cfg := config{
		Token:          os.Getenv("BOT_TOKEN"),
		ApplicationID:  os.Getenv("APPLICATION_ID"),
		GuildID:        os.Getenv("GUILD_ID"),
		AlertChannelID: os.Getenv("ALERT_CHANNEL_ID"),
	}
	b, _ := json.Marshal(cfg)
	log.Println("[DEBUG] config: " + string(b))

	return cfg, nil
}
