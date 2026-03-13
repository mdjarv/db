package schema

// Table describes a database table, view, or materialized view.
type Table struct {
	Name        string
	Schema      string
	Type        string // "table", "view", "materialized view"
	RowEstimate int64
	Size        string // human-readable, e.g. "8192 bytes"
}

// ColumnInfo describes a single column in a table.
type ColumnInfo struct {
	Name     string
	TypeName string
	Nullable bool
	Default  string
	IsPK     bool
	Position int
}

// Index describes an index on a table.
type Index struct {
	Name       string
	Columns    []string
	Unique     bool
	Type       string // btree, hash, gin, gist
	Size       string
	Definition string
}

// Constraint describes a table constraint.
type Constraint struct {
	Name       string
	Type       string // PRIMARY KEY, FOREIGN KEY, UNIQUE, CHECK, EXCLUDE
	Columns    []string
	Definition string
}

// ForeignKey describes a foreign key relationship.
type ForeignKey struct {
	Name              string
	Columns           []string
	ReferencedTable   string
	ReferencedSchema  string
	ReferencedColumns []string
	OnDelete          string
	OnUpdate          string
}
