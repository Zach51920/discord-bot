package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Bot         BotConfig   `yaml:"bot"`
	Permissions Permissions `yaml:"permissions"`
	Ranna       RannaConfig `yaml:"ranna"`
}

type BotConfig struct {
	ApplicationID  string   `yaml:"application_id"`
	AlertChannelID string   `yaml:"alert_channel_id"`
	LogLevel       string   `yaml:"log_level"`
	Alerts         bool     `yaml:"alerts"`
	Commands       []string `yaml:"commands"`
}

type Permissions struct {
	CodeExecutor CodeExecutorConfig `yaml:"code_executor"`
}

type CodeExecutorConfig struct {
	Roles []string `yaml:"roles"`
}

type RannaConfig struct {
	Endpoint  string `yaml:"endpoint"`
	Version   string `yaml:"version"`
	UserAgent string `yaml:"user_agent"`
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
