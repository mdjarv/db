package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// AppConfig holds application-level settings from config.yaml.
type AppConfig struct {
	Theme string `yaml:"theme"`
}

// Load reads the app config from the config file.
// Returns zero-value config if file doesn't exist.
func Load() (AppConfig, error) {
	var cfg AppConfig
	data, err := os.ReadFile(File())
	if err != nil {
		return cfg, nil // file not existing is not an error
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("config: %w", err)
	}
	return cfg, nil
}
