package export

import (
	"fmt"
	"io"
	"strings"

	"github.com/mdjarv/db/internal/db"
)

type sqlExporter struct {
	opts Options
}

func (e *sqlExporter) Export(w io.Writer, result *db.Result) error {
	tableName := e.opts.TableName
	if tableName == "" {
		tableName = "table_name"
	}

	colNames := make([]string, len(result.Columns))
	for i, col := range result.Columns {
		colNames[i] = col.Name
	}
	colList := strings.Join(colNames, ", ")

	var batch []string
	for result.Rows.Next() {
		vals, err := result.Rows.Values()
		if err != nil {
			return err
		}
		parts := make([]string, len(vals))
		for i, v := range vals {
			parts[i] = sqlLiteral(v)
		}
		batch = append(batch, "("+strings.Join(parts, ", ")+")")

		if len(batch) >= 100 {
			if err := writeBatch(w, tableName, colList, batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}
	if err := result.Rows.Err(); err != nil {
		return err
	}

	if len(batch) > 0 {
		return writeBatch(w, tableName, colList, batch)
	}
	return nil
}

func writeBatch(w io.Writer, tableName, colList string, rows []string) error {
	_, err := fmt.Fprintf(w, "INSERT INTO %s (%s) VALUES\n  %s;\n",
		tableName, colList, strings.Join(rows, ",\n  "))
	return err
}

func sqlLiteral(v any) string {
	if v == nil {
		return "NULL"
	}
	switch val := v.(type) {
	case bool:
		if val {
			return "TRUE"
		}
		return "FALSE"
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
	case string:
		return "'" + strings.ReplaceAll(val, "'", "''") + "'"
	case []byte:
		return "'" + strings.ReplaceAll(string(val), "'", "''") + "'"
	default:
		s := formatValue(v, "NULL")
		return "'" + strings.ReplaceAll(s, "'", "''") + "'"
	}
}
