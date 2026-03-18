package editor

import (
	"sync"
	"testing"
)

func TestChangeBuffer_Add(t *testing.T) {
	b := NewChangeBuffer()
	b.Add(Change{
		Kind:     ChangeUpdate,
		Table:    "users",
		Schema:   "public",
		PK:       PKValue{Columns: []string{"id"}, Values: []any{1}},
		Column:   "name",
		OldValue: "Alice",
		NewValue: "Bob",
	})
	if b.Len() != 1 {
		t.Fatalf("expected 1 change, got %d", b.Len())
	}
}

func TestChangeBuffer_Collapse(t *testing.T) {
	b := NewChangeBuffer()
	pk := PKValue{Columns: []string{"id"}, Values: []any{1}}

	b.Add(Change{
		Kind: ChangeUpdate, Table: "users", Schema: "public",
		PK: pk, Column: "name", OldValue: "Alice", NewValue: "Bob",
	})
	b.Add(Change{
		Kind: ChangeUpdate, Table: "users", Schema: "public",
		PK: pk, Column: "name", OldValue: "Bob", NewValue: "Charlie",
	})

	if b.Len() != 1 {
		t.Fatalf("expected 1 collapsed change, got %d", b.Len())
	}
	c := b.Changes()[0]
	if c.OldValue != "Alice" {
		t.Errorf("expected original OldValue 'Alice', got %v", c.OldValue)
	}
	if c.NewValue != "Charlie" {
		t.Errorf("expected NewValue 'Charlie', got %v", c.NewValue)
	}
}

func TestChangeBuffer_NoCollapseDifferentColumn(t *testing.T) {
	b := NewChangeBuffer()
	pk := PKValue{Columns: []string{"id"}, Values: []any{1}}

	b.Add(Change{
		Kind: ChangeUpdate, Table: "users", Schema: "public",
		PK: pk, Column: "name", OldValue: "Alice", NewValue: "Bob",
	})
	b.Add(Change{
		Kind: ChangeUpdate, Table: "users", Schema: "public",
		PK: pk, Column: "email", OldValue: "a@b.c", NewValue: "x@y.z",
	})

	if b.Len() != 2 {
		t.Fatalf("expected 2 changes, got %d", b.Len())
	}
}

func TestChangeBuffer_Remove(t *testing.T) {
	b := NewChangeBuffer()
	b.Add(Change{Kind: ChangeInsert, Table: "t", Schema: "s"})
	b.Add(Change{Kind: ChangeDelete, Table: "t", Schema: "s"})
	b.Remove(0)
	if b.Len() != 1 {
		t.Fatalf("expected 1 change, got %d", b.Len())
	}
	if b.Changes()[0].Kind != ChangeDelete {
		t.Error("expected remaining change to be delete")
	}
}

func TestChangeBuffer_RemoveLast(t *testing.T) {
	b := NewChangeBuffer()
	b.Add(Change{Kind: ChangeInsert, Table: "t", Schema: "s"})
	b.Add(Change{Kind: ChangeDelete, Table: "t", Schema: "s"})

	last := b.RemoveLast()
	if last == nil || last.Kind != ChangeDelete {
		t.Fatal("expected last change to be delete")
	}
	if b.Len() != 1 {
		t.Fatalf("expected 1 change, got %d", b.Len())
	}
}

func TestChangeBuffer_RemoveLastEmpty(t *testing.T) {
	b := NewChangeBuffer()
	if b.RemoveLast() != nil {
		t.Error("expected nil from empty buffer")
	}
}

func TestChangeBuffer_Clear(t *testing.T) {
	b := NewChangeBuffer()
	b.Add(Change{Kind: ChangeInsert, Table: "t", Schema: "s"})
	b.Add(Change{Kind: ChangeInsert, Table: "t", Schema: "s"})
	b.Clear()
	if b.Len() != 0 {
		t.Fatalf("expected 0 changes, got %d", b.Len())
	}
}

func TestChangeBuffer_HasChangesForCell(t *testing.T) {
	b := NewChangeBuffer()
	pk := PKValue{Columns: []string{"id"}, Values: []any{42}}
	b.Add(Change{
		Kind: ChangeUpdate, Table: "users", Schema: "public",
		PK: pk, Column: "name", OldValue: "x", NewValue: "y",
	})

	if !b.HasChangesForCell("users", "public", "name", pk) {
		t.Error("expected cell to have changes")
	}
	if b.HasChangesForCell("users", "public", "email", pk) {
		t.Error("expected no changes for different column")
	}

	otherPK := PKValue{Columns: []string{"id"}, Values: []any{99}}
	if b.HasChangesForCell("users", "public", "name", otherPK) {
		t.Error("expected no changes for different PK")
	}
}

func TestChangeBuffer_HasDeleteForRow(t *testing.T) {
	b := NewChangeBuffer()
	pk := PKValue{Columns: []string{"id"}, Values: []any{1}}
	b.Add(Change{Kind: ChangeDelete, Table: "t", Schema: "s", PK: pk})

	if !b.HasDeleteForRow("t", "s", pk) {
		t.Error("expected delete for row")
	}
	if b.HasDeleteForRow("t", "s", PKValue{Columns: []string{"id"}, Values: []any{2}}) {
		t.Error("expected no delete for different PK")
	}
}

