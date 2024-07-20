package bot

import "github.com/bwmarrin/discordgo"

func (b *Bot) RegisterIntents() {
	return
	b.sess.Identify.Intents = discordgo.IntentGuildMessages |
		discordgo.IntentsDirectMessageReactions |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsDirectMessageReactions
}
