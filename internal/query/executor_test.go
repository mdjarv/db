package query

import (
	"context"
	"fmt"
	"testing"

	"github.com/mdjarv/db/internal/db"
)

type mockConn struct {
	queryFn func(ctx context.Context, sql string, args ...any) (db.Result, error)
	execFn  func(ctx context.Context, sql string, args ...any) (db.ExecResult, error)
	beginFn func(ctx context.Context) (db.Tx, error)
	closeFn func(ctx context.Context) error
}

func (m *mockConn) Query(ctx context.Context, sql string, args ...any) (db.Result, error) {
	return m.queryFn(ctx, sql, args...)
}

func (m *mockConn) Exec(ctx context.Context, sql string, args ...any) (db.ExecResult, error) {
	return m.execFn(ctx, sql, args...)
}

func (m *mockConn) Begin(ctx context.Context) (db.Tx, error) {
	return m.beginFn(ctx)
}

func (m *mockConn) Dialect() db.Dialect { return db.PostgresDialect() }

func (m *mockConn) Close(ctx context.Context) error {
	return m.closeFn(ctx)
}

type mockTx struct {
	queryFn    func(ctx context.Context, sql string, args ...any) (db.Result, error)
	execFn     func(ctx context.Context, sql string, args ...any) (db.ExecResult, error)
	commitFn   func(ctx context.Context) error
	rollbackFn func(ctx context.Context) error
}

func (m *mockTx) Query(ctx context.Context, sql string, args ...any) (db.Result, error) {
	return m.queryFn(ctx, sql, args...)
}

func (m *mockTx) Exec(ctx context.Context, sql string, args ...any) (db.ExecResult, error) {
	return m.execFn(ctx, sql, args...)
}

func (m *mockTx) Commit(ctx context.Context) error {
	return m.commitFn(ctx)
}

func (m *mockTx) Rollback(ctx context.Context) error {
	return m.rollbackFn(ctx)
}

func (m *mockTx) Dialect() db.Dialect { return db.PostgresDialect() }

type sliceIter struct {
	rows [][]any
	pos  int
}

func (s *sliceIter) Next() bool {
	s.pos++
	return s.pos <= len(s.rows)
}

func (s *sliceIter) Values() ([]any, error) {
	return s.rows[s.pos-1], nil
}

func (s *sliceIter) Err() error { return nil }
func (s *sliceIter) Close()     {}

func newMockResult() db.Result {
	return db.Result{
		Columns: []db.Column{{Name: "id", TypeName: "int4"}},
		Rows:    &sliceIter{rows: [][]any{{1}, {2}}},
	}
}

func TestExecuteSelect(t *testing.T) {
	called := false
	c := &mockConn{
		queryFn: func(_ context.Context, _ string, _ ...any) (db.Result, error) {
			called = true
			return newMockResult(), nil
		},
	}
	e := NewExecutor(c, AutoCommit)
	res, err := e.Execute(context.Background(), "SELECT * FROM users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected queryFn to be called")
	}
	if !res.IsQuery {
		t.Fatal("expected IsQuery=true")
	}
	if res.Result == nil {
		t.Fatal("expected Result to be non-nil")
	}
	if res.ExecResult != nil {
		t.Fatal("expected ExecResult to be nil")
	}
	if res.Duration <= 0 {
		t.Fatal("expected Duration > 0")
	}
}

func TestExecuteInsert(t *testing.T) {
	called := false
	c := &mockConn{
		execFn: func(_ context.Context, _ string, _ ...any) (db.ExecResult, error) {
			called = true
			return db.ExecResult{RowsAffected: 3}, nil
		},
	}
	e := NewExecutor(c, AutoCommit)
	res, err := e.Execute(context.Background(), "INSERT INTO users (name) VALUES ('a')")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected execFn to be called")
	}
	if res.IsQuery {
		t.Fatal("expected IsQuery=false")
	}
	if res.ExecResult == nil {
		t.Fatal("expected ExecResult to be non-nil")
	}
	if res.ExecResult.RowsAffected != 3 {
		t.Fatalf("expected 3 rows affected, got %d", res.ExecResult.RowsAffected)
	}
	if res.Result != nil {
		t.Fatal("expected Result to be nil")
	}
}

func TestSQLTypeDetection(t *testing.T) {
	tests := []struct {
		sql     string
		isQuery bool
	}{
		{"SELECT 1", true},
		{"select 1", true},
		{"  SELECT 1", true},
		{"\n\tSELECT 1", true},
		{"WITH cte AS (SELECT 1) SELECT * FROM cte", true},
		{"EXPLAIN SELECT 1", true},
		{"SHOW search_path", true},
		{"TABLE users", true},
		{"VALUES (1), (2)", true},
		{"INSERT INTO t VALUES (1)", false},
		{"UPDATE t SET x=1", false},
		{"DELETE FROM t", false},
		{"CREATE TABLE t (id int)", false},
		{"DROP TABLE t", false},
		{"ALTER TABLE t ADD COLUMN x int", false},
		{"TRUNCATE t", false},
		{"", false},
	}
	for _, tt := range tests {
		got := isQuerySQL(tt.sql)
		if got != tt.isQuery {
			t.Errorf("isQuerySQL(%q) = %v, want %v", tt.sql, got, tt.isQuery)
		}
	}
}

