// Package query provides SQL execution with transaction management.
package query

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mdjarv/db/internal/db"
)

// TransactionMode controls how statements are executed.
type TransactionMode int

// TransactionMode values.
const (
	AutoCommit TransactionMode = iota
	Explicit
)

// ExecResult holds the outcome of a single SQL execution.
type ExecResult struct {
	Result     *db.Result
	ExecResult *db.ExecResult
	Duration   time.Duration
	IsQuery    bool
}

// Executor runs SQL statements against a connection, optionally within a transaction.
type Executor struct {
	conn db.Conn
	tx   db.Tx
	mode TransactionMode
}

// NewExecutor creates an Executor for the given connection.
func NewExecutor(conn db.Conn, mode TransactionMode) *Executor {
	return &Executor{conn: conn, mode: mode}
}

// Execute runs a single SQL statement and returns the result with timing.
func (e *Executor) Execute(ctx context.Context, sql string) (*ExecResult, error) {
	isQuery := isQuerySQL(sql)
	start := time.Now()

	if isQuery {
		var (
			result db.Result
			err    error
		)
		if e.tx != nil {
			result, err = e.tx.Query(ctx, sql)
		} else {
			result, err = e.conn.Query(ctx, sql)
		}
		if err != nil {
			return nil, err
		}
		return &ExecResult{
			Result:   &result,
			Duration: time.Since(start),
			IsQuery:  true,
		}, nil
	}

	var (
		execRes db.ExecResult
		err     error
	)
	if e.tx != nil {
		execRes, err = e.tx.Exec(ctx, sql)
	} else {
		execRes, err = e.conn.Exec(ctx, sql)
	}
	if err != nil {
		return nil, err
	}
	return &ExecResult{
		ExecResult: &execRes,
		Duration:   time.Since(start),
		IsQuery:    false,
	}, nil
}

// Begin starts an explicit transaction.
func (e *Executor) Begin(ctx context.Context) error {
	if e.tx != nil {
		return fmt.Errorf("query: already in transaction")
	}
	tx, err := e.conn.Begin(ctx)
	if err != nil {
		return err
	}
	e.tx = tx
	return nil
}

// Commit commits the current transaction.
func (e *Executor) Commit(ctx context.Context) error {
	if e.tx == nil {
		return fmt.Errorf("query: not in transaction")
	}
	err := e.tx.Commit(ctx)
	e.tx = nil
	return err
}

// Rollback aborts the current transaction.
func (e *Executor) Rollback(ctx context.Context) error {
	if e.tx == nil {
		return fmt.Errorf("query: not in transaction")
	}
	err := e.tx.Rollback(ctx)
	e.tx = nil
	return err
}

// InTransaction returns true if an explicit transaction is active.
func (e *Executor) InTransaction() bool {
	return e.tx != nil
}

// isQuerySQL returns true if the SQL is a query (returns rows).
func isQuerySQL(sql string) bool {
	s := strings.TrimSpace(sql)
	if s == "" {
		return false
	}
	// Extract first word, case-insensitive
	end := strings.IndexAny(s, " \t\n\r(;")
	word := s
	if end > 0 {
		word = s[:end]
	}
	switch strings.ToUpper(word) {
	case "SELECT", "WITH", "EXPLAIN", "SHOW", "TABLE", "VALUES":
		return true
	}
	return false
}
