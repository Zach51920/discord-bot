package events

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/ranna-go/ranna/pkg/models"
	"log/slog"
	"strings"
)

const codeExecEmoji = "⚡"

func (h *Handler) HandleMessageCreate(s *discordgo.Session, e *discordgo.MessageCreate) {
	slog.Debug("intercepted message create", "message", e.ID)
	h.HandleMessage(s, e.Message)
}

func (h *Handler) HandleMessageUpdate(s *discordgo.Session, e *discordgo.MessageUpdate) {
	slog.Debug("intercepted message update", "message", e.ID)
	h.HandleMessage(s, e.Message)
}

func (h *Handler) HandleReactionAdd(s *discordgo.Session, e *discordgo.MessageReactionAdd) {
	slog.Debug("intercepted reaction add", "message", e.MessageID)
	// if the code exec emoji was added by a non-bot user, continue
	if e.Emoji.Name != codeExecEmoji || e.Member.User.Bot {
		return
	}

	message, err := h.sess.ChannelMessage(e.ChannelID, e.MessageID)
	if err != nil {
		slog.Error("failed to get message", "message", e.MessageID)
		return
	}
	message.GuildID = e.GuildID // this is null for some reason :/

	// check if the bot already acknowledged the message by adding its emoji
	if !messageHasReaction(codeExecEmoji, message.Reactions) {
		return
	}
	h.executeCodeBlock(message, e.UserID)
}

func (h *Handler) HandleMessage(s *discordgo.Session, e *discordgo.Message) {
	if e.Author.Bot || e.Content == "" {
		return
	}

	slog.Debug("adding message to messageCh", "message", e.ID)
	h.messageCh <- e
}

func (h *Handler) handleMessage(e *discordgo.Message) {
	h.wg.Add(1)
	defer h.wg.Done()

	if isCodeBlock(e.Content) {
		h.handleCodeBlock(e)
		return
	}
	slog.Debug("not a message of interest, discarding...")
	return
}

func (h *Handler) handleCodeBlock(e *discordgo.Message) {
	// check what the code execution mode is for this channel
	var execMode string
	query := `SELECT COALESCE(
				(SELECT code_exec::text FROM channels WHERE guild_id = $1 AND channel_id = $2),
				'DISABLED') as code_exec`
	if err := h.db.Get(&execMode, query, e.GuildID, e.ChannelID); err != nil {
		slog.Error("failed to get code execution mode", "message", e.ID, "error", err)
		return
	}
	switch execMode {
	case "DISABLED":
		slog.Debug("code execution is disabled for this channel", "channel", e.ChannelID)
		return
	case "AUTO":
		h.executeCodeBlock(e, e.Author.ID)
		return
	case "MANUAL":
		if err := h.sess.MessageReactionAdd(e.ChannelID, e.ID, codeExecEmoji); err != nil {
			slog.Error("failed to acknowledge code block", "message", e.ID, "error", err)
		}
		return
	}
}

func (h *Handler) executeCodeBlock(e *discordgo.Message, requestor string) {
	slog.Debug("executing code block", "message", e.ID)

	// check if the requestor is authorized to execute code
	member, err := h.sess.GuildMember(e.GuildID, requestor)
	if err != nil {
		slog.Error("failed to get member", "message", e.ID, "guild_id", e.GuildID, "requestor", requestor, "error", err)
		h.writeReply(e, "Unable to execute code block: an unexpected error has occurred")
		return
	}
	// todo: this should check the database for what roles are 'developer' roles
	if !hasDeveloperRole(member.Roles) {
		h.writeReply(e, "Unable to execute code block: invalid permissions")
		return
	}

	block, ok := parseCodeBlock(e.Content)
	if !ok {
		slog.Warn("message is not a code block... how'd it make it this far?", "message", e.ID)
		return
	}

	// all checks passed, execute the code
	res, err := h.rClient.Exec(block)
	if err != nil {
		slog.Error("code execution failed", "message", e.ID, "error", err)
		h.writeReply(e, "Unable to execute code block: an unexpected error has occurred")
		return
	}

	embed := mapCodeBlock(block, res)
	if _, err = h.sess.ChannelMessageSendEmbedReply(e.ChannelID, embed, e.Reference()); err != nil {
		slog.Error("failed to send reply", "message", e.ID, "error", err)
		return
	}
}

func (h *Handler) writeReply(e *discordgo.Message, content string) {
	h.sess.Lock()
	defer h.sess.Unlock()

	if _, err := h.sess.ChannelMessageSendReply(e.ChannelID, content, e.Reference()); err != nil {
		slog.Error("failed to send message reply", "message", e.ID, "error", err)
		return
	}
}

func messageHasReaction(emoji string, reactions []*discordgo.MessageReactions) bool {
	for _, reaction := range reactions {
		if reaction.Emoji.Name == emoji && reaction.Me {
			return true
		}
	}
	return false
}

func mapCodeBlock(block models.ExecutionRequest, res models.ExecutionResponse) *discordgo.MessageEmbed {
	color := 0x42f56c // green
	if res.StdErr != "" {
		color = 0xff0000 // red
	}

	return &discordgo.MessageEmbed{
		Type:  "rich",
		Title: block.Language,
		Color: color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "StdOut",
				Value:  res.StdOut,
				Inline: false,
			},
			{
				Name:   "StdErr",
				Value:  res.StdErr,
				Inline: false,
			},
			{
				Name:   "Exec Time",
				Value:  fmt.Sprintf("%vms", res.ExecTimeMS),
				Inline: false,
			},
		},
	}
}

func isCodeBlock(content string) bool {
	spl := strings.Split(content, "```")
	if len(spl) < 3 {
		return false
	}

	first := strings.Index(spl[1], "\n")
	if first < 0 || len(spl[1]) < first {
		return false
	}
	return true
}

func parseCodeBlock(content string) (models.ExecutionRequest, bool) {
	if content == "" {
		return models.ExecutionRequest{}, false
	}

	spl := strings.Split(content, "```")
	if len(spl) < 3 {
		return models.ExecutionRequest{}, false
	}

	inner := spl[1]
	first := strings.Index(inner, "\n")
	if first < 0 || len(inner) < first {
		return models.ExecutionRequest{}, false
	}

	language := inner[:first]
	code := inner[first+1:]

	return models.ExecutionRequest{
		Language:         mapLang(language),
		Code:             code,
		InlineExpression: false,
		Arguments:        make([]string, 0),
		Environment:      make(map[string]string),
	}, true
}

func mapLang(lang string) string {
	if lang == "go" || lang == "golang" {
		return "gotip"
	}
	return lang
}
