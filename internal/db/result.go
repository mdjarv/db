package db

// Result holds the columns and a streaming row iterator from a query.
type Result struct {
	Columns []Column
	Rows    RowIterator
}

// RowIterator streams rows one at a time.
type RowIterator interface {
	Next() bool
	Values() ([]any, error)
	Err() error
	Close()
}

// ExecResult holds the outcome of a non-query statement.
type ExecResult struct {
	RowsAffected int64
}

// CompositeField describes a field within a composite type.
type CompositeField struct {
	Name     string
	TypeName string
}

// Column describes a single result column. Driver-agnostic.
//
// EnumValues and CompositeFields are optional: drivers that support these
// concepts (e.g. PostgreSQL) populate them; others leave them nil.
type Column struct {
	Name            string
	TypeName        string
	EnumValues      []string
	CompositeFields []CompositeField
}

// IsArray reports whether the column's type name indicates an array
// (by the trailing `[]` convention).
func (c Column) IsArray() bool {
	n := len(c.TypeName)
	return n >= 2 && c.TypeName[n-2:] == "[]"
}

// ElemTypeName returns the element type name for an array column,
// or the column's TypeName if it is not an array.
func (c Column) ElemTypeName() string {
	if c.IsArray() {
		return c.TypeName[:len(c.TypeName)-2]
	}
	return c.TypeName
}
