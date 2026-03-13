// Package schema provides database schema introspection.
package schema

import "context"

// Inspector queries database metadata.
type Inspector interface {
	Tables(ctx context.Context, schema string) ([]Table, error)
	Columns(ctx context.Context, schema, table string) ([]ColumnInfo, error)
	Indexes(ctx context.Context, schema, table string) ([]Index, error)
	Constraints(ctx context.Context, schema, table string) ([]Constraint, error)
	ForeignKeys(ctx context.Context, schema, table string) ([]ForeignKey, error)
}
