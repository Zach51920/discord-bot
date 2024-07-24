package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Bot    BotConfig    `yaml:"bot"`
	Ranna  RannaConfig  `yaml:"ranna"`
	Logger LoggerConfig `yaml:"logger"`
}

type BotConfig struct {
	ApplicationID  string   `yaml:"application_id"`
	AlertChannelID string   `yaml:"alert_channel_id"`
	Alerts         bool     `yaml:"alerts"`
	Intents        []string `yaml:"intents"`
}

type RannaConfig struct {
	Endpoint  string `yaml:"endpoint"`
	Version   string `yaml:"version"`
	UserAgent string `yaml:"user_agent"`
}

type LoggerConfig struct {
	Level   string `yaml:"level"`
	Outfile string `yaml:"outfile"`
	Format  string `yaml:"format"`
}

func Load(filepath string) (Config, error) {
	yamlFile, err := os.ReadFile(filepath)
	if err != nil {
		return Config{}, fmt.Errorf("read file: %v", err)
	}

	var config Config
	if err = yaml.Unmarshal(yamlFile, &config); err != nil {
		return Config{}, fmt.Errorf("unmarshall: %v", err)
	}
	return config, nil
}
