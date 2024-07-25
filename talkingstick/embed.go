package talkingstick

import (
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"time"
)

func (tss *tsSession) CreateControlPanel() error {
	slog.Debug("creating control panel", "channel_id", tss.channelID)

	message := &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{getEmbed(tss)},
		Components: getComponents(tss),
	}
	msg, err := tss.sess.ChannelMessageSendComplex(tss.channelID, message)
	tss.embed = msg
	return err
}

func (tss *tsSession) RefreshControlPanel() {
	slog.Debug("refreshing control panel", "channel_id", tss.channelID)

	edit := &discordgo.MessageEdit{
		Channel:    tss.channelID,
		ID:         tss.embed.ID,
		Embeds:     []*discordgo.MessageEmbed{getEmbed(tss)},
		Components: getComponents(tss),
	}
	if _, err := tss.sess.ChannelMessageEditComplex(edit); err != nil {
		slog.Error("failed to refresh tss control panel", "channel_id", tss.channelID, "error", err)
	}
}

func (tss *tsSession) DecommissionControlPanel() error {
	slog.Debug("decommissioning control panel", "channel_id", tss.channelID)

	duration := time.Since(tss.startTime).Round(time.Second)
	edit := &discordgo.MessageEdit{
		Channel: tss.channelID,
		ID:      tss.embed.ID,
		Embeds: []*discordgo.MessageEmbed{{
			Title:       "Talking Stick Session Ended",
			Description: "The talking stick session is over",
			Color:       0x0000FF, // Blue
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Session Duration",
					Value:  duration.String(),
					Inline: false,
				},
			},
		}},
		Components: []discordgo.MessageComponent{},
	}
	_, err := tss.sess.ChannelMessageEditComplex(edit)
	return err
}

func getEmbed(tss *tsSession) *discordgo.MessageEmbed {
	stickholder := tss.stickholder
	return &discordgo.MessageEmbed{
		Title: "Talking Stick Session",
		Color: 0x00FF00, // Green color
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Current Speaker",
				Value:  stickholder.data.Mention(),
				Inline: false,
			},
			{
				Name:   "Next Speaker",
				Value:  stickholder.next.data.Mention(),
				Inline: false,
			},
			{
				Name:   "Session Status",
				Value:  getStatusEmoji(tss.Running()),
				Inline: false,
			},
		},
	}
}

func getComponents(tss *tsSession) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Quit",
					Style:    discordgo.DangerButton,
					CustomID: "talking_stick_quit",
					Emoji: discordgo.ComponentEmoji{
						Name: "❌",
					},
				},
				discordgo.Button{
					Label:    getPlayPauseLabel(tss.Running()),
					Style:    discordgo.PrimaryButton,
					CustomID: "talking_stick_playpause",
					Emoji: discordgo.ComponentEmoji{
						Name: getPlayPauseEmoji(tss.Running()),
					},
				},
				discordgo.Button{
					Label:    "Next Speaker",
					Style:    discordgo.SecondaryButton,
					CustomID: "talking_stick_next",
					Emoji: discordgo.ComponentEmoji{
						Name: "⏭️",
					},
				},
			},
		},
		//discordgo.ActionsRow{
		//	Components: []discordgo.MessageComponent{
		//		discordgo.SelectMenu{
		//			MenuType:    discordgo.UserSelectMenu,
		//			CustomID:    "talking_stick_pass",
		//			Placeholder: "Pass Talking Stick to...",
		//		},
		//	},
		//},
	}
}

func getStatusEmoji(isRunning bool) string {
	if isRunning {
		return "🟢 Running"
	}
	return "🔴 Paused"
}

func getPlayPauseLabel(isRunning bool) string {
	if isRunning {
		return "Pause"
	}
	return "Play"
}

func getPlayPauseEmoji(isRunning bool) string {
	if isRunning {
		return "⏸️"
	}
	return "▶️"
}
