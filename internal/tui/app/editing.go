package app

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mdjarv/db/internal/editor"
	"github.com/mdjarv/db/internal/tui/components/dialog"
	"github.com/mdjarv/db/internal/tui/components/editdialog"
	"github.com/mdjarv/db/internal/tui/components/table"
	"github.com/mdjarv/db/internal/tui/core"
)

// editingEnabled returns true if the current result has PK information.
func (m *Model) editingEnabled() bool {
	return len(m.editPKCols) > 0
}

// rowPK extracts the PK value for a given result row.
func (m *Model) rowPK(row int) editor.PKValue {
	cols, rows := m.resultView.ResultData()
	if cols == nil || row >= len(rows) {
		return editor.PKValue{}
	}
	pk := editor.PKValue{Columns: m.editPKCols}
	for _, idx := range m.editPKIdx {
		val := rows[row][idx]
		if table.IsNull(val) {
			pk.Values = append(pk.Values, nil)
		} else {
			pk.Values = append(pk.Values, val)
		}
	}
	return pk
}

// columnName returns the result column name at index.
func (m *Model) columnName(col int) string {
	cols, _ := m.resultView.ResultData()
	if col < len(cols) {
		return cols[col].Name
	}
	return ""
}

// handleEditRequest opens the edit dialog for a cell.
func (m *Model) handleEditRequest(msg core.EditRequestMsg) (Model, tea.Cmd) {
	if !m.editingEnabled() {
		m.statusBar.SetMessage("editing disabled: no primary key")
		return *m, nil
	}
	m.mode = core.ModeEdit
	m.statusBar.SetMode(m.mode)
	nullable := m.columnNullable(msg.ColName)
	m.editDialog.Open(editdialog.OpenOpts{
		Row:             msg.Row,
		Col:             msg.Col,
		ColName:         msg.ColName,
		TypeName:        msg.TypeName,
		Value:           msg.Value,
		Nullable:        nullable,
		EnumValues:      msg.EnumValues,
		CompositeFields: convertEditCompositeFields(msg.CompositeFields),
	})
	return *m, nil
}

// columnNullable returns whether the named column allows NULL.
func (m *Model) columnNullable(name string) bool {
	for _, c := range m.editCols {
		if c.Name == name {
			return c.Nullable
		}
	}
	return false
}

// handleEditSubmit processes a confirmed cell edit from the dialog.
func (m *Model) handleEditSubmit(msg editdialog.SubmitMsg) (Model, tea.Cmd) {
	m.mode = core.ModeNormal
	m.statusBar.SetMode(m.mode)

	newVal := msg.NewValue
	if msg.IsNull {
		newVal = table.NullPlaceholder
	}
	oldVal := msg.OldValue
	if oldVal == "" {
		oldVal = table.NullPlaceholder
	}

	// skip if value unchanged
	if oldVal == newVal {
		return *m, nil
	}

	pk := m.rowPK(msg.Row)
	colName := m.columnName(msg.Col)

	var oldAny, newAny any
	if table.IsNull(oldVal) {
		oldAny = nil
	} else {
		oldAny = oldVal
	}
	if table.IsNull(newVal) {
		newAny = nil
	} else {
		newAny = newVal
	}

	m.changeBuf.Add(editor.Change{
		Kind:     editor.ChangeUpdate,
		Table:    m.editTable,
		Schema:   m.editSchema,
		PK:       pk,
		Column:   colName,
		OldValue: oldAny,
		NewValue: newAny,
	})

	// update display value in result view
	if msg.Row < len(m.resultView.TableRows()) && msg.Col < len(m.resultView.TableRows()[msg.Row]) {
		m.resultView.TableRows()[msg.Row][msg.Col] = newVal
	}

	m.resultView.MarkModified(msg.Row, msg.Col)
	m.updatePendingStatus()
	return *m, nil
}

// handleEditCancel processes a cancelled cell edit.
func (m *Model) handleEditCancel() (Model, tea.Cmd) {
	m.mode = core.ModeNormal
	m.statusBar.SetMode(m.mode)
	return *m, nil
}

