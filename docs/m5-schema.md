# M5: Schema Inspection

## Goal

Introspect PostgreSQL schema: tables, columns, indexes, constraints, foreign keys. CLI commands for quick schema lookup.

## Tasks

### Types (`internal/schema/types.go`)

- [ ] `Table`: name, schema, type (table/view/materialized view), row estimate, size
- [ ] `ColumnInfo`: name, type, nullable, default, is_pk, position
- [ ] `Index`: name, columns, unique, type (btree/hash/gin/gist), size
- [ ] `Constraint`: name, type (PK/FK/UNIQUE/CHECK/EXCLUDE), columns, definition
- [ ] `ForeignKey`: name, columns, referenced table, referenced columns, on_delete, on_update

### Inspector Interface (`internal/schema/inspector.go`)

- [ ] `Inspector` interface (see architecture.md)
- [ ] Accept schema filter (default `public`)

### PostgreSQL Implementation (`internal/schema/postgres.go`)

- [ ] `Tables()` — query `information_schema.tables` + `pg_stat_user_tables` for row estimates
- [ ] `Columns()` — query `information_schema.columns` + `pg_constraint` for PK detection
- [ ] `Indexes()` — query `pg_indexes` + `pg_stat_user_indexes` for size
- [ ] `Constraints()` — query `pg_constraint` with type classification
- [ ] `ForeignKeys()` — query `pg_constraint` + resolve referenced table/columns
- [ ] Schema listing: `Schemas(ctx) ([]string, error)`
- [ ] View detection: distinguish tables, views, materialized views

### CLI Commands

- [ ] `db tables` — list tables with row count and size
- [ ] `db tables --schema <name>` — filter by schema
- [ ] `db describe <table>` — show columns, indexes, constraints, FKs
- [ ] `db describe <table> --format json` — machine-readable output
- [ ] Pretty table output with alignment

### Tests

- [ ] Integration: create schema with tables, views, indexes, FKs, constraints
- [ ] Integration: verify Tables() returns correct list
- [ ] Integration: verify Columns() returns correct types, nullability, defaults
- [ ] Integration: verify Indexes() includes all index types
- [ ] Integration: verify ForeignKeys() resolves references correctly
- [ ] Integration: CLI `db tables` and `db describe` end-to-end

## Test Schema

```sql
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

CREATE VIEW active_users AS
    SELECT * FROM users WHERE active = true;
```

## Acceptance Criteria

- `db tables` lists all tables with row count estimates
- `db describe users` shows columns, indexes, constraints, FKs
- Views and materialized views appear in table list with type indicator
- Partial indexes show their WHERE clause
- FK relationships display both sides (source and referenced)
- All queries work on `public` schema by default, `--schema` overrides

## Dependencies

- M1 (database layer — needs db.Conn for raw queries)

## Can Be Parallelized With

- M3, M4, M6 — all Phase 2 milestones
