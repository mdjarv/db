//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/mdjarv/db/internal/db"
	_ "github.com/mdjarv/db/internal/db/postgres"
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

func setupPostgres(t *testing.T) (string, func()) {
	t.Helper()
	ctx := context.Background()

	ctr, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		postgres.WithInitScripts(), // no init scripts, we'll run schema manually
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

	// Apply schema
	conn, err := db.Open(ctx, "postgres", dsn)
	if err != nil {
		ctr.Terminate(ctx) //nolint:errcheck
		t.Fatalf("connect for schema: %v", err)
	}
	if _, err := conn.Exec(ctx, testSchema); err != nil {
		conn.Close(ctx)    //nolint:errcheck
		ctr.Terminate(ctx) //nolint:errcheck
		t.Fatalf("apply schema: %v", err)
	}
	conn.Close(ctx) //nolint:errcheck

	return dsn, func() { ctr.Terminate(ctx) } //nolint:errcheck
}

func TestConnect(t *testing.T) {
	dsn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	conn, err := db.Open(ctx, "postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer conn.Close(ctx) //nolint:errcheck
}

func TestConnectBadDSN(t *testing.T) {
	ctx := context.Background()
	// pgxpool connects lazily, so Open succeeds — error surfaces on first use
	conn, err := db.Open(ctx, "postgres", "postgres://bad:bad@localhost:1/nope?sslmode=disable&connect_timeout=1")
	if err != nil {
		return // some environments fail at parse/connect — that's fine
	}
	defer conn.Close(ctx) //nolint:errcheck

	_, err = conn.Query(ctx, "SELECT 1")
	if err == nil {
		t.Fatal("expected error querying bad connection")
	}
}

func TestQuery(t *testing.T) {
	dsn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	conn, err := db.Open(ctx, "postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer conn.Close(ctx) //nolint:errcheck

	result, err := conn.Query(ctx, "SELECT id, name, email, active FROM users ORDER BY id")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer result.Rows.Close()

	if len(result.Columns) != 4 {
		t.Fatalf("expected 4 columns, got %d", len(result.Columns))
	}
	if result.Columns[0].Name != "id" {
		t.Errorf("expected column 0 name 'id', got %q", result.Columns[0].Name)
	}
	if result.Columns[3].TypeName != "bool" {
		t.Errorf("expected column 3 type 'bool', got %q", result.Columns[3].TypeName)
	}

	var count int
	for result.Rows.Next() {
		vals, err := result.Rows.Values()
		if err != nil {
			t.Fatalf("values: %v", err)
		}
		if len(vals) != 4 {
			t.Fatalf("expected 4 values, got %d", len(vals))
		}
		count++
	}
	if err := result.Rows.Err(); err != nil {
		t.Fatalf("rows err: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected 3 rows, got %d", count)
	}
}

func TestQueryWithArgs(t *testing.T) {
	dsn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	conn, err := db.Open(ctx, "postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer conn.Close(ctx) //nolint:errcheck

	result, err := conn.Query(ctx, "SELECT name FROM users WHERE active = $1", true)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer result.Rows.Close()

	var count int
	for result.Rows.Next() {
		count++
	}
	if count != 2 {
		t.Fatalf("expected 2 active users, got %d", count)
	}
}

func TestExec(t *testing.T) {
	dsn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	conn, err := db.Open(ctx, "postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer conn.Close(ctx) //nolint:errcheck

	res, err := conn.Exec(ctx, "UPDATE users SET active = false WHERE name = $1", "alice")
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	if res.RowsAffected != 1 {
		t.Fatalf("expected 1 row affected, got %d", res.RowsAffected)
	}

	res, err = conn.Exec(ctx, "DELETE FROM posts WHERE user_id = (SELECT id FROM users WHERE name = $1)", "bob")
	if err != nil {
		t.Fatalf("exec delete: %v", err)
	}
	if res.RowsAffected != 1 {
		t.Fatalf("expected 1 row deleted, got %d", res.RowsAffected)
	}
}

func TestTransactionCommit(t *testing.T) {
	dsn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	conn, err := db.Open(ctx, "postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer conn.Close(ctx) //nolint:errcheck

	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	_, err = tx.Exec(ctx, "INSERT INTO users (name, email) VALUES ($1, $2)", "dave", "dave@example.com")
	if err != nil {
		t.Fatalf("tx exec: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Verify committed
	result, err := conn.Query(ctx, "SELECT name FROM users WHERE name = $1", "dave")
	if err != nil {
		t.Fatalf("query after commit: %v", err)
	}
	defer result.Rows.Close()
	if !result.Rows.Next() {
		t.Fatal("expected dave to exist after commit")
	}
}

func TestTransactionRollback(t *testing.T) {
	dsn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	conn, err := db.Open(ctx, "postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer conn.Close(ctx) //nolint:errcheck

	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	_, err = tx.Exec(ctx, "INSERT INTO users (name, email) VALUES ($1, $2)", "eve", "eve@example.com")
	if err != nil {
		t.Fatalf("tx exec: %v", err)
	}

	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("rollback: %v", err)
	}

	// Verify rolled back
	result, err := conn.Query(ctx, "SELECT name FROM users WHERE name = $1", "eve")
	if err != nil {
		t.Fatalf("query after rollback: %v", err)
	}
	defer result.Rows.Close()
	if result.Rows.Next() {
		t.Fatal("expected eve to not exist after rollback")
	}
}

func TestTransactionQuery(t *testing.T) {
	dsn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	conn, err := db.Open(ctx, "postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer conn.Close(ctx) //nolint:errcheck

	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	result, err := tx.Query(ctx, "SELECT id, name FROM users ORDER BY id")
	if err != nil {
		t.Fatalf("tx query: %v", err)
	}
	defer result.Rows.Close()

	var count int
	for result.Rows.Next() {
		vals, err := result.Rows.Values()
		if err != nil {
			t.Fatalf("values: %v", err)
		}
		if len(vals) != 2 {
			t.Fatalf("expected 2 values, got %d", len(vals))
		}
		count++
	}
	if count != 3 {
		t.Fatalf("expected 3 rows in tx, got %d", count)
	}
}

func TestRowIteratorEarlyClose(t *testing.T) {
	dsn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	conn, err := db.Open(ctx, "postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer conn.Close(ctx) //nolint:errcheck

	result, err := conn.Query(ctx, "SELECT id FROM users ORDER BY id")
	if err != nil {
		t.Fatalf("query: %v", err)
	}

	// Read one row then close early
	if !result.Rows.Next() {
		t.Fatal("expected at least one row")
	}
	result.Rows.Close()

	if err := result.Rows.Err(); err != nil {
		t.Fatalf("err after early close: %v", err)
	}
}

func TestContextCancellation(t *testing.T) {
	dsn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	conn, err := db.Open(ctx, "postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer conn.Close(ctx) //nolint:errcheck

	cancelCtx, cancel := context.WithCancel(ctx)
	cancel() // cancel immediately

	_, err = conn.Query(cancelCtx, "SELECT pg_sleep(10)")
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestColumnTypes(t *testing.T) {
	dsn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	conn, err := db.Open(ctx, "postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer conn.Close(ctx) //nolint:errcheck

	result, err := conn.Query(ctx, "SELECT id, name, email, active, created_at FROM users LIMIT 1")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer result.Rows.Close()

	expected := map[string]string{
		"id":         "int4",
		"name":       "varchar",
		"email":      "varchar",
		"active":     "bool",
		"created_at": "timestamptz",
	}

	for _, col := range result.Columns {
		want, ok := expected[col.Name]
		if !ok {
			t.Errorf("unexpected column %q", col.Name)
			continue
		}
		if col.TypeName != want {
			t.Errorf("column %q: expected type %q, got %q", col.Name, want, col.TypeName)
		}
	}
}

func TestRegistryUnknownDriver(t *testing.T) {
	_, err := db.Open(context.Background(), "nonexistent", "whatever")
	if err == nil {
		t.Fatal("expected error for unknown driver")
	}
}