// handleDeleteRow processes a delete row request — shows confirmation dialog.
func (m *Model) handleDeleteRow(msg core.DeleteRowMsg) (Model, tea.Cmd) {
	if !m.editingEnabled() {
		m.statusBar.SetMessage("editing disabled: no primary key")
		return *m, nil
	}
	pk := m.rowPK(msg.Row)
	pkDesc := formatPKDesc(pk)
	m.dialog.Open("delete", "Delete row?", pkDesc)
	return *m, nil
}

// handleInsertRow adds a blank row for editing.
func (m *Model) handleInsertRow() (Model, tea.Cmd) {
	if !m.editingEnabled() {
		m.statusBar.SetMessage("editing disabled: no primary key")
		return *m, nil
	}
	m.changeBuf.Add(editor.Change{
		Kind:   editor.ChangeInsert,
		Table:  m.editTable,
		Schema: m.editSchema,
		Row:    make(map[string]any),
	})
	m.updatePendingStatus()
	m.statusBar.SetMessage("insert row added to pending changes")
	return *m, nil
}

// handleUndo removes the last change from the buffer.
func (m *Model) handleUndo() (Model, tea.Cmd) {
	last := m.changeBuf.RemoveLast()
	if last == nil {
		m.statusBar.SetMessage("nothing to undo")
		return *m, nil
	}
	if last.Kind == editor.ChangeUpdate {
		row := m.findRowForPK(last.PK)
		if row >= 0 {
			cols, _ := m.resultView.ResultData()
			for i, c := range cols {
				if c.Name == last.Column {
					oldStr := ""
					if last.OldValue == nil {
						oldStr = table.NullPlaceholder
					} else {
						oldStr = fmt.Sprintf("%v", last.OldValue)
					}
					m.resultView.RestoreCell(row, i, oldStr)
					break
				}
			}
		}
	}
	if last.Kind == editor.ChangeDelete {
		row := m.findRowForPK(last.PK)
		if row >= 0 {
			m.resultView.UnmarkDeleted(row)
		}
	}
	m.updatePendingStatus()
	m.statusBar.SetMessage("undone")
	return *m, nil
}

// handleCommit shows confirmation dialog with change summary.
func (m *Model) handleCommit() (Model, tea.Cmd) {
	if m.changeBuf.Len() == 0 {
		m.statusBar.SetMessage("no pending changes")
		return *m, nil
	}
	summary := m.changeSummary()
	m.dialog.Open("commit", "Commit changes?", summary)
	return *m, nil
}

// handleRollback discards all pending changes.
func (m *Model) handleRollback() (Model, tea.Cmd) {
	if m.changeBuf.Len() == 0 {
		m.statusBar.SetMessage("no pending changes")
		return *m, nil
	}
	// restore all modified cells
	for _, c := range m.changeBuf.Changes() {
		row := m.findRowForPK(c.PK)
		if row < 0 {
			continue
		}
		if c.Kind == editor.ChangeUpdate {
			cols, _ := m.resultView.ResultData()
			for i, col := range cols {
				if col.Name == c.Column {
					oldStr := ""
					if c.OldValue == nil {
						oldStr = table.NullPlaceholder
					} else {
						oldStr = fmt.Sprintf("%v", c.OldValue)
					}
					m.resultView.RestoreCell(row, i, oldStr)
					break
				}
			}
		}
		if c.Kind == editor.ChangeDelete {
			m.resultView.UnmarkDeleted(row)
		}
	}
	m.changeBuf.Clear()
	m.resultView.ClearModified()
	m.updatePendingStatus()
	m.statusBar.SetMessage("changes discarded")
	return *m, nil
}

// handleChanges shows list of pending changes.
func (m *Model) handleChanges() (Model, tea.Cmd) {
	if m.changeBuf.Len() == 0 {
		m.statusBar.SetMessage("no pending changes")
		return *m, nil
	}
	m.statusBar.SetMessage(m.changeSummary())
	return *m, nil
}

