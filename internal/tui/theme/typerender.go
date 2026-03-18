package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TypeRenderer styles type names and cell values for a PostgreSQL type.
type TypeRenderer interface {
	// BaseStyle returns the base lipgloss.Style for this type (used by
	// RenderType and by arrayRenderer to derive fainted variants).
	BaseStyle() lipgloss.Style
	RenderType(name string) string
	RenderValue(text string) string
}

// Field describes a composite type field.
type Field struct {
	Name     string
	TypeName string
}

// ForType returns a TypeRenderer for the given PostgreSQL type name.
// Returns nil for unknown types.
func ForType(typeName string) TypeRenderer {
	lower := strings.ToLower(typeName)
	if strings.HasSuffix(lower, "[]") {
		base := typeName[:len(typeName)-2]
		elem := ForType(base)
		if elem == nil {
			elem = sString
		}
		return &arrayRenderer{element: elem}
	}
	base := lower
	if idx := strings.IndexByte(base, '('); idx != -1 {
		base = base[:idx]
	}
	switch base {
	case "serial", "bigserial", "smallserial",
		"integer", "int", "int2", "int4", "int8",
		"bigint", "smallint",
		"numeric", "decimal",
		"real", "float4",
		"double precision", "float8",
		"money", "oid":
		return sNumeric
	case "boolean", "bool":
		return sBool
	case "varchar", "character varying",
		"char", "character",
		"text", "name",
		"citext", "bpchar":
		return sString
	case "date", "time", "timetz",
		"timestamp", "timestamptz",
		"timestamp without time zone",
		"timestamp with time zone",
		"time without time zone",
		"time with time zone",
		"interval":
		return sDate
	case "uuid":
		return sUUID
	case "json", "jsonb":
		return sJSON
	}
	return nil
}

// ForComposite returns a TypeRenderer for a composite type.
func ForComposite(fields []Field) TypeRenderer {
	renderers := make([]TypeRenderer, len(fields))
	for i, f := range fields {
		renderers[i] = ForType(f.TypeName)
	}
	return &compositeRenderer{fields: renderers}
}

var (
	sNumeric = &numericRenderer{}
	sBool    = &boolRenderer{}
	sString  = &stringRenderer{}
	sDate    = &dateRenderer{}
	sUUID    = &uuidRenderer{}
	sJSON    = &jsonRenderer{}
)

// numericRenderer styles numeric types.
type numericRenderer struct{}

func (r *numericRenderer) BaseStyle() lipgloss.Style      { return Current().Styles.DataNumber }
func (r *numericRenderer) RenderType(name string) string  { return r.BaseStyle().Render(name) }
func (r *numericRenderer) RenderValue(text string) string { return r.BaseStyle().Render(text) }

// boolRenderer styles boolean types.
type boolRenderer struct{}

func (r *boolRenderer) BaseStyle() lipgloss.Style     { return Current().Styles.DataBoolTrue }
func (r *boolRenderer) RenderType(name string) string { return r.BaseStyle().Render(name) }
func (r *boolRenderer) RenderValue(text string) string {
	if strings.TrimSpace(text) == "true" {
		return Current().Styles.DataBoolTrue.Render(text)
	}
	return Current().Styles.DataBoolFalse.Render(text)
}

// stringRenderer styles string/text types.
type stringRenderer struct{}

func (r *stringRenderer) BaseStyle() lipgloss.Style      { return Current().Styles.DataString }
func (r *stringRenderer) RenderType(name string) string  { return r.BaseStyle().Render(name) }
func (r *stringRenderer) RenderValue(text string) string { return r.BaseStyle().Render(text) }

// dateRenderer styles date/time types with digit-group highlighting.
type dateRenderer struct{}

func (r *dateRenderer) BaseStyle() lipgloss.Style     { return Current().Styles.DataDate }
func (r *dateRenderer) RenderType(name string) string { return r.BaseStyle().Render(name) }
func (r *dateRenderer) RenderValue(text string) string {
	s := Current().Styles
	return styledGroups(text, isDigit, s.DataDate, s.Dim)
}

// uuidRenderer styles UUID types with hex-group highlighting.
type uuidRenderer struct{}

func (r *uuidRenderer) BaseStyle() lipgloss.Style     { return Current().Styles.DataUUID }
func (r *uuidRenderer) RenderType(name string) string { return r.BaseStyle().Render(name) }
func (r *uuidRenderer) RenderValue(text string) string {
	s := Current().Styles
	return styledGroups(text, isHexDigit, s.DataUUID, s.Dim)
}

// jsonRenderer styles JSON types.
type jsonRenderer struct{}

func (r *jsonRenderer) BaseStyle() lipgloss.Style      { return Current().Styles.Keyword }
func (r *jsonRenderer) RenderType(name string) string  { return r.BaseStyle().Render(name) }
func (r *jsonRenderer) RenderValue(text string) string { return Current().Styles.Dim.Render(text) }

// arrayRenderer styles array types, delegating element rendering.
type arrayRenderer struct {
	element TypeRenderer
}

func (r *arrayRenderer) BaseStyle() lipgloss.Style {
	return r.element.BaseStyle().Faint(true)
}

