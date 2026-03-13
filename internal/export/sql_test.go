package export

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mdjarv/db/internal/db"
)

func TestSQLBasic(t *testing.T) {
	r := mockResult(
		[]db.Column{{Name: "id", TypeName: "int4"}, {Name: "name", TypeName: "text"}},
		[][]any{{int64(1), "alice"}, {int64(2), "bob"}},
	)
	var buf bytes.Buffer
	exp := NewExporter(FormatSQL, Options{TableName: "users"})
	if err := exp.Export(&buf, r); err != nil {
		t.Fatal(err)
	}
	want := "INSERT INTO users (id, name) VALUES\n  (1, 'alice'),\n  (2, 'bob');\n"
	if buf.String() != want {
		t.Errorf("got:\n%s\nwant:\n%s", buf.String(), want)
	}
}

func TestSQLNulls(t *testing.T) {
	r := mockResult(
		[]db.Column{{Name: "a", TypeName: "text"}},
		[][]any{{nil}},
	)
	var buf bytes.Buffer
	exp := NewExporter(FormatSQL, Options{TableName: "t"})
	if err := exp.Export(&buf, r); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "NULL") {
		t.Error("expected NULL in output")
	}
}

func TestSQLStringEscaping(t *testing.T) {
	r := mockResult(
		[]db.Column{{Name: "v", TypeName: "text"}},
		[][]any{{"it's a test"}},
	)
	var buf bytes.Buffer
	exp := NewExporter(FormatSQL, Options{TableName: "t"})
	if err := exp.Export(&buf, r); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "''") {
		t.Errorf("expected escaped single quote, got:\n%s", buf.String())
	}
}

func TestSQLBooleans(t *testing.T) {
	r := mockResult(
		[]db.Column{{Name: "b", TypeName: "bool"}},
		[][]any{{true}, {false}},
	)
	var buf bytes.Buffer
	exp := NewExporter(FormatSQL, Options{TableName: "t"})
	if err := exp.Export(&buf, r); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "TRUE") || !strings.Contains(buf.String(), "FALSE") {
		t.Errorf("expected TRUE/FALSE, got:\n%s", buf.String())
	}
}

func TestSQLBatchMultiRow(t *testing.T) {
	// 150 rows should produce 2 INSERT statements (100 + 50).
	data := make([][]any, 150)
	for i := range data {
		data[i] = []any{int64(i)}
	}
	r := mockResult(
		[]db.Column{{Name: "id", TypeName: "int4"}},
		data,
	)
	var buf bytes.Buffer
	exp := NewExporter(FormatSQL, Options{TableName: "t"})
	if err := exp.Export(&buf, r); err != nil {
		t.Fatal(err)
	}
	count := strings.Count(buf.String(), "INSERT INTO")
	if count != 2 {
		t.Errorf("expected 2 INSERT statements, got %d", count)
	}
}

func TestSQLEmpty(t *testing.T) {
	r := mockResult(
		[]db.Column{{Name: "a", TypeName: "text"}},
		nil,
	)
	var buf bytes.Buffer
	exp := NewExporter(FormatSQL, Options{TableName: "t"})
	if err := exp.Export(&buf, r); err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output for no rows, got: %s", buf.String())
	}
}
