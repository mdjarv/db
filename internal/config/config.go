package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// AppConfig holds application-level settings from config.yaml.
type AppConfig struct {
	Theme string `yaml:"theme"`
}

// Load reads the app config from the config file.
// Returns zero-value config if file doesn't exist or is invalid.
func Load() AppConfig {
	var cfg AppConfig
	data, err := os.ReadFile(File())
	if err != nil {
		return cfg
	}
	_ = yaml.Unmarshal(data, &cfg)
	return cfg
}
