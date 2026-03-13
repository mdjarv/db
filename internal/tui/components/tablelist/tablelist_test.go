package tablelist

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mdjarv/db/internal/schema"
	"github.com/mdjarv/db/internal/tui/core"
)

func testTables() []schema.Table {
	return []schema.Table{
		{Name: "users", Schema: "public", Type: "table", RowEstimate: 1000},
		{Name: "orders", Schema: "public", Type: "table", RowEstimate: 5000},
		{Name: "products", Schema: "public", Type: "table", RowEstimate: 200},
		{Name: "user_sessions", Schema: "public", Type: "view", RowEstimate: 0},
		{Name: "order_totals", Schema: "public", Type: "materialized view", RowEstimate: 5000},
	}
}

func loadedModel() *Model {
	m := New()
	m.SetSize(40, 20)
	m.SetFocused(true)
	m.Update(core.SchemaLoadedMsg{Tables: testTables()})
	return m
}

func sendKey(m *Model, key string) tea.Cmd {
	return m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
}

func TestSchemaLoaded(t *testing.T) {
	m := loadedModel()
	if len(m.Tables()) != 5 {
		t.Fatalf("tables = %d, want 5", len(m.Tables()))
	}
	if len(m.Filtered()) != 5 {
		t.Fatalf("filtered = %d, want 5", len(m.Filtered()))
	}
	if m.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", m.Cursor())
	}
}

func TestSchemaLoadedError(t *testing.T) {
	m := New()
	m.SetFocused(true)
	m.Update(core.SchemaLoadedMsg{Err: errTest})
	if len(m.Tables()) != 0 {
		t.Error("tables should be empty on error")
	}
}

var errTest = &testError{}

type testError struct{}

func (e *testError) Error() string { return "test" }

func TestNavigation(t *testing.T) {
	m := loadedModel()

	sendKey(m, "j")
	if m.Cursor() != 1 {
		t.Errorf("after j: cursor = %d, want 1", m.Cursor())
	}

	sendKey(m, "k")
	if m.Cursor() != 0 {
		t.Errorf("after k: cursor = %d, want 0", m.Cursor())
	}

	// k at top should stay
	sendKey(m, "k")
	if m.Cursor() != 0 {
		t.Errorf("k at top: cursor = %d, want 0", m.Cursor())
	}
}

func TestGG(t *testing.T) {
	m := loadedModel()

	// go to bottom
	sendKey(m, "G")
	if m.Cursor() != 4 {
		t.Errorf("G: cursor = %d, want 4", m.Cursor())
	}

	// gg to top
	sendKey(m, "g")
	sendKey(m, "g")
	if m.Cursor() != 0 {
		t.Errorf("gg: cursor = %d, want 0", m.Cursor())
	}
}

func TestSingleG(t *testing.T) {
	m := loadedModel()
	sendKey(m, "G") // bottom
	sendKey(m, "g") // first g, should not move yet
	if m.Cursor() != 4 {
		t.Errorf("single g should not move, cursor = %d", m.Cursor())
	}
	// any other key resets the g flag
	sendKey(m, "j")
	sendKey(m, "g")
	sendKey(m, "j") // resets g
	if m.Cursor() != 4 {
		t.Errorf("g then j: cursor = %d, want 4", m.Cursor())
	}
}

func TestFilter(t *testing.T) {
	m := loadedModel()

	sendKey(m, "/")
	if !m.IsFiltering() {
		t.Fatal("should be in filter mode")
	}

	sendKey(m, "u")
	sendKey(m, "s")
	sendKey(m, "e")
	sendKey(m, "r")

	if len(m.Filtered()) != 2 {
		t.Errorf("filter 'user': filtered = %d, want 2 (users, user_sessions)", len(m.Filtered()))
	}

	// enter confirms filter
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.IsFiltering() {
		t.Error("enter should exit filter mode")
	}
	if len(m.Filtered()) != 2 {
		t.Error("filter should persist after enter")
	}
}

func TestFilterEscClears(t *testing.T) {
	m := loadedModel()

	sendKey(m, "/")
	sendKey(m, "x")
	sendKey(m, "y")
	sendKey(m, "z")

	if len(m.Filtered()) != 0 {
		t.Errorf("filter 'xyz' should match 0, got %d", len(m.Filtered()))
	}

	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.IsFiltering() {
		t.Error("esc should exit filter mode")
	}
	if len(m.Filtered()) != 5 {
		t.Errorf("esc should clear filter, got %d", len(m.Filtered()))
	}
}

func TestFilterBackspace(t *testing.T) {
	m := loadedModel()

	sendKey(m, "/")
	sendKey(m, "x")
	sendKey(m, "y")
	m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	// should have filter "x" now
	if len(m.Filtered()) != 0 {
		t.Errorf("filter 'x' should match 0 tables, got %d", len(m.Filtered()))
	}
	m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	// filter empty
	if len(m.Filtered()) != 5 {
		t.Errorf("empty filter should show all, got %d", len(m.Filtered()))
	}
}

