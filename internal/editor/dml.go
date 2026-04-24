package editor

import (
	"fmt"
	"strings"

	"github.com/mdjarv/db/internal/db"
)

// DMLResult holds a generated SQL statement and its parameters.
type DMLResult struct {
	SQL  string
	Args []any
}

// GenerateUpdate produces an UPDATE statement for a single cell change
// using the given SQL dialect.
func GenerateUpdate(d db.Dialect, schema, table string, pk PKValue, column string, newValue any) DMLResult {
	var args []any
	idx := 1

	set := fmt.Sprintf("%s = %s", d.QuoteIdent(column), d.Placeholder(idx))
	args = append(args, newValue)
	idx++

	where := buildWhere(d, pk, &idx, &args)

	sql := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		d.QualifyTable(schema, table), set, where)

	return DMLResult{SQL: sql, Args: args}
}

// GenerateInsert produces an INSERT statement using the given SQL dialect.
func GenerateInsert(d db.Dialect, schema, table string, values map[string]any) DMLResult {
	if len(values) == 0 {
		return DMLResult{
			SQL: fmt.Sprintf("INSERT INTO %s DEFAULT VALUES", d.QualifyTable(schema, table)),
		}
	}

	var cols []string
	var placeholders []string
	var args []any
	idx := 1
	for col, val := range values {
		cols = append(cols, d.QuoteIdent(col))
		if val == nil {
			placeholders = append(placeholders, "NULL")
		} else {
			placeholders = append(placeholders, d.Placeholder(idx))
			args = append(args, val)
			idx++
		}
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		d.QualifyTable(schema, table),
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "))

	return DMLResult{SQL: sql, Args: args}
}

// GenerateDelete produces a DELETE statement using the given SQL dialect.
func GenerateDelete(d db.Dialect, schema, table string, pk PKValue) DMLResult {
	var args []any
	idx := 1
	where := buildWhere(d, pk, &idx, &args)

	sql := fmt.Sprintf("DELETE FROM %s WHERE %s",
		d.QualifyTable(schema, table), where)

	return DMLResult{SQL: sql, Args: args}
}

func buildWhere(d db.Dialect, pk PKValue, idx *int, args *[]any) string {
	var clauses []string
	for i, col := range pk.Columns {
		val := pk.Values[i]
		if val == nil {
			clauses = append(clauses, fmt.Sprintf("%s IS NULL", d.QuoteIdent(col)))
		} else {
			clauses = append(clauses, fmt.Sprintf("%s = %s", d.QuoteIdent(col), d.Placeholder(*idx)))
			*args = append(*args, val)
			(*idx)++
		}
	}
	return strings.Join(clauses, " AND ")
}
