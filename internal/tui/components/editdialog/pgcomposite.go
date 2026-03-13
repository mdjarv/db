package editdialog

import "strings"

// parseComposite parses a PostgreSQL composite literal like "(val1,val2)" into
// a map keyed by field names.
func parseComposite(s string, fieldNames []string) map[string]string {
	s = strings.TrimSpace(s)
	if len(s) < 2 || s[0] != '(' || s[len(s)-1] != ')' {
		// return empty map with field names
		m := make(map[string]string, len(fieldNames))
		for _, n := range fieldNames {
			m[n] = ""
		}
		return m
	}
	inner := s[1 : len(s)-1]

	vals := splitCompositeFields(inner)
	result := make(map[string]string, len(fieldNames))
	for i, name := range fieldNames {
		if i < len(vals) {
			result[name] = vals[i]
		} else {
			result[name] = ""
		}
	}
	return result
}

func splitCompositeFields(s string) []string {
	var fields []string
	var buf strings.Builder
	inQuote := false
	escaped := false

	for i := 0; i < len(s); i++ {
		ch := s[i]
		if escaped {
			buf.WriteByte(ch)
			escaped = false
			continue
		}
		if ch == '\\' && inQuote {
			escaped = true
			continue
		}
		if ch == '"' {
			inQuote = !inQuote
			continue
		}
		if ch == ',' && !inQuote {
			fields = append(fields, buf.String())
			buf.Reset()
			continue
		}
		buf.WriteByte(ch)
	}
	fields = append(fields, buf.String())
	return fields
}

// formatComposite formats field values back into a PostgreSQL composite literal.
func formatComposite(fields []CompositeField, values map[string]string) string {
	var sb strings.Builder
	sb.WriteByte('(')
	for i, f := range fields {
		if i > 0 {
			sb.WriteByte(',')
		}
		v := values[f.Name]
		if needsCompositeQuoting(v) {
			sb.WriteByte('"')
			sb.WriteString(escapeCompositeVal(v))
			sb.WriteByte('"')
		} else {
			sb.WriteString(v)
		}
	}
	sb.WriteByte(')')
	return sb.String()
}

func needsCompositeQuoting(s string) bool {
	if s == "" {
		return true
	}
	for _, ch := range s {
		if ch == ',' || ch == '"' || ch == '\\' || ch == '(' || ch == ')' || ch == ' ' {
			return true
		}
	}
	return false
}

func escapeCompositeVal(s string) string {
	var sb strings.Builder
	for _, ch := range s {
		if ch == '"' || ch == '\\' {
			sb.WriteByte('\\')
		}
		sb.WriteRune(ch)
	}
	return sb.String()
}
