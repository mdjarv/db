package editdialog

import "strings"

// parseArray parses a PostgreSQL array literal like "{a,b,c}" into elements.
func parseArray(s string) []string {
	s = strings.TrimSpace(s)
	if len(s) < 2 || s[0] != '{' || s[len(s)-1] != '}' {
		return nil
	}
	inner := s[1 : len(s)-1]
	if inner == "" {
		return []string{}
	}

	var elems []string
	var buf strings.Builder
	inQuote := false
	escaped := false

	for i := 0; i < len(inner); i++ {
		ch := inner[i]
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
			elems = append(elems, buf.String())
			buf.Reset()
			continue
		}
		buf.WriteByte(ch)
	}
	elems = append(elems, buf.String())
	return elems
}

// formatArray formats elements into a PostgreSQL array literal.
func formatArray(elems []string) string {
	var sb strings.Builder
	sb.WriteByte('{')
	for i, e := range elems {
		if i > 0 {
			sb.WriteByte(',')
		}
		if e == "NULL" || !needsQuoting(e) {
			sb.WriteString(e)
		} else {
			sb.WriteByte('"')
			sb.WriteString(escapeArrayElem(e))
			sb.WriteByte('"')
		}
	}
	sb.WriteByte('}')
	return sb.String()
}

func needsQuoting(s string) bool {
	if s == "" {
		return true
	}
	for _, ch := range s {
		if ch == ',' || ch == '"' || ch == '\\' || ch == '{' || ch == '}' || ch == ' ' {
			return true
		}
	}
	return false
}

func escapeArrayElem(s string) string {
	var sb strings.Builder
	for _, ch := range s {
		if ch == '"' || ch == '\\' {
			sb.WriteByte('\\')
		}
		sb.WriteRune(ch)
	}
	return sb.String()
}
