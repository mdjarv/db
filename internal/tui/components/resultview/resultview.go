// Package resultview wraps the table component for the result display pane.
package resultview

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/tui/components/table"
	"github.com/mdjarv/db/internal/tui/core"
	"github.com/mdjarv/db/internal/tui/theme"
)

const defaultSeparator = ","

// Model is the result view state.
type Model struct {
	table     table.Model
	scroll    *ScrollState
	focused   bool
	width     int
	height    int
	separator string

	// result metadata
	columns  []core.ResultColumn
	duration time.Duration
	errMsg   string
	hasData  bool

	// operator pending (for dR, oR sequences)
	pendingOp string
}

// New creates a result view with empty state.
func New() *Model {
	return &Model{
		table:     table.New(nil, nil),
		scroll:    NewScrollState(),
		separator: defaultSeparator,
	}
}

// SetResult sets query result data.
func (m *Model) SetResult(columns []core.ResultColumn, rows [][]string, duration time.Duration) {
	cols := make([]table.Column, len(columns))
	for i, c := range columns {
		title := c.Name
		if c.TypeName != "" {
			title += " [" + c.TypeName + "]"
		}
		cols[i] = table.Column{
			Title: title,
			Width: autoWidth(c.Name, c.TypeName, rows, i),
		}
	}
	m.columns = columns
	m.duration = duration
	m.errMsg = ""
	m.hasData = true

	m.scroll.SetRows(rows)
	m.scroll.AllLoaded = true
	m.scroll.TotalEstimate = len(rows)

	m.table = table.New(cols, rows)
	m.table.Width = max(m.width-4, 1)
	m.table.Height = max(m.height-4, 1)
}

// SetError displays an error message.
func (m *Model) SetError(err error) {
	m.errMsg = err.Error()
	m.hasData = false
	m.columns = nil
	m.table = table.New(nil, nil)
	m.scroll.Reset()
}

// Clear resets to empty state.
func (m *Model) Clear() {
	m.errMsg = ""
	m.hasData = false
	m.columns = nil
	m.duration = 0
	m.table = table.New(nil, nil)
	m.scroll.Reset()
}

// ResultData returns the current columns and rows for export.
func (m *Model) ResultData() ([]core.ResultColumn, [][]string) {
	if !m.hasData {
		return nil, nil
	}
	return m.columns, m.scroll.Rows()
}

// EnterVisualLine starts V mode (row selection).
func (m *Model) EnterVisualLine() { m.table.EnterVisualLine() }

// EnterVisualBlock starts v mode (rectangular selection).
func (m *Model) EnterVisualBlock() { m.table.EnterVisualBlock() }

// ExitVisual cancels visual selection.
func (m *Model) ExitVisual() { m.table.ExitVisual() }

// Update handles input messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if !m.focused {
		return nil
	}
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	switch m.table.Visual {
	case table.VisualLine:
		return m.updateVisualLine(km)
	case table.VisualBlock:
		return m.updateVisualBlock(km)
	default:
		return m.updateNormal(km)
	}
}

func (m *Model) updateNormal(km tea.KeyMsg) tea.Cmd {
	key := km.String()

	// handle pending operator
	if m.pendingOp != "" {
		op := m.pendingOp
		m.pendingOp = ""
		if key == "R" {
			switch op {
			case "d":
				row := m.table.CursorRow
				return func() tea.Msg { return core.DeleteRowMsg{Row: row} }
			case "o":
				return func() tea.Msg { return core.InsertRowMsg{} }
			}
		}
		// invalid sequence, fall through
	}

	switch key {
	case "j", "down":
		m.table.MoveDown()
	case "k", "up":
		m.table.MoveUp()
	case "h", "left":
		m.table.MoveLeft()
	case "l", "right":
		m.table.MoveRight()
	case "g":
		m.table.GotoTop()
	case "G":
		m.table.GotoBottom()
	case "0":
		m.table.GotoFirstCol()
	case "$":
		m.table.GotoLastCol()
	case "ctrl+d":
		m.table.HalfPageDown()
	case "ctrl+u":
		m.table.HalfPageUp()
	case "ctrl+f":
		m.table.FullPageDown()
	case "ctrl+b":
		m.table.FullPageUp()
	case "enter", "e":
		return m.startEdit()
	case "y":
		content := m.table.YankCell()
		return func() tea.Msg { return core.YankMsg{Content: content} }
	case "Y":
		content := m.table.YankRow(m.separator)
		return func() tea.Msg { return core.YankMsg{Content: content} }
	case "d":
		m.pendingOp = "d"
	case "o":
		m.pendingOp = "o"
	case "ctrl+z":
		return func() tea.Msg { return core.UndoMsg{} }
	}
	return nil
}

