package export

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mdjarv/db/internal/db"
)

func TestTableBasic(t *testing.T) {
	r := mockResult(
		[]db.Column{{Name: "id", TypeName: "int4"}, {Name: "name", TypeName: "text"}},
		[][]any{{int64(1), "alice"}, {int64(2), "bob"}},
	)
	var buf bytes.Buffer
	exp := NewExporter(FormatTable, Options{})
	if err := exp.Export(&buf, r); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "┌") || !strings.Contains(out, "┘") {
		t.Error("expected box-drawing borders")
	}
	if !strings.Contains(out, "id") || !strings.Contains(out, "name") {
		t.Error("expected column headers")
	}
	if !strings.Contains(out, "alice") || !strings.Contains(out, "bob") {
		t.Error("expected row data")
	}
	if !strings.Contains(out, "(2 rows)") {
		t.Error("expected row count footer")
	}
}

func TestTableNumericAlignment(t *testing.T) {
	r := mockResult(
		[]db.Column{{Name: "val", TypeName: "int4"}},
		[][]any{{int64(1)}, {int64(100)}},
	)
	var buf bytes.Buffer
	exp := NewExporter(FormatTable, Options{})
	if err := exp.Export(&buf, r); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(buf.String(), "\n")
	// Find data rows (contain │).
	var dataLines []string
	for _, line := range lines {
		// Skip header/border, find lines with data values.
		if strings.Contains(line, "│") && !strings.Contains(line, "val") {
			dataLines = append(dataLines, line)
		}
	}
	if len(dataLines) != 2 {
		t.Fatalf("expected 2 data lines, got %d", len(dataLines))
	}
	// "1" should be right-aligned: preceded by spaces.
	if !strings.Contains(dataLines[0], "  1") {
		t.Errorf("expected right-aligned 1, got: %s", dataLines[0])
	}
}

func TestTableTruncation(t *testing.T) {
	r := mockResult(
		[]db.Column{{Name: "v", TypeName: "text"}},
		[][]any{{"abcdefghij"}},
	)
	var buf bytes.Buffer
	exp := NewExporter(FormatTable, Options{MaxColWidth: 6})
	if err := exp.Export(&buf, r); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "abc...") {
		t.Errorf("expected truncation with ellipsis, got:\n%s", buf.String())
	}
}

func TestTableEmpty(t *testing.T) {
	r := mockResult(
		[]db.Column{{Name: "a", TypeName: "text"}},
		nil,
	)
	var buf bytes.Buffer
	exp := NewExporter(FormatTable, Options{})
	if err := exp.Export(&buf, r); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "(0 rows)") {
		t.Error("expected (0 rows) footer")
	}
}
