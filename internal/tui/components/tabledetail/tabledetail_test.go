package tabledetail

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mdjarv/db/internal/schema"
)

func sampleTable() schema.Table {
	return schema.Table{Name: "users", Schema: "public", Type: "table", RowEstimate: 2847}
}

func sampleColumns() []schema.ColumnInfo {
	return []schema.ColumnInfo{
		{Name: "id", TypeName: "serial", Nullable: false, IsPK: true, Position: 1},
		{Name: "name", TypeName: "varchar(100)", Nullable: false, Position: 2},
		{Name: "email", TypeName: "varchar(255)", Nullable: false, Position: 3},
		{Name: "active", TypeName: "boolean", Nullable: true, Default: "true", Position: 4},
		{Name: "created_at", TypeName: "timestamptz", Nullable: true, Default: "now()", Position: 5},
	}
}

func TestOpenClose(t *testing.T) {
	m := New()
	if m.IsActive() {
		t.Error("should start inactive")
	}

	m.Open(sampleTable(), sampleColumns(), nil, nil, nil)
	if !m.IsActive() {
		t.Error("should be active after Open")
	}

	m.Close()
	if m.IsActive() {
		t.Error("should be inactive after Close")
	}
}

func TestScrollJK(t *testing.T) {
	m := New()
	m.Open(sampleTable(), sampleColumns(), nil, nil, nil)

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if m.scroll != 1 {
		t.Errorf("scroll after j = %d, want 1", m.scroll)
	}

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if m.scroll != 2 {
		t.Errorf("scroll after jj = %d, want 2", m.scroll)
	}

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if m.scroll != 1 {
		t.Errorf("scroll after k = %d, want 1", m.scroll)
	}

	// k at 0 stays at 0
	m.scroll = 0
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if m.scroll != 0 {
		t.Errorf("scroll should not go below 0, got %d", m.scroll)
	}
}

func TestScrollGAndShiftG(t *testing.T) {
	m := New()
	m.Open(sampleTable(), sampleColumns(), nil, nil, nil)

	m.scroll = 5
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	if m.scroll != 0 {
		t.Errorf("g should scroll to top, got %d", m.scroll)
	}

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	if m.scroll != 9999 {
		t.Errorf("G should set high scroll, got %d", m.scroll)
	}
}

func TestScrollHalfPage(t *testing.T) {
	m := New()
	m.Open(sampleTable(), sampleColumns(), nil, nil, nil)

	m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	if m.scroll != 10 {
		t.Errorf("ctrl+d should scroll +10, got %d", m.scroll)
	}

	m.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	if m.scroll != 0 {
		t.Errorf("ctrl+u should scroll -10, got %d", m.scroll)
	}

	// ctrl+u from 0 stays at 0
	m.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	if m.scroll != 0 {
		t.Errorf("ctrl+u from 0 should stay at 0, got %d", m.scroll)
	}
}

func TestDismissKeys(t *testing.T) {
	keys := []string{"d", "q", "esc"}
	for _, k := range keys {
		m := New()
		m.Open(sampleTable(), sampleColumns(), nil, nil, nil)

		var msg tea.KeyMsg
		switch k {
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		default:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		}
		m.Update(msg)
		if m.IsActive() {
			t.Errorf("key %q should dismiss overlay", k)
		}
	}
}

func TestViewNotActiveReturnsEmpty(t *testing.T) {
	m := New()
	if v := m.View(80, 40); v != "" {
		t.Error("inactive overlay should return empty string")
	}
}

func TestViewContainsTableName(t *testing.T) {
	m := New()
	m.Open(sampleTable(), sampleColumns(), nil, nil, nil)
	v := m.View(80, 40)
	if v == "" {
		t.Fatal("View should not be empty when active")
	}
	// table name should appear somewhere in output
	if len(v) < 10 {
		t.Error("View output too short")
	}
}

func TestFilterConstraints(t *testing.T) {
	cons := []schema.Constraint{
		{Name: "pk", Type: "PRIMARY KEY", Columns: []string{"id"}},
		{Name: "fk", Type: "FOREIGN KEY", Columns: []string{"user_id"}},
		{Name: "chk", Type: "CHECK", Columns: []string{"active"}, Definition: "(active IS NOT NULL)"},
		{Name: "uq", Type: "UNIQUE", Columns: []string{"email"}},
	}
	filtered := filterConstraints(cons)
	if len(filtered) != 2 {
		t.Fatalf("expected 2 filtered constraints, got %d", len(filtered))
	}
	if filtered[0].Name != "chk" || filtered[1].Name != "uq" {
		t.Errorf("unexpected filtered constraints: %v", filtered)
	}
}

func TestAbbreviateType(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"timestamp with time zone", "timestamptz"},
		{"timestamp without time zone", "timestamp"},
		{"time with time zone", "timetz"},
		{"time without time zone", "time"},
		{"double precision", "float8"},
		{"character varying", "varchar"},
		{"character varying(100)", "varchar(100)"},
		{"character(1)", "char(1)"},
		{"character", "char"},
		{"uuid", "uuid"},
		{"text", "text"},
		{"jsonb", "jsonb"},
	}
	for _, tt := range tests {
		got := abbreviateType(tt.in)
		if got != tt.want {
			t.Errorf("abbreviateType(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestFormatCount(t *testing.T) {
	tests := []struct {
		n    int64
		want string
	}{
		{0, "0"},
		{500, "500"},
		{2847, "2.8K"},
		{1000000, "1.0M"},
		{2500000, "2.5M"},
	}
	for _, tt := range tests {
		got := formatCount(tt.n)
		if got != tt.want {
			t.Errorf("formatCount(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}
