package dump

import (
	"strings"
	"testing"
	"time"
)

func TestFormatString(t *testing.T) {
	tests := []struct {
		f    Format
		want string
	}{
		{Plain, "plain"},
		{Custom, "custom"},
		{Directory, "directory"},
		{Tar, "tar"},
		{Format(99), "unknown(99)"},
	}
	for _, tt := range tests {
		if got := tt.f.String(); got != tt.want {
			t.Errorf("Format(%d).String() = %q, want %q", tt.f, got, tt.want)
		}
	}
}

func TestFormatFlag(t *testing.T) {
	tests := []struct {
		f    Format
		want string
	}{
		{Plain, "p"},
		{Custom, "c"},
		{Directory, "d"},
		{Tar, "t"},
	}
	for _, tt := range tests {
		if got := tt.f.flag(); got != tt.want {
			t.Errorf("Format(%d).flag() = %q, want %q", tt.f, got, tt.want)
		}
	}
}

func TestBuildArgs(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want []string
	}{
		{
			name: "minimal",
			cfg:  Config{DBName: "mydb"},
			want: []string{"--dbname", "mydb", "--format=p", "--verbose", "--no-password"},
		},
		{
			name: "full connection",
			cfg: Config{
				Host:   "db.example.com",
				Port:   "5433",
				User:   "admin",
				DBName: "prod",
			},
			want: []string{
				"--host", "db.example.com",
				"--port", "5433",
				"--username", "admin",
				"--dbname", "prod",
				"--format=p",
				"--verbose", "--no-password",
			},
		},
		{
			name: "custom format",
			cfg:  Config{DBName: "mydb", Format: Custom},
			want: []string{"--dbname", "mydb", "--format=c", "--verbose", "--no-password"},
		},
		{
			name: "plain format",
			cfg:  Config{DBName: "mydb", Format: Plain},
			want: []string{"--dbname", "mydb", "--format=p", "--verbose", "--no-password"},
		},
		{
			name: "directory format",
			cfg:  Config{DBName: "mydb", Format: Directory},
			want: []string{"--dbname", "mydb", "--format=d", "--verbose", "--no-password"},
		},
		{
			name: "tar format",
			cfg:  Config{DBName: "mydb", Format: Tar},
			want: []string{"--dbname", "mydb", "--format=t", "--verbose", "--no-password"},
		},
		{
			name: "schema only",
			cfg:  Config{DBName: "mydb", SchemaOnly: true},
			want: []string{"--dbname", "mydb", "--format=p", "--schema-only", "--verbose", "--no-password"},
		},
		{
			name: "single table",
			cfg:  Config{DBName: "mydb", Tables: []string{"users"}},
			want: []string{"--dbname", "mydb", "--format=p", "-t", "users", "--verbose", "--no-password"},
		},
		{
			name: "multiple tables",
			cfg:  Config{DBName: "mydb", Tables: []string{"users", "posts"}},
			want: []string{"--dbname", "mydb", "--format=p", "-t", "users", "-t", "posts", "--verbose", "--no-password"},
		},
		{
			name: "output path",
			cfg:  Config{DBName: "mydb", OutputPath: "/tmp/dump.sql"},
			want: []string{"--dbname", "mydb", "--format=p", "-f", "/tmp/dump.sql", "--verbose", "--no-password"},
		},
		{
			name: "everything",
			cfg: Config{
				Host:       "h",
				Port:       "5432",
				User:       "u",
				DBName:     "d",
				Format:     Plain,
				SchemaOnly: true,
				Tables:     []string{"t1", "t2"},
				OutputPath: "/out.sql",
			},
			want: []string{
				"--host", "h",
				"--port", "5432",
				"--username", "u",
				"--dbname", "d",
				"--format=p",
				"--schema-only",
				"-t", "t1", "-t", "t2",
				"-f", "/out.sql",
				"--verbose", "--no-password",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildArgs(tt.cfg)
			if len(got) != len(tt.want) {
				t.Fatalf("BuildArgs() len = %d, want %d\ngot:  %v\nwant: %v", len(got), len(tt.want), got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("arg[%d] = %q, want %q\ngot:  %v\nwant: %v", i, got[i], tt.want[i], got, tt.want)
					break
				}
			}
		})
	}
}

func TestBuildArgsAlwaysHasVerboseAndNoPassword(t *testing.T) {
	args := BuildArgs(Config{})
	hasVerbose := false
	hasNoPw := false
	for _, a := range args {
		if a == "--verbose" {
			hasVerbose = true
		}
		if a == "--no-password" {
			hasNoPw = true
		}
	}
	if !hasVerbose {
		t.Error("missing --verbose")
	}
	if !hasNoPw {
		t.Error("missing --no-password")
	}
}

func TestDefaultOutputPath(t *testing.T) {
	date := time.Now().Format("20060102")
	tests := []struct {
		dbname string
		format Format
		want   string
	}{
		{"mydb", Custom, "mydb_" + date + ".dump"},
		{"mydb", Plain, "mydb_" + date + ".sql"},
		{"mydb", Tar, "mydb_" + date + ".tar"},
		{"mydb", Directory, "mydb_" + date},
	}
	for _, tt := range tests {
		got := DefaultOutputPath(tt.dbname, tt.format)
		if got != tt.want {
			t.Errorf("DefaultOutputPath(%q, %v) = %q, want %q", tt.dbname, tt.format, got, tt.want)
		}
	}
}

func TestDefaultOutputPathContainsDate(t *testing.T) {
	got := DefaultOutputPath("testdb", Custom)
	date := time.Now().Format("20060102")
	if !strings.Contains(got, date) {
		t.Errorf("DefaultOutputPath should contain today's date %q, got %q", date, got)
	}
}
