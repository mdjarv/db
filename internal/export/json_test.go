package export

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mdjarv/db/internal/db"
)

func TestJSONBasicArray(t *testing.T) {
	r := mockResult(
		[]db.Column{{Name: "id", TypeName: "int4"}, {Name: "name", TypeName: "text"}},
		[][]any{{int64(1), "alice"}, {int64(2), "bob"}},
	)
	var buf bytes.Buffer
	exp := NewExporter(FormatJSON, Options{})
	if err := exp.Export(&buf, r); err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(buf.String())
	want := `[{"id":1,"name":"alice"}` + "\n" + `,{"id":2,"name":"bob"}` + "\n" + `]`
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestJSONPretty(t *testing.T) {
	r := mockResult(
		[]db.Column{{Name: "x", TypeName: "int4"}},
		[][]any{{int64(42)}},
	)
	var buf bytes.Buffer
	exp := NewExporter(FormatJSON, Options{Pretty: true})
	if err := exp.Export(&buf, r); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "  ") {
		t.Error("expected indented output")
	}
	if !strings.Contains(buf.String(), `"x"`) {
		t.Error("expected column name in output")
	}
}

func TestJSONLines(t *testing.T) {
	r := mockResult(
		[]db.Column{{Name: "a", TypeName: "text"}},
		[][]any{{"one"}, {"two"}},
	)
	var buf bytes.Buffer
	exp := NewExporter(FormatJSON, Options{JSONLines: true})
	if err := exp.Export(&buf, r); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), buf.String())
	}
	if lines[0] != `{"a":"one"}` {
		t.Errorf("line 0: %s", lines[0])
	}
}

func TestJSONTypePreservation(t *testing.T) {
	r := mockResult(
		[]db.Column{
			{Name: "i", TypeName: "int4"},
			{Name: "f", TypeName: "float8"},
			{Name: "b", TypeName: "bool"},
			{Name: "n", TypeName: "text"},
			{Name: "s", TypeName: "text"},
		},
		[][]any{{int64(10), float64(3.14), true, nil, "hello"}},
	)
	var buf bytes.Buffer
	exp := NewExporter(FormatJSON, Options{JSONLines: true})
	if err := exp.Export(&buf, r); err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(buf.String())
	want := `{"i":10,"f":3.14,"b":true,"n":null,"s":"hello"}`
	if got != want {
		t.Errorf("got:  %s\nwant: %s", got, want)
	}
}

func TestJSONEmpty(t *testing.T) {
	r := mockResult(
		[]db.Column{{Name: "a", TypeName: "text"}},
		nil,
	)
	var buf bytes.Buffer
	exp := NewExporter(FormatJSON, Options{})
	if err := exp.Export(&buf, r); err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(buf.String())
	if got != "[]" {
		t.Errorf("got: %s, want: []", got)
	}
}
