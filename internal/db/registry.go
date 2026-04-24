package db

import (
	"context"
	"fmt"
	"net/url"
	"sync"
)

var (
	mu      sync.RWMutex
	drivers = make(map[string]Driver)
)

// Register adds a named driver to the registry.
func Register(name string, d Driver) {
	mu.Lock()
	defer mu.Unlock()
	drivers[name] = d
}

// Open connects to a database using the named driver.
func Open(ctx context.Context, name, dsn string) (Conn, error) {
	mu.RLock()
	d, ok := drivers[name]
	mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("db: unknown driver %q", name)
	}
	return d.Connect(ctx, dsn)
}

// OpenDSN connects using the driver implied by the DSN's URL scheme.
func OpenDSN(ctx context.Context, dsn string) (Conn, error) {
	name, err := driverFromDSN(dsn)
	if err != nil {
		return nil, err
	}
	return Open(ctx, name, dsn)
}

func driverFromDSN(dsn string) (string, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "", fmt.Errorf("db: parse dsn: %w", err)
	}
	switch u.Scheme {
	case "postgres", "postgresql":
		return "postgres", nil
	case "sqlite", "sqlite3", "file":
		return "sqlite", nil
	}
	return "", fmt.Errorf("db: unknown dsn scheme %q", u.Scheme)
}
