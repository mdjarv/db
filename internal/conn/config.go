// Package conn manages database connection configuration, storage, and resolution.
package conn

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Driver identifiers.
const (
	DriverPostgres = "postgres"
	DriverSQLite   = "sqlite"
)

// ConnectionConfig holds parameters for a database connection.
type ConnectionConfig struct {
	Name     string `yaml:"name,omitempty"`
	Driver   string `yaml:"driver,omitempty"` // "postgres" (default) or "sqlite"
	Host     string `yaml:"host,omitempty"`
	Port     int    `yaml:"port,omitempty"`
	User     string `yaml:"user,omitempty"`
	DBName   string `yaml:"dbname,omitempty"`
	SSLMode  string `yaml:"sslmode,omitempty"`
	Path     string `yaml:"path,omitempty"` // for file-based drivers (sqlite)
	Password string `yaml:"-"`
}

// DriverOrDefault returns the configured driver, defaulting to postgres.
func (c ConnectionConfig) DriverOrDefault() string {
	if c.Driver == "" {
		return DriverPostgres
	}
	return c.Driver
}

// DSN builds a connection URL appropriate for the configured driver.
func (c ConnectionConfig) DSN() string {
	switch c.DriverOrDefault() {
	case DriverSQLite:
		return sqliteDSN(c)
	default:
		return postgresDSN(c)
	}
}

func postgresDSN(c ConnectionConfig) string {
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
	sslmode := c.SSLMode
	if sslmode == "" && isLocalHost(c.Host) {
		sslmode = "disable"
	}
	if sslmode != "" {
		q := u.Query()
		q.Set("sslmode", sslmode)
		u.RawQuery = q.Encode()
	}
	return u.String()
}

func sqliteDSN(c ConnectionConfig) string {
	if c.Path == "" {
		return "sqlite://:memory:"
	}
	return "sqlite://" + c.Path
}

func isLocalHost(host string) bool {
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

// DriverFromScheme maps a URL scheme to a driver name. Returns "" if unknown.
func DriverFromScheme(scheme string) string {
	switch scheme {
	case "postgres", "postgresql":
		return DriverPostgres
	case "sqlite", "sqlite3", "file":
		return DriverSQLite
	}
	return ""
}

// ParseDSN parses a connection URL into a ConnectionConfig. Supports
// postgres://, postgresql://, sqlite://, sqlite3://, and file: schemes.
func ParseDSN(dsn string) (ConnectionConfig, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return ConnectionConfig{}, fmt.Errorf("parse dsn: %w", err)
	}

	driver := DriverFromScheme(u.Scheme)
	if driver == "" {
		return ConnectionConfig{}, fmt.Errorf("unsupported scheme %q", u.Scheme)
	}

	switch driver {
	case DriverSQLite:
		return parseSQLiteDSN(dsn, u)
	default:
		return parsePostgresDSN(u)
	}
}

func parsePostgresDSN(u *url.URL) (ConnectionConfig, error) {
	host := u.Hostname()
	port := 5432
	if p := u.Port(); p != "" {
		v, err := strconv.Atoi(p)
		if err != nil {
			return ConnectionConfig{}, fmt.Errorf("invalid port %q: %w", p, err)
		}
		port = v
	}

	var user, password string
	if u.User != nil {
		user = u.User.Username()
		password, _ = u.User.Password()
	}

	dbname := strings.TrimPrefix(u.Path, "/")

	return ConnectionConfig{
		Driver:   DriverPostgres,
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbname,
		SSLMode:  u.Query().Get("sslmode"),
	}, nil
}

func parseSQLiteDSN(raw string, u *url.URL) (ConnectionConfig, error) {
	// sqlite://path/to/file.db → Host="path" Path="/to/file.db"
	// sqlite:///abs/path.db → Host="" Path="/abs/path.db"
	// sqlite://:memory:
	// file:./rel.db
	var path string
	switch u.Scheme {
	case "file":
		path = strings.TrimPrefix(raw, "file:")
	default:
		if u.Opaque != "" {
			path = u.Opaque
		} else if u.Host == ":memory:" {
			path = ":memory:"
		} else {
			path = u.Host + u.Path
		}
	}
	if path == "" {
		return ConnectionConfig{}, fmt.Errorf("sqlite dsn missing path")
	}
	return ConnectionConfig{Driver: DriverSQLite, Path: path}, nil
}
