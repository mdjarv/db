package conn

import (
	"os"
	"path/filepath"
	"testing"
)

func writeStoreFile(t *testing.T, path string, conns map[string]ConnectionConfig, defaultName string) {
	t.Helper()
	s := NewStore(path)
	for _, cfg := range conns {
		if err := s.Add(cfg); err != nil {
			t.Fatal(err)
		}
	}
	if defaultName != "" {
		if err := s.SetDefault(defaultName); err != nil {
			t.Fatal(err)
		}
	}
}

func TestParseDotEnv(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    map[string]string
	}{
		{
			name:    "basic key=value",
			content: "FOO=bar\nBAZ=qux",
			want:    map[string]string{"FOO": "bar", "BAZ": "qux"},
		},
		{
			name:    "comments and blanks",
			content: "# comment\n\nKEY=val\n  \n# another",
			want:    map[string]string{"KEY": "val"},
		},
		{
			name:    "export prefix",
			content: "export HOST=localhost\nexport PORT=5432",
			want:    map[string]string{"HOST": "localhost", "PORT": "5432"},
		},
		{
			name:    "double quoted",
			content: `DB_URL="postgres://u@h:5432/d"`,
			want:    map[string]string{"DB_URL": "postgres://u@h:5432/d"},
		},
		{
			name:    "single quoted",
			content: "SECRET='s3cret'",
			want:    map[string]string{"SECRET": "s3cret"},
		},
		{
			name:    "unquoted",
			content: "PLAIN=value",
			want:    map[string]string{"PLAIN": "value"},
		},
		{
			name:    "empty value",
			content: "EMPTY=",
			want:    map[string]string{"EMPTY": ""},
		},
		{
			name:    "malformed lines skipped",
			content: "GOOD=yes\nno_equals\n=no_key\nALSO=ok",
			want:    map[string]string{"GOOD": "yes", "ALSO": "ok"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), ".env")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}
			got, err := parseDotEnv(path)
			if err != nil {
				t.Fatalf("parseDotEnv: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("len = %d, want %d; got %v", len(got), len(tt.want), got)
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("key %q = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}

func TestExpandVars(t *testing.T) {
	vars := map[string]string{"USER": "dev", "HOST": "localhost", "PORT": "5432"}

	tests := []struct {
		input string
		want  string
	}{
		{"${USER}@${HOST}", "dev@localhost"},
		{"$USER@$HOST:$PORT", "dev@localhost:5432"},
		{"no vars here", "no vars here"},
		{"${MISSING}", ""},
		{"prefix_${USER}_suffix", "prefix_dev_suffix"},
		{"$$USER", "$dev"},
	}
	for _, tt := range tests {
		got := expandVars(tt.input, vars)
		if got != tt.want {
			t.Errorf("expandVars(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseDotEnv_VarInterpolation(t *testing.T) {
	content := `DB_HOST=localhost
DB_PORT=5432
DB_USER=dev
DB_PASSWORD=secret
DB_NAME=mydb
DATABASE_URL=postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable
`
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := parseDotEnv(path)
	if err != nil {
		t.Fatalf("parseDotEnv: %v", err)
	}
	want := "postgresql://dev:secret@localhost:5432/mydb?sslmode=disable"
	if got["DATABASE_URL"] != want {
		t.Errorf("DATABASE_URL = %q, want %q", got["DATABASE_URL"], want)
	}
}

func TestParseDotEnv_SingleQuoteNoExpand(t *testing.T) {
	content := "VAR=hello\nLITERAL='${VAR}'\n"
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := parseDotEnv(path)
	if err != nil {
		t.Fatalf("parseDotEnv: %v", err)
	}
	if got["LITERAL"] != "${VAR}" {
		t.Errorf("LITERAL = %q, want literal ${VAR}", got["LITERAL"])
	}
}

func TestParseDotEnv_MissingFile(t *testing.T) {
	_, err := parseDotEnv(filepath.Join(t.TempDir(), "nope"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestDiscover_Stores(t *testing.T) {
	projDir := t.TempDir()
	globalDir := t.TempDir()
	projPath := filepath.Join(projDir, "connections.yaml")
	globalPath := filepath.Join(globalDir, "connections.yaml")

	writeStoreFile(t, projPath, map[string]ConnectionConfig{
		"dev": {Name: "dev", Host: "localhost", Port: 5432, User: "dev", DBName: "devdb"},
	}, "dev")
	writeStoreFile(t, globalPath, map[string]ConnectionConfig{
		"prod": {Name: "prod", Host: "db.prod", Port: 5432, User: "admin", DBName: "proddb"},
	}, "prod")

	kr := NewMemoryKeyring()
	cs := NewCredentialStore(kr)
	_ = cs.SetPassword("dev", "devpass")
	_ = cs.SetPassword("prod", "prodpass")

	candidates := Discover(DiscoverOptions{
		Stores: []*Store{NewStore(projPath), NewStore(globalPath)},
		Creds:  cs,
	})

	if len(candidates) != 2 {
		t.Fatalf("len = %d, want 2", len(candidates))
	}

	// Find project and global candidates.
	var proj, global *Candidate
	for i := range candidates {
		switch candidates[i].Source {
		case SourceProjectStore:
			proj = &candidates[i]
		case SourceGlobalStore:
			global = &candidates[i]
		}
	}

	if proj == nil {
		t.Fatal("no project candidate")
	}
	if proj.Config.Host != "localhost" || proj.Config.Password != "devpass" {
		t.Errorf("project candidate = %+v", proj.Config)
	}
	if !proj.IsDefault {
		t.Error("project candidate should be default")
	}
	if proj.Label != "project" {
		t.Errorf("project label = %q", proj.Label)
	}

	if global == nil {
		t.Fatal("no global candidate")
	}
	if global.Config.Host != "db.prod" || global.Config.Password != "prodpass" {
		t.Errorf("global candidate = %+v", global.Config)
	}
	if !global.IsDefault {
		t.Error("global candidate should be default")
	}
	if global.Label != "global" {
		t.Errorf("global label = %q", global.Label)
	}
}

func TestDiscover_SingleStore(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "connections.yaml")
	writeStoreFile(t, path, map[string]ConnectionConfig{
		"solo": {Name: "solo", Host: "h", Port: 5432, User: "u", DBName: "d"},
	}, "")

	candidates := Discover(DiscoverOptions{
		Stores: []*Store{NewStore(path)},
	})

	if len(candidates) != 1 {
		t.Fatalf("len = %d, want 1", len(candidates))
	}
	if candidates[0].Source != SourceGlobalStore {
		t.Errorf("single store source = %q, want global", candidates[0].Source)
	}
	if candidates[0].IsDefault {
		t.Error("should not be default when none set")
	}
}

func TestDiscover_EnvPatterns(t *testing.T) {
	t.Run("DATABASE_URL", func(t *testing.T) {
		lookup := func(key string) string {
			if key == "DATABASE_URL" {
				return "postgres://eu@eh:5433/edb"
			}
			return ""
		}
		if !databaseURLPattern.Match(lookup) {
			t.Fatal("should match")
		}
		cfg := databaseURLPattern.Extract(lookup)
		if cfg.Host != "eh" || cfg.Port != 5433 || cfg.User != "eu" || cfg.DBName != "edb" {
			t.Errorf("got %+v", cfg)
		}
	})

	t.Run("DATABASE_URL invalid", func(t *testing.T) {
		lookup := func(key string) string {
			if key == "DATABASE_URL" {
				return "not-a-url"
			}
			return ""
		}
		cfg := databaseURLPattern.Extract(lookup)
		if cfg.Host != "" {
			t.Errorf("expected empty config for invalid URL, got %+v", cfg)
		}
	})

	t.Run("DB_DSN", func(t *testing.T) {
		lookup := func(key string) string {
			m := map[string]string{"DB_DSN": "postgres://du@dh:5432/ddb"}
			return m[key]
		}
		if !dbPrefixPattern.Match(lookup) {
			t.Fatal("should match")
		}
		cfg := dbPrefixPattern.Extract(lookup)
		if cfg.Host != "dh" || cfg.User != "du" || cfg.DBName != "ddb" {
			t.Errorf("got %+v", cfg)
		}
	})

	t.Run("DB_HOST family", func(t *testing.T) {
		lookup := func(key string) string {
			m := map[string]string{
				"DB_HOST":     "dbhost",
				"DB_PORT":     "5433",
				"DB_USER":     "dbuser",
				"DB_PASSWORD": "dbpass",
				"DB_NAME":     "dbname",
			}
			return m[key]
		}
		if !dbPrefixPattern.Match(lookup) {
			t.Fatal("should match")
		}
		cfg := dbPrefixPattern.Extract(lookup)
		if cfg.Host != "dbhost" || cfg.Port != 5433 || cfg.User != "dbuser" ||
			cfg.Password != "dbpass" || cfg.DBName != "dbname" {
			t.Errorf("got %+v", cfg)
		}
	})

	t.Run("DB_HOST default port", func(t *testing.T) {
		lookup := func(key string) string {
			if key == "DB_HOST" {
				return "h"
			}
			return ""
		}
		cfg := dbPrefixPattern.Extract(lookup)
		if cfg.Port != 5432 {
			t.Errorf("port = %d, want 5432", cfg.Port)
		}
	})

	t.Run("PGHOST family", func(t *testing.T) {
		lookup := func(key string) string {
			m := map[string]string{
				"PGHOST":     "pghost",
				"PGPORT":     "5434",
				"PGUSER":     "pguser",
				"PGPASSWORD": "pgpass",
				"PGDATABASE": "pgdb",
			}
			return m[key]
		}
		if !pgPattern.Match(lookup) {
			t.Fatal("should match")
		}
		cfg := pgPattern.Extract(lookup)
		if cfg.Host != "pghost" || cfg.Port != 5434 || cfg.User != "pguser" ||
			cfg.Password != "pgpass" || cfg.DBName != "pgdb" {
			t.Errorf("got %+v", cfg)
		}
	})

	t.Run("PGHOST default port", func(t *testing.T) {
		lookup := func(key string) string {
			if key == "PGHOST" {
				return "h"
			}
			return ""
		}
		cfg := pgPattern.Extract(lookup)
		if cfg.Port != 5432 {
			t.Errorf("port = %d, want 5432", cfg.Port)
		}
	})

	t.Run("no match", func(t *testing.T) {
		lookup := func(string) string { return "" }
		for _, pat := range defaultPatterns {
			if pat.Match(lookup) {
				t.Errorf("%s should not match empty lookup", pat.Name)
			}
		}
	})
}

func TestDiscover_DotEnv(t *testing.T) {
	root := t.TempDir()
	content := "DATABASE_URL=postgres://du@dh:5432/ddb\n"
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	candidates := Discover(DiscoverOptions{GitRoot: root})
	if len(candidates) != 1 {
		t.Fatalf("len = %d, want 1", len(candidates))
	}
	c := candidates[0]
	if c.Source != SourceDotEnv {
		t.Errorf("source = %q, want dotenv", c.Source)
	}
	if c.Config.Host != "dh" || c.Config.User != "du" || c.Config.DBName != "ddb" {
		t.Errorf("config = %+v", c.Config)
	}
	if c.Label != ".env:DATABASE_URL" {
		t.Errorf("label = %q", c.Label)
	}
}

func TestDiscover_Dedup(t *testing.T) {
	// Store and dotenv with same DSN — store wins.
	dir := t.TempDir()
	storePath := filepath.Join(dir, "store", "connections.yaml")
	writeStoreFile(t, storePath, map[string]ConnectionConfig{
		"saved": {Name: "saved", Host: "h", Port: 5432, User: "u", DBName: "d"},
	}, "")

	root := t.TempDir()
	content := "PGHOST=h\nPGPORT=5432\nPGUSER=u\nPGDATABASE=d\n"
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	candidates := Discover(DiscoverOptions{
		Stores:  []*Store{NewStore(storePath)},
		GitRoot: root,
	})

	if len(candidates) != 1 {
		t.Fatalf("len = %d, want 1 (dedup)", len(candidates))
	}
	if candidates[0].Source != SourceGlobalStore {
		t.Errorf("source = %q, want global (first wins)", candidates[0].Source)
	}
}

func TestDiscover_Empty(t *testing.T) {
	candidates := Discover(DiscoverOptions{})
	if len(candidates) != 0 {
		t.Errorf("expected empty, got %d", len(candidates))
	}
}
