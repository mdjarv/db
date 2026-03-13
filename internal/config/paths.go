// Package config provides XDG-compliant path resolution.
package config

import (
	"os"
	"path/filepath"
)

const appName = "db"

// Dir returns the application config directory (~/.config/db).
func Dir() string {
	base, err := os.UserConfigDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, appName)
}

// DataDir returns the application data directory (~/.local/share/db).
func DataDir() string {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return filepath.Join(dir, appName)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", appName)
}

// ConnectionsFile returns the path to the connections config file.
func ConnectionsFile() string {
	return filepath.Join(Dir(), "connections.yaml")
}
