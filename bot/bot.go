package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/kkdai/youtube/v2"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Bot struct {
	youtube youtube.Client
	session *discordgo.Session

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
		return Bot{}, fmt.Errorf("failed to create youpirate: %w", err)
	}

	return Bot{
		youtube: youtube.Client{},
		session: sess,
		config:  &cfg,
		wg:      &sync.WaitGroup{},
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
	if err := b.session.Open(); err != nil {
		return err
	}
	log.Println("bot started...")
	return nil
}

// Stop triggers a graceful shutdown. If the shutdown takes more than 5 seconds pull the plug.
func (b *Bot) Stop() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)

	b.sendAlert("shutting down bot")
	log.Println("received shutdown signal, waiting for processes to finish...")
	go func() {
		b.wg.Wait() // wait for processes to finish
		log.Println("processes finished successfully")
		ctxCancel()
	}()

	<-ctx.Done()
	if err := b.session.Close(); err != nil {
		b.sendAlert("failed to close session: " + err.Error())
	}
}

// sendAlert sends an alert to the alert channel
func (b *Bot) sendAlert(alert string) {
	log.Println("[ALERT]:", alert)

	b.wg.Add(1)
	defer b.wg.Done()

	b.session.Lock()
	defer b.session.Unlock()
	if _, err := b.session.ChannelMessageSend(b.config.AlertChannelID, alert); err != nil {
		log.Println("[WARNING] failed to write alert:", err)
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
	fmt.Println("config:", string(b))

	return cfg, nil
}
