package logger

import (
	"github.com/Zach51920/discord-bot/config"
	"log/slog"
	"os"
	"strings"
)

func Init(cfg config.LoggerConfig) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: getLevel(cfg),
	}))
	slog.SetDefault(logger)
}

func getLevel(cfg config.LoggerConfig) *slog.LevelVar {
	levels := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}
	level, ok := levels[strings.ToLower(cfg.Level)]
	if !ok {
		level = slog.LevelInfo
	}

	lvl := new(slog.LevelVar)
	lvl.Set(level)
	return lvl
}
