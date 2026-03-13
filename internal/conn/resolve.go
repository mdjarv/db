package conn

import (
	"fmt"
	"os"
	"strconv"
)

// ResolveOptions controls connection resolution priority.
type ResolveOptions struct {
	DSN       string
	ConnName  string
	Host      string
	Port      int
	User      string
	Password  string
	DBName    string
	SSLMode   string
	EnvPrefix string
}

// Resolve determines the connection config from flags, stores, env, or default.
// Stores are checked in order (project-local before global).
func Resolve(opts ResolveOptions, stores []*Store, creds *CredentialStore) (ConnectionConfig, error) {
	// 1. Explicit DSN
	if opts.DSN != "" {
		return ParseDSN(opts.DSN)
	}

	// 2. Individual flags
	if opts.Host != "" || opts.User != "" || opts.DBName != "" {
		port := opts.Port
		if port == 0 {
			port = 5432
		}
		return ConnectionConfig{
			Host:     opts.Host,
			Port:     port,
			User:     opts.User,
			Password: opts.Password,
			DBName:   opts.DBName,
			SSLMode:  opts.SSLMode,
		}, nil
	}

	// 3. Named connection from stores (first match wins)
	if opts.ConnName != "" {
		for _, store := range stores {
			if store == nil {
				continue
			}
			cfg, err := store.Get(opts.ConnName)
			if err == nil {
				if creds != nil {
					if pw, err := creds.GetPassword(opts.ConnName); err == nil {
						cfg.Password = pw
					}
				}
				return cfg, nil
			}
		}
		return ConnectionConfig{}, fmt.Errorf("connection %q not found", opts.ConnName)
	}

	// 4. Environment variables
	prefix := opts.EnvPrefix
	if prefix == "" {
		prefix = "DB"
	}
	if dsn := os.Getenv(prefix + "_DSN"); dsn != "" {
		return ParseDSN(dsn)
	}
	if envHost := os.Getenv(prefix + "_HOST"); envHost != "" {
		port := 5432
		if p := os.Getenv(prefix + "_PORT"); p != "" {
			if v, err := strconv.Atoi(p); err == nil {
				port = v
			}
		}
		return ConnectionConfig{
			Host:     envHost,
			Port:     port,
			User:     os.Getenv(prefix + "_USER"),
			Password: os.Getenv(prefix + "_PASSWORD"),
			DBName:   os.Getenv(prefix + "_NAME"),
		}, nil
	}

	// 5. Store default (first match wins)
	for _, store := range stores {
		if store == nil {
			continue
		}
		cfg, err := store.Default()
		if err == nil {
			if creds != nil {
				if pw, err := creds.GetPassword(cfg.Name); err == nil {
					cfg.Password = pw
				}
			}
			return cfg, nil
		}
	}

	return ConnectionConfig{}, fmt.Errorf("no connection configured: use --dsn, --connection, env vars, or set a default")
}
