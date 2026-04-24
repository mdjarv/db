package editor

import (
	"context"
	"fmt"
	"testing"

	"github.com/mdjarv/db/internal/db"
)

type mockTx struct {
	execCalls  []mockExecCall
	committed  bool
	rolledBack bool
	execErr    error // if set, Exec returns this error on the nth call
	execErrAt  int
}

type mockExecCall struct {
	sql  string
	args []any
}

func (tx *mockTx) Query(_ context.Context, _ string, _ ...any) (db.Result, error) {
	return db.Result{}, nil
}

func (tx *mockTx) Exec(_ context.Context, sql string, args ...any) (db.ExecResult, error) {
	idx := len(tx.execCalls)
	tx.execCalls = append(tx.execCalls, mockExecCall{sql: sql, args: args})
	if tx.execErr != nil && idx == tx.execErrAt {
		return db.ExecResult{}, tx.execErr
	}
	return db.ExecResult{RowsAffected: 1}, nil
}

func (tx *mockTx) Commit(_ context.Context) error {
	tx.committed = true
	return nil
}

func (tx *mockTx) Rollback(_ context.Context) error {
	tx.rolledBack = true
	return nil
}

func (tx *mockTx) Dialect() db.Dialect { return db.PostgresDialect() }

type mockConn struct {
	tx       *mockTx
	beginErr error
	// for autocommit mode
	execCalls []mockExecCall
	execErr   error
	execErrAt int
}

func (c *mockConn) Query(_ context.Context, _ string, _ ...any) (db.Result, error) {
	return db.Result{}, nil
}

func (c *mockConn) Exec(_ context.Context, sql string, args ...any) (db.ExecResult, error) {
	idx := len(c.execCalls)
	c.execCalls = append(c.execCalls, mockExecCall{sql: sql, args: args})
	if c.execErr != nil && idx == c.execErrAt {
		return db.ExecResult{}, c.execErr
	}
	return db.ExecResult{RowsAffected: 1}, nil
}

func (c *mockConn) Begin(_ context.Context) (db.Tx, error) {
	if c.beginErr != nil {
		return nil, c.beginErr
	}
	return c.tx, nil
}

func (c *mockConn) Close(_ context.Context) error { return nil }

func (c *mockConn) Dialect() db.Dialect { return db.PostgresDialect() }

func TestApply_ExplicitCommit(t *testing.T) {
	tx := &mockTx{}
	conn := &mockConn{tx: tx}

	changes := []Change{
		{
			Kind: ChangeUpdate, Table: "users", Schema: "public",
			PK:     PKValue{Columns: []string{"id"}, Values: []any{1}},
			Column: "name", NewValue: "Bob",
		},
		{
			Kind: ChangeDelete, Table: "users", Schema: "public",
			PK: PKValue{Columns: []string{"id"}, Values: []any{2}},
		},
	}

	result := Apply(context.Background(), conn, changes, false)
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.Applied != 2 {
		t.Errorf("expected 2 applied, got %d", result.Applied)
	}
	if result.Tx == nil {
		t.Fatal("expected non-nil Tx")
	}
	if len(tx.execCalls) != 2 {
		t.Errorf("expected 2 exec calls, got %d", len(tx.execCalls))
	}
	if tx.committed || tx.rolledBack {
		t.Error("Tx should not be committed/rolled back yet")
	}

	// caller commits
	if err := result.Tx.Commit(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !tx.committed {
		t.Error("expected committed")
	}
}

func TestApply_ExplicitRollbackOnError(t *testing.T) {
	tx := &mockTx{execErr: fmt.Errorf("constraint violation"), execErrAt: 1}
	conn := &mockConn{tx: tx}

	changes := []Change{
		{
			Kind: ChangeUpdate, Table: "t", Schema: "s",
			PK:     PKValue{Columns: []string{"id"}, Values: []any{1}},
			Column: "a", NewValue: "x",
		},
		{
			Kind: ChangeUpdate, Table: "t", Schema: "s",
			PK:     PKValue{Columns: []string{"id"}, Values: []any{2}},
			Column: "a", NewValue: "y",
		},
	}

	result := Apply(context.Background(), conn, changes, false)
	if result.Err == nil {
		t.Fatal("expected error")
	}
	if result.Applied != 1 {
		t.Errorf("expected 1 applied before error, got %d", result.Applied)
	}
	if result.Tx != nil {
		t.Error("expected nil Tx on error")
	}
	if !tx.rolledBack {
		t.Error("expected rollback on error")
	}
}

func TestApply_BeginError(t *testing.T) {
	conn := &mockConn{beginErr: fmt.Errorf("connection lost")}
	result := Apply(context.Background(), conn, []Change{{Kind: ChangeInsert}}, false)
	if result.Err == nil {
		t.Fatal("expected error")
	}
}

func TestApply_Autocommit(t *testing.T) {
	conn := &mockConn{}

	changes := []Change{
		{
			Kind: ChangeUpdate, Table: "t", Schema: "s",
			PK:     PKValue{Columns: []string{"id"}, Values: []any{1}},
			Column: "a", NewValue: "x",
		},
	}

	result := Apply(context.Background(), conn, changes, true)
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.Applied != 1 {
		t.Errorf("expected 1 applied, got %d", result.Applied)
	}
	if result.Tx != nil {
		t.Error("expected nil Tx in autocommit mode")
	}
	if len(conn.execCalls) != 1 {
		t.Errorf("expected 1 direct exec call, got %d", len(conn.execCalls))
	}
}

func TestApply_AutocommitError(t *testing.T) {
	conn := &mockConn{execErr: fmt.Errorf("bad"), execErrAt: 0}

	changes := []Change{
		{
			Kind: ChangeDelete, Table: "t", Schema: "s",
			PK: PKValue{Columns: []string{"id"}, Values: []any{1}},
		},
	}

	result := Apply(context.Background(), conn, changes, true)
	if result.Err == nil {
		t.Fatal("expected error")
	}
	if result.Applied != 0 {
		t.Errorf("expected 0 applied, got %d", result.Applied)
	}
}

func TestApply_Empty(t *testing.T) {
	conn := &mockConn{}
	result := Apply(context.Background(), conn, nil, false)
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.Applied != 0 {
		t.Errorf("expected 0 applied, got %d", result.Applied)
	}
}
