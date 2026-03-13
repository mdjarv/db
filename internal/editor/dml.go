package editor

import (
	"fmt"
	"strings"
)

// DMLResult holds a generated SQL statement and its parameters.
type DMLResult struct {
	SQL  string
	Args []any
}

// GenerateUpdate produces an UPDATE statement for a single cell change.
// Uses parameterized queries: SET col = $1 WHERE pk1 = $2 AND pk2 = $3
func GenerateUpdate(schema, table string, pk PKValue, column string, newValue any) DMLResult {
	var args []any
	idx := 1

	set := fmt.Sprintf("%s = $%d", quoteIdent(column), idx)
	args = append(args, newValue)
	idx++

	where := buildWhere(pk, &idx, &args)

	sql := fmt.Sprintf("UPDATE %s.%s SET %s WHERE %s",
		quoteIdent(schema), quoteIdent(table), set, where)

	return DMLResult{SQL: sql, Args: args}
}

// GenerateInsert produces an INSERT statement.
func GenerateInsert(schema, table string, values map[string]any) DMLResult {
	if len(values) == 0 {
		return DMLResult{
			SQL: fmt.Sprintf("INSERT INTO %s.%s DEFAULT VALUES", quoteIdent(schema), quoteIdent(table)),
		}
	}

	var cols []string
	var placeholders []string
	var args []any
	idx := 1
	for col, val := range values {
		cols = append(cols, quoteIdent(col))
		if val == nil {
			placeholders = append(placeholders, "NULL")
		} else {
			placeholders = append(placeholders, fmt.Sprintf("$%d", idx))
			args = append(args, val)
			idx++
		}
	}

	sql := fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (%s)",
		quoteIdent(schema), quoteIdent(table),
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "))

	return DMLResult{SQL: sql, Args: args}
}

// GenerateDelete produces a DELETE statement.
func GenerateDelete(schema, table string, pk PKValue) DMLResult {
	var args []any
	idx := 1
	where := buildWhere(pk, &idx, &args)

	sql := fmt.Sprintf("DELETE FROM %s.%s WHERE %s",
		quoteIdent(schema), quoteIdent(table), where)

	return DMLResult{SQL: sql, Args: args}
}

func buildWhere(pk PKValue, idx *int, args *[]any) string {
	var clauses []string
	for i, col := range pk.Columns {
		val := pk.Values[i]
		if val == nil {
			clauses = append(clauses, fmt.Sprintf("%s IS NULL", quoteIdent(col)))
		} else {
			clauses = append(clauses, fmt.Sprintf("%s = $%d", quoteIdent(col), *idx))
			*args = append(*args, val)
			(*idx)++
		}
	}
	return strings.Join(clauses, " AND ")
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}
