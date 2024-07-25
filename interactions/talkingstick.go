package interactions

import (
	"errors"
	"fmt"
	talkingstick "github.com/Zach51920/discord-bot/talkingstick"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"time"
)

func (h *Handlers) TalkingStick(s *discordgo.Session, i *discordgo.InteractionCreate) {
	handlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"start": h.talkingStickStart,
		"pass":  h.talkingStickPass,
		"end":   h.talkingStickEnd,
	}
	opts := NewRequestOptions(i.ApplicationCommandData().Options)
	subcommand, _ := opts.GetSubcommand()
	handler, exists := handlers[subcommand]
	if !exists {
		writeMessage(s, i, fmt.Sprintf("Unknown subcommand: %s", subcommand))
		return
	}
	handler(s, i)
}

func (h *Handlers) talkingStickStart(s *discordgo.Session, i *discordgo.InteractionCreate) {
	_, opts := NewRequestOptions(i.ApplicationCommandData().Options).GetSubcommand()
	turnDuration := opts.GetIntDefault("duration", 15)
	duration := time.Duration(turnDuration) * time.Second

	// get the channels active voice members
	vs, err := getVoiceState(s, i.GuildID, i.Member.User.ID)
	if err != nil {
		writeMessage(s, i, "Failed to get voice state. Are you in a voice channel?")
		return
	}

	// create a session if one doesn't already exist
	if err = h.tsManager.Create(vs.GuildID, vs.ChannelID, duration); err != nil && !errors.Is(err, talkingstick.ErrSessionExists) {
		slog.Error("failed to create talking stick session", "channel_id", vs.ChannelID, "error", err)
		return
	}
	if err = h.tsManager.Start(vs.ChannelID); err != nil {
		slog.Error("failed to start talking stick session", "channel_id", vs.ChannelID, "error", err)
		return
	}
	writeMessage(s, i, "Talking stick initiated")
}

func (h *Handlers) talkingStickPass(s *discordgo.Session, i *discordgo.InteractionCreate) {
	vs, err := getVoiceState(s, i.GuildID, i.Member.User.ID)
	if err != nil {
		writeMessage(s, i, "Failed to get voice state. Are you in a voice channel?")
		return
	}

	_, opts := NewRequestOptions(i.ApplicationCommandData().Options).GetSubcommand()
	user, ok := opts.GetUser(s)
	if ok {
		err = h.tsManager.Pass(vs.ChannelID, user.ID)
	} else {
		err = h.tsManager.Skip(vs.ChannelID)
	}

	if err != nil {
		writeMessage(s, i, err.Error())
		return
	}
	writeMessage(s, i, "Passing the talking stick")
}

func (h *Handlers) talkingStickEnd(s *discordgo.Session, i *discordgo.InteractionCreate) {
	vs, err := getVoiceState(s, i.GuildID, i.Member.User.ID)
	if err != nil {
		writeMessage(s, i, "Failed to get voice state. Are you in a voice channel?")
		return
	}

	if err = h.tsManager.End(vs.ChannelID); err != nil {
		if errors.Is(err, talkingstick.ErrNoSession) {
			writeMessage(s, i, "There are no active talking stick sessions in your current channel")
			return
		}
		slog.Error("failed to end talking stick session", "channel_id", vs.ChannelID, "error", err)
		return
	}
	writeMessage(s, i, "Terminating the talking stick")
}
