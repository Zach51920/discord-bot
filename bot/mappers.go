package bot

import (
	"github.com/Zach51920/discord-bot/handlers"
	"github.com/bwmarrin/discordgo"
)

func mapFiles(files []*handlers.File) []*discordgo.File {
	if len(files) == 0 {
		return make([]*discordgo.File, 0)
	}
	retFiles := make([]*discordgo.File, len(files))
	for i, f := range files {
		retFiles[i] = mapFile(f)
	}
	return retFiles
}

func mapFile(f *handlers.File) *discordgo.File {
	if f == nil {
		return nil
	}
	return &discordgo.File{
		Name:        f.Name,
		ContentType: f.ContentType,
		Reader:      f.ReaderCloser,
	}
}
