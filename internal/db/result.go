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

// Column describes a single result column.
type Column struct {
	Name     string
	TypeName string
	TypeOID  uint32
}
