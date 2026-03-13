package conn

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type connectionsFile struct {
	Default     string                      `yaml:"default,omitempty"`
	Connections map[string]ConnectionConfig `yaml:"connections"`
}

// Store manages connection configs on disk.
type Store struct {
	path string
}

// NewStore creates a Store backed by the given file path.
func NewStore(path string) *Store {
	return &Store{path: path}
}

func (s *Store) load() (*connectionsFile, error) {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return &connectionsFile{Connections: make(map[string]ConnectionConfig)}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read connections: %w", err)
	}

	var f connectionsFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse connections: %w", err)
	}
	if f.Connections == nil {
		f.Connections = make(map[string]ConnectionConfig)
	}
	return &f, nil
}

func (s *Store) save(f *connectionsFile) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := yaml.Marshal(f)
	if err != nil {
		return fmt.Errorf("marshal connections: %w", err)
	}
	return os.WriteFile(s.path, data, 0o600)
}

// Add saves a connection config. Uses ConnectionConfig.Name as the key.
func (s *Store) Add(cfg ConnectionConfig) error {
	if cfg.Name == "" {
		return fmt.Errorf("connection name required")
	}
	f, err := s.load()
	if err != nil {
		return err
	}
	f.Connections[cfg.Name] = cfg
	return s.save(f)
}

// Get returns a connection config by name.
func (s *Store) Get(name string) (ConnectionConfig, error) {
	f, err := s.load()
	if err != nil {
		return ConnectionConfig{}, err
	}
	cfg, ok := f.Connections[name]
	if !ok {
		return ConnectionConfig{}, fmt.Errorf("connection %q not found", name)
	}
	cfg.Name = name
	return cfg, nil
}

// List returns all saved connections.
func (s *Store) List() ([]ConnectionConfig, error) {
	f, err := s.load()
	if err != nil {
		return nil, err
	}
	conns := make([]ConnectionConfig, 0, len(f.Connections))
	for name, cfg := range f.Connections {
		cfg.Name = name
		conns = append(conns, cfg)
	}
	return conns, nil
}

// Remove deletes a connection by name.
func (s *Store) Remove(name string) error {
	f, err := s.load()
	if err != nil {
		return err
	}
	if _, ok := f.Connections[name]; !ok {
		return fmt.Errorf("connection %q not found", name)
	}
	delete(f.Connections, name)
	if f.Default == name {
		f.Default = ""
	}
	return s.save(f)
}

// Default returns the default connection config.
func (s *Store) Default() (ConnectionConfig, error) {
	f, err := s.load()
	if err != nil {
		return ConnectionConfig{}, err
	}
	if f.Default == "" {
		return ConnectionConfig{}, fmt.Errorf("no default connection set")
	}
	cfg, ok := f.Connections[f.Default]
	if !ok {
		return ConnectionConfig{}, fmt.Errorf("default connection %q not found", f.Default)
	}
	cfg.Name = f.Default
	return cfg, nil
}

// DefaultName returns the name of the default connection, or empty string.
func (s *Store) DefaultName() string {
	f, err := s.load()
	if err != nil {
		return ""
	}
	return f.Default
}

// Rename changes a connection's key. Updates default if it pointed to the old name.
func (s *Store) Rename(oldName, newName string) error {
	if newName == "" {
		return fmt.Errorf("new name required")
	}
	f, err := s.load()
	if err != nil {
		return err
	}
	cfg, ok := f.Connections[oldName]
	if !ok {
		return fmt.Errorf("connection %q not found", oldName)
	}
	if _, exists := f.Connections[newName]; exists {
		return fmt.Errorf("connection %q already exists", newName)
	}
	delete(f.Connections, oldName)
	cfg.Name = newName
	f.Connections[newName] = cfg
	if f.Default == oldName {
		f.Default = newName
	}
	return s.save(f)
}

// SetDefault sets the default connection name.
func (s *Store) SetDefault(name string) error {
	f, err := s.load()
	if err != nil {
		return err
	}
	if _, ok := f.Connections[name]; !ok {
		return fmt.Errorf("connection %q not found", name)
	}
	f.Default = name
	return s.save(f)
}