// handleDialogResult processes confirmation dialog results.
func (m *Model) handleDialogResult(msg dialog.ResultMsg) (Model, tea.Cmd) {
	switch msg.Action {
	case "delete":
		if msg.Confirmed {
			return m.confirmDelete()
		}
		m.statusBar.SetMessage("delete cancelled")
	case "commit":
		if msg.Confirmed {
			return m.doCommit()
		}
		m.statusBar.SetMessage("commit cancelled")
	case "switch-conn":
		if msg.Confirmed {
			m.changeBuf.Clear()
			m.resultView.ClearModified()
			return *m, m.discoverConnections()
		}
		m.statusBar.SetMessage("switch cancelled")
	}
	return *m, nil
}

// confirmDelete executes the pending delete.
func (m *Model) confirmDelete() (Model, tea.Cmd) {
	cols, rows := m.resultView.ResultData()
	if cols == nil {
		return *m, nil
	}
	row := m.resultView.TableCursorRow()
	if row >= len(rows) {
		return *m, nil
	}
	pk := m.rowPK(row)
	m.changeBuf.Add(editor.Change{
		Kind:   editor.ChangeDelete,
		Table:  m.editTable,
		Schema: m.editSchema,
		PK:     pk,
	})
	m.resultView.MarkDeleted(row)
	m.updatePendingStatus()
	m.statusBar.SetMessage("row marked for deletion")
	return *m, nil
}

// doCommit applies all changes to the database.
func (m *Model) doCommit() (Model, tea.Cmd) {
	if m.conn == nil {
		m.statusBar.SetMessage("not connected")
		return *m, nil
	}

	conn := m.conn
	changes := m.changeBuf.Changes()
	autocommit := m.autocommit

	return *m, func() tea.Msg {
		result := editor.Apply(context.Background(), conn, changes, autocommit)
		if result.Err != nil {
			return commitResultMsg{err: result.Err, applied: result.Applied}
		}
		if result.Tx != nil {
			err := result.Tx.Commit(context.Background())
			if err != nil {
				return commitResultMsg{err: fmt.Errorf("commit: %w", err), applied: result.Applied}
			}
		}
		return commitResultMsg{applied: result.Applied}
	}
}

type commitResultMsg struct {
	applied int
	err     error
}

// handleCommitResult processes the result of applying changes.
func (m *Model) handleCommitResult(msg commitResultMsg) (Model, tea.Cmd) {
	if msg.err != nil {
		m.statusBar.SetError(fmt.Sprintf("commit failed (%d applied): %s", msg.applied, msg.err))
		return *m, nil
	}
	m.changeBuf.Clear()
	m.resultView.ClearModified()
	m.updatePendingStatus()
	m.statusBar.SetSuccess(fmt.Sprintf("committed %d changes", msg.applied))

	// re-run the query to refresh data
	sql := m.queryEditor.Content()
	if sql != "" {
		return *m, m.executeQuery(sql)
	}
	return *m, nil
}

// handleSetCommand processes `:set` args for autocommit/noautocommit.
func (m *Model) handleSetCommand(args string) {
	switch strings.TrimSpace(args) {
	case "autocommit":
		if m.changeBuf.Len() > 0 {
			m.statusBar.SetMessage("warning: switching with pending changes! Changes will be applied immediately.")
		}
		m.autocommit = true
		m.statusBar.SetTxMode("auto")
		m.statusBar.SetMessage("autocommit enabled")
	case "noautocommit":
		if m.changeBuf.Len() > 0 {
			m.statusBar.SetMessage("warning: switching with pending changes!")
		}
		m.autocommit = false
		m.statusBar.SetTxMode("txn")
		m.statusBar.SetMessage("manual commit mode")
	default:
		m.statusBar.SetMessage("set: " + args)
	}
}

// updatePendingStatus updates the status bar with pending change count.
func (m *Model) updatePendingStatus() {
	n := m.changeBuf.Len()
	if n > 0 {
		m.statusBar.SetMessage(fmt.Sprintf("%d changes pending", n))
	}
}

