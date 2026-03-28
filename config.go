package main

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the ~/.standup.yaml config file.
type Config struct {
	Dirs         []string `yaml:"dirs"`
	SlackWebhook string   `yaml:"slack_webhook"`
	DefaultSince string   `yaml:"default_since"`
	NoClip       bool     `yaml:"no_clip"`
	NoSlack      bool     `yaml:"no_slack"`
	Author       string   `yaml:"author"`
}

func loadConfig() Config {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}
	}
	data, err := os.ReadFile(filepath.Join(home, ".standup.yaml"))
	if err != nil {
		return Config{} // no config file is fine
	}
	var cfg Config
	_ = yaml.Unmarshal(data, &cfg)

	// Expand ~ in dir paths
	for i, d := range cfg.Dirs {
		if strings.HasPrefix(d, "~/") {
			cfg.Dirs[i] = filepath.Join(home, d[2:])
		}
	}
	return cfg
}
