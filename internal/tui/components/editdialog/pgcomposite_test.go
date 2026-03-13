package editdialog

import (
	"testing"
)

func TestParseCompositeSimple(t *testing.T) {
	fields := []string{"a", "b", "c"}
	got := parseComposite("(1,2,3)", fields)
	if got["a"] != "1" || got["b"] != "2" || got["c"] != "3" {
		t.Errorf("parseComposite simple = %v", got)
	}
}

func TestParseCompositeQuoted(t *testing.T) {
	fields := []string{"name", "val"}
	got := parseComposite(`("hello, world",42)`, fields)
	if got["name"] != "hello, world" || got["val"] != "42" {
		t.Errorf("parseComposite quoted = %v", got)
	}
}

func TestParseCompositeEmpty(t *testing.T) {
	fields := []string{"a", "b"}
	got := parseComposite("(,)", fields)
	if got["a"] != "" || got["b"] != "" {
		t.Errorf("parseComposite empty = %v", got)
	}
}

func TestParseCompositeInvalid(t *testing.T) {
	fields := []string{"a", "b"}
	got := parseComposite("not composite", fields)
	if got["a"] != "" || got["b"] != "" {
		t.Errorf("parseComposite invalid = %v", got)
	}
}

func TestFormatCompositeSimple(t *testing.T) {
	fields := []CompositeField{
		{Name: "a", TypeName: "int4"},
		{Name: "b", TypeName: "text"},
	}
	vals := map[string]string{"a": "1", "b": "hello"}
	got := formatComposite(fields, vals)
	if got != `(1,hello)` {
		t.Errorf("formatComposite simple = %q", got)
	}
}

func TestFormatCompositeQuoted(t *testing.T) {
	fields := []CompositeField{
		{Name: "name", TypeName: "text"},
		{Name: "val", TypeName: "int4"},
	}
	vals := map[string]string{"name": "hello, world", "val": "42"}
	got := formatComposite(fields, vals)
	if got != `("hello, world",42)` {
		t.Errorf("formatComposite quoted = %q", got)
	}
}

func TestCompositeRoundTrip(t *testing.T) {
	fields := []CompositeField{
		{Name: "name", TypeName: "text"},
		{Name: "val", TypeName: "int4"},
	}
	fieldNames := []string{"name", "val"}

	cases := []string{
		`("hello, world",42)`,
		`("",42)`,
	}
	for _, s := range cases {
		vals := parseComposite(s, fieldNames)
		got := formatComposite(fields, vals)
		if got != s {
			t.Errorf("composite round-trip %q: got %q", s, got)
		}
	}
}
