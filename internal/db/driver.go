// Package db defines the database driver abstraction.
package db

import "context"

// Driver connects to a database given a DSN.
type Driver interface {
	Connect(ctx context.Context, dsn string) (Conn, error)
}

// Conn is a database connection (backed by a pool).
type Conn interface {
	Query(ctx context.Context, sql string, args ...any) (Result, error)
	Exec(ctx context.Context, sql string, args ...any) (ExecResult, error)
	Begin(ctx context.Context) (Tx, error)
	Close(ctx context.Context) error
}

// TypeIntrospector provides detailed type information for debugging.
type TypeIntrospector interface {
	TypeDetail(oid uint32) TypeDetail
}

// Tx is a database transaction.
type Tx interface {
	Query(ctx context.Context, sql string, args ...any) (Result, error)
	Exec(ctx context.Context, sql string, args ...any) (ExecResult, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
