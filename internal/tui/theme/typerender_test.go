package theme

import (
	"strings"
	"testing"
)

func TestForTypeNumeric(t *testing.T) {
	for _, ty := range []string{"integer", "bigint", "serial", "numeric", "real", "double precision", "money", "int2", "int4", "int8", "smallint", "decimal", "float4", "float8", "oid"} {
		r := ForType(ty)
		if r == nil {
			t.Errorf("ForType(%q) = nil", ty)
			continue
		}
		if _, ok := r.(*numericRenderer); !ok {
			t.Errorf("ForType(%q) = %T, want *numericRenderer", ty, r)
		}
	}
}

func TestForTypeBool(t *testing.T) {
	for _, ty := range []string{"boolean", "bool"} {
		r := ForType(ty)
		if _, ok := r.(*boolRenderer); !ok {
			t.Errorf("ForType(%q) = %T, want *boolRenderer", ty, r)
		}
	}
}

func TestForTypeString(t *testing.T) {
	for _, ty := range []string{"varchar", "text", "name", "char", "citext", "bpchar", "varchar(100)", "character varying", "character"} {
		r := ForType(ty)
		if _, ok := r.(*stringRenderer); !ok {
			t.Errorf("ForType(%q) = %T, want *stringRenderer", ty, r)
		}
	}
}

func TestForTypeDate(t *testing.T) {
	for _, ty := range []string{"date", "timestamp", "timestamptz", "time", "timetz", "interval"} {
		r := ForType(ty)
		if _, ok := r.(*dateRenderer); !ok {
			t.Errorf("ForType(%q) = %T, want *dateRenderer", ty, r)
		}
	}
}

func TestForTypeUUID(t *testing.T) {
	r := ForType("uuid")
	if _, ok := r.(*uuidRenderer); !ok {
		t.Errorf("ForType(uuid) = %T, want *uuidRenderer", r)
	}
}

func TestForTypeJSON(t *testing.T) {
	for _, ty := range []string{"json", "jsonb"} {
		r := ForType(ty)
		if _, ok := r.(*jsonRenderer); !ok {
			t.Errorf("ForType(%q) = %T, want *jsonRenderer", ty, r)
		}
	}
}

func TestForTypeNil(t *testing.T) {
	for _, ty := range []string{"hstore", "geometry", "custom_type", ""} {
		if r := ForType(ty); r != nil {
			t.Errorf("ForType(%q) = %T, want nil", ty, r)
		}
	}
}

func TestForTypeArray(t *testing.T) {
	r := ForType("integer[]")
	ar, ok := r.(*arrayRenderer)
	if !ok {
		t.Fatalf("ForType(integer[]) = %T, want *arrayRenderer", r)
	}
	if _, ok := ar.element.(*numericRenderer); !ok {
		t.Errorf("element = %T, want *numericRenderer", ar.element)
	}
}

func TestForTypeArrayNested(t *testing.T) {
	r := ForType("integer[][]")
	ar, ok := r.(*arrayRenderer)
	if !ok {
		t.Fatalf("ForType(integer[][]) = %T, want *arrayRenderer", r)
	}
	inner, ok := ar.element.(*arrayRenderer)
	if !ok {
		t.Fatalf("element = %T, want *arrayRenderer", ar.element)
	}
	if _, ok := inner.element.(*numericRenderer); !ok {
		t.Errorf("inner element = %T, want *numericRenderer", inner.element)
	}
}

func TestForTypeArrayUnknownBase(t *testing.T) {
	r := ForType("hstore[]")
	ar, ok := r.(*arrayRenderer)
	if !ok {
		t.Fatalf("ForType(hstore[]) = %T, want *arrayRenderer", r)
	}
	if _, ok := ar.element.(*stringRenderer); !ok {
		t.Errorf("unknown base element = %T, want *stringRenderer (fallback)", ar.element)
	}
}

