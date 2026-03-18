package schema

import (
	"context"
	"errors"
	"sync"
	"testing"
)

// mockInspector tracks call counts per method.
type mockInspector struct {
	mu     sync.Mutex
	calls  map[string]int
	tables []Table
	cols   []ColumnInfo
	idxs   []Index
	cons   []Constraint
	fks    []ForeignKey
	err    error
}

func newMockInspector() *mockInspector {
	return &mockInspector{calls: make(map[string]int)}
}

func (m *mockInspector) callCount(method string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls[method]
}

func (m *mockInspector) record(method string) {
	m.mu.Lock()
	m.calls[method]++
	m.mu.Unlock()
}

func (m *mockInspector) Tables(_ context.Context, _ string) ([]Table, error) {
	m.record("Tables")
	return m.tables, m.err
}

func (m *mockInspector) Columns(_ context.Context, _, _ string) ([]ColumnInfo, error) {
	m.record("Columns")
	return m.cols, m.err
}

func (m *mockInspector) Indexes(_ context.Context, _, _ string) ([]Index, error) {
	m.record("Indexes")
	return m.idxs, m.err
}

func (m *mockInspector) Constraints(_ context.Context, _, _ string) ([]Constraint, error) {
	m.record("Constraints")
	return m.cons, m.err
}

func (m *mockInspector) ForeignKeys(_ context.Context, _, _ string) ([]ForeignKey, error) {
	m.record("ForeignKeys")
	return m.fks, m.err
}

func TestCachedInspector_Tables_CachesResult(t *testing.T) {
	mock := newMockInspector()
	mock.tables = []Table{{Name: "users", Schema: "public"}}
	ci := NewCachedInspector(mock)

	ctx := context.Background()
	got, err := ci.Tables(ctx, "public")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "users" {
		t.Fatalf("unexpected tables: %v", got)
	}

	// second call should use cache
	got2, err := ci.Tables(ctx, "public")
	if err != nil {
		t.Fatal(err)
	}
	if len(got2) != 1 {
		t.Fatalf("unexpected tables: %v", got2)
	}
	if mock.callCount("Tables") != 1 {
		t.Fatalf("expected 1 call, got %d", mock.callCount("Tables"))
	}
}

func TestCachedInspector_Columns_CachesResult(t *testing.T) {
	mock := newMockInspector()
	mock.cols = []ColumnInfo{{Name: "id", TypeName: "int"}}
	ci := NewCachedInspector(mock)

	ctx := context.Background()
	got, err := ci.Columns(ctx, "public", "users")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("unexpected: %v", got)
	}

	_, _ = ci.Columns(ctx, "public", "users")
	if mock.callCount("Columns") != 1 {
		t.Fatalf("expected 1 call, got %d", mock.callCount("Columns"))
	}
}

func TestCachedInspector_DifferentKeys(t *testing.T) {
	mock := newMockInspector()
	mock.cols = []ColumnInfo{{Name: "id"}}
	ci := NewCachedInspector(mock)

	ctx := context.Background()
	_, _ = ci.Columns(ctx, "public", "users")
	_, _ = ci.Columns(ctx, "public", "orders")
	if mock.callCount("Columns") != 2 {
		t.Fatalf("expected 2 calls, got %d", mock.callCount("Columns"))
	}
}

func TestCachedInspector_Invalidate(t *testing.T) {
	mock := newMockInspector()
	mock.tables = []Table{{Name: "users"}}
	mock.cols = []ColumnInfo{{Name: "id"}}
	ci := NewCachedInspector(mock)

	ctx := context.Background()
	_, _ = ci.Tables(ctx, "public")
	_, _ = ci.Columns(ctx, "public", "users")

	ci.Invalidate()

	_, _ = ci.Tables(ctx, "public")
	_, _ = ci.Columns(ctx, "public", "users")

	if mock.callCount("Tables") != 2 {
		t.Fatalf("expected 2 Tables calls, got %d", mock.callCount("Tables"))
	}
	if mock.callCount("Columns") != 2 {
		t.Fatalf("expected 2 Columns calls, got %d", mock.callCount("Columns"))
	}
}