func TestTransactionCommitLifecycle(t *testing.T) {
	txExecCalled := false
	mtx := &mockTx{
		execFn: func(_ context.Context, _ string, _ ...any) (db.ExecResult, error) {
			txExecCalled = true
			return db.ExecResult{RowsAffected: 1}, nil
		},
		commitFn: func(_ context.Context) error { return nil },
	}
	c := &mockConn{
		beginFn: func(_ context.Context) (db.Tx, error) { return mtx, nil },
	}

	e := NewExecutor(c, Explicit)
	ctx := context.Background()

	if e.InTransaction() {
		t.Fatal("should not be in transaction initially")
	}
	if err := e.Begin(ctx); err != nil {
		t.Fatalf("begin: %v", err)
	}
	if !e.InTransaction() {
		t.Fatal("should be in transaction after Begin")
	}

	_, err := e.Execute(ctx, "INSERT INTO t VALUES (1)")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !txExecCalled {
		t.Fatal("expected tx.Exec to be called, not conn.Exec")
	}

	if err := e.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}
	if e.InTransaction() {
		t.Fatal("should not be in transaction after Commit")
	}
}

func TestTransactionRollbackLifecycle(t *testing.T) {
	mtx := &mockTx{
		execFn: func(_ context.Context, _ string, _ ...any) (db.ExecResult, error) {
			return db.ExecResult{RowsAffected: 1}, nil
		},
		rollbackFn: func(_ context.Context) error { return nil },
	}
	c := &mockConn{
		beginFn: func(_ context.Context) (db.Tx, error) { return mtx, nil },
	}

	e := NewExecutor(c, Explicit)
	ctx := context.Background()

	if err := e.Begin(ctx); err != nil {
		t.Fatalf("begin: %v", err)
	}
	if _, err := e.Execute(ctx, "DELETE FROM t"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if err := e.Rollback(ctx); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if e.InTransaction() {
		t.Fatal("should not be in transaction after Rollback")
	}
}

func TestExecuteInTransactionUsesTx(t *testing.T) {
	connQueryCalled := false
	txQueryCalled := false
	mtx := &mockTx{
		queryFn: func(_ context.Context, _ string, _ ...any) (db.Result, error) {
			txQueryCalled = true
			return newMockResult(), nil
		},
		commitFn: func(_ context.Context) error { return nil },
	}
	c := &mockConn{
		queryFn: func(_ context.Context, _ string, _ ...any) (db.Result, error) {
			connQueryCalled = true
			return newMockResult(), nil
		},
		beginFn: func(_ context.Context) (db.Tx, error) { return mtx, nil },
	}

	e := NewExecutor(c, Explicit)
	ctx := context.Background()

	if err := e.Begin(ctx); err != nil {
		t.Fatalf("begin: %v", err)
	}
	if _, err := e.Execute(ctx, "SELECT 1"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if err := e.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	if connQueryCalled {
		t.Fatal("should not call conn.Query when in transaction")
	}
	if !txQueryCalled {
		t.Fatal("should call tx.Query when in transaction")
	}
}

func TestErrorPropagation(t *testing.T) {
	expectedErr := fmt.Errorf("connection lost")
	c := &mockConn{
		queryFn: func(_ context.Context, _ string, _ ...any) (db.Result, error) {
			return db.Result{}, expectedErr
		},
		execFn: func(_ context.Context, _ string, _ ...any) (db.ExecResult, error) {
			return db.ExecResult{}, expectedErr
		},
	}

	e := NewExecutor(c, AutoCommit)
	ctx := context.Background()

	_, err := e.Execute(ctx, "SELECT 1")
	if err != expectedErr {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	_, err = e.Execute(ctx, "INSERT INTO t VALUES (1)")
	if err != expectedErr {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestBeginAlreadyInTransaction(t *testing.T) {
	mtx := &mockTx{}
	c := &mockConn{
		beginFn: func(_ context.Context) (db.Tx, error) { return mtx, nil },
	}
	e := NewExecutor(c, Explicit)
	ctx := context.Background()

	if err := e.Begin(ctx); err != nil {
		t.Fatalf("first begin: %v", err)
	}
	if err := e.Begin(ctx); err == nil {
		t.Fatal("expected error on second Begin")
	}
}

func TestCommitWithoutTransaction(t *testing.T) {
	c := &mockConn{}
	e := NewExecutor(c, AutoCommit)
	if err := e.Commit(context.Background()); err == nil {
		t.Fatal("expected error on Commit without transaction")
	}
}

func TestRollbackWithoutTransaction(t *testing.T) {
	c := &mockConn{}
	e := NewExecutor(c, AutoCommit)
	if err := e.Rollback(context.Background()); err == nil {
		t.Fatal("expected error on Rollback without transaction")
	}
}
