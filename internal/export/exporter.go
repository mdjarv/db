// Package export provides formatters for query results (CSV, JSON, SQL, table).
package export

import (
	"fmt"
	"io"
	"time"

	"github.com/mdjarv/db/internal/db"
)

// Format selects the output format.
type Format int

// Output formats.
const (
	FormatTable Format = iota
	FormatCSV
	FormatJSON
	FormatSQL
)

// Options configures exporter behavior.
type Options struct {
	Delimiter   rune   // CSV: comma, tab, pipe
	Pretty      bool   // JSON: indented
	JSONLines   bool   // JSON: one object per line, no array
	TableName   string // SQL: target table name
	NoHeader    bool   // CSV/Table: omit header row
	NullString  string // representation of NULL values
	MaxColWidth int    // Table: max column width before truncation
}

// Exporter writes a query result to a writer.
type Exporter interface {
	Export(w io.Writer, result *db.Result) error
}

// NewExporter returns an exporter for the given format.
func NewExporter(format Format, opts Options) Exporter {
	switch format {
	case FormatCSV:
		return &csvExporter{opts: opts}
	case FormatJSON:
		return &jsonExporter{opts: opts}
	case FormatSQL:
		return &sqlExporter{opts: opts}
	case FormatTable:
		return &tableExporter{opts: opts}
	default:
		return &tableExporter{opts: opts}
	}
}

// formatValue converts a value to its string representation.
func formatValue(v any, nullStr string) string {
	if v == nil {
		return nullStr
	}
	switch val := v.(type) {
	case bool:
		if val {
			return "true"
		}
		return "false"
	case int:
		return fmt.Sprintf("%d", val)
	case int32:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	case float32:
		return fmt.Sprintf("%g", val)
	case float64:
		return fmt.Sprintf("%g", val)
	case []byte:
		return string(val)
	case string:
		return val
	case time.Time:
		return val.Format(time.RFC3339)
	case fmt.Stringer:
		return val.String()
	default:
		return fmt.Sprintf("%v", val)
	}
}
