// Package editor implements a change buffer and DML generator for inline data editing.
// This package has no TUI dependencies.
package editor

// ChangeKind identifies the type of data change.
type ChangeKind int

// Change kinds.
const (
	ChangeUpdate ChangeKind = iota
	ChangeInsert
	ChangeDelete
)

// PKValue identifies a row by its primary key columns and values.
type PKValue struct {
	Columns []string
	Values  []any
}

// Change represents a single pending data modification.
type Change struct {
	Kind     ChangeKind
	Table    string
	Schema   string
	PK       PKValue
	Column   string         // UPDATE only: which column changed
	OldValue any            // UPDATE only
	NewValue any            // UPDATE only
	Row      map[string]any // INSERT only: column -> value
}

// ChangeBuffer holds an ordered list of pending changes with collapse support.
type ChangeBuffer struct {
	changes []Change
}

// NewChangeBuffer creates an empty change buffer.
func NewChangeBuffer() *ChangeBuffer {
	return &ChangeBuffer{}
}

// Add appends a change, collapsing duplicate cell edits.
func (b *ChangeBuffer) Add(c Change) {
	if c.Kind == ChangeUpdate {
		// collapse: if same cell already has a pending update, replace it
		for i, existing := range b.changes {
			if existing.Kind == ChangeUpdate &&
				existing.Table == c.Table &&
				existing.Schema == c.Schema &&
				existing.Column == c.Column &&
				pkEqual(existing.PK, c.PK) {
				// keep original OldValue, update NewValue
				b.changes[i].NewValue = c.NewValue
				return
			}
		}
	}
	b.changes = append(b.changes, c)
}

// Remove deletes the change at index i.
func (b *ChangeBuffer) Remove(i int) {
	if i < 0 || i >= len(b.changes) {
		return
	}
	b.changes = append(b.changes[:i], b.changes[i+1:]...)
}

// RemoveLast removes the most recent change and returns it.
// Returns nil if buffer is empty.
func (b *ChangeBuffer) RemoveLast() *Change {
	if len(b.changes) == 0 {
		return nil
	}
	last := b.changes[len(b.changes)-1]
	b.changes = b.changes[:len(b.changes)-1]
	return &last
}

// Changes returns a copy of all pending changes.
func (b *ChangeBuffer) Changes() []Change {
	out := make([]Change, len(b.changes))
	copy(out, b.changes)
	return out
}

// Len returns the number of pending changes.
func (b *ChangeBuffer) Len() int {
	return len(b.changes)
}

// Clear removes all pending changes.
func (b *ChangeBuffer) Clear() {
	b.changes = nil
}

// HasChangesForCell returns true if there's a pending update for the given cell.
func (b *ChangeBuffer) HasChangesForCell(table, schema, column string, pk PKValue) bool {
	for _, c := range b.changes {
		if c.Kind == ChangeUpdate &&
			c.Table == table &&
			c.Schema == schema &&
			c.Column == column &&
			pkEqual(c.PK, pk) {
			return true
		}
	}
	return false
}

// HasDeleteForRow returns true if there's a pending delete for the given PK.
func (b *ChangeBuffer) HasDeleteForRow(table, schema string, pk PKValue) bool {
	for _, c := range b.changes {
		if c.Kind == ChangeDelete &&
			c.Table == table &&
			c.Schema == schema &&
			pkEqual(c.PK, pk) {
			return true
		}
	}
	return false
}

// IsInsertedRow returns true if there's a pending insert at the given index.
// insertIdx is relative to the insert order in the buffer.
func (b *ChangeBuffer) IsInsertedRow(table, schema string, idx int) bool {
	count := 0
	for _, c := range b.changes {
		if c.Kind == ChangeInsert && c.Table == table && c.Schema == schema {
			if count == idx {
				return true
			}
			count++
		}
	}
	return false
}

func pkEqual(a, b PKValue) bool {
	if len(a.Columns) != len(b.Columns) {
		return false
	}
	for i := range a.Columns {
		if a.Columns[i] != b.Columns[i] {
			return false
		}
	}
	if len(a.Values) != len(b.Values) {
		return false
	}
	for i := range a.Values {
		if a.Values[i] != b.Values[i] {
			return false
		}
	}
	return true
}
