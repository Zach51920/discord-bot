package config

import (
	"github.com/joho/godotenv"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

var configMap map[string]string

func init() {
	_ = godotenv.Load(".env")
	configMap = map[string]string{}

	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		configMap[pair[0]] = pair[1]
	}
}

func GetString(key string) string {
	v, ok := configMap[key]
	if !ok {
		slog.Warn("missing configMap value", "key", key)
	}
	return v
}

func GetInt(key string) int {
	v, ok := configMap[key]
	if !ok {
		slog.Warn("missing configMap value", "key", key)
	}
	val, err := strconv.Atoi(v)
	if err != nil {
		slog.Warn("configMap value is not an integer", "key", key)
	}
	return val
}
