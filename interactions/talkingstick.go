package interactions

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"time"
)

func (h *Handlers) TalkingStick(s *discordgo.Session, i *discordgo.InteractionCreate) {
	handlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"start": h.talkingStickStart,
		"pass":  h.talkingStickPass,
		"end":   h.talkingStickEnd,
	}
	opts := NewRequestOptions(i.ApplicationCommandData().Options)
	handler, exists := handlers[opts.GetSubcommand()]
	if !exists {
		writeMessage(s, i, fmt.Sprintf("Unknown subcommand: %s", opts.GetSubcommand()))
		return
	}
	handler(s, i)
}

func (h *Handlers) talkingStickStart(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := NewRequestOptions(i.ApplicationCommandData().Options)
	turnDuration := opts.GetIntDefault("duration", 15)
	duration := time.Duration(turnDuration) * time.Second

	// get the channels active voice members
	vs, err := getVoiceState(s, i.GuildID, i.Member.User.ID)
	if err != nil {
		writeMessage(s, i, "Failed to get voice state. Are you in a voice channel?")
		return
	}

	if err = h.tsManager.Create(vs.GuildID, vs.ChannelID, duration); err != nil {
		writeMessage(s, i, err.Error())
		return
	}
	if err = h.tsManager.Run(vs.ChannelID); err != nil {
		writeMessage(s, i, err.Error())
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

	tss := h.tsManager.Get(vs.ChannelID)
	if tss == nil {
		writeMessage(s, i, "There are no active talking stick sessions in your current channel")
		return
	}
	ok := tss.Pass()
	if !ok {
		tss.End()
	}
	writeMessage(s, i, "Passing the talking stick")
}

func (h *Handlers) talkingStickEnd(s *discordgo.Session, i *discordgo.InteractionCreate) {
	vs, err := getVoiceState(s, i.GuildID, i.Member.User.ID)
	if err != nil {
		writeMessage(s, i, "Failed to get voice state. Are you in a voice channel?")
		return
	}

	tss := h.tsManager.Get(vs.ChannelID)
	if tss == nil {
		writeMessage(s, i, "There are no active talking stick sessions in your current channel")
		return
	}
	tss.End()
	writeMessage(s, i, "Terminating the talking stick")
}