func (m *Model) startEdit() tea.Cmd {
	if !m.hasData || len(m.table.Rows) == 0 {
		return nil
	}
	row := m.table.CursorRow
	col := m.table.CursorCol
	if row >= len(m.table.Rows) || col >= len(m.table.Rows[row]) {
		return nil
	}
	val := m.table.Rows[row][col]
	if val == table.NullPlaceholder {
		val = ""
	}
	colName := ""
	typeName := ""
	var enumValues []string
	var compositeFields []core.CompositeField
	if col < len(m.columns) {
		colName = m.columns[col].Name
		typeName = m.columns[col].TypeName
		enumValues = m.columns[col].EnumValues
		compositeFields = m.columns[col].CompositeFields
	}
	return func() tea.Msg {
		return core.EditRequestMsg{Row: row, Col: col, ColName: colName, TypeName: typeName, Value: val, EnumValues: enumValues, CompositeFields: compositeFields}
	}
}

// MarkModified marks a cell as modified (yellow highlight).
func (m *Model) MarkModified(row, col int) {
	if m.table.ModifiedCells == nil {
		m.table.ModifiedCells = make(map[table.CellKey]bool)
	}
	m.table.ModifiedCells[table.CellKey{Row: row, Col: col}] = true
}

// MarkDeleted marks a row as pending deletion.
func (m *Model) MarkDeleted(row int) {
	if m.table.DeletedRows == nil {
		m.table.DeletedRows = make(map[int]bool)
	}
	m.table.DeletedRows[row] = true
}

// UnmarkDeleted removes the deletion mark from a row.
func (m *Model) UnmarkDeleted(row int) {
	delete(m.table.DeletedRows, row)
}

// UnmarkModified removes the modified mark from a cell.
func (m *Model) UnmarkModified(row, col int) {
	delete(m.table.ModifiedCells, table.CellKey{Row: row, Col: col})
}

// ClearModified removes all modified/deleted marks.
func (m *Model) ClearModified() {
	m.table.ModifiedCells = nil
	m.table.DeletedRows = nil
}

// RestoreCell restores the original value of a cell.
func (m *Model) RestoreCell(row, col int, value string) {
	if row < len(m.table.Rows) && col < len(m.table.Rows[row]) {
		m.table.Rows[row][col] = value
	}
	m.UnmarkModified(row, col)
}

func (m *Model) updateVisualLine(km tea.KeyMsg) tea.Cmd {
	switch km.String() {
	case "j", "down":
		if m.table.IsLineRowAxis() {
			m.table.MoveDown()
		}
	case "k", "up":
		if m.table.IsLineRowAxis() {
			m.table.MoveUp()
		}
	case "h", "left":
		if !m.table.IsLineRowAxis() {
			m.table.MoveLeft()
			m.table.UpdateLineColRange()
		}
	case "l", "right":
		if !m.table.IsLineRowAxis() {
			m.table.MoveRight()
			m.table.UpdateLineColRange()
		}
	case "tab":
		m.table.ToggleLineAxis()
	case "y":
		content := m.table.YankSelection(m.separator)
		m.table.ExitVisual()
		return func() tea.Msg { return core.YankMsg{Content: content} }
	}
	return nil
}

func (m *Model) updateVisualBlock(km tea.KeyMsg) tea.Cmd {
	switch km.String() {
	case "j", "down":
		m.table.MoveDown()
	case "k", "up":
		m.table.MoveUp()
	case "h", "left":
		m.table.MoveLeft()
	case "l", "right":
		m.table.MoveRight()
	case "y":
		content := m.table.YankSelection(m.separator)
		m.table.ExitVisual()
		return func() tea.Msg { return core.YankMsg{Content: content} }
	}
	return nil
}

