package bot

import (
	"github.com/Zach51920/discord-bot/events"
	"github.com/Zach51920/discord-bot/interactions"
	"github.com/bwmarrin/discordgo"
)

func (b *Bot) RegisterHandlers() {
	interaction := interactions.New(b.sess)
	event := events.New(b.sess, b.rClient, b.dbProvider.Get())

	b.closers = append(b.closers, interaction, event)

	b.sess.AddHandler(interaction.Handle)
	b.sess.AddHandler(event.HandleMessageCreate)
	b.sess.AddHandler(event.HandleMessageUpdate)
	b.sess.AddHandler(event.HandleReactionAdd)
	b.sess.AddHandler(b.handleLeaveGuild)
	b.sess.AddHandler(b.handleJoinGuild)
}

func (b *Bot) handleLeaveGuild(s *discordgo.Session, event *discordgo.GuildDelete) {
	b.sendAlert("bot was removed from guild \"%s\"", event.Guild.ID)
}

func (b *Bot) handleJoinGuild(s *discordgo.Session, event *discordgo.GuildCreate) {
	b.overwriteCommands(event.Guild.ID, event.Guild.Name)
	b.sendAlert("bot was added to guild \"%s\"", event.Guild.Name)
}