// changeSummary returns a human-readable summary of pending changes.
func (m *Model) changeSummary() string {
	var updates, deletes, inserts int
	for _, c := range m.changeBuf.Changes() {
		switch c.Kind {
		case editor.ChangeUpdate:
			updates++
		case editor.ChangeDelete:
			deletes++
		case editor.ChangeInsert:
			inserts++
		}
	}
	var parts []string
	if updates > 0 {
		parts = append(parts, fmt.Sprintf("%d update(s)", updates))
	}
	if deletes > 0 {
		parts = append(parts, fmt.Sprintf("%d delete(s)", deletes))
	}
	if inserts > 0 {
		parts = append(parts, fmt.Sprintf("%d insert(s)", inserts))
	}
	return strings.Join(parts, ", ")
}

// findRowForPK finds the display row index for a given PK.
func (m *Model) findRowForPK(pk editor.PKValue) int {
	_, rows := m.resultView.ResultData()
	for r, row := range rows {
		match := true
		for i, idx := range m.editPKIdx {
			if idx >= len(row) {
				match = false
				break
			}
			val := row[idx]
			pkVal := pk.Values[i]
			if pkVal == nil {
				if val != table.NullPlaceholder {
					match = false
					break
				}
			} else {
				if val != fmt.Sprintf("%v", pkVal) {
					match = false
					break
				}
			}
		}
		if match {
			return r
		}
	}
	return -1
}

func convertEditCompositeFields(fields []core.CompositeField) []editdialog.CompositeField {
	if fields == nil {
		return nil
	}
	out := make([]editdialog.CompositeField, len(fields))
	for i, f := range fields {
		out[i] = editdialog.CompositeField{Name: f.Name, TypeName: f.TypeName}
	}
	return out
}

func formatPKDesc(pk editor.PKValue) string {
	var parts []string
	for i, col := range pk.Columns {
		val := pk.Values[i]
		if val == nil {
			parts = append(parts, fmt.Sprintf("%s=NULL", col))
		} else {
			parts = append(parts, fmt.Sprintf("%s=%v", col, val))
		}
	}
	return strings.Join(parts, ", ")
}

// parseTableFromSQL extracts the table name from a simple SELECT query.
// Returns (table, schema). Schema defaults to "public" if not specified.
func parseTableFromSQL(sql string) (string, string) {
	sql = strings.TrimSpace(sql)
	lower := strings.ToLower(sql)

	// simple heuristic: look for FROM <table>
	idx := strings.Index(lower, "from ")
	if idx < 0 {
		return "", ""
	}
	rest := strings.TrimSpace(sql[idx+5:])

	// skip subqueries and function calls
	if len(rest) > 0 && rest[0] == '(' {
		return "", ""
	}

	// take the first token after FROM
	var tablePart string
	for i, ch := range rest {
		if ch == ' ' || ch == '\n' || ch == '\t' || ch == ';' {
			tablePart = rest[:i]
			break
		}
		if i == len(rest)-1 {
			tablePart = rest
		}
	}
	tablePart = strings.Trim(tablePart, `"`)

	// handle schema.table
	if parts := strings.SplitN(tablePart, ".", 2); len(parts) == 2 {
		return strings.Trim(parts[1], `"`), strings.Trim(parts[0], `"`)
	}
	return tablePart, "public"
}

// setEditContext sets up editing metadata from table detail (PK columns, etc.).
func (m *Model) setEditContext(tableName, schemaName string, columns []core.ResultColumn) {
	m.editTable = tableName
	m.editSchema = schemaName
	m.editPKCols = nil
	m.editPKIdx = nil

	// look up PK columns from schema
	if m.inspector == nil {
		return
	}
	colInfos, err := m.inspector.Columns(context.Background(), schemaName, tableName)
	if err != nil {
		return
	}
	m.editCols = colInfos

	// find PK columns and map to result column indices
	for _, ci := range colInfos {
		if ci.IsPK {
			// find this column in the result set
			for j, rc := range columns {
				if rc.Name == ci.Name {
					m.editPKCols = append(m.editPKCols, ci.Name)
					m.editPKIdx = append(m.editPKIdx, j)
					break
				}
			}
		}
	}
}
