//go:build integration

package cmd_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const testSchema = `
CREATE TABLE users (
	id SERIAL PRIMARY KEY,
	name VARCHAR(100) NOT NULL,
	email VARCHAR(255) UNIQUE NOT NULL,
	active BOOLEAN DEFAULT true,
	created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE posts (
	id SERIAL PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES users(id),
	title TEXT NOT NULL,
	body TEXT,
	published_at TIMESTAMPTZ
);

INSERT INTO users (name, email, active) VALUES
	('alice', 'alice@example.com', true),
	('bob', 'bob@example.com', true),
	('carol', 'carol@example.com', false);

INSERT INTO posts (user_id, title, body) VALUES
	(1, 'Hello World', 'First post'),
	(2, 'Go Tips', 'Use interfaces');
`

var testBinary string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "db-cli-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmp)

	testBinary = filepath.Join(tmp, "db")
	cmd := exec.Command("go", "build", "-o", testBinary, "..")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "build binary: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

type pgSetup struct {
	dsn     string
	cleanup func()
}

func setupPostgres(t *testing.T) pgSetup {
	t.Helper()
	ctx := context.Background()

	ctr, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		postgres.WithInitScripts(),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start container: %v", err)
	}

	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		ctr.Terminate(ctx) //nolint:errcheck
		t.Fatalf("connection string: %v", err)
	}

	// Apply schema via psql inside the container
	code, _, err := ctr.Exec(ctx, []string{
		"psql", "-U", "testuser", "-d", "testdb", "-c", testSchema,
	})
	if err != nil || code != 0 {
		ctr.Terminate(ctx) //nolint:errcheck
		t.Fatalf("apply schema: exit=%d err=%v", code, err)
	}

	return pgSetup{
		dsn:     dsn,
		cleanup: func() { ctr.Terminate(ctx) }, //nolint:errcheck
	}
}

// runCLI executes the binary with a 30s timeout.
func runCLI(t *testing.T, dsn string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fullArgs := append([]string{"--dsn", dsn}, args...)
	cmd := exec.CommandContext(ctx, testBinary, fullArgs...)
	cmd.Env = append(os.Environ(), "HOME="+t.TempDir())
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

// runCLIRaw executes the binary with custom env and no --dsn flag.
func runCLIRaw(t *testing.T, env []string, stdin string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, testBinary, args...)
	cmd.Env = env
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

// --- Task #1: CLI integration tests ---

func TestCLITables(t *testing.T) {
	pg := setupPostgres(t)
	defer pg.cleanup()

	stdout, _, err := runCLI(t, pg.dsn, "tables")
	if err != nil {
		t.Fatalf("tables failed: %v", err)
	}

	for _, name := range []string{"users", "posts"} {
		if !strings.Contains(stdout, name) {
			t.Errorf("tables output missing %q:\n%s", name, stdout)
		}
	}
}

func TestCLIQuerySelect(t *testing.T) {
	pg := setupPostgres(t)
	defer pg.cleanup()

	stdout, _, err := runCLI(t, pg.dsn, "query", "SELECT name, email FROM users ORDER BY name")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	for _, want := range []string{"alice", "bob", "carol"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("query output missing %q:\n%s", want, stdout)
		}
	}
}

