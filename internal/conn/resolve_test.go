package conn

import (
	"path/filepath"
	"testing"
)

func setupStoreWithDefault(t *testing.T) (*Store, *CredentialStore) {
	t.Helper()
	s := NewStore(filepath.Join(t.TempDir(), "connections.yaml"))
	cs := NewCredentialStore(NewMemoryKeyring())

	cfg := ConnectionConfig{Name: "prod", Host: "db.prod", Port: 5432, User: "admin", DBName: "proddb"}
	if err := s.Add(cfg); err != nil {
		t.Fatal(err)
	}
	if err := s.SetDefault("prod"); err != nil {
		t.Fatal(err)
	}
	if err := cs.SetPassword("prod", "prodpass"); err != nil {
		t.Fatal(err)
	}
	return s, cs
}

func TestResolve_DSN(t *testing.T) {
	cfg, err := Resolve(ResolveOptions{DSN: "postgres://u:p@h:5432/d"}, nil, nil)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if cfg.Host != "h" || cfg.User != "u" || cfg.Password != "p" || cfg.DBName != "d" {
		t.Errorf("Resolve = %+v", cfg)
	}
}

func TestResolve_Flags(t *testing.T) {
	cfg, err := Resolve(ResolveOptions{Host: "myhost", Port: 5433, User: "me", DBName: "mydb"}, nil, nil)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if cfg.Host != "myhost" || cfg.Port != 5433 || cfg.User != "me" || cfg.DBName != "mydb" {
		t.Errorf("Resolve = %+v", cfg)
	}
}

func TestResolve_FlagsDefaultPort(t *testing.T) {
	cfg, err := Resolve(ResolveOptions{Host: "h", User: "u"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 5432 {
		t.Errorf("port = %d, want 5432", cfg.Port)
	}
}

func TestResolve_NamedConnection(t *testing.T) {
	s, cs := setupStoreWithDefault(t)

	cfg, err := Resolve(ResolveOptions{ConnName: "prod"}, s, cs)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if cfg.Host != "db.prod" || cfg.Password != "prodpass" {
		t.Errorf("Resolve = %+v", cfg)
	}
}

func TestResolve_EnvDSN(t *testing.T) {
	t.Setenv("DB_DSN", "postgres://envuser@envhost:5432/envdb")

	cfg, err := Resolve(ResolveOptions{}, nil, nil)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if cfg.Host != "envhost" || cfg.User != "envuser" || cfg.DBName != "envdb" {
		t.Errorf("Resolve = %+v", cfg)
	}
}

func TestResolve_EnvFields(t *testing.T) {
	t.Setenv("DB_DSN", "")
	t.Setenv("DB_HOST", "envhost")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_USER", "envuser")
	t.Setenv("DB_PASSWORD", "envpass")
	t.Setenv("DB_NAME", "envdb")

	cfg, err := Resolve(ResolveOptions{}, nil, nil)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if cfg.Host != "envhost" || cfg.Port != 5433 || cfg.User != "envuser" ||
		cfg.Password != "envpass" || cfg.DBName != "envdb" {
		t.Errorf("Resolve = %+v", cfg)
	}
}

func TestResolve_StoreDefault(t *testing.T) {
	t.Setenv("DB_DSN", "")
	t.Setenv("DB_HOST", "")

	s, cs := setupStoreWithDefault(t)

	cfg, err := Resolve(ResolveOptions{}, s, cs)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if cfg.Host != "db.prod" || cfg.Password != "prodpass" {
		t.Errorf("Resolve = %+v", cfg)
	}
}

func TestResolve_Nothing(t *testing.T) {
	t.Setenv("DB_DSN", "")
	t.Setenv("DB_HOST", "")

	_, err := Resolve(ResolveOptions{}, nil, nil)
	if err == nil {
		t.Fatal("expected error when nothing configured")
	}
}

func TestResolve_DSNOverridesAll(t *testing.T) {
	t.Setenv("DB_HOST", "envhost")

	cfg, err := Resolve(ResolveOptions{
		DSN:  "postgres://dsnuser@dsnhost:5432/dsndb",
		Host: "flaghost",
	}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "dsnhost" {
		t.Errorf("DSN should take priority, got host = %q", cfg.Host)
	}
}

func TestResolve_FlagsOverrideEnv(t *testing.T) {
	t.Setenv("DB_HOST", "envhost")
	t.Setenv("DB_DSN", "")

	cfg, err := Resolve(ResolveOptions{Host: "flaghost", User: "flaguser"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "flaghost" {
		t.Errorf("flags should take priority, got host = %q", cfg.Host)
	}
}
