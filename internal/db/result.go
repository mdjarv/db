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

// Column describes a single result column.
type Column struct {
	Name            string
	TypeName        string
	TypeOID         uint32
	EnumValues      []string         // non-nil for enum types
	CompositeFields []CompositeField // non-nil for composite types
}

// TypeDetail provides full introspection info for a type OID.
type TypeDetail struct {
	OID             uint32
	Name            string
	IsArray         bool
	ElemOID         uint32
	ElemTypeName    string
	EnumValues      []string
	CompositeFields []CompositeField
}
