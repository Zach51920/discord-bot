package talkingstick

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"time"
)

type DisplayType int

const (
	DisplayTSSettings DisplayType = iota
	DisplayTSPause
	DisplayTSActive
	DisplayTSNew
	DisplayTSDecommissioned
)

const (
	colorGreen  = 0x00FF00
	colorBlue   = 0x0000FF
	colorYellow = 0xFFFF00
	colorRed    = 0xFF0000
)

const displayActiveGifUrl = "https://media1.tenor.com/m/kLMNr09gHxYAAAAC/rolblox-letsdrink.gif"

type ControlPanel struct {
	prevDisplay DisplayType
	display     DisplayType
	sess        *discordgo.Session
	tss         *tsSession
	channelID   string
	messageID   string
}

func NewControlPanel(tss *tsSession) (*ControlPanel, error) {
	panel := &ControlPanel{
		sess:      tss.sess,
		channelID: tss.channelID,
		tss:       tss,
		display:   DisplayTSNew}
	return panel, panel.Create()
}

func (c *ControlPanel) Set(d DisplayType) {
	c.prevDisplay = c.display
	c.display = d
}

func (c *ControlPanel) Refresh() {
	slog.Debug("refreshing control panel", "channel_id", c.channelID, "display_type", c.display)

	edit := &discordgo.MessageEdit{
		Channel:    c.channelID,
		ID:         c.messageID,
		Embeds:     c.getEmbeds(),
		Components: c.getComponents(),
	}
	if _, err := c.sess.ChannelMessageEditComplex(edit); err != nil {
		slog.Error("failed to refresh tss control panel", "channel_id", c.channelID, "error", err)
	}
}

func (c *ControlPanel) Create() error {
	msg, err := c.sess.ChannelMessageSendComplex(c.channelID, &discordgo.MessageSend{
		Embeds:     c.getEmbeds(),
		Components: c.getComponents(),
	})
	if err != nil {
		return err
	}
	c.messageID = msg.ID
	return nil
}

func (c *ControlPanel) getEmbeds() []*discordgo.MessageEmbed {
	switch c.display {
	case DisplayTSSettings:
		return c.displaySettingsEmbed()
	case DisplayTSPause:
		return c.displayPausedEmbed()
	case DisplayTSActive:
		return c.displayActiveEmbed()
	case DisplayTSNew:
		return c.displayNewEmbed()
	case DisplayTSDecommissioned:
		return c.displayDecommissionedEmbed()
	}
	return make([]*discordgo.MessageEmbed, 0)
}

func (c *ControlPanel) getComponents() []discordgo.MessageComponent {
	switch c.display {
	case DisplayTSSettings:
		return c.displaySettingsComponents()
	case DisplayTSPause:
		return c.displayPausedComponents()
	case DisplayTSActive:
		return c.displayActiveComponents()
	case DisplayTSNew:
		return c.displayNewComponents()
	case DisplayTSDecommissioned:
	}
	return make([]discordgo.MessageComponent, 0)
}

func (c *ControlPanel) displayNewEmbed() []*discordgo.MessageEmbed {
	return []*discordgo.MessageEmbed{{
		Title:       "Talking Stick Session",
		Description: "The talking stick session is ready to begin",
		Color:       colorYellow,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Click the Start button to begin the session",
		},
	}}
}

func (c *ControlPanel) displayActiveEmbed() []*discordgo.MessageEmbed {
	stickholder := c.tss.stickholder
	return []*discordgo.MessageEmbed{{
		Title: "Talking Stick Session",
		Color: colorGreen,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    fmt.Sprintf("Current Speaker: %s", stickholder.data.User.Username),
			IconURL: stickholder.data.AvatarURL(""),
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: displayActiveGifUrl},
		Fields: []*discordgo.MessageEmbedField{{
			Name:   "Next",
			Value:  fmt.Sprintf("🔜 %s", stickholder.next.data.Mention()),
			Inline: false,
		}},
		Footer: &discordgo.MessageEmbedFooter{Text: "Use the buttons below to control the session"},
	}}
}

func (c *ControlPanel) displayPausedEmbed() []*discordgo.MessageEmbed {
	return []*discordgo.MessageEmbed{{
		Title:       "Talking Stick Session",
		Description: "The talking stick session has been paused",
		Color:       colorYellow,
	}}
}

func (c *ControlPanel) displayDecommissionedEmbed() []*discordgo.MessageEmbed {
	duration := time.Since(c.tss.startTime).Round(time.Second)
	return []*discordgo.MessageEmbed{{
		Title:       "Talking Stick Session Ended",
		Description: "The talking stick session is over",
		Color:       colorBlue,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Session Duration",
				Value:  fmt.Sprintf("⏱️ %s", duration.String()),
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Thank you for participating!",
		},
	}}
}

