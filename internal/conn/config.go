// Package conn manages database connection configuration, storage, and resolution.
package conn

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ConnectionConfig holds parameters for a database connection.
type ConnectionConfig struct {
	Name     string `yaml:"name,omitempty"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode,omitempty"`
	Password string `yaml:"-"`
}

// DSN builds a postgres:// connection URL.
func (c ConnectionConfig) DSN() string {
	u := url.URL{
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%d", c.Host, c.Port),
		Path:   c.DBName,
	}
	if c.User != "" {
		if c.Password != "" {
			u.User = url.UserPassword(c.User, c.Password)
		} else {
			u.User = url.User(c.User)
		}
	}
	if c.SSLMode != "" {
		q := u.Query()
		q.Set("sslmode", c.SSLMode)
		u.RawQuery = q.Encode()
	}
	return u.String()
}

// ParseDSN parses a postgres:// URL into a ConnectionConfig.
func ParseDSN(dsn string) (ConnectionConfig, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return ConnectionConfig{}, fmt.Errorf("parse dsn: %w", err)
	}
	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return ConnectionConfig{}, fmt.Errorf("unsupported scheme %q", u.Scheme)
	}

	host := u.Hostname()
	port := 5432
	if p := u.Port(); p != "" {
		port, err = strconv.Atoi(p)
		if err != nil {
			return ConnectionConfig{}, fmt.Errorf("invalid port %q: %w", p, err)
		}
	}

	var user, password string
	if u.User != nil {
		user = u.User.Username()
		password, _ = u.User.Password()
	}

	dbname := strings.TrimPrefix(u.Path, "/")

	cfg := ConnectionConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbname,
		SSLMode:  u.Query().Get("sslmode"),
	}
	return cfg, nil
}
