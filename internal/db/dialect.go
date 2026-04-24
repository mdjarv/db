package db

import (
	"fmt"
	"strings"
)

// Dialect holds SQL dialect details that vary between drivers.
type Dialect struct {
	// Name identifies the dialect (for diagnostics).
	Name string
	// Placeholder renders a positional parameter marker.
	// Postgres: "$1", "$2"; SQLite/MySQL: "?".
	Placeholder func(n int) string
	// QuoteIdent quotes a SQL identifier (column/table name).
	QuoteIdent func(s string) string
	// DefaultSchema is the implicit schema namespace (e.g. "public").
	// Empty for drivers without schemas.
	DefaultSchema string
	// SupportsSchemas is false for drivers where tables have no schema
	// namespace (SQLite). When false, QualifyTable omits the schema prefix.
	SupportsSchemas bool
}

// QualifyTable formats a schema-qualified table identifier according to the
// dialect. If schema is empty or the dialect does not support schemas,
// only the table is quoted.
func (d Dialect) QualifyTable(schema, table string) string {
	if !d.SupportsSchemas || schema == "" {
		return d.QuoteIdent(table)
	}
	return d.QuoteIdent(schema) + "." + d.QuoteIdent(table)
}

// PostgresDialect returns the PostgreSQL dialect description.
func PostgresDialect() Dialect {
	return Dialect{
		Name:            "postgres",
		Placeholder:     func(n int) string { return fmt.Sprintf("$%d", n) },
		QuoteIdent:      doubleQuoteIdent,
		DefaultSchema:   "public",
		SupportsSchemas: true,
	}
}

// SQLiteDialect returns the SQLite dialect description.
func SQLiteDialect() Dialect {
	return Dialect{
		Name:            "sqlite",
		Placeholder:     func(_ int) string { return "?" },
		QuoteIdent:      doubleQuoteIdent,
		DefaultSchema:   "",
		SupportsSchemas: false,
	}
}

func doubleQuoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}