func TestCLIQueryCSV(t *testing.T) {
	pg := setupPostgres(t)
	defer pg.cleanup()

	stdout, _, err := runCLI(t, pg.dsn, "query", "-F", "csv", "SELECT name, email FROM users ORDER BY name")
	if err != nil {
		t.Fatalf("query csv failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) < 4 { // header + 3 rows
		t.Fatalf("expected at least 4 lines, got %d:\n%s", len(lines), stdout)
	}

	// Header should have column names
	if !strings.Contains(lines[0], "name") || !strings.Contains(lines[0], "email") {
		t.Errorf("CSV header unexpected: %q", lines[0])
	}

	// Data rows should be comma-separated
	for _, line := range lines[1:] {
		if !strings.Contains(line, ",") {
			t.Errorf("expected CSV comma-separated line, got %q", line)
		}
	}

	if !strings.Contains(stdout, "alice") {
		t.Errorf("CSV output missing 'alice':\n%s", stdout)
	}
}

func TestCLIQueryJSON(t *testing.T) {
	pg := setupPostgres(t)
	defer pg.cleanup()

	stdout, _, err := runCLI(t, pg.dsn, "query", "-F", "json", "SELECT name FROM users WHERE name = 'alice'")
	if err != nil {
		t.Fatalf("query json failed: %v", err)
	}
	if !strings.Contains(stdout, "alice") {
		t.Errorf("JSON output missing 'alice':\n%s", stdout)
	}
}

func TestCLIQueryExec(t *testing.T) {
	pg := setupPostgres(t)
	defer pg.cleanup()

	_, stderr, err := runCLI(t, pg.dsn, "query", "UPDATE users SET active = false WHERE name = 'alice'")
	if err != nil {
		t.Fatalf("query exec failed: %v", err)
	}
	if !strings.Contains(stderr, "1 row(s) affected") {
		t.Errorf("expected '1 row(s) affected' in stderr, got %q", stderr)
	}
}

func TestCLIQueryFromStdin(t *testing.T) {
	pg := setupPostgres(t)
	defer pg.cleanup()

	env := append(os.Environ(), "HOME="+t.TempDir())
	stdout, stderr, err := runCLIRaw(t, env, "SELECT name FROM users WHERE name = 'bob'",
		"--dsn", pg.dsn, "query")
	if err != nil {
		t.Fatalf("query from stdin failed: %v\nstderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "bob") {
		t.Errorf("stdin query output missing 'bob':\n%s", stdout)
	}
}

func TestCLIQueryFromFile(t *testing.T) {
	pg := setupPostgres(t)
	defer pg.cleanup()

	sqlFile := filepath.Join(t.TempDir(), "test.sql")
	if err := os.WriteFile(sqlFile, []byte("SELECT name FROM users ORDER BY name"), 0o644); err != nil {
		t.Fatalf("write sql file: %v", err)
	}

	stdout, _, err := runCLI(t, pg.dsn, "query", "-f", sqlFile)
	if err != nil {
		t.Fatalf("query from file failed: %v", err)
	}
	if !strings.Contains(stdout, "alice") {
		t.Errorf("file query output missing 'alice':\n%s", stdout)
	}
}

func TestCLIConnectAddAndList(t *testing.T) {
	home := t.TempDir()
	env := append(os.Environ(), "HOME="+home, "XDG_CONFIG_HOME="+filepath.Join(home, ".config"))

	// Add a connection via stdin prompts
	stdout, stderr, err := runCLIRaw(t, env,
		"myconn\nlocalhost\n5432\ntestuser\n\ntestdb\ndisable\n",
		"connect", "add")
	if err != nil {
		t.Fatalf("connect add failed: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
	}
	if !strings.Contains(stdout, "saved") {
		t.Errorf("expected 'saved' in output, got %q", stdout)
	}

	// List connections
	stdout, stderr, err = runCLIRaw(t, env, "", "connect", "list")
	if err != nil {
		t.Fatalf("connect list failed: %v\nstderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "myconn") {
		t.Errorf("connect list missing 'myconn':\n%s", stdout)
	}
}

// --- Task #2: Connection failure scenarios ---

func TestCLIConnectionRefused(t *testing.T) {
	// Port 1 — connection refused
	dsn := "postgres://testuser:testpass@localhost:1/testdb?sslmode=disable&connect_timeout=2"
	_, _, err := runCLI(t, dsn, "query", "SELECT 1")
	if err == nil {
		t.Fatal("expected error for connection refused")
	}
}

func TestCLIConnectionTimeout(t *testing.T) {
	// 192.0.2.1 is TEST-NET-1 (RFC 5737), routed to nowhere — triggers timeout
	dsn := "postgres://testuser:testpass@192.0.2.1:5432/testdb?sslmode=disable&connect_timeout=2"
	_, _, err := runCLI(t, dsn, "query", "SELECT 1")
	if err == nil {
		t.Fatal("expected error for connection timeout")
	}
}

func TestCLIAuthFailure(t *testing.T) {
	pg := setupPostgres(t)
	defer pg.cleanup()

	badDSN := strings.Replace(pg.dsn, "testpass", "wrongpass", 1)
	_, _, err := runCLI(t, badDSN, "query", "SELECT 1")
	if err == nil {
		t.Fatal("expected error for auth failure")
	}
}

func TestCLISSLMismatch(t *testing.T) {
	pg := setupPostgres(t)
	defer pg.cleanup()

	// Container has no SSL — requiring it should fail
	sslDSN := strings.Replace(pg.dsn, "sslmode=disable", "sslmode=require", 1)
	_, _, err := runCLI(t, sslDSN, "query", "SELECT 1")
	if err == nil {
		t.Fatal("expected error for SSL mismatch")
	}
}