// View renders the result pane.
func (m *Model) View() string {
	t := theme.Current().Styles

	borderColor := t.BorderUnfocused
	if m.focused {
		borderColor = t.BorderFocused
	}
	if m.table.Visual != table.VisualNone {
		borderColor = t.BorderVisual
	}

	innerH := m.height - 2
	innerW := m.width - 2

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(innerW).
		Height(innerH)

	if m.errMsg != "" {
		return style.Render(t.Error.Render("Error: " + m.errMsg))
	}

	if !m.hasData {
		return style.Render(t.Dim.Render("No results"))
	}

	// render with custom bottom border containing status
	noBottom := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderBottom(false).
		BorderForeground(borderColor).
		Width(innerW).
		Height(innerH)

	content := m.table.View(m.focused)
	top := noBottom.Render(content)

	status := m.statusLine()
	bottom := buildBottomBorder(status, m.width, borderColor)

	return top + "\n" + bottom
}

func buildBottomBorder(status string, width int, color lipgloss.Color) string {
	style := lipgloss.NewStyle().Foreground(color)
	// ╰── status ─────────────╯
	maxStatus := width - 6
	if len(status) > maxStatus {
		status = status[:maxStatus]
	}
	pad := width - len(status) - 5
	if pad < 0 {
		pad = 0
	}
	return style.Render("╰─ " + status + " " + strings.Repeat("─", pad) + "╯")
}

func (m *Model) statusLine() string {
	if len(m.table.Rows) == 0 {
		return "0 rows"
	}
	vh := m.table.ViewHeight()
	startRow := m.table.RowOffset + 1
	endRow := min(m.table.RowOffset+vh, len(m.table.Rows))

	rowInfo := fmt.Sprintf("rows %d-%d of %d", startRow, endRow, len(m.table.Rows))
	colInfo := fmt.Sprintf("%d cols", len(m.columns))
	durInfo := formatDuration(m.duration)

	return fmt.Sprintf("%s | %s | %s", rowInfo, colInfo, durInfo)
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%d\u00b5s", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%.1fms", float64(d.Microseconds())/1000)
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

func autoWidth(name, typeName string, rows [][]string, colIdx int) int {
	headerLen := len(name)
	if typeName != "" {
		headerLen += len(typeName) + 3 // " [type]"
	}
	w := headerLen

	sampleSize := min(len(rows), 100)
	for i := range sampleSize {
		if colIdx < len(rows[i]) {
			val := rows[i][colIdx]
			if val == table.NullPlaceholder {
				if 4 > w {
					w = 4
				}
			} else if len(val) > w {
				w = len(val)
			}
		}
	}

	if w < 4 {
		w = 4
	}
	if w > 50 {
		w = 50
	}
	return w
}

// Focused returns whether the pane is focused.
func (m *Model) Focused() bool { return m.focused }

// SetFocused sets the focus state.
func (m *Model) SetFocused(f bool) { m.focused = f }

// SetSize sets the render dimensions.
func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.table.Width = max(w-4, 1)
	m.table.Height = max(h-4, 1)
}

// SetSeparator sets the CSV separator.
func (m *Model) SetSeparator(sep string) { m.separator = sep }

// TableRows returns the underlying table rows slice.
func (m *Model) TableRows() [][]string { return m.table.Rows }

// TableCursorRow returns the result table cursor row.
func (m *Model) TableCursorRow() int { return m.table.CursorRow }

// TableCursorCol returns the result table cursor column.
func (m *Model) TableCursorCol() int { return m.table.CursorCol }

// TableRowOffset returns the result table row scroll offset.
func (m *Model) TableRowOffset() int { return m.table.RowOffset }

// TableColOffset returns the result table column scroll offset.
func (m *Model) TableColOffset() int { return m.table.ColOffset }

// SetTableCursor restores the result table cursor and scroll positions.
func (m *Model) SetTableCursor(cursorRow, cursorCol, rowOffset, colOffset int) {
	m.table.CursorRow = cursorRow
	m.table.CursorCol = cursorCol
	m.table.RowOffset = rowOffset
	m.table.ColOffset = colOffset
}