func TestCachedInspector_InvalidateTable(t *testing.T) {
	mock := newMockInspector()
	mock.cols = []ColumnInfo{{Name: "id"}}
	mock.idxs = []Index{{Name: "pk"}}
	ci := NewCachedInspector(mock)

	ctx := context.Background()
	_, _ = ci.Columns(ctx, "public", "users")
	_, _ = ci.Indexes(ctx, "public", "users")
	_, _ = ci.Columns(ctx, "public", "orders")

	ci.InvalidateTable("public", "users")

	// users should re-query, orders should still be cached
	_, _ = ci.Columns(ctx, "public", "users")
	_, _ = ci.Indexes(ctx, "public", "users")
	_, _ = ci.Columns(ctx, "public", "orders")

	if mock.callCount("Columns") != 3 {
		t.Fatalf("expected 3 Columns calls, got %d", mock.callCount("Columns"))
	}
	if mock.callCount("Indexes") != 2 {
		t.Fatalf("expected 2 Indexes calls, got %d", mock.callCount("Indexes"))
	}
}

func TestCachedInspector_ErrorNotCached(t *testing.T) {
	mock := newMockInspector()
	mock.err = errors.New("db down")
	ci := NewCachedInspector(mock)

	ctx := context.Background()
	_, err := ci.Tables(ctx, "public")
	if err == nil {
		t.Fatal("expected error")
	}

	// error should not be cached — retry should hit inner again
	mock.err = nil
	mock.tables = []Table{{Name: "users"}}
	got, err := ci.Tables(ctx, "public")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("unexpected: %v", got)
	}
	if mock.callCount("Tables") != 2 {
		t.Fatalf("expected 2 calls, got %d", mock.callCount("Tables"))
	}
}

func TestCachedInspector_AllMethods(t *testing.T) {
	mock := newMockInspector()
	mock.tables = []Table{{Name: "t"}}
	mock.cols = []ColumnInfo{{Name: "c"}}
	mock.idxs = []Index{{Name: "i"}}
	mock.cons = []Constraint{{Name: "con"}}
	mock.fks = []ForeignKey{{Name: "fk"}}
	ci := NewCachedInspector(mock)

	ctx := context.Background()

	// first round: all hit inner
	_, _ = ci.Tables(ctx, "public")
	_, _ = ci.Columns(ctx, "public", "t")
	_, _ = ci.Indexes(ctx, "public", "t")
	_, _ = ci.Constraints(ctx, "public", "t")
	_, _ = ci.ForeignKeys(ctx, "public", "t")

	// second round: all cached
	_, _ = ci.Tables(ctx, "public")
	_, _ = ci.Columns(ctx, "public", "t")
	_, _ = ci.Indexes(ctx, "public", "t")
	_, _ = ci.Constraints(ctx, "public", "t")
	_, _ = ci.ForeignKeys(ctx, "public", "t")

	for _, method := range []string{"Tables", "Columns", "Indexes", "Constraints", "ForeignKeys"} {
		if mock.callCount(method) != 1 {
			t.Errorf("%s: expected 1 call, got %d", method, mock.callCount(method))
		}
	}
}

func TestCachedInspector_ConcurrentAccess(t *testing.T) {
	mock := newMockInspector()
	mock.cols = []ColumnInfo{{Name: "id"}}
	ci := NewCachedInspector(mock)

	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = ci.Columns(ctx, "public", "users")
		}()
	}
	wg.Wait()

	// should have queried at least once, but not necessarily 50 times
	count := mock.callCount("Columns")
	if count < 1 || count > 50 {
		t.Fatalf("unexpected call count: %d", count)
	}
}