func (r *arrayRenderer) RenderType(name string) string {
	return r.BaseStyle().Render(name)
}

func (r *arrayRenderer) RenderValue(text string) string {
	s := Current().Styles
	var sb strings.Builder
	var elem strings.Builder
	depth := 0
	inQuotes := false

	for _, ch := range text {
		switch {
		case ch == '"':
			if depth == 1 {
				inQuotes = !inQuotes
			}
			elem.WriteRune(ch)
		case !inQuotes && ch == '{':
			depth++
			if depth == 1 {
				sb.WriteString(s.Dim.Render(string(ch)))
			} else {
				elem.WriteRune(ch)
			}
		case !inQuotes && ch == '}':
			depth--
			if depth == 0 {
				if elem.Len() > 0 {
					sb.WriteString(r.element.RenderValue(elem.String()))
					elem.Reset()
				}
				sb.WriteString(s.Dim.Render(string(ch)))
			} else {
				elem.WriteRune(ch)
			}
		case !inQuotes && ch == ',' && depth == 1:
			if elem.Len() > 0 {
				sb.WriteString(r.element.RenderValue(elem.String()))
				elem.Reset()
			}
			sb.WriteString(s.Dim.Render(string(ch)))
		default:
			if depth >= 1 {
				elem.WriteRune(ch)
			}
		}
	}
	if elem.Len() > 0 {
		sb.WriteString(r.element.RenderValue(elem.String()))
	}
	return sb.String()
}

// compositeRenderer styles composite/record types.
type compositeRenderer struct {
	fields []TypeRenderer
}

func (r *compositeRenderer) BaseStyle() lipgloss.Style     { return Current().Styles.Dim }
func (r *compositeRenderer) RenderType(name string) string { return r.BaseStyle().Render(name) }

func (r *compositeRenderer) RenderValue(text string) string {
	if len(r.fields) > 0 {
		return r.renderPositional(text)
	}
	return r.renderFSM(text)
}

func (r *compositeRenderer) renderPositional(text string) string {
	s := Current().Styles
	var sb strings.Builder
	var elem strings.Builder
	fieldIdx := 0
	inQuotes := false

	flushElem := func() {
		if elem.Len() == 0 {
			return
		}
		content := elem.String()
		elem.Reset()
		if fieldIdx < len(r.fields) && r.fields[fieldIdx] != nil {
			sb.WriteString(r.fields[fieldIdx].RenderValue(content))
			return
		}
		if inQuotes {
			sb.WriteString(s.DataString.Render(content))
		} else {
			sb.WriteString(s.DataNumber.Render(content))
		}
	}

	for _, ch := range text {
		switch {
		case ch == '"':
			flushElem()
			sb.WriteString(s.Dim.Render(string(ch)))
			inQuotes = !inQuotes
		case !inQuotes && (ch == '(' || ch == ')'):
			flushElem()
			sb.WriteString(s.Dim.Render(string(ch)))
		case !inQuotes && ch == ',':
			flushElem()
			sb.WriteString(s.Dim.Render(string(ch)))
			fieldIdx++
		default:
			elem.WriteRune(ch)
		}
	}
	flushElem()
	return sb.String()
}

func (r *compositeRenderer) renderFSM(text string) string {
	s := Current().Styles

	type kind int
	const (
		kStruct kind = iota
		kString
		kValue
	)

	var sb strings.Builder
	var buf strings.Builder
	cur := kStruct
	inQuotes := false

	flush := func(k kind) {
		if buf.Len() == 0 {
			return
		}
		switch k {
		case kStruct:
			sb.WriteString(s.Dim.Render(buf.String()))
		case kString:
			sb.WriteString(s.DataString.Render(buf.String()))
		case kValue:
			sb.WriteString(s.DataNumber.Render(buf.String()))
		}
		buf.Reset()
	}

	for _, ch := range text {
		var k kind
		switch ch {
		case '"':
			k = kStruct
			if inQuotes {
				flush(kString)
			}
			inQuotes = !inQuotes
		case '(', ')', ',':
			if inQuotes {
				k = kString
			} else {
				k = kStruct
			}
		default:
			if inQuotes {
				k = kString
			} else {
				k = kValue
			}
		}

		if k != cur {
			flush(cur)
			cur = k
		}
		buf.WriteRune(ch)
	}
	flush(cur)
	return sb.String()
}

// styledGroups renders text by alternating two styles based on a character predicate.
func styledGroups(text string, groupA func(rune) bool, styleA, styleB lipgloss.Style) string {
	if len(text) == 0 {
		return text
	}
	var sb strings.Builder
	var buf strings.Builder
	inA := groupA(rune(text[0]))
	for _, ch := range text {
		a := groupA(ch)
		if a != inA {
			if inA {
				sb.WriteString(styleA.Render(buf.String()))
			} else {
				sb.WriteString(styleB.Render(buf.String()))
			}
			buf.Reset()
			inA = a
		}
		buf.WriteRune(ch)
	}
	if buf.Len() > 0 {
		if inA {
			sb.WriteString(styleA.Render(buf.String()))
		} else {
			sb.WriteString(styleB.Render(buf.String()))
		}
	}
	return sb.String()
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isHexDigit(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}
