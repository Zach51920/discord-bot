package bot

import (
	googletts "cloud.google.com/go/texttospeech/apiv1"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/kkdai/youtube/v2"
	"google.golang.org/api/option"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Bot struct {
	ttsClient *googletts.Client
	ytClient  youtube.Client
	session   *discordgo.Session

	wg     *sync.WaitGroup
	config *config
}

type config struct {
	Token          string `json:"botToken"`
	ApplicationID  string `json:"applicationID"`
	GuildID        string `json:"guildID"`
	AlertChannelID string `json:"alertChannelID"`

	TTSKey          string `json:"ttsAPIKey"`
	TTSLanguageCode string `json:"ttsLanguageCode"`
}

func New() (Bot, error) {
	cfg, err := loadConfig()
	if err != nil {
		return Bot{}, fmt.Errorf("failed to read config: %w", err)
	}

	ttsClient, err := googletts.NewClient(context.Background(), option.WithAPIKey(cfg.TTSKey))
	if err != nil {
		log.Println("Unable to create text to speech client. Some features may be unavailable.")
	}

	sess, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		return Bot{}, fmt.Errorf("failed to create bot: %w", err)
	}
	return Bot{
		ytClient:  youtube.Client{},
		ttsClient: ttsClient,
		session:   sess,
		config:    &cfg,
		wg:        &sync.WaitGroup{},
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
	if err := b.ttsClient.Close(); err != nil {
		b.sendAlert("failed to close tts client: " + err.Error())
	}
}

func loadConfig() (config, error) {
	_ = godotenv.Load(".env")
	cfg := config{
		Token:           os.Getenv("BOT_TOKEN"),
		ApplicationID:   os.Getenv("APPLICATION_ID"),
		GuildID:         os.Getenv("GUILD_ID"),
		AlertChannelID:  os.Getenv("ALERT_CHANNEL_ID"),
		TTSKey:          os.Getenv("TTS_API_KEY"),
		TTSLanguageCode: os.Getenv("TTS_LANGUAGE_CODE"),
	}
	b, _ := json.Marshal(cfg)
	log.Println("config: " + string(b))

	return cfg, nil
}
