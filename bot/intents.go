package bot

import (
	"github.com/bwmarrin/discordgo"
	"log/slog"
)

func (b *Bot) RegisterIntents() {
	intents := map[string]discordgo.Intent{
		"guilds":                        discordgo.IntentGuilds,
		"guild_members":                 discordgo.IntentGuildMembers,
		"guild_bans":                    discordgo.IntentGuildBans,
		"guild_emojis":                  discordgo.IntentGuildEmojis,
		"guild_integrations":            discordgo.IntentGuildIntegrations,
		"guild_webhooks":                discordgo.IntentGuildWebhooks,
		"guild_invites":                 discordgo.IntentGuildInvites,
		"guild_voice_states":            discordgo.IntentGuildVoiceStates,
		"guild_presences":               discordgo.IntentGuildPresences,
		"guild_messages":                discordgo.IntentGuildMessages,
		"guild_message_reactions":       discordgo.IntentGuildMessageReactions,
		"guild_message_typing":          discordgo.IntentGuildMessageTyping,
		"direct_message":                discordgo.IntentDirectMessages,
		"direct_message_reactions":      discordgo.IntentDirectMessageReactions,
		"direct_message_typing":         discordgo.IntentDirectMessageTyping,
		"message_content":               discordgo.IntentMessageContent,
		"guild_scheduled_events":        discordgo.IntentGuildScheduledEvents,
		"auto_moderation_configuration": discordgo.IntentAutoModerationConfiguration,
		"auto_moderation_execution":     discordgo.IntentAutoModerationExecution,
	}

	for _, v := range b.config.Bot.Intents {
		intent, ok := intents[v]
		if !ok {
			slog.Warn("unknown intent", "intent", v)
			continue
		}
		b.sess.Identify.Intents = b.sess.Identify.Intents | intent
	}
}