func TestPKEqual(t *testing.T) {
	a := PKValue{Columns: []string{"id", "seq"}, Values: []any{1, "a"}}
	b := PKValue{Columns: []string{"id", "seq"}, Values: []any{1, "a"}}
	if !pkEqual(a, b) {
		t.Error("expected equal")
	}

	c := PKValue{Columns: []string{"id"}, Values: []any{1}}
	if pkEqual(a, c) {
		t.Error("expected not equal (different length)")
	}
}

func TestChangeBuffer_ChangesCopyIsolation(t *testing.T) {
	b := NewChangeBuffer()
	b.Add(Change{Kind: ChangeInsert, Table: "t", Schema: "s"})
	snapshot := b.Changes()

	b.Add(Change{Kind: ChangeInsert, Table: "t2", Schema: "s"})

	if len(snapshot) != 1 {
		t.Fatalf("snapshot should be isolated from buffer mutations, got len=%d", len(snapshot))
	}
	if b.Len() != 2 {
		t.Fatalf("expected buffer len=2, got %d", b.Len())
	}
}

func TestChangeBuffer_RemoveOutOfBounds(t *testing.T) {
	b := NewChangeBuffer()
	b.Add(Change{Kind: ChangeInsert, Table: "t", Schema: "s"})

	// No-op, no panic
	b.Remove(-1)
	b.Remove(5)
	if b.Len() != 1 {
		t.Fatalf("expected 1 after out-of-bounds removes, got %d", b.Len())
	}
}

func TestChangeBuffer_IsInsertedRow(t *testing.T) {
	b := NewChangeBuffer()
	b.Add(Change{Kind: ChangeInsert, Table: "t", Schema: "s", Row: map[string]any{"a": 1}})
	b.Add(Change{Kind: ChangeDelete, Table: "t", Schema: "s"})
	b.Add(Change{Kind: ChangeInsert, Table: "t", Schema: "s", Row: map[string]any{"a": 2}})

	if !b.IsInsertedRow("t", "s", 0) {
		t.Error("expected idx 0 to be inserted")
	}
	if !b.IsInsertedRow("t", "s", 1) {
		t.Error("expected idx 1 to be inserted")
	}
	if b.IsInsertedRow("t", "s", 2) {
		t.Error("expected idx 2 to not exist")
	}
}

func TestChangeBuffer_SerializedConcurrentAccess(t *testing.T) {
	// ChangeBuffer is not thread-safe by design; this test verifies
	// correct behavior under serialized concurrent access (mutex-guarded).
	b := NewChangeBuffer()
	var mu sync.Mutex
	var wg sync.WaitGroup

	const writers = 100
	wg.Add(writers)
	for i := 0; i < writers; i++ {
		go func(id int) {
			defer wg.Done()
			mu.Lock()
			defer mu.Unlock()
			b.Add(Change{
				Kind: ChangeInsert, Table: "t", Schema: "s",
				Row: map[string]any{"id": id},
			})
		}(i)
	}
	wg.Wait()

	if b.Len() != writers {
		t.Fatalf("expected %d changes, got %d", writers, b.Len())
	}

	// Concurrent reads with lock
	var wg2 sync.WaitGroup
	const readers = 50
	wg2.Add(readers)
	for i := 0; i < readers; i++ {
		go func() {
			defer wg2.Done()
			mu.Lock()
			defer mu.Unlock()
			changes := b.Changes()
			if len(changes) != writers {
				t.Errorf("expected %d changes in snapshot, got %d", writers, len(changes))
			}
		}()
	}
	wg2.Wait()
}

func TestChangeBuffer_CollapseCompositePK(t *testing.T) {
	b := NewChangeBuffer()
	pk := PKValue{Columns: []string{"a", "b"}, Values: []any{1, 2}}

	b.Add(Change{Kind: ChangeUpdate, Table: "t", Schema: "s", PK: pk, Column: "val", OldValue: "x", NewValue: "y"})
	b.Add(Change{Kind: ChangeUpdate, Table: "t", Schema: "s", PK: pk, Column: "val", OldValue: "x", NewValue: "z"})

	if b.Len() != 1 {
		t.Fatalf("expected 1 collapsed, got %d", b.Len())
	}
	if b.Changes()[0].NewValue != "z" {
		t.Fatalf("expected NewValue=z, got %v", b.Changes()[0].NewValue)
	}

	// Different composite PK should not collapse
	pk2 := PKValue{Columns: []string{"a", "b"}, Values: []any{1, 3}}
	b.Add(Change{Kind: ChangeUpdate, Table: "t", Schema: "s", PK: pk2, Column: "val", NewValue: "w"})
	if b.Len() != 2 {
		t.Fatalf("expected 2, got %d", b.Len())
	}
}
