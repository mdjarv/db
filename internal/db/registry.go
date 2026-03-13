package db

import (
	"context"
	"fmt"
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
