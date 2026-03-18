package core

import (
	"time"

	"github.com/mdjarv/db/internal/conn"
	"github.com/mdjarv/db/internal/db"
	"github.com/mdjarv/db/internal/dump"
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

// CompositeField describes a field within a composite type.
type CompositeField struct {
	Name     string
	TypeName string
}

// ResultColumn describes a column in the result set.
type ResultColumn struct {
	Name            string
	TypeName        string
	EnumValues      []string
	CompositeFields []CompositeField
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

// EditRequestMsg requests opening the edit dialog for a cell.
type EditRequestMsg struct {
	Row             int
	Col             int
	ColName         string
	TypeName        string
	Value           string
	EnumValues      []string
	CompositeFields []CompositeField
}

// DeleteRowMsg requests deletion of the current row.
type DeleteRowMsg struct {
	Row int
}

// InsertRowMsg requests insertion of a new blank row.
type InsertRowMsg struct{}

// CommitMsg requests applying and committing pending changes.
type CommitMsg struct{}

// RollbackMsg requests discarding pending changes.
type RollbackMsg struct{}

// UndoMsg requests undoing the last change.
type UndoMsg struct{}

// ChangesMsg requests listing pending changes.
type ChangesMsg struct{}

// ConfirmMsg carries the result of a confirmation dialog.
type ConfirmMsg struct {
	Action    string
	Confirmed bool
}

// PendingChangesMsg updates the status bar with change count.
type PendingChangesMsg struct {
	Count int
}

// EditingDisabledMsg signals that editing is not available.
type EditingDisabledMsg struct {
	Reason string
}

// ConnSelectorMsg triggers opening the connection selector with candidates.
type ConnSelectorMsg struct {
	Candidates []conn.Candidate
}

// ConnectedMsg signals a successful connection switch.
type ConnectedMsg struct {
	Conn      db.Conn
	Inspector schema.Inspector
	ConnInfo  string
	Candidate conn.Candidate
}

// ConnectErrorMsg signals a connection attempt failed.
type ConnectErrorMsg struct {
	Err error
}

// DumpTableMsg requests a table data dump.
type DumpTableMsg struct {
	Table string
}

// DumpSchemaMsg requests a schema-only dump.
type DumpSchemaMsg struct {
	Table string
}

// DumpStartMsg carries a dump configuration to begin execution.
type DumpStartMsg struct {
	Config dump.Config
}

// DumpProgressMsg carries a progress update from a running dump.
type DumpProgressMsg struct {
	Event dump.ProgressEvent
}

// DumpCompleteMsg signals a dump has finished.
type DumpCompleteMsg struct {
	Path     string
	Size     int64
	Duration time.Duration
	Err      error
}

// DumpCancelMsg signals the user cancelled an in-progress dump.
type DumpCancelMsg struct{}

// ClearErrorMsg signals that a timed error message should be cleared.
type ClearErrorMsg struct {
	// ID matches the error instance so stale clears are ignored.
	ID int
}
