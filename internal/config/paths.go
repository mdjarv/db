// Package config provides XDG-compliant path resolution.
package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// ConnectionsFile returns the path to the global connections config file.
func ConnectionsFile() string {
	return filepath.Join(Dir(), "connections.yaml")
}

// ProjectConnectionsFile returns the project-scoped connections file,
// or empty string if not inside a git repository.
func ProjectConnectionsFile() string {
	root := GitRoot()
	if root == "" {
		return ""
	}
	// Sanitize: /home/user/git/myapp → home-user-git-myapp
	sanitized := strings.TrimPrefix(root, "/")
	sanitized = strings.ReplaceAll(sanitized, string(filepath.Separator), "-")
	return filepath.Join(Dir(), "projects", sanitized, "connections.yaml")
}

// GitRoot returns the git repository root for the current directory,
// or empty string if not inside a git repo.
func GitRoot() string {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// File returns the path to the main config file.
func File() string {
	return filepath.Join(Dir(), "config.yaml")
}

// ThemesDir returns the path to the custom themes directory.
func ThemesDir() string {
	return filepath.Join(Dir(), "themes")
}
