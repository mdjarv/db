package conn

import "testing"

func TestDSN(t *testing.T) {
	tests := []struct {
		name string
		cfg  ConnectionConfig
		want string
	}{
		{
			name: "basic",
			cfg:  ConnectionConfig{Host: "localhost", Port: 5432, User: "app", DBName: "mydb"},
			want: "postgres://app@localhost:5432/mydb",
		},
		{
			name: "with password",
			cfg:  ConnectionConfig{Host: "db.example.com", Port: 5432, User: "admin", Password: "secret", DBName: "prod"},
			want: "postgres://admin:secret@db.example.com:5432/prod",
		},
		{
			name: "with sslmode",
			cfg:  ConnectionConfig{Host: "localhost", Port: 5432, User: "u", DBName: "d", SSLMode: "require"},
			want: "postgres://u@localhost:5432/d?sslmode=require",
		},
		{
			name: "no user",
			cfg:  ConnectionConfig{Host: "localhost", Port: 5432, DBName: "test"},
			want: "postgres://localhost:5432/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.DSN()
			if got != tt.want {
				t.Errorf("DSN() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseDSN(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		want    ConnectionConfig
		wantErr bool
	}{
		{
			name: "full url",
			dsn:  "postgres://admin:secret@db.example.com:5433/prod?sslmode=require",
			want: ConnectionConfig{Host: "db.example.com", Port: 5433, User: "admin", Password: "secret", DBName: "prod", SSLMode: "require"},
		},
		{
			name: "minimal",
			dsn:  "postgres://localhost:5432/test",
			want: ConnectionConfig{Host: "localhost", Port: 5432, DBName: "test"},
		},
		{
			name: "postgresql scheme",
			dsn:  "postgresql://user@host:5432/db",
			want: ConnectionConfig{Host: "host", Port: 5432, User: "user", DBName: "db"},
		},
		{
			name: "default port",
			dsn:  "postgres://localhost/test",
			want: ConnectionConfig{Host: "localhost", Port: 5432, DBName: "test"},
		},
		{
			name:    "bad scheme",
			dsn:     "mysql://localhost/test",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDSN(tt.dsn)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Host != tt.want.Host || got.Port != tt.want.Port || got.User != tt.want.User ||
				got.Password != tt.want.Password || got.DBName != tt.want.DBName || got.SSLMode != tt.want.SSLMode {
				t.Errorf("ParseDSN() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestDSNRoundTrip(t *testing.T) {
	orig := ConnectionConfig{
		Host:     "db.example.com",
		Port:     5433,
		User:     "admin",
		Password: "s3cret",
		DBName:   "production",
		SSLMode:  "require",
	}

	parsed, err := ParseDSN(orig.DSN())
	if err != nil {
		t.Fatalf("round-trip parse: %v", err)
	}
	if parsed.Host != orig.Host || parsed.Port != orig.Port || parsed.User != orig.User ||
		parsed.Password != orig.Password || parsed.DBName != orig.DBName || parsed.SSLMode != orig.SSLMode {
		t.Errorf("round-trip mismatch: got %+v, want %+v", parsed, orig)
	}
}
