//go:build integration

package editor_test

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	pgmod "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/mdjarv/db/internal/db"
	_ "github.com/mdjarv/db/internal/db/postgres"
	"github.com/mdjarv/db/internal/editor"
)

const testSchema = `
CREATE TABLE users (
	id SERIAL PRIMARY KEY,
	name VARCHAR(100) NOT NULL,
	email VARCHAR(255) UNIQUE NOT NULL,
	active BOOLEAN DEFAULT true
);

INSERT INTO users (name, email, active) VALUES
	('alice', 'alice@example.com', true),
	('bob', 'bob@example.com', true),
	('carol', 'carol@example.com', false);

CREATE TABLE items (
	id SERIAL PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES users(id),
	label TEXT NOT NULL,
	qty INTEGER DEFAULT 0
);

INSERT INTO items (user_id, label, qty) VALUES
	(1, 'widget', 10),
	(2, 'gadget', 5);

CREATE TABLE composite_pk (
	a INTEGER NOT NULL,
	b INTEGER NOT NULL,
	val TEXT,
	PRIMARY KEY (a, b)
);

INSERT INTO composite_pk (a, b, val) VALUES
	(1, 1, 'one-one'),
	(1, 2, 'one-two'),
	(2, 1, 'two-one');
`

func setupPostgres(t *testing.T) (db.Conn, func()) {
	t.Helper()
	ctx := context.Background()

	ctr, err := pgmod.Run(ctx,
		"postgres:16-alpine",
		pgmod.WithDatabase("testdb"),
		pgmod.WithUsername("testuser"),
		pgmod.WithPassword("testpass"),
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

// queryScalar runs a single-row single-column query and returns the value.
func queryScalar(t *testing.T, conn db.Conn, sql string, args ...any) any {
	t.Helper()
	ctx := context.Background()
	result, err := conn.Query(ctx, sql, args...)
	if err != nil {
		t.Fatalf("queryScalar: %v", err)
	}
	defer result.Rows.Close()
	if !result.Rows.Next() {
		t.Fatalf("queryScalar: no rows for %q", sql)
	}
	vals, err := result.Rows.Values()
	if err != nil {
		t.Fatalf("queryScalar values: %v", err)
	}
	return vals[0]
}

// queryCount returns COUNT(*) for a given query.
func queryCount(t *testing.T, conn db.Conn, sql string, args ...any) int64 {
	t.Helper()
	v := queryScalar(t, conn, sql, args...)
	return v.(int64)
}

// --- Autocommit tests ---

func TestIntegrationApplyUpdate_Autocommit(t *testing.T) {
	conn, cleanup := setupPostgres(t)
	defer cleanup()

	buf := editor.NewChangeBuffer()
	buf.Add(editor.Change{
		Kind: editor.ChangeUpdate, Table: "users", Schema: "public",
		PK:     editor.PKValue{Columns: []string{"id"}, Values: []any{int32(1)}},
		Column: "name", NewValue: "alice_updated",
	})

	result := editor.Apply(context.Background(), conn, buf.Changes(), true)
	if result.Err != nil {
		t.Fatalf("apply: %v", result.Err)
	}
	if result.Applied != 1 {
		t.Fatalf("expected 1 applied, got %d", result.Applied)
	}

	got := queryScalar(t, conn, "SELECT name FROM users WHERE id = 1")
	if got != "alice_updated" {
		t.Fatalf("expected alice_updated, got %v", got)
	}
}

func TestIntegrationApplyInsert_Autocommit(t *testing.T) {
	conn, cleanup := setupPostgres(t)
	defer cleanup()

	buf := editor.NewChangeBuffer()
	buf.Add(editor.Change{
		Kind: editor.ChangeInsert, Table: "users", Schema: "public",
		Row: map[string]any{"name": "dave", "email": "dave@example.com", "active": true},
	})

	result := editor.Apply(context.Background(), conn, buf.Changes(), true)
	if result.Err != nil {
		t.Fatalf("apply: %v", result.Err)
	}
	if result.Applied != 1 {
		t.Fatalf("expected 1 applied, got %d", result.Applied)
	}

	cnt := queryCount(t, conn, "SELECT count(*) FROM users WHERE name = 'dave'")
	if cnt != 1 {
		t.Fatalf("expected 1 dave, got %d", cnt)
	}
}

func TestIntegrationApplyDelete_Autocommit(t *testing.T) {
	conn, cleanup := setupPostgres(t)
	defer cleanup()

	buf := editor.NewChangeBuffer()
	buf.Add(editor.Change{
		Kind: editor.ChangeDelete, Table: "users", Schema: "public",
		PK: editor.PKValue{Columns: []string{"id"}, Values: []any{int32(3)}},
	})

	result := editor.Apply(context.Background(), conn, buf.Changes(), true)
	if result.Err != nil {
		t.Fatalf("apply: %v", result.Err)
	}

	cnt := queryCount(t, conn, "SELECT count(*) FROM users WHERE id = 3")
	if cnt != 0 {
		t.Fatalf("expected carol deleted, got count=%d", cnt)
	}
}

func TestIntegrationApplyMixed_Autocommit(t *testing.T) {
	conn, cleanup := setupPostgres(t)
	defer cleanup()

	buf := editor.NewChangeBuffer()
	// UPDATE alice
	buf.Add(editor.Change{
		Kind: editor.ChangeUpdate, Table: "users", Schema: "public",
		PK:     editor.PKValue{Columns: []string{"id"}, Values: []any{int32(1)}},
		Column: "active", NewValue: false,
	})
	// INSERT dave
	buf.Add(editor.Change{
		Kind: editor.ChangeInsert, Table: "users", Schema: "public",
		Row: map[string]any{"name": "dave", "email": "dave@example.com"},
	})
	// DELETE carol
	buf.Add(editor.Change{
		Kind: editor.ChangeDelete, Table: "users", Schema: "public",
		PK: editor.PKValue{Columns: []string{"id"}, Values: []any{int32(3)}},
	})

	result := editor.Apply(context.Background(), conn, buf.Changes(), true)
	if result.Err != nil {
		t.Fatalf("apply: %v", result.Err)
	}
	if result.Applied != 3 {
		t.Fatalf("expected 3 applied, got %d", result.Applied)
	}

	// alice inactive
	active := queryScalar(t, conn, "SELECT active FROM users WHERE id = 1")
	if active != false {
		t.Fatalf("expected alice inactive, got %v", active)
	}
	// dave exists
	cnt := queryCount(t, conn, "SELECT count(*) FROM users WHERE name = 'dave'")
	if cnt != 1 {
		t.Fatalf("expected dave, got %d", cnt)
	}
	// carol gone
	cnt = queryCount(t, conn, "SELECT count(*) FROM users WHERE id = 3")
	if cnt != 0 {
		t.Fatalf("expected carol gone, got %d", cnt)
	}
}

func TestIntegrationApplyCompositePK_Autocommit(t *testing.T) {
	conn, cleanup := setupPostgres(t)
	defer cleanup()

	buf := editor.NewChangeBuffer()
	buf.Add(editor.Change{
		Kind: editor.ChangeUpdate, Table: "composite_pk", Schema: "public",
		PK:     editor.PKValue{Columns: []string{"a", "b"}, Values: []any{int32(1), int32(2)}},
		Column: "val", NewValue: "updated",
	})

	result := editor.Apply(context.Background(), conn, buf.Changes(), true)
	if result.Err != nil {
		t.Fatalf("apply: %v", result.Err)
	}

	got := queryScalar(t, conn, "SELECT val FROM composite_pk WHERE a = 1 AND b = 2")
	if got != "updated" {
		t.Fatalf("expected updated, got %v", got)
	}
	// other rows untouched
	got = queryScalar(t, conn, "SELECT val FROM composite_pk WHERE a = 1 AND b = 1")
	if got != "one-one" {
		t.Fatalf("expected one-one untouched, got %v", got)
	}
}

// --- Explicit transaction commit tests ---

func TestIntegrationApplyExplicit_Commit(t *testing.T) {
	conn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	buf := editor.NewChangeBuffer()
	buf.Add(editor.Change{
		Kind: editor.ChangeUpdate, Table: "users", Schema: "public",
		PK:     editor.PKValue{Columns: []string{"id"}, Values: []any{int32(1)}},
		Column: "name", NewValue: "alice_tx",
	})
	buf.Add(editor.Change{
		Kind: editor.ChangeInsert, Table: "users", Schema: "public",
		Row: map[string]any{"name": "eve", "email": "eve@example.com"},
	})
	// Delete carol (id=3, no FK references from items)
	buf.Add(editor.Change{
		Kind: editor.ChangeDelete, Table: "users", Schema: "public",
		PK: editor.PKValue{Columns: []string{"id"}, Values: []any{int32(3)}},
	})

	result := editor.Apply(ctx, conn, buf.Changes(), false)
	if result.Err != nil {
		t.Fatalf("apply: %v", result.Err)
	}
	if result.Applied != 3 {
		t.Fatalf("expected 3 applied, got %d", result.Applied)
	}
	if result.Tx == nil {
		t.Fatal("expected non-nil Tx in explicit mode")
	}

	// Commit
	if err := result.Tx.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Verify all changes persisted
	name := queryScalar(t, conn, "SELECT name FROM users WHERE id = 1")
	if name != "alice_tx" {
		t.Fatalf("expected alice_tx, got %v", name)
	}
	cnt := queryCount(t, conn, "SELECT count(*) FROM users WHERE name = 'eve'")
	if cnt != 1 {
		t.Fatalf("expected eve, got %d", cnt)
	}
	cnt = queryCount(t, conn, "SELECT count(*) FROM users WHERE id = 3")
	if cnt != 0 {
		t.Fatalf("expected carol deleted, got %d", cnt)
	}
}

// --- Explicit transaction rollback tests ---

func TestIntegrationApplyExplicit_Rollback(t *testing.T) {
	conn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	buf := editor.NewChangeBuffer()
	buf.Add(editor.Change{
		Kind: editor.ChangeUpdate, Table: "users", Schema: "public",
		PK:     editor.PKValue{Columns: []string{"id"}, Values: []any{int32(1)}},
		Column: "name", NewValue: "alice_gone",
	})
	buf.Add(editor.Change{
		Kind: editor.ChangeInsert, Table: "users", Schema: "public",
		Row: map[string]any{"name": "phantom", "email": "phantom@example.com"},
	})
	buf.Add(editor.Change{
		Kind: editor.ChangeDelete, Table: "users", Schema: "public",
		PK: editor.PKValue{Columns: []string{"id"}, Values: []any{int32(3)}},
	})

	result := editor.Apply(ctx, conn, buf.Changes(), false)
	if result.Err != nil {
		t.Fatalf("apply: %v", result.Err)
	}
	if result.Tx == nil {
		t.Fatal("expected non-nil Tx")
	}

	// Rollback
	if err := result.Tx.Rollback(ctx); err != nil {
		t.Fatalf("rollback: %v", err)
	}

	// Verify everything reverted
	name := queryScalar(t, conn, "SELECT name FROM users WHERE id = 1")
	if name != "alice" {
		t.Fatalf("expected alice unchanged, got %v", name)
	}
	cnt := queryCount(t, conn, "SELECT count(*) FROM users WHERE name = 'phantom'")
	if cnt != 0 {
		t.Fatalf("expected phantom absent, got %d", cnt)
	}
	cnt = queryCount(t, conn, "SELECT count(*) FROM users WHERE id = 3")
	if cnt != 1 {
		t.Fatalf("expected carol still present, got %d", cnt)
	}
}

func TestIntegrationApplyExplicit_ErrorAutoRollback(t *testing.T) {
	conn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	buf := editor.NewChangeBuffer()
	// Valid update
	buf.Add(editor.Change{
		Kind: editor.ChangeUpdate, Table: "users", Schema: "public",
		PK:     editor.PKValue{Columns: []string{"id"}, Values: []any{int32(1)}},
		Column: "name", NewValue: "alice_temp",
	})
	// Invalid: duplicate email violates UNIQUE constraint
	buf.Add(editor.Change{
		Kind: editor.ChangeInsert, Table: "users", Schema: "public",
		Row: map[string]any{"name": "dup", "email": "alice@example.com"},
	})

	result := editor.Apply(ctx, conn, buf.Changes(), false)
	if result.Err == nil {
		t.Fatal("expected constraint violation error")
	}
	if result.Applied != 1 {
		t.Fatalf("expected 1 applied before error, got %d", result.Applied)
	}
	if result.Tx != nil {
		t.Error("expected nil Tx after auto-rollback on error")
	}

	// First change should also be rolled back
	name := queryScalar(t, conn, "SELECT name FROM users WHERE id = 1")
	if name != "alice" {
		t.Fatalf("expected alice unchanged after rollback, got %v", name)
	}
}

func TestIntegrationApplyDeleteCompositePK_Explicit(t *testing.T) {
	conn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	buf := editor.NewChangeBuffer()
	buf.Add(editor.Change{
		Kind: editor.ChangeDelete, Table: "composite_pk", Schema: "public",
		PK: editor.PKValue{Columns: []string{"a", "b"}, Values: []any{int32(2), int32(1)}},
	})

	result := editor.Apply(ctx, conn, buf.Changes(), false)
	if result.Err != nil {
		t.Fatalf("apply: %v", result.Err)
	}
	if err := result.Tx.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	cnt := queryCount(t, conn, "SELECT count(*) FROM composite_pk WHERE a = 2 AND b = 1")
	if cnt != 0 {
		t.Fatalf("expected row deleted, got %d", cnt)
	}
	// other rows still there
	cnt = queryCount(t, conn, "SELECT count(*) FROM composite_pk")
	if cnt != 2 {
		t.Fatalf("expected 2 remaining rows, got %d", cnt)
	}
}

func TestIntegrationApplyEmpty(t *testing.T) {
	conn, cleanup := setupPostgres(t)
	defer cleanup()

	result := editor.Apply(context.Background(), conn, nil, false)
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.Applied != 0 {
		t.Fatalf("expected 0 applied, got %d", result.Applied)
	}
}

func TestIntegrationChangeBufferCollapse(t *testing.T) {
	conn, cleanup := setupPostgres(t)
	defer cleanup()

	buf := editor.NewChangeBuffer()
	pk := editor.PKValue{Columns: []string{"id"}, Values: []any{int32(1)}}
	// Multiple updates to same cell collapse to final value
	buf.Add(editor.Change{
		Kind: editor.ChangeUpdate, Table: "users", Schema: "public",
		PK: pk, Column: "name", OldValue: "alice", NewValue: "alice2",
	})
	buf.Add(editor.Change{
		Kind: editor.ChangeUpdate, Table: "users", Schema: "public",
		PK: pk, Column: "name", OldValue: "alice", NewValue: "alice_final",
	})

	if buf.Len() != 1 {
		t.Fatalf("expected 1 collapsed change, got %d", buf.Len())
	}

	result := editor.Apply(context.Background(), conn, buf.Changes(), true)
	if result.Err != nil {
		t.Fatalf("apply: %v", result.Err)
	}

	got := queryScalar(t, conn, "SELECT name FROM users WHERE id = 1")
	if got != "alice_final" {
		t.Fatalf("expected alice_final, got %v", got)
	}
}
