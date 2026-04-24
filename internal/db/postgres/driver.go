// Package postgres implements the db.Driver interface using pgx/v5.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mdjarv/db/internal/db"
	"github.com/mdjarv/db/internal/schema"
)

func init() {
	db.Register("postgres", &Driver{})
}

// Driver connects to PostgreSQL via pgxpool.
type Driver struct{}

// Connect establishes a pooled connection to PostgreSQL.
func (d *Driver) Connect(ctx context.Context, dsn string) (db.Conn, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: parse config: %w", err)
	}
	cfg.MaxConns = 5

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("postgres: connect: %w", err)
	}

	tm := LoadTypeMap(ctx, pool)
	return &Conn{pool: pool, tm: tm}, nil
}

// Conn is the postgres db.Conn implementation. Exported so that callers can
// type-assert to access postgres-specific extensions (e.g. TypeMap).
type Conn struct {
	pool *pgxpool.Pool
	tm   *TypeMap
}

// Query runs a SQL query and returns a streaming result.
func (c *Conn) Query(ctx context.Context, sql string, args ...any) (db.Result, error) {
	rows, err := c.pool.Query(ctx, sql, args...)
	if err != nil {
		return db.Result{}, fmt.Errorf("postgres: query: %w", err)
	}
	return buildResult(rows, c.tm), nil
}

// Exec runs a SQL statement and returns the affected-row count.
func (c *Conn) Exec(ctx context.Context, sql string, args ...any) (db.ExecResult, error) {
	tag, err := c.pool.Exec(ctx, sql, args...)
	if err != nil {
		return db.ExecResult{}, fmt.Errorf("postgres: exec: %w", err)
	}
	return db.ExecResult{RowsAffected: tag.RowsAffected()}, nil
}

// Begin starts a transaction.
func (c *Conn) Begin(ctx context.Context) (db.Tx, error) {
	pgTx, err := c.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("postgres: begin: %w", err)
	}
	return &tx{pgTx: pgTx, tm: c.tm}, nil
}

// Dialect returns the PostgreSQL SQL dialect.
func (c *Conn) Dialect() db.Dialect { return db.PostgresDialect() }

// Inspector returns a schema inspector for this connection.
func (c *Conn) Inspector() schema.Inspector {
	return schema.NewPostgresInspector(c)
}

// TypeMap returns the live OID→type-name map loaded from pg_catalog.
// Exposed for postgres-specific introspection (e.g. `db introspect`).
func (c *Conn) TypeMap() *TypeMap { return c.tm }

// Close tears down the pool.
func (c *Conn) Close(_ context.Context) error {
	c.pool.Close()
	return nil
}

type tx struct {
	pgTx pgx.Tx
	tm   *TypeMap
}

func (t *tx) Query(ctx context.Context, sql string, args ...any) (db.Result, error) {
	rows, err := t.pgTx.Query(ctx, sql, args...)
	if err != nil {
		return db.Result{}, fmt.Errorf("postgres: tx query: %w", err)
	}
	return buildResult(rows, t.tm), nil
}

func (t *tx) Exec(ctx context.Context, sql string, args ...any) (db.ExecResult, error) {
	tag, err := t.pgTx.Exec(ctx, sql, args...)
	if err != nil {
		return db.ExecResult{}, fmt.Errorf("postgres: tx exec: %w", err)
	}
	return db.ExecResult{RowsAffected: tag.RowsAffected()}, nil
}

func (t *tx) Dialect() db.Dialect { return db.PostgresDialect() }

func (t *tx) Commit(ctx context.Context) error {
	return t.pgTx.Commit(ctx)
}

func (t *tx) Rollback(ctx context.Context) error {
	return t.pgTx.Rollback(ctx)
}
