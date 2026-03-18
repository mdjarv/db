package contextmenu

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func testItems() []MenuItem {
	return []MenuItem{
		{Label: "Query", ActionID: "query"},
		{Label: "Describe", ActionID: "describe"},
		{Label: "Dump table", ActionID: "dump", Hint: "..."},
		{Label: "Copy name", ActionID: "copy"},
	}
}

func key(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestOpenClose(t *testing.T) {
	m := New()
	if m.IsActive() {
		t.Fatal("should start inactive")
	}
	m.Open(testItems())
	if !m.IsActive() {
		t.Fatal("should be active after Open")
	}
	if m.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", m.Cursor())
	}
	m.Close()
	if m.IsActive() {
		t.Fatal("should be inactive after Close")
	}
}

func TestNavigationWraps(t *testing.T) {
	m := New()
	m.Open(testItems())

	// k at top wraps to bottom
	m.Update(key("k"))
	if m.Cursor() != 3 {
		t.Errorf("k at top: cursor = %d, want 3", m.Cursor())
	}

	// j at bottom wraps to top
	m.Update(key("j"))
	if m.Cursor() != 0 {
		t.Errorf("j at bottom: cursor = %d, want 0", m.Cursor())
	}

	// normal j
	m.Update(key("j"))
	if m.Cursor() != 1 {
		t.Errorf("j: cursor = %d, want 1", m.Cursor())
	}
}

func TestSelectionReturnsID(t *testing.T) {
	m := New()
	m.Open(testItems())

	m.Update(key("j")) // cursor=1 "Describe"
	id, selected := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !selected {
		t.Fatal("enter should select")
	}
	if id != "describe" {
		t.Errorf("id = %q, want 'describe'", id)
	}
	if m.IsActive() {
		t.Error("should close after selection")
	}
}

func TestEscCloses(t *testing.T) {
	m := New()
	m.Open(testItems())

	id, selected := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if selected {
		t.Error("esc should not select")
	}
	if id != "" {
		t.Errorf("id = %q, want empty", id)
	}
	if m.IsActive() {
		t.Error("should close on esc")
	}
}

func TestQCloses(t *testing.T) {
	m := New()
	m.Open(testItems())

	_, selected := m.Update(key("q"))
	if selected {
		t.Error("q should not select")
	}
	if m.IsActive() {
		t.Error("should close on q")
	}
}

func TestInactiveNoOp(t *testing.T) {
	m := New()
	id, selected := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if selected || id != "" {
		t.Error("inactive menu should not respond")
	}
}

func TestViewWhenInactive(t *testing.T) {
	m := New()
	v := m.View(80, 24)
	if v != "" {
		t.Error("inactive menu should render empty")
	}
}

func TestViewWhenActive(t *testing.T) {
	m := New()
	m.Open(testItems())
	v := m.View(80, 24)
	if v == "" {
		t.Error("active menu should render content")
	}
}
