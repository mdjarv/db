package conn

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Source identifies where a connection candidate was found.
type Source string

// Connection source constants.
const (
	SourceProjectStore Source = "project"
	SourceGlobalStore  Source = "global"
	SourceEnvVar       Source = "env"
	SourceDotEnv       Source = "dotenv"
)

// Candidate is a discovered connection with provenance metadata.
type Candidate struct {
	Config    ConnectionConfig
	Source    Source
	Label     string
	IsDefault bool
}

// DiscoverOptions configures connection discovery.
type DiscoverOptions struct {
	Stores  []*Store
	Creds   *CredentialStore
	GitRoot string
}

type envPattern struct {
	Name    string
	Match   func(lookup func(string) string) bool
	Extract func(lookup func(string) string) ConnectionConfig
}

var defaultPatterns = []envPattern{
	databaseURLPattern,
	dbPrefixPattern,
	pgPattern,
}

var databaseURLPattern = envPattern{
	Name: "DATABASE_URL",
	Match: func(lookup func(string) string) bool {
		return lookup("DATABASE_URL") != ""
	},
	Extract: func(lookup func(string) string) ConnectionConfig {
		cfg, err := ParseDSN(lookup("DATABASE_URL"))
		if err != nil {
			return ConnectionConfig{}
		}
		return cfg
	},
}

var dbPrefixPattern = envPattern{
	Name: "DB_*",
	Match: func(lookup func(string) string) bool {
		return lookup("DB_DSN") != "" || lookup("DB_HOST") != ""
	},
	Extract: func(lookup func(string) string) ConnectionConfig {
		if dsn := lookup("DB_DSN"); dsn != "" {
			cfg, err := ParseDSN(dsn)
			if err != nil {
				return ConnectionConfig{}
			}
			return cfg
		}
		port := 5432
		if p := lookup("DB_PORT"); p != "" {
			if v, err := strconv.Atoi(p); err == nil {
				port = v
			}
		}
		return ConnectionConfig{
			Host:     lookup("DB_HOST"),
			Port:     port,
			User:     lookup("DB_USER"),
			Password: lookup("DB_PASSWORD"),
			DBName:   lookup("DB_NAME"),
			SSLMode:  lookup("DB_SSLMODE"),
		}
	},
}

var pgPattern = envPattern{
	Name: "PG*",
	Match: func(lookup func(string) string) bool {
		return lookup("PGHOST") != ""
	},
	Extract: func(lookup func(string) string) ConnectionConfig {
		port := 5432
		if p := lookup("PGPORT"); p != "" {
			if v, err := strconv.Atoi(p); err == nil {
				port = v
			}
		}
		return ConnectionConfig{
			Host:     lookup("PGHOST"),
			Port:     port,
			User:     lookup("PGUSER"),
			Password: lookup("PGPASSWORD"),
			DBName:   lookup("PGDATABASE"),
			SSLMode:  lookup("PGSSLMODE"),
		}
	},
}

// Discover collects connection candidates from stores, env vars, and dotenv files.
func Discover(opts DiscoverOptions) []Candidate {
	var candidates []Candidate
	seen := make(map[string]bool)

	add := func(c Candidate) {
		dsn := c.Config.DSN()
		if seen[dsn] {
			return
		}
		seen[dsn] = true
		candidates = append(candidates, c)
	}

	// 1. Saved connections from stores.
	for i, store := range opts.Stores {
		if store == nil {
			continue
		}
		source := SourceGlobalStore
		labelPrefix := "global"
		if len(opts.Stores) >= 2 && i == 0 {
			source = SourceProjectStore
			labelPrefix = "project"
		}
		defaultName := store.DefaultName()
		conns, err := store.List()
		if err != nil {
			continue
		}
		for _, cfg := range conns {
			if opts.Creds != nil && cfg.Name != "" {
				if pw, err := opts.Creds.GetPassword(cfg.Name); err == nil {
					cfg.Password = pw
				}
			}
			add(Candidate{
				Config:    cfg,
				Source:    source,
				Label:     labelPrefix,
				IsDefault: cfg.Name != "" && cfg.Name == defaultName,
			})
		}
	}

	// 2. Process env vars.
	envLookup := func(key string) string { return os.Getenv(key) }
	for _, pat := range defaultPatterns {
		if pat.Match(envLookup) {
			cfg := pat.Extract(envLookup)
			add(Candidate{
				Config: cfg,
				Source: SourceEnvVar,
				Label:  "env:" + pat.Name,
			})
		}
	}

	// 3. Dotenv files.
	if opts.GitRoot != "" {
		dotenvFiles := []string{".env", ".env.local", ".env.development", ".env.development.local"}
		for _, name := range dotenvFiles {
			path := filepath.Join(opts.GitRoot, name)
			vars, err := parseDotEnv(path)
			if err != nil || len(vars) == 0 {
				continue
			}
			lookup := func(key string) string { return vars[key] }
			for _, pat := range defaultPatterns {
				if pat.Match(lookup) {
					cfg := pat.Extract(lookup)
					add(Candidate{
						Config: cfg,
						Source: SourceDotEnv,
						Label:  name + ":" + pat.Name,
					})
				}
			}
		}
	}

	return candidates
}

// parseDotEnv reads a .env file into key-value pairs.
// Returns error only on file-read failure; malformed lines are skipped.
func parseDotEnv(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	vars := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		idx := strings.IndexByte(line, '=')
		if idx < 1 {
			continue
		}
		key := line[:idx]
		val := line[idx+1:]
		// Strip surrounding quotes.
		singleQuoted := false
		if len(val) >= 2 {
			if val[0] == '\'' && val[len(val)-1] == '\'' {
				singleQuoted = true
				val = val[1 : len(val)-1]
			} else if val[0] == '"' && val[len(val)-1] == '"' {
				val = val[1 : len(val)-1]
			}
		}
		// Expand ${VAR} and $VAR references (skip single-quoted values).
		if !singleQuoted {
			val = expandVars(val, vars)
		}
		vars[key] = val
	}
	return vars, nil
}

// expandVars replaces ${VAR} and $VAR references with values from the map.
func expandVars(s string, vars map[string]string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if s[i] != '$' {
			b.WriteByte(s[i])
			i++
			continue
		}
		i++ // skip '$'
		if i >= len(s) {
			b.WriteByte('$')
			break
		}
		// ${VAR} form
		if s[i] == '{' {
			end := strings.IndexByte(s[i:], '}')
			if end < 0 {
				b.WriteString("${")
				i++
				continue
			}
			name := s[i+1 : i+end]
			b.WriteString(vars[name])
			i += end + 1
			continue
		}
		// $VAR form — take [A-Za-z0-9_]+
		start := i
		for i < len(s) && (s[i] == '_' || (s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= 'a' && s[i] <= 'z') || (s[i] >= '0' && s[i] <= '9')) {
			i++
		}
		if i == start {
			b.WriteByte('$')
			continue
		}
		b.WriteString(vars[s[start:i]])
	}
	return b.String()
}
