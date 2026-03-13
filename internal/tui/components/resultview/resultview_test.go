package resultview

import (
	"testing"
	"time"

	"github.com/mdjarv/db/internal/tui/components/table"
	"github.com/mdjarv/db/internal/tui/core"
)

func TestAutoWidth(t *testing.T) {
	rows := [][]string{
		{"1", "Alice", "alice@example.com"},
		{"2", "Bob", "bob@example.com"},
		{"300", "Carol Longname", "carol@longdomain.example.com"},
	}

	w := autoWidth("id", "int4", rows, 0)
	if w < 4 {
		t.Errorf("width for 'id' = %d, want >= 4", w)
	}

	w = autoWidth("email", "text", rows, 2)
	expected := len("carol@longdomain.example.com")
	if w < expected {
		t.Errorf("width for 'email' = %d, want >= %d", w, expected)
	}
}

func TestAutoWidth_MaxCap(t *testing.T) {
	longVal := ""
	for range 100 {
		longVal += "x"
	}
	rows := [][]string{{longVal}}

	w := autoWidth("col", "", rows, 0)
	if w > 50 {
		t.Errorf("width = %d, should be capped at 50", w)
	}
}

func TestAutoWidth_MinCap(t *testing.T) {
	rows := [][]string{{"a"}}
	w := autoWidth("x", "", rows, 0)
	if w < 4 {
		t.Errorf("width = %d, should be at least 4", w)
	}
}

func TestAutoWidth_NullValues(t *testing.T) {
	rows := [][]string{{table.NullPlaceholder}}
	w := autoWidth("x", "", rows, 0)
	if w < 4 {
		t.Errorf("width = %d, should be >= 4 for NULL display", w)
	}
}

func TestSetResult(t *testing.T) {
	m := New()
	m.SetSize(80, 24)

	cols := []core.ResultColumn{
		{Name: "id", TypeName: "int4"},
		{Name: "name", TypeName: "text"},
	}
	rows := [][]string{
		{"1", "Alice"},
		{"2", "Bob"},
	}
	dur := 15 * time.Millisecond

	m.SetResult(cols, rows, dur)

	if !m.hasData {
		t.Error("hasData should be true")
	}
	if m.errMsg != "" {
		t.Errorf("errMsg = %q, want empty", m.errMsg)
	}
	if len(m.table.Rows) != 2 {
		t.Errorf("table rows = %d, want 2", len(m.table.Rows))
	}
	if len(m.table.Columns) != 2 {
		t.Errorf("table columns = %d, want 2", len(m.table.Columns))
	}
}

func TestSetError(t *testing.T) {
	m := New()
	m.SetSize(80, 24)
	m.SetResult(
		[]core.ResultColumn{{Name: "id"}},
		[][]string{{"1"}},
		0,
	)
	m.SetError(errTest)

	if m.hasData {
		t.Error("hasData should be false after error")
	}
	if m.errMsg != "test error" {
		t.Errorf("errMsg = %q, want 'test error'", m.errMsg)
	}
}

func TestClear(t *testing.T) {
	m := New()
	m.SetSize(80, 24)
	m.SetResult(
		[]core.ResultColumn{{Name: "id"}},
		[][]string{{"1"}},
		5*time.Millisecond,
	)
	m.Clear()

	if m.hasData {
		t.Error("hasData should be false after clear")
	}
	if m.duration != 0 {
		t.Error("duration should be 0 after clear")
	}
}

func TestResultData(t *testing.T) {
	m := New()
	cols, rows := m.ResultData()
	if cols != nil || rows != nil {
		t.Error("ResultData should return nil when no data")
	}

	m.SetSize(80, 24)
	m.SetResult(
		[]core.ResultColumn{{Name: "id", TypeName: "int4"}},
		[][]string{{"1"}, {"2"}},
		0,
	)
	cols, rows = m.ResultData()
	if len(cols) != 1 {
		t.Errorf("cols = %d, want 1", len(cols))
	}
	if len(rows) != 2 {
		t.Errorf("rows = %d, want 2", len(rows))
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{500 * time.Microsecond, "500\u00b5s"},
		{15 * time.Millisecond, "15.0ms"},
		{1500 * time.Millisecond, "1.50s"},
	}
	for _, tt := range tests {
		got := formatDuration(tt.d)
		if got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

type testError struct{}

func (testError) Error() string { return "test error" }

var errTest = testError{}
