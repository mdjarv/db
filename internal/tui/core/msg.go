package core

import (
	"time"

	"github.com/mdjarv/db/internal/schema"
)

// ModeChangedMsg signals a vim mode transition.
type ModeChangedMsg struct {
	Mode Mode
}

// StatusMsg carries a status bar message.
type StatusMsg struct {
	Text string
}

// QuerySubmittedMsg carries a submitted SQL query.
type QuerySubmittedMsg struct {
	SQL string
}

// YankMsg carries content to copy to clipboard.
type YankMsg struct {
	Content string
}

// SchemaLoadedMsg delivers table list data from Inspector.
type SchemaLoadedMsg struct {
	Tables []schema.Table
	Err    error
}

// TableSelectedMsg signals cursor moved to a new table.
type TableSelectedMsg struct {
	Table schema.Table
}

// TableDetailMsg delivers schema detail for the selected table.
type TableDetailMsg struct {
	Table       schema.Table
	Columns     []schema.ColumnInfo
	Indexes     []schema.Index
	Constraints []schema.Constraint
	ForeignKeys []schema.ForeignKey
	Err         error
}

// QueryRequestMsg requests a query be placed in the editor and executed.
type QueryRequestMsg struct {
	SQL string
}

// RefreshSchemaMsg requests a schema reload from Inspector.
type RefreshSchemaMsg struct{}

// QueryResultMsg carries query results to the result viewer.
type QueryResultMsg struct {
	Columns  []ResultColumn
	Rows     [][]string
	Duration time.Duration
}

// ResultColumn describes a column in the result set.
type ResultColumn struct {
	Name     string
	TypeName string
}

// QueryErrorMsg carries a query error.
type QueryErrorMsg struct {
	Err error
}

// ExportRequestMsg triggers an export from the command bar.
type ExportRequestMsg struct {
	Format string // "csv", "json", "sql"
	Path   string
}

// CellInspectMsg opens the cell inspector popup.
type CellInspectMsg struct {
	Column   string
	TypeName string
	Value    string
}