func (c *ControlPanel) displaySettingsEmbed() []*discordgo.MessageEmbed {
	return []*discordgo.MessageEmbed{{
		Title:       "Talking Stick Settings",
		Description: "",
		Color:       colorYellow,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Turn Duration",
				Value:  c.tss.turnDuration.String(),
				Inline: true,
			},
		},
	}}
}

func (c *ControlPanel) displayActiveComponents() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			getQuitButton(),
			getPauseButton(),
			getNextSpeakerButton(),
		},
	}}
}

func (c *ControlPanel) displayPausedComponents() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			getQuitButton(),
			getResumeButton(),
			getSettingsButton(),
		},
	}}
}

func (c *ControlPanel) displayNewComponents() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			getQuitButton(),
			getStartButton(),
			getSettingsButton(),
		},
	}}
}

func (c *ControlPanel) displaySettingsComponents() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{getDurationInput(c.tss), discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			getQuitButton(),
			getBackButton(),
		},
	}}
}

func getQuitButton() discordgo.Button {
	return discordgo.Button{
		Label:    "Quit",
		Style:    discordgo.DangerButton,
		Emoji:    discordgo.ComponentEmoji{Name: "❌"},
		CustomID: string(ActionQuitSession),
	}
}

func getStartButton() discordgo.Button {
	return discordgo.Button{
		Label:    "Start",
		Style:    discordgo.SuccessButton,
		CustomID: string(ActionTogglePlayPause),
		Emoji:    discordgo.ComponentEmoji{Name: "▶️"},
	}
}

func getResumeButton() discordgo.Button {
	return discordgo.Button{
		Label:    "Resume",
		Style:    discordgo.PrimaryButton,
		CustomID: string(ActionTogglePlayPause),
		Emoji:    discordgo.ComponentEmoji{Name: "▶️"},
	}
}

func getPauseButton() discordgo.Button {
	return discordgo.Button{
		Label:    "Pause",
		Style:    discordgo.PrimaryButton,
		CustomID: string(ActionTogglePlayPause),
		Emoji:    discordgo.ComponentEmoji{Name: "⏸️"},
	}
}

func getNextSpeakerButton() discordgo.Button {
	return discordgo.Button{
		Label:    "Next",
		Style:    discordgo.SecondaryButton,
		CustomID: string(ActionSkipUser),
		Emoji:    discordgo.ComponentEmoji{Name: "⏭️"},
	}
}

func getSettingsButton() discordgo.Button {
	return discordgo.Button{
		Label:    "Settings",
		Style:    discordgo.SecondaryButton,
		CustomID: string(ActionOpenSettings),
		Emoji:    discordgo.ComponentEmoji{Name: "⚙️"},
	}
}

func getBackButton() discordgo.Button {
	return discordgo.Button{
		Label:    "Back",
		Style:    discordgo.SecondaryButton,
		CustomID: string(ActionDisplayBack),
		Emoji:    discordgo.ComponentEmoji{Name: "🔙"},
	}
}

func getDurationInput(tss *tsSession) discordgo.ActionsRow {
	return discordgo.ActionsRow{Components: []discordgo.MessageComponent{
		discordgo.TextInput{
			CustomID:    string(ActionSetDuration),
			Label:       "Turn Duration",
			Style:       discordgo.TextInputShort,
			Placeholder: tss.turnDuration.String(),
		}}}
}

func getDurationDropdown() discordgo.ActionsRow {
	return discordgo.ActionsRow{Components: []discordgo.MessageComponent{
		discordgo.SelectMenu{
			MenuType:    discordgo.StringSelectMenu,
			CustomID:    string(ActionSetDuration),
			Placeholder: "Set turn duration...",
			MaxValues:   1,
			Options: []discordgo.SelectMenuOption{
				{
					Label: "15 seconds",
					Value: "15",
					Emoji: discordgo.ComponentEmoji{Name: "⏱️"},
				},
				{
					Label: "30 seconds",
					Value: "30",
					Emoji: discordgo.ComponentEmoji{Name: "⏱️"},
				},
				{
					Label: "1 minute",
					Value: "60",
					Emoji: discordgo.ComponentEmoji{Name: "⏱️"},
				},
				{
					Label: "2 minutes",
					Value: "120",
					Emoji: discordgo.ComponentEmoji{Name: "⏱️"},
				},
				{
					Label: "5 minutes",
					Value: "300",
					Emoji: discordgo.ComponentEmoji{Name: "⏱️"},
				},
			},
			Disabled:     false,
			ChannelTypes: nil,
		},
	}}
}
