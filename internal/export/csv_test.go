package export

import (
	"bytes"
	"testing"

	"github.com/mdjarv/db/internal/db"
)

func TestCSVBasic(t *testing.T) {
	r := mockResult(
		[]db.Column{{Name: "id", TypeName: "int4"}, {Name: "name", TypeName: "text"}},
		[][]any{{1, "alice"}, {2, "bob"}},
	)
	var buf bytes.Buffer
	exp := NewExporter(FormatCSV, Options{})
	if err := exp.Export(&buf, r); err != nil {
		t.Fatal(err)
	}
	want := "id,name\n1,alice\n2,bob\n"
	if buf.String() != want {
		t.Errorf("got:\n%s\nwant:\n%s", buf.String(), want)
	}
}

func TestCSVTabDelimiter(t *testing.T) {
	r := mockResult(
		[]db.Column{{Name: "a", TypeName: "text"}, {Name: "b", TypeName: "text"}},
		[][]any{{"x", "y"}},
	)
	var buf bytes.Buffer
	exp := NewExporter(FormatCSV, Options{Delimiter: '\t'})
	if err := exp.Export(&buf, r); err != nil {
		t.Fatal(err)
	}
	want := "a\tb\nx\ty\n"
	if buf.String() != want {
		t.Errorf("got:\n%s\nwant:\n%s", buf.String(), want)
	}
}

func TestCSVNoHeader(t *testing.T) {
	r := mockResult(
		[]db.Column{{Name: "a", TypeName: "text"}},
		[][]any{{"val"}},
	)
	var buf bytes.Buffer
	exp := NewExporter(FormatCSV, Options{NoHeader: true})
	if err := exp.Export(&buf, r); err != nil {
		t.Fatal(err)
	}
	want := "val\n"
	if buf.String() != want {
		t.Errorf("got:\n%s\nwant:\n%s", buf.String(), want)
	}
}

func TestCSVNullHandling(t *testing.T) {
	r := mockResult(
		[]db.Column{{Name: "a", TypeName: "text"}},
		[][]any{{nil}},
	)
	var buf bytes.Buffer
	exp := NewExporter(FormatCSV, Options{NullString: "\\N"})
	if err := exp.Export(&buf, r); err != nil {
		t.Fatal(err)
	}
	want := "a\n\\N\n"
	if buf.String() != want {
		t.Errorf("got:\n%s\nwant:\n%s", buf.String(), want)
	}
}

func TestCSVQuoting(t *testing.T) {
	r := mockResult(
		[]db.Column{{Name: "v", TypeName: "text"}},
		[][]any{{"has,comma"}, {`has"quote`}},
	)
	var buf bytes.Buffer
	exp := NewExporter(FormatCSV, Options{})
	if err := exp.Export(&buf, r); err != nil {
		t.Fatal(err)
	}
	want := "v\n\"has,comma\"\n\"has\"\"quote\"\n"
	if buf.String() != want {
		t.Errorf("got:\n%s\nwant:\n%s", buf.String(), want)
	}
}