func TestForComposite(t *testing.T) {
	fields := []Field{
		{Name: "id", TypeName: "integer"},
		{Name: "name", TypeName: "text"},
	}
	r := ForComposite(fields)
	cr, ok := r.(*compositeRenderer)
	if !ok {
		t.Fatalf("ForComposite = %T, want *compositeRenderer", r)
	}
	if len(cr.fields) != 2 {
		t.Fatalf("fields len = %d, want 2", len(cr.fields))
	}
	if _, ok := cr.fields[0].(*numericRenderer); !ok {
		t.Errorf("field[0] = %T, want *numericRenderer", cr.fields[0])
	}
	if _, ok := cr.fields[1].(*stringRenderer); !ok {
		t.Errorf("field[1] = %T, want *stringRenderer", cr.fields[1])
	}
}

func TestForCompositeUnknownField(t *testing.T) {
	fields := []Field{
		{Name: "data", TypeName: "hstore"},
	}
	r := ForComposite(fields)
	cr := r.(*compositeRenderer)
	if cr.fields[0] != nil {
		t.Errorf("unknown field type should be nil, got %T", cr.fields[0])
	}
}

func TestBoolRenderValue(t *testing.T) {
	r := ForType("boolean")
	tr := r.RenderValue("true ")
	fr := r.RenderValue("false")
	if !strings.Contains(tr, "true") {
		t.Errorf("bool true render missing text: %q", tr)
	}
	if !strings.Contains(fr, "false") {
		t.Errorf("bool false render missing text: %q", fr)
	}
}

func TestDateRenderValue(t *testing.T) {
	r := ForType("timestamp")
	result := r.RenderValue("2024-01-15 10:30:00")
	if !strings.Contains(result, "2024") {
		t.Errorf("date render missing digits: %q", result)
	}
}

func TestUUIDRenderValue(t *testing.T) {
	r := ForType("uuid")
	result := r.RenderValue("550e8400-e29b-41d4-a716-446655440000")
	if !strings.Contains(result, "550e8400") {
		t.Errorf("uuid render missing hex: %q", result)
	}
}

func TestArrayRenderValue(t *testing.T) {
	r := ForType("integer[]")
	result := r.RenderValue("{1,2,3}")
	if !strings.Contains(result, "1") || !strings.Contains(result, "2") || !strings.Contains(result, "3") {
		t.Errorf("array render missing elements: %q", result)
	}
}

func TestArrayRenderValueEmpty(t *testing.T) {
	r := ForType("text[]")
	result := r.RenderValue("{}")
	if !strings.Contains(result, "{") || !strings.Contains(result, "}") {
		t.Errorf("empty array render: %q", result)
	}
}

func TestCompositePositionalRenderValue(t *testing.T) {
	fields := []Field{
		{Name: "name", TypeName: "text"},
		{Name: "age", TypeName: "integer"},
	}
	r := ForComposite(fields)
	result := r.RenderValue(`("hello",42)`)
	if !strings.Contains(result, "hello") || !strings.Contains(result, "42") {
		t.Errorf("composite positional missing content: %q", result)
	}
}

func TestCompositeFSMRenderValue(t *testing.T) {
	r := ForComposite(nil)
	result := r.RenderValue(`("hello",123)`)
	if !strings.Contains(result, "hello") || !strings.Contains(result, "123") {
		t.Errorf("composite FSM missing content: %q", result)
	}
}

func TestRenderTypeNonEmpty(t *testing.T) {
	tests := []struct {
		typeName string
		input    string
	}{
		{"integer", "integer"},
		{"boolean", "boolean"},
		{"text", "text"},
		{"timestamp", "timestamp"},
		{"uuid", "uuid"},
		{"json", "json"},
		{"integer[]", "integer[]"},
	}
	for _, tt := range tests {
		r := ForType(tt.typeName)
		if r == nil {
			t.Errorf("ForType(%q) = nil", tt.typeName)
			continue
		}
		result := r.RenderType(tt.input)
		if !strings.Contains(result, tt.input) {
			t.Errorf("RenderType(%q) missing input text: %q", tt.typeName, result)
		}
	}
}
