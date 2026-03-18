package dump

import (
	"strings"
	"testing"
)

func TestFindPgDumpConfigPath(t *testing.T) {
	got, err := FindPgDump("/usr/local/bin/pg_dump")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/usr/local/bin/pg_dump" {
		t.Errorf("FindPgDump with config path = %q, want /usr/local/bin/pg_dump", got)
	}
}

func TestFindPgDumpConfigPathPrecedence(t *testing.T) {
	// Config path should be returned as-is regardless of PATH contents.
	got, err := FindPgDump("/nonexistent/pg_dump")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/nonexistent/pg_dump" {
		t.Errorf("got %q, want /nonexistent/pg_dump", got)
	}
}

func TestFindPgDumpEmptyConfigNotInPath(t *testing.T) {
	// Save and clear PATH to force LookPath failure.
	t.Setenv("PATH", "")
	_, err := FindPgDump("")
	if err == nil {
		t.Fatal("expected error when pg_dump not in PATH")
	}
	if !strings.Contains(err.Error(), "pg_dump not found in PATH") {
		t.Errorf("error = %q, want message about pg_dump not found", err.Error())
	}
	if !strings.Contains(err.Error(), "postgresql-client") {
		t.Errorf("error = %q, want install hint", err.Error())
	}
}