func TestEnterEmitsQueryRequest(t *testing.T) {
	m := loadedModel()
	cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter should return a cmd")
	}
	msg := cmd()
	qr, ok := msg.(core.QueryRequestMsg)
	if !ok {
		t.Fatalf("expected QueryRequestMsg, got %T", msg)
	}
	if !strings.Contains(qr.SQL, "users") {
		t.Errorf("query should reference 'users', got %q", qr.SQL)
	}
	if !strings.Contains(qr.SQL, "LIMIT 100") {
		t.Errorf("query should have LIMIT 100, got %q", qr.SQL)
	}
}

func TestYankEmitsYankMsg(t *testing.T) {
	m := loadedModel()
	cmd := sendKey(m, "y")
	if cmd == nil {
		t.Fatal("y should return a cmd")
	}
	msg := cmd()
	ym, ok := msg.(core.YankMsg)
	if !ok {
		t.Fatalf("expected YankMsg, got %T", msg)
	}
	if ym.Content != "users" {
		t.Errorf("yanked = %q, want 'users'", ym.Content)
	}
}

func TestRefreshEmitsRefreshMsg(t *testing.T) {
	m := loadedModel()
	cmd := sendKey(m, "R")
	if cmd == nil {
		t.Fatal("R should return a cmd")
	}
	msg := cmd()
	if _, ok := msg.(core.RefreshSchemaMsg); !ok {
		t.Fatalf("expected RefreshSchemaMsg, got %T", msg)
	}
}

func TestDescribeView(t *testing.T) {
	m := loadedModel()

	m.Update(core.TableDetailMsg{
		Table:   testTables()[0],
		Columns: []schema.ColumnInfo{{Name: "id", TypeName: "integer", IsPK: true}},
	})

	sendKey(m, "d")
	if !m.InDetailView() {
		t.Fatal("d should switch to detail view")
	}

	v := m.View()
	if !strings.Contains(v, "id") {
		t.Error("detail view should show column name 'id'")
	}

	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.InDetailView() {
		t.Error("esc should return to list view")
	}
}

func TestTypeIcons(t *testing.T) {
	m := loadedModel()
	v := m.View()
	if !strings.Contains(v, "T ") {
		t.Error("table should show T icon")
	}
	// navigate to the view entry
	sendKey(m, "j")
	sendKey(m, "j")
	sendKey(m, "j")
	v = m.View()
	if !strings.Contains(v, "V ") {
		t.Error("view should show V icon")
	}
}

func TestNotFocusedIgnoresKeys(t *testing.T) {
	m := loadedModel()
	m.SetFocused(false)
	sendKey(m, "j")
	if m.Cursor() != 0 {
		t.Error("unfocused model should not respond to keys")
	}
}

func TestEmptyTablesView(t *testing.T) {
	m := New()
	m.SetSize(40, 20)
	m.SetFocused(true)
	v := m.View()
	if !strings.Contains(v, "no tables") {
		t.Error("empty state should show 'no tables'")
	}
}

func TestSchemaTableNonPublic(t *testing.T) {
	m := New()
	m.SetSize(40, 20)
	m.SetFocused(true)
	m.Update(core.SchemaLoadedMsg{Tables: []schema.Table{
		{Name: "audit_log", Schema: "audit", Type: "table", RowEstimate: 100},
	}})

	cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter should return cmd")
	}
	msg := cmd()
	qr := msg.(core.QueryRequestMsg)
	if !strings.Contains(qr.SQL, "audit.audit_log") {
		t.Errorf("non-public schema should be qualified, got %q", qr.SQL)
	}
}

func TestTableSelectedOnLoad(t *testing.T) {
	m := New()
	m.SetSize(40, 20)
	m.SetFocused(true)
	cmd := m.Update(core.SchemaLoadedMsg{Tables: testTables()})
	if cmd == nil {
		t.Fatal("schema load should emit TableSelectedMsg")
	}
	msg := cmd()
	ts, ok := msg.(core.TableSelectedMsg)
	if !ok {
		t.Fatalf("expected TableSelectedMsg, got %T", msg)
	}
	if ts.Table.Name != "users" {
		t.Errorf("selected = %q, want 'users'", ts.Table.Name)
	}
}

func TestTableSelectedOnNavigate(t *testing.T) {
	m := loadedModel()
	cmd := sendKey(m, "j")
	if cmd == nil {
		t.Fatal("j should emit TableSelectedMsg")
	}
	msg := cmd()
	ts, ok := msg.(core.TableSelectedMsg)
	if !ok {
		t.Fatalf("expected TableSelectedMsg, got %T", msg)
	}
	if ts.Table.Name != "orders" {
		t.Errorf("selected = %q, want 'orders'", ts.Table.Name)
	}
}
