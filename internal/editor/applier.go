package editor

import (
	"context"
	"fmt"

	"github.com/mdjarv/db/internal/db"
)

// ApplyResult reports the outcome of applying changes.
type ApplyResult struct {
	Applied int
	Tx      db.Tx // non-nil in explicit-commit mode; caller must Commit/Rollback
	Err     error
}

// Apply executes all pending changes against the database.
// In explicit mode (autocommit=false): wraps in a transaction, returns Tx for caller to commit/rollback.
// In auto mode (autocommit=true): executes each change immediately, rolls back nothing on error.
func Apply(ctx context.Context, conn db.Conn, changes []Change, autocommit bool) ApplyResult {
	if len(changes) == 0 {
		return ApplyResult{}
	}

	d := conn.Dialect()
	if autocommit {
		return applyAutocommit(ctx, conn, d, changes)
	}
	return applyExplicit(ctx, conn, d, changes)
}

func applyExplicit(ctx context.Context, conn db.Conn, d db.Dialect, changes []Change) ApplyResult {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return ApplyResult{Err: fmt.Errorf("begin transaction: %w", err)}
	}

	for i, c := range changes {
		dml := changeToDML(d, c)
		_, err := tx.Exec(ctx, dml.SQL, dml.Args...)
		if err != nil {
			_ = tx.Rollback(ctx)
			return ApplyResult{Applied: i, Err: fmt.Errorf("change %d: %w", i+1, err)}
		}
	}

	return ApplyResult{Applied: len(changes), Tx: tx}
}

func applyAutocommit(ctx context.Context, conn db.Conn, d db.Dialect, changes []Change) ApplyResult {
	for i, c := range changes {
		dml := changeToDML(d, c)
		_, err := conn.Exec(ctx, dml.SQL, dml.Args...)
		if err != nil {
			return ApplyResult{Applied: i, Err: fmt.Errorf("change %d: %w", i+1, err)}
		}
	}
	return ApplyResult{Applied: len(changes)}
}

func changeToDML(d db.Dialect, c Change) DMLResult {
	switch c.Kind {
	case ChangeUpdate:
		return GenerateUpdate(d, c.Schema, c.Table, c.PK, c.Column, c.NewValue)
	case ChangeInsert:
		return GenerateInsert(d, c.Schema, c.Table, c.Row)
	case ChangeDelete:
		return GenerateDelete(d, c.Schema, c.Table, c.PK)
	default:
		return DMLResult{}
	}
}
