package export

import (
	"encoding/json"
	"io"

	"github.com/mdjarv/db/internal/db"
)

type jsonExporter struct {
	opts Options
}

func (e *jsonExporter) Export(w io.Writer, result *db.Result) error {
	colNames := make([]string, len(result.Columns))
	for i, col := range result.Columns {
		colNames[i] = col.Name
	}

	enc := json.NewEncoder(w)
	if e.opts.Pretty {
		enc.SetIndent("", "  ")
	}

	if e.opts.JSONLines {
		return e.exportLines(enc, result, colNames)
	}
	return e.exportArray(w, enc, result, colNames)
}

func (e *jsonExporter) exportLines(enc *json.Encoder, result *db.Result, colNames []string) error {
	for result.Rows.Next() {
		vals, err := result.Rows.Values()
		if err != nil {
			return err
		}
		obj := e.buildObject(colNames, vals)
		if err := enc.Encode(obj); err != nil {
			return err
		}
	}
	return result.Rows.Err()
}

func (e *jsonExporter) exportArray(w io.Writer, enc *json.Encoder, result *db.Result, colNames []string) error {
	if _, err := io.WriteString(w, "["); err != nil {
		return err
	}

	first := true
	for result.Rows.Next() {
		vals, err := result.Rows.Values()
		if err != nil {
			return err
		}
		if !first {
			if _, err := io.WriteString(w, ","); err != nil {
				return err
			}
		}
		first = false

		if e.opts.Pretty {
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}

		obj := e.buildObject(colNames, vals)
		if err := enc.Encode(obj); err != nil {
			return err
		}
	}
	if err := result.Rows.Err(); err != nil {
		return err
	}

	if e.opts.Pretty && !first {
		if _, err := io.WriteString(w, "\n"); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, "]\n")
	return err
}

// buildObject creates an ordered map from column names to values.
// Uses orderedMap to preserve column order in JSON output.
func (e *jsonExporter) buildObject(colNames []string, vals []any) *orderedMap {
	om := &orderedMap{
		keys: colNames,
		vals: make(map[string]any, len(colNames)),
	}
	for i, name := range colNames {
		var v any
		if i < len(vals) {
			v = vals[i]
		}
		om.vals[name] = jsonValue(v)
	}
	return om
}

// jsonValue converts a Go value to a JSON-compatible type.
func jsonValue(v any) any {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case bool:
		return val
	case int:
		return val
	case int32:
		return val
	case int64:
		return val
	case float32:
		return float64(val)
	case float64:
		return val
	case []byte:
		return string(val)
	case string:
		return val
	default:
		return formatValue(v, "")
	}
}

// orderedMap preserves key insertion order during JSON marshaling.
type orderedMap struct {
	keys []string
	vals map[string]any
}

func (om *orderedMap) MarshalJSON() ([]byte, error) {
	buf := []byte{'{'}
	for i, key := range om.keys {
		if i > 0 {
			buf = append(buf, ',')
		}
		kb, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		buf = append(buf, kb...)
		buf = append(buf, ':')
		vb, err := json.Marshal(om.vals[key])
		if err != nil {
			return nil, err
		}
		buf = append(buf, vb...)
	}
	buf = append(buf, '}')
	return buf, nil
}
