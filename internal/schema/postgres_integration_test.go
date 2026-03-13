//go:build integration

package schema_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	pgmod "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/mdjarv/db/internal/db"
	_ "github.com/mdjarv/db/internal/db/postgres"
	"github.com/mdjarv/db/internal/schema"
)

const testSchema = `
CREATE TABLE users (
	id SERIAL PRIMARY KEY,
	name VARCHAR(100) NOT NULL,
	email VARCHAR(255) UNIQUE NOT NULL,
	active BOOLEAN DEFAULT true,
	created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE posts (
	id SERIAL PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	title VARCHAR(200) NOT NULL,
	body TEXT,
	published BOOLEAN DEFAULT false,
	created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_posts_user_id ON posts(user_id);
CREATE INDEX idx_posts_published ON posts(published) WHERE published = true;

CREATE VIEW active_users AS SELECT * FROM users WHERE active = true;
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
		ctr.Terminate(ctx)
		t.Fatalf("connection string: %v", err)
	}

	conn, err := db.Open(ctx, "postgres", dsn)
	if err != nil {
		ctr.Terminate(ctx)
		t.Fatalf("connect: %v", err)
	}
	if _, err := conn.Exec(ctx, testSchema); err != nil {
		conn.Close(ctx)
		ctr.Terminate(ctx)
		t.Fatalf("apply schema: %v", err)
	}

	// ANALYZE so pg_stat_user_tables has row estimates
	if _, err := conn.Exec(ctx, "ANALYZE"); err != nil {
		conn.Close(ctx)
		ctr.Terminate(ctx)
		t.Fatalf("analyze: %v", err)
	}

	return conn, func() {
		conn.Close(ctx)
		ctr.Terminate(ctx)
	}
}

func TestTables(t *testing.T) {
	conn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	insp := schema.NewPostgresInspector(conn)

	tables, err := insp.Tables(ctx, "public")
	if err != nil {
		t.Fatalf("Tables: %v", err)
	}

	byName := make(map[string]schema.Table)
	for _, tbl := range tables {
		byName[tbl.Name] = tbl
	}

	for _, name := range []string{"users", "posts", "active_users"} {
		if _, ok := byName[name]; !ok {
			t.Errorf("missing table %q", name)
		}
	}

	if tbl := byName["users"]; tbl.Type != "table" {
		t.Errorf("users type = %q, want table", tbl.Type)
	}
	if tbl := byName["active_users"]; tbl.Type != "view" {
		t.Errorf("active_users type = %q, want view", tbl.Type)
	}
}

func TestColumns(t *testing.T) {
	conn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	insp := schema.NewPostgresInspector(conn)

	cols, err := insp.Columns(ctx, "public", "users")
	if err != nil {
		t.Fatalf("Columns: %v", err)
	}

	if len(cols) != 5 {
		t.Fatalf("expected 5 columns, got %d", len(cols))
	}

	// id column
	id := cols[0]
	if id.Name != "id" {
		t.Errorf("col[0] name = %q, want id", id.Name)
	}
	if !id.IsPK {
		t.Error("id should be PK")
	}
	if id.Nullable {
		t.Error("id should not be nullable")
	}

	// email column
	var email schema.ColumnInfo
	for _, c := range cols {
		if c.Name == "email" {
			email = c
			break
		}
	}
	if email.Nullable {
		t.Error("email should not be nullable")
	}

	// active column
	var active schema.ColumnInfo
	for _, c := range cols {
		if c.Name == "active" {
			active = c
			break
		}
	}
	if active.Default == "" {
		t.Error("active should have a default")
	}
}

func TestIndexes(t *testing.T) {
	conn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	insp := schema.NewPostgresInspector(conn)

	indexes, err := insp.Indexes(ctx, "public", "posts")
	if err != nil {
		t.Fatalf("Indexes: %v", err)
	}

	byName := make(map[string]schema.Index)
	for _, idx := range indexes {
		byName[idx.Name] = idx
	}

	if _, ok := byName["idx_posts_user_id"]; !ok {
		t.Error("missing idx_posts_user_id")
	}

	partial, ok := byName["idx_posts_published"]
	if !ok {
		t.Fatal("missing idx_posts_published")
	}
	if !strings.Contains(partial.Definition, "WHERE") {
		t.Error("partial index should contain WHERE clause")
	}
	if partial.Type != "btree" {
		t.Errorf("idx_posts_published type = %q, want btree", partial.Type)
	}
}

func TestConstraints(t *testing.T) {
	conn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	insp := schema.NewPostgresInspector(conn)

	constraints, err := insp.Constraints(ctx, "public", "users")
	if err != nil {
		t.Fatalf("Constraints: %v", err)
	}

	types := make(map[string]bool)
	for _, c := range constraints {
		types[c.Type] = true
	}

	if !types["PRIMARY KEY"] {
		t.Error("missing PRIMARY KEY constraint")
	}
	if !types["UNIQUE"] {
		t.Error("missing UNIQUE constraint")
	}

	// posts should have FK constraint
	postConstraints, err := insp.Constraints(ctx, "public", "posts")
	if err != nil {
		t.Fatalf("Constraints(posts): %v", err)
	}
	hasFk := false
	for _, c := range postConstraints {
		if c.Type == "FOREIGN KEY" {
			hasFk = true
			break
		}
	}
	if !hasFk {
		t.Error("posts should have FOREIGN KEY constraint")
	}
}

func TestForeignKeys(t *testing.T) {
	conn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	insp := schema.NewPostgresInspector(conn)

	fks, err := insp.ForeignKeys(ctx, "public", "posts")
	if err != nil {
		t.Fatalf("ForeignKeys: %v", err)
	}

	if len(fks) != 1 {
		t.Fatalf("expected 1 FK, got %d", len(fks))
	}

	fk := fks[0]
	if len(fk.Columns) != 1 || fk.Columns[0] != "user_id" {
		t.Errorf("FK columns = %v, want [user_id]", fk.Columns)
	}
	if fk.ReferencedTable != "users" {
		t.Errorf("FK ref table = %q, want users", fk.ReferencedTable)
	}
	if len(fk.ReferencedColumns) != 1 || fk.ReferencedColumns[0] != "id" {
		t.Errorf("FK ref columns = %v, want [id]", fk.ReferencedColumns)
	}
	if fk.OnDelete != "CASCADE" {
		t.Errorf("FK on delete = %q, want CASCADE", fk.OnDelete)
	}
}
