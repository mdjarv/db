// Package resultview wraps the table component for the result display pane.
package resultview

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/tui/components/table"
	"github.com/mdjarv/db/internal/tui/core"
)

const defaultSeparator = ","

var statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

// Model is the result view state.
type Model struct {
	table     table.Model
	scroll    *ScrollState
	inspector Inspector
	focused   bool
	width     int
	height    int
	separator string

	// result metadata
	columns  []core.ResultColumn
	duration time.Duration
	errMsg   string
	hasData  bool
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
	m.table.Height = max(m.height-5, 1) // extra row for status line
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

	if m.inspector.IsActive() {
		return m.inspector.Update(km)
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
	switch km.String() {
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
	case "enter":
		return m.inspectCell()
	case "y":
		content := m.table.YankCell()
		return func() tea.Msg { return core.YankMsg{Content: content} }
	case "Y":
		content := m.table.YankRow(m.separator)
		return func() tea.Msg { return core.YankMsg{Content: content} }
	}
	return nil
}

func (m *Model) inspectCell() tea.Cmd {
	if !m.hasData || len(m.columns) == 0 {
		return nil
	}
	col := m.table.CursorCol
	if col >= len(m.columns) {
		return nil
	}
	val := m.table.YankCell()
	if val == table.NullPlaceholder {
		val = "NULL"
	}
	m.inspector.Open(m.columns[col].Name, m.columns[col].TypeName, val)
	return nil
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
	borderColor := lipgloss.Color("240")
	if m.focused {
		borderColor = lipgloss.Color("62")
	}
	if m.table.Visual != table.VisualNone {
		borderColor = lipgloss.Color("208")
	}

	innerH := m.height - 2
	innerW := m.width - 2

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(innerW).
		Height(innerH)

	if m.inspector.IsActive() {
		return style.Render(m.inspector.View(innerW, innerH))
	}

	if m.errMsg != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		return style.Render(errStyle.Render("Error: " + m.errMsg))
	}

	if !m.hasData {
		return style.Render(statusStyle.Render("No results"))
	}

	content := m.table.View(m.focused)
	status := m.statusLine()
	full := content + "\n" + statusStyle.Render(status)

	return style.Render(full)
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
	m.table.Height = max(h-5, 1) // status line
}

// SetSeparator sets the CSV separator.
func (m *Model) SetSeparator(sep string) { m.separator = sep }
