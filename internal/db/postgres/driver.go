// Package postgres implements the db.Driver interface using pgx/v5.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mdjarv/db/internal/db"
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
	return &conn{pool: pool, tm: tm}, nil
}

type conn struct {
	pool *pgxpool.Pool
	tm   *TypeMap
}

func (c *conn) Query(ctx context.Context, sql string, args ...any) (db.Result, error) {
	rows, err := c.pool.Query(ctx, sql, args...)
	if err != nil {
		return db.Result{}, fmt.Errorf("postgres: query: %w", err)
	}
	return buildResult(rows, c.tm), nil
}

func (c *conn) Exec(ctx context.Context, sql string, args ...any) (db.ExecResult, error) {
	tag, err := c.pool.Exec(ctx, sql, args...)
	if err != nil {
		return db.ExecResult{}, fmt.Errorf("postgres: exec: %w", err)
	}
	return db.ExecResult{RowsAffected: tag.RowsAffected()}, nil
}

func (c *conn) Begin(ctx context.Context) (db.Tx, error) {
	pgTx, err := c.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("postgres: begin: %w", err)
	}
	return &tx{pgTx: pgTx, tm: c.tm}, nil
}

func (c *conn) TypeDetail(oid uint32) db.TypeDetail {
	d := db.TypeDetail{
		OID:        oid,
		Name:       c.tm.Resolve(oid),
		EnumValues: c.tm.EnumValues(oid),
	}
	if pgFields := c.tm.CompositeFields(oid); pgFields != nil {
		d.CompositeFields = make([]db.CompositeField, len(pgFields))
		for i, f := range pgFields {
			d.CompositeFields[i] = db.CompositeField{Name: f.Name, TypeName: f.TypeName}
		}
	}
	if elemOID, ok := c.tm.ElemOID(oid); ok {
		d.IsArray = true
		d.ElemOID = elemOID
		d.ElemTypeName = c.tm.Resolve(elemOID)
	}
	return d
}

func (c *conn) Close(_ context.Context) error {
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

func (t *tx) Commit(ctx context.Context) error {
	return t.pgTx.Commit(ctx)
}

func (t *tx) Rollback(ctx context.Context) error {
	return t.pgTx.Rollback(ctx)
}
