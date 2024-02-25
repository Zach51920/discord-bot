package handlers

import "github.com/bwmarrin/discordgo"

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

func (opts RequestOptions) GetStringPtr(key string) *string {
	if val, ok := opts.GetString(key); ok {
		return &val
	}
	return nil
}
