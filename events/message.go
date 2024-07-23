package events

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/ranna-go/ranna/pkg/models"
	"log/slog"
	"strings"
)

func (h *Handler) HandleMessageCreate(s *discordgo.Session, e *discordgo.MessageCreate) {
	h.HandleMessage(s, e.Message)
}

func (h *Handler) HandleMessageUpdate(s *discordgo.Session, e *discordgo.MessageUpdate) {
	h.HandleMessage(s, e.Message)
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

	// check if the message is a code block
	block, isCode := parseCodeBlock(e.Content)
	if !isCode {
		return
	}

	if !h.isAuthorizedToExecuteCode(e) {
		return
	}

	// all checks passed, execute the code
	res, err := h.rClient.Exec(block)
	if err != nil {
		slog.Error("code execution failed", "message_id", e.ID, "error", err)
		return
	}
	embed := mapCodeBlock(block, res)

	if _, err = h.sess.ChannelMessageSendEmbedReply(e.ChannelID, embed, e.Reference()); err != nil {
		slog.Error("failed to send reply", "message_id", e.ID, "error", err)
		return
	}
	return
}

func (h *Handler) isAuthorizedToExecuteCode(e *discordgo.Message) bool {
	// todo: check if the developer assistant has been added to the channel
	// this will require some type of db to keep track of which channels it has access to

	// check if the user has the developer role
	member, err := h.sess.GuildMember(e.GuildID, e.Author.ID)
	if err != nil {
		slog.Error("failed to get member", "message_id", e.ID, "error", err)
		return false
	}
	if !hasDeveloperRole(member.Roles) {
		slog.Debug("unable to execute code", "message_id", e.ID, "error", "not a developer")
		return false
	}

	return true
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
