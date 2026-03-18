//go:build integration

package query_test

import (
	"context"
	"fmt"
	"runtime"
	"testing"

	"github.com/mdjarv/db/internal/query"
)

func TestIntegrationLargeResultSet(t *testing.T) {
	conn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	e := query.NewExecutor(conn, query.AutoCommit)

	// generate_series produces 100k rows without needing table inserts
	res, err := e.Execute(ctx, "SELECT i, 'row_' || i::text AS label FROM generate_series(1, 100000) AS i")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !res.IsQuery {
		t.Fatal("expected IsQuery=true")
	}

	var count int
	for res.Result.Rows.Next() {
		vals, err := res.Result.Rows.Values()
		if err != nil {
			t.Fatalf("values at row %d: %v", count, err)
		}
		if len(vals) != 2 {
			t.Fatalf("expected 2 values, got %d at row %d", len(vals), count)
		}
		count++
	}
	res.Result.Rows.Close()

	if count != 100000 {
		t.Fatalf("expected 100000 rows, got %d", count)
	}
}

func TestIntegrationLargeResultMemory(t *testing.T) {
	conn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	e := query.NewExecutor(conn, query.AutoCommit)

	// Force GC and measure baseline
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	// Stream 100k rows — pgx streams by default, so memory should stay bounded
	res, err := e.Execute(ctx, "SELECT i, repeat('x', 100) AS payload FROM generate_series(1, 100000) AS i")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	var count int
	for res.Result.Rows.Next() {
		if _, err := res.Result.Rows.Values(); err != nil {
			t.Fatalf("values: %v", err)
		}
		count++
	}
	res.Result.Rows.Close()

	runtime.GC()
	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	if count != 100000 {
		t.Fatalf("expected 100000 rows, got %d", count)
	}

	// Heap growth should be well under the total data size.
	// 100k rows * ~100 bytes = ~10MB raw data; streaming should keep heap growth under 50MB.
	heapGrowth := int64(after.HeapAlloc) - int64(before.HeapAlloc)
	const maxGrowth = 50 * 1024 * 1024 // 50 MB
	if heapGrowth > maxGrowth {
		t.Fatalf("heap grew %d bytes (%.1f MB), expected < 50 MB — likely not streaming",
			heapGrowth, float64(heapGrowth)/1024/1024)
	}
	t.Logf("heap growth: %.1f MB for 100k rows", float64(heapGrowth)/1024/1024)
}

func TestIntegrationLargeResultEarlyClose(t *testing.T) {
	conn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	e := query.NewExecutor(conn, query.AutoCommit)

	res, err := e.Execute(ctx, "SELECT i FROM generate_series(1, 100000) AS i")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	// Read only 10 rows then close
	for i := 0; i < 10; i++ {
		if !res.Result.Rows.Next() {
			t.Fatalf("expected row %d", i)
		}
	}
	res.Result.Rows.Close()

	if err := res.Result.Rows.Err(); err != nil {
		t.Fatalf("err after early close: %v", err)
	}

	// Connection should still be usable
	res2, err := e.Execute(ctx, "SELECT 1")
	if err != nil {
		t.Fatalf("subsequent query: %v", err)
	}
	res2.Result.Rows.Close()
}

func TestIntegrationConcurrentQueries(t *testing.T) {
	conn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()
	const goroutines = 10
	const queriesPerGoroutine = 20

	errs := make(chan error, goroutines*queriesPerGoroutine)
	done := make(chan struct{}, goroutines)

	for g := 0; g < goroutines; g++ {
		go func(id int) {
			defer func() { done <- struct{}{} }()
			e := query.NewExecutor(conn, query.AutoCommit)
			for q := 0; q < queriesPerGoroutine; q++ {
				sql := fmt.Sprintf("SELECT %d AS gid, %d AS qid", id, q)
				res, err := e.Execute(ctx, sql)
				if err != nil {
					errs <- fmt.Errorf("g%d q%d: %w", id, q, err)
					return
				}
				if !res.Result.Rows.Next() {
					errs <- fmt.Errorf("g%d q%d: no rows", id, q)
					res.Result.Rows.Close()
					return
				}
				res.Result.Rows.Close()
			}
		}(g)
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}
	close(errs)

	for err := range errs {
		t.Errorf("concurrent query error: %v", err)
	}
}
