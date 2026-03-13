//go:build integration

package query_test

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/mdjarv/db/internal/db"
	_ "github.com/mdjarv/db/internal/db/postgres"
	"github.com/mdjarv/db/internal/query"
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

func setupPostgres(t *testing.T) (db.Conn, func()) {
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

	conn, err := db.Open(ctx, "postgres", dsn)
	if err != nil {
		ctr.Terminate(ctx) //nolint:errcheck
		t.Fatalf("connect: %v", err)
	}
	if _, err := conn.Exec(ctx, testSchema); err != nil {
		conn.Close(ctx)    //nolint:errcheck
		ctr.Terminate(ctx) //nolint:errcheck
		t.Fatalf("apply schema: %v", err)
	}

	return conn, func() {
		conn.Close(ctx)    //nolint:errcheck
		ctr.Terminate(ctx) //nolint:errcheck
	}
}

func TestIntegrationSelectReturnsRows(t *testing.T) {
	conn, cleanup := setupPostgres(t)
	defer cleanup()

	e := query.NewExecutor(conn, query.AutoCommit)
	res, err := e.Execute(context.Background(), "SELECT id, name FROM users ORDER BY id")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !res.IsQuery {
		t.Fatal("expected IsQuery=true")
	}
	if res.Result == nil {
		t.Fatal("expected Result non-nil")
	}
	if len(res.Result.Columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(res.Result.Columns))
	}

	var count int
	for res.Result.Rows.Next() {
		count++
	}
	res.Result.Rows.Close()
	if count != 3 {
		t.Fatalf("expected 3 rows, got %d", count)
	}
	if res.Duration <= 0 {
		t.Fatal("expected Duration > 0")
	}
}

func TestIntegrationInsertReturnsRowsAffected(t *testing.T) {
	conn, cleanup := setupPostgres(t)
	defer cleanup()

	e := query.NewExecutor(conn, query.AutoCommit)
	res, err := e.Execute(context.Background(), "INSERT INTO users (name, email) VALUES ('dave', 'dave@example.com')")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if res.IsQuery {
		t.Fatal("expected IsQuery=false")
	}
	if res.ExecResult == nil {
		t.Fatal("expected ExecResult non-nil")
	}
	if res.ExecResult.RowsAffected != 1 {
		t.Fatalf("expected 1 row affected, got %d", res.ExecResult.RowsAffected)
	}
}

func TestIntegrationTransactionCommit(t *testing.T) {
	conn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	e := query.NewExecutor(conn, query.Explicit)

	if err := e.Begin(ctx); err != nil {
		t.Fatalf("begin: %v", err)
	}
	if _, err := e.Execute(ctx, "INSERT INTO users (name, email) VALUES ('eve', 'eve@example.com')"); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if err := e.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Verify persisted
	res, err := e.Execute(ctx, "SELECT name FROM users WHERE name = 'eve'")
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	defer res.Result.Rows.Close()
	if !res.Result.Rows.Next() {
		t.Fatal("expected eve to exist after commit")
	}
}

func TestIntegrationTransactionRollback(t *testing.T) {
	conn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	e := query.NewExecutor(conn, query.Explicit)

	if err := e.Begin(ctx); err != nil {
		t.Fatalf("begin: %v", err)
	}
	if _, err := e.Execute(ctx, "INSERT INTO users (name, email) VALUES ('frank', 'frank@example.com')"); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if err := e.Rollback(ctx); err != nil {
		t.Fatalf("rollback: %v", err)
	}

	// Verify discarded
	res, err := e.Execute(ctx, "SELECT name FROM users WHERE name = 'frank'")
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	defer res.Result.Rows.Close()
	if res.Result.Rows.Next() {
		t.Fatal("expected frank to not exist after rollback")
	}
}

func TestIntegrationContextCancellation(t *testing.T) {
	conn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	e := query.NewExecutor(conn, query.AutoCommit)
	_, err := e.Execute(ctx, "SELECT pg_sleep(10)")
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}
