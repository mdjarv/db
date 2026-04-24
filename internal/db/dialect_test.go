package db

import "testing"

func TestPostgresDialect(t *testing.T) {
	d := PostgresDialect()
	if got := d.Placeholder(1); got != "$1" {
		t.Errorf("Placeholder(1) = %q, want $1", got)
	}
	if got := d.Placeholder(42); got != "$42" {
		t.Errorf("Placeholder(42) = %q, want $42", got)
	}
	if got := d.QuoteIdent(`weird"name`); got != `"weird""name"` {
		t.Errorf("QuoteIdent: %q", got)
	}
	if got := d.QualifyTable("public", "users"); got != `"public"."users"` {
		t.Errorf("QualifyTable = %q", got)
	}
	if got := d.QualifyTable("", "users"); got != `"users"` {
		t.Errorf("QualifyTable empty schema = %q, want just table", got)
	}
}

func TestSQLiteDialect(t *testing.T) {
	d := SQLiteDialect()
	if got := d.Placeholder(1); got != "?" {
		t.Errorf("Placeholder = %q, want ?", got)
	}
	if got := d.Placeholder(99); got != "?" {
		t.Errorf("Placeholder always ?: got %q", got)
	}
	if got := d.QualifyTable("main", "users"); got != `"users"` {
		t.Errorf("SQLite QualifyTable should drop schema: got %q", got)
	}
}

func TestOpenDSN_UnknownScheme(t *testing.T) {
	_, err := OpenDSN(nil, "mysql://localhost/db") //nolint:staticcheck
	if err == nil {
		t.Fatal("expected unknown-scheme error")
	}
}

func TestOpenDSN_ParseError(t *testing.T) {
	_, err := OpenDSN(nil, "://broken") //nolint:staticcheck
	if err == nil {
		t.Fatal("expected parse error")
	}
}
