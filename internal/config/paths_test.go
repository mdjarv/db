package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDir(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		dir := Dir()
		if !strings.HasSuffix(dir, filepath.Join("db")) {
			t.Errorf("Dir() = %q, want suffix 'db'", dir)
		}
	})

	t.Run("respects XDG_CONFIG_HOME", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-config")
		dir := Dir()
		want := filepath.Join("/tmp/xdg-config", "db")
		if dir != want {
			t.Errorf("Dir() = %q, want %q", dir, want)
		}
	})
}

func TestDataDir(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", "")
		dir := DataDir()
		home, _ := os.UserHomeDir()
		want := filepath.Join(home, ".local", "share", "db")
		if dir != want {
			t.Errorf("DataDir() = %q, want %q", dir, want)
		}
	})

	t.Run("respects XDG_DATA_HOME", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", "/tmp/xdg-data")
		dir := DataDir()
		want := filepath.Join("/tmp/xdg-data", "db")
		if dir != want {
			t.Errorf("DataDir() = %q, want %q", dir, want)
		}
	})
}

func TestConnectionsFile(t *testing.T) {
	f := ConnectionsFile()
	if !strings.HasSuffix(f, "connections.yaml") {
		t.Errorf("ConnectionsFile() = %q, want suffix 'connections.yaml'", f)
	}
}
