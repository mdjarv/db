package table

import "testing"

func newTestTable() *Model {
	cols := []Column{
		{Title: "a", Width: 4},
		{Title: "b", Width: 4},
		{Title: "c", Width: 4},
	}
	rows := make([][]string, 20)
	for i := range rows {
		rows[i] = []string{"a", "b", "c"}
	}
	m := New(cols, rows)
	m.Width = 40
	m.Height = 12
	return &m
}

func TestHalfPageDown(t *testing.T) {
	m := newTestTable()
	vh := m.ViewHeight()
	half := vh / 2

	m.HalfPageDown()
	if m.CursorRow != half {
		t.Errorf("CursorRow = %d, want %d", m.CursorRow, half)
	}
}

func TestHalfPageUp(t *testing.T) {
	m := newTestTable()
	m.CursorRow = 10
	vh := m.ViewHeight()
	half := vh / 2

	m.HalfPageUp()
	if m.CursorRow != 10-half {
		t.Errorf("CursorRow = %d, want %d", m.CursorRow, 10-half)
	}
}

func TestHalfPageUp_ClampZero(t *testing.T) {
	m := newTestTable()
	m.CursorRow = 2

	m.HalfPageUp()
	if m.CursorRow != 0 {
		t.Errorf("CursorRow = %d, want 0", m.CursorRow)
	}
}

func TestFullPageDown(t *testing.T) {
	m := newTestTable()
	vh := m.ViewHeight()

	m.FullPageDown()
	if m.CursorRow != vh {
		t.Errorf("CursorRow = %d, want %d", m.CursorRow, vh)
	}
}

func TestFullPageDown_ClampEnd(t *testing.T) {
	m := newTestTable()
	m.CursorRow = 18

	m.FullPageDown()
	if m.CursorRow != 19 {
		t.Errorf("CursorRow = %d, want 19 (last row)", m.CursorRow)
	}
}

func TestFullPageUp(t *testing.T) {
	m := newTestTable()
	m.CursorRow = 15
	vh := m.ViewHeight()

	m.FullPageUp()
	if m.CursorRow != 15-vh {
		t.Errorf("CursorRow = %d, want %d", m.CursorRow, 15-vh)
	}
}

func TestGotoFirstCol(t *testing.T) {
	m := newTestTable()
	m.CursorCol = 2
	m.ColOffset = 1

	m.GotoFirstCol()
	if m.CursorCol != 0 {
		t.Errorf("CursorCol = %d, want 0", m.CursorCol)
	}
	if m.ColOffset != 0 {
		t.Errorf("ColOffset = %d, want 0", m.ColOffset)
	}
}

func TestGotoLastCol(t *testing.T) {
	m := newTestTable()

	m.GotoLastCol()
	if m.CursorCol != 2 {
		t.Errorf("CursorCol = %d, want 2", m.CursorCol)
	}
}

func TestTruncateCell(t *testing.T) {
	tests := []struct {
		val   string
		width int
		want  string
	}{
		{"hello", 10, "hello     "},
		{"hello world", 5, "hell\u2026"},
		{"ab", 2, "ab"},
		{"abc", 2, "a\u2026"},
	}
	for _, tt := range tests {
		got := truncateCell(tt.val, tt.width)
		if got != tt.want {
			t.Errorf("truncateCell(%q, %d) = %q, want %q", tt.val, tt.width, got, tt.want)
		}
	}
}

func TestNullPlaceholder_Rendering(t *testing.T) {
	cols := []Column{{Title: "col", Width: 6}}
	rows := [][]string{{NullPlaceholder}}
	m := New(cols, rows)
	m.Width = 20
	m.Height = 5

	view := m.View(false)
	if view == "" {
		t.Error("view should not be empty")
	}
	// NULL should appear in the rendered output
	if !containsNull(view) {
		t.Error("rendered view should contain NULL display")
	}
}

func containsNull(s string) bool {
	for i := range len(s) - 3 {
		if s[i:i+4] == "NULL" {
			return true
		}
	}
	return false
}
