package interactions

import (
	"github.com/bwmarrin/discordgo"
)

type RequestOptions map[string]*discordgo.ApplicationCommandInteractionDataOption

func NewRequestOptions(opts []*discordgo.ApplicationCommandInteractionDataOption) RequestOptions {
	reqOptions := make(RequestOptions, len(opts))
	for _, opt := range opts {
		reqOptions[opt.Name] = opt
	}
	return reqOptions
}

func (opts RequestOptions) GetString(key string) (string, bool) {
	if opt, ok := opts[key]; ok && opt.Type == discordgo.ApplicationCommandOptionString {
		return opt.Value.(string), true
	}
	return "", false
}

func (opts RequestOptions) GetInt(key string) (int, bool) {
	if opt, ok := opts[key]; ok && opt.Type == discordgo.ApplicationCommandOptionNumber {
		return int(opt.Value.(float64)), true
	}
	return -1, false
}

func (opts RequestOptions) GetStringPtr(key string) *string {
	if val, ok := opts.GetString(key); ok {
		return &val
	}
	return nil
}

func (opts RequestOptions) GetUser(s *discordgo.Session) (*discordgo.User, bool) {
	if opt, ok := opts["user"]; ok && opt.Type == discordgo.ApplicationCommandOptionUser {
		return opt.UserValue(s), true
	}
	return nil, false
}

func (opts RequestOptions) GetRole(s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.Role, bool) {
	if opt, ok := opts["role"]; ok && opt.Type == discordgo.ApplicationCommandOptionRole {
		return opt.RoleValue(s, i.GuildID), true
	}
	return nil, false
}
