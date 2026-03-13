package editdialog

import (
	"reflect"
	"testing"
)

func TestParseArrayEmpty(t *testing.T) {
	got := parseArray("{}")
	if got == nil || len(got) != 0 {
		t.Errorf("parseArray({}) = %v, want []", got)
	}
}

func TestParseArraySimple(t *testing.T) {
	got := parseArray("{a,b,c}")
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseArray({a,b,c}) = %v, want %v", got, want)
	}
}

func TestParseArrayQuoted(t *testing.T) {
	got := parseArray(`{"a,b","c"}`)
	want := []string{"a,b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseArray quoted = %v, want %v", got, want)
	}
}

func TestParseArrayNULL(t *testing.T) {
	got := parseArray("{a,NULL,b}")
	want := []string{"a", "NULL", "b"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseArray NULL = %v, want %v", got, want)
	}
}

func TestParseArrayEscapedQuotes(t *testing.T) {
	got := parseArray(`{"he said \"hi\""}`)
	want := []string{`he said "hi"`}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseArray escaped = %v, want %v", got, want)
	}
}

func TestFormatArrayEmpty(t *testing.T) {
	got := formatArray([]string{})
	if got != "{}" {
		t.Errorf("formatArray([]) = %q, want {}", got)
	}
}

func TestFormatArraySimple(t *testing.T) {
	got := formatArray([]string{"a", "b", "c"})
	if got != "{a,b,c}" {
		t.Errorf("formatArray simple = %q, want {a,b,c}", got)
	}
}

func TestFormatArrayNULL(t *testing.T) {
	got := formatArray([]string{"a", "NULL", "b"})
	if got != "{a,NULL,b}" {
		t.Errorf("formatArray NULL = %q, want {a,NULL,b}", got)
	}
}

func TestRoundTrip(t *testing.T) {
	cases := []string{
		"{}",
		"{a,b,c}",
		`{"a,b",c}`,
		"{a,NULL,b}",
		`{"he said \"hi\""}`,
	}
	for _, s := range cases {
		elems := parseArray(s)
		got := formatArray(elems)
		if got != s {
			t.Errorf("round-trip %q: got %q", s, got)
		}
	}
}

func TestParseArrayInvalid(t *testing.T) {
	got := parseArray("not an array")
	if got != nil {
		t.Errorf("parseArray(invalid) = %v, want nil", got)
	}
}
