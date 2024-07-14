package config

import (
	"github.com/joho/godotenv"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

var config map[string]string

func init() {
	_ = godotenv.Load(".env")
	config = map[string]string{}

	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		config[pair[0]] = pair[1]
	}
}

func GetString(key string) string {
	v, ok := config[key]
	if !ok {
		slog.Warn("missing config value", "key", key)
	}
	return v
}

func GetInt(key string) int {
	v, ok := config[key]
	if !ok {
		slog.Warn("missing config value", "key", key)
	}
	val, err := strconv.Atoi(v)
	if err != nil {
		slog.Warn("config value is not an integer", "key", key)
	}
	return val
}
