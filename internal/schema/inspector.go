// Package schema provides database schema introspection.
package schema

import (
	"context"
	"fmt"

	"github.com/mdjarv/db/internal/db"
)

// Inspector queries database metadata.
//
// The schema parameter names a namespace. Drivers without schema support
// (e.g. SQLite) should accept an empty string and ignore non-default values.
// For PostgreSQL, an empty string is treated as the dialect's default schema.
type Inspector interface {
	Tables(ctx context.Context, schema string) ([]Table, error)
	Columns(ctx context.Context, schema, table string) ([]ColumnInfo, error)
	Indexes(ctx context.Context, schema, table string) ([]Index, error)
	Constraints(ctx context.Context, schema, table string) ([]Constraint, error)
	ForeignKeys(ctx context.Context, schema, table string) ([]ForeignKey, error)
}

// inspectorProvider is implemented by connections that expose a native
// Inspector. Each driver's *Conn type satisfies this.
type inspectorProvider interface {
	Inspector() Inspector
}

// NewInspector returns the driver-native Inspector for the given connection.
// Returns an error if the driver does not support introspection.
func NewInspector(c db.Conn) (Inspector, error) {
	if p, ok := c.(inspectorProvider); ok {
		return p.Inspector(), nil
	}
	return nil, fmt.Errorf("schema: driver does not support introspection")
}

// MustInspector is NewInspector with panic-on-error for call sites that
// require an inspector (e.g. TUI setup where failure is fatal).
func MustInspector(c db.Conn) Inspector {
	insp, err := NewInspector(c)
	if err != nil {
		panic(err)
	}
	return insp
}
