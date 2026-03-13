// Package table provides a reusable table component with cell cursor,
// horizontal scrolling, and visual selection.
package table

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Column defines a table column.
type Column struct {
	Title string
	Width int
}

// VisualKind identifies the type of visual selection.
type VisualKind int

// Visual selection modes.
const (
	VisualNone  VisualKind = iota
	VisualLine             // V — row selection, Tab toggles column axis
	VisualBlock            // v — rectangular selection
)

type axis int

const (
	axisRow axis = iota
	axisCol
)

// Model holds the table state.
type Model struct {
	Columns   []Column
	Rows      [][]string
	CursorRow int
	CursorCol int
	RowOffset int
	ColOffset int
	Width     int
	Height    int

	Visual    VisualKind
	AnchorRow int
	AnchorCol int

	// V-LINE column selection
	LineAxis      axis
	LineAnchorCol int
	LineColStart  int
	LineColEnd    int
}

// New creates a table model.
func New(cols []Column, rows [][]string) Model {
	return Model{
		Columns: cols,
		Rows:    rows,
	}
}

// EnterVisualLine starts V mode.
func (m *Model) EnterVisualLine() {
	m.Visual = VisualLine
	m.AnchorRow = m.CursorRow
	m.LineAxis = axisRow
	m.LineColStart = 0
	m.LineColEnd = len(m.Columns) - 1
}

// EnterVisualBlock starts v mode.
func (m *Model) EnterVisualBlock() {
	m.Visual = VisualBlock
	m.AnchorRow = m.CursorRow
	m.AnchorCol = m.CursorCol
}

// ExitVisual cancels selection.
func (m *Model) ExitVisual() {
	m.Visual = VisualNone
}

// Navigation

// MoveDown moves the cursor down.
func (m *Model) MoveDown() {
	if m.CursorRow < len(m.Rows)-1 {
		m.CursorRow++
		m.ensureRowVisible()
	}
}

// MoveUp moves the cursor up.
func (m *Model) MoveUp() {
	if m.CursorRow > 0 {
		m.CursorRow--
		m.ensureRowVisible()
	}
}

// MoveLeft moves the cursor left.
func (m *Model) MoveLeft() {
	if m.CursorCol > 0 {
		m.CursorCol--
		m.ensureColVisible()
	}
}

// MoveRight moves the cursor right.
func (m *Model) MoveRight() {
	if m.CursorCol < len(m.Columns)-1 {
		m.CursorCol++
		m.ensureColVisible()
	}
}

// GotoTop moves cursor to first row.
func (m *Model) GotoTop() {
	m.CursorRow = 0
	m.RowOffset = 0
}

// GotoBottom moves cursor to last row.
func (m *Model) GotoBottom() {
	m.CursorRow = len(m.Rows) - 1
	m.ensureRowVisible()
}

// GotoFirstCol moves cursor to first column.
func (m *Model) GotoFirstCol() {
	m.CursorCol = 0
	m.ColOffset = 0
}

// GotoLastCol moves cursor to last column.
func (m *Model) GotoLastCol() {
	if len(m.Columns) > 0 {
		m.CursorCol = len(m.Columns) - 1
		m.ensureColVisible()
	}
}

// HalfPageDown moves cursor half a page down.
func (m *Model) HalfPageDown() {
	half := m.ViewHeight() / 2
	if half < 1 {
		half = 1
	}
	m.CursorRow = min(m.CursorRow+half, len(m.Rows)-1)
	m.ensureRowVisible()
}

// HalfPageUp moves cursor half a page up.
func (m *Model) HalfPageUp() {
	half := m.ViewHeight() / 2
	if half < 1 {
		half = 1
	}
	m.CursorRow = max(m.CursorRow-half, 0)
	m.ensureRowVisible()
}

// FullPageDown moves cursor a full page down.
func (m *Model) FullPageDown() {
	vh := m.ViewHeight()
	if vh < 1 {
		vh = 1
	}
	m.CursorRow = min(m.CursorRow+vh, len(m.Rows)-1)
	m.ensureRowVisible()
}

// FullPageUp moves cursor a full page up.
func (m *Model) FullPageUp() {
	vh := m.ViewHeight()
	if vh < 1 {
		vh = 1
	}
	m.CursorRow = max(m.CursorRow-vh, 0)
	m.ensureRowVisible()
}

// ToggleLineAxis switches between row and column axis in V-LINE mode.
func (m *Model) ToggleLineAxis() {
	if m.Visual != VisualLine {
		return
	}
	if m.LineAxis == axisRow {
		m.LineAxis = axisCol
		m.LineAnchorCol = m.CursorCol
		m.LineColStart = m.CursorCol
		m.LineColEnd = m.CursorCol
	} else {
		m.LineAxis = axisRow
	}
}

// UpdateLineColRange recalculates column range from anchor.
func (m *Model) UpdateLineColRange() {
	if m.LineAnchorCol <= m.CursorCol {
		m.LineColStart = m.LineAnchorCol
		m.LineColEnd = m.CursorCol
	} else {
		m.LineColStart = m.CursorCol
		m.LineColEnd = m.LineAnchorCol
	}
}

// IsLineRowAxis returns true if V-LINE is on row axis.
func (m *Model) IsLineRowAxis() bool {
	return m.LineAxis == axisRow
}

// Scrolling

func (m *Model) ensureRowVisible() {
	vh := m.ViewHeight()
	if m.CursorRow < m.RowOffset {
		m.RowOffset = m.CursorRow
	} else if m.CursorRow >= m.RowOffset+vh {
		m.RowOffset = m.CursorRow - vh + 1
	}
}

func (m *Model) ensureColVisible() {
	vc := m.VisibleCols()
	if m.CursorCol < m.ColOffset {
		m.ColOffset = m.CursorCol
	} else if m.CursorCol >= m.ColOffset+vc {
		m.ColOffset = m.CursorCol - vc + 1
	}
}

// ViewHeight returns the number of visible data rows.
func (m *Model) ViewHeight() int {
	return max(m.Height-2, 1) // header + separator
}

// VisibleCols returns the number of columns that fit in the width.
func (m *Model) VisibleCols() int {
	avail := m.Width
	count := 0
	used := 0
	for i := m.ColOffset; i < len(m.Columns); i++ {
		w := m.Columns[i].Width
		if i > m.ColOffset {
			w += 3 // " │ "
		}
		if used+w > avail && count > 0 {
			break
		}
		used += w
		count++
	}
	return max(count, 1)
}

// Selection ranges

// RowRange returns the selected row range (start, end inclusive).
func (m *Model) RowRange() (int, int) {
	if m.AnchorRow <= m.CursorRow {
		return m.AnchorRow, m.CursorRow
	}
	return m.CursorRow, m.AnchorRow
}

// BlockColRange returns the selected column range for block mode.
func (m *Model) BlockColRange() (int, int) {
	if m.AnchorCol <= m.CursorCol {
		return m.AnchorCol, m.CursorCol
	}
	return m.CursorCol, m.AnchorCol
}

// Yank

// YankCell returns the current cell value.
func (m *Model) YankCell() string {
	if m.CursorRow < len(m.Rows) && m.CursorCol < len(m.Rows[m.CursorRow]) {
		return m.Rows[m.CursorRow][m.CursorCol]
	}
	return ""
}

// YankRow returns the current row as CSV (data only, no headers).
func (m *Model) YankRow(sep string) string {
	return m.formatRows(m.CursorRow, m.CursorRow, 0, len(m.Columns)-1, sep)
}

// YankSelection returns the selected data as CSV (data only, no headers).
func (m *Model) YankSelection(sep string) string {
	switch m.Visual {
	case VisualLine:
		startRow, endRow := m.RowRange()
		return m.formatRows(startRow, endRow, m.LineColStart, m.LineColEnd, sep)
	case VisualBlock:
		startRow, endRow := m.RowRange()
		startCol, endCol := m.BlockColRange()
		return m.formatRows(startRow, endRow, startCol, endCol, sep)
	default:
		return ""
	}
}

func (m *Model) formatRows(startRow, endRow, startCol, endCol int, sep string) string {
	var sb strings.Builder
	for r := startRow; r <= endRow; r++ {
		if r > startRow {
			sb.WriteByte('\n')
		}
		for c := startCol; c <= endCol; c++ {
			if c > startCol {
				sb.WriteString(sep)
			}
			sb.WriteString(escapeCSV(m.Rows[r][c], sep))
		}
	}
	return sb.String()
}

func escapeCSV(val, sep string) string {
	if strings.ContainsAny(val, sep+"\"\n") {
		return "\"" + strings.ReplaceAll(val, "\"", "\"\"") + "\""
	}
	return val
}

// Rendering

// NullPlaceholder is the display string for NULL values.
const NullPlaceholder = "\x00NULL\x00"

var (
	headerStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("75"))
	separatorColor = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57"))
	selectStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("208"))
	colSelectStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("223")).Background(lipgloss.Color("94"))
	dimStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorRowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	nullStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
)

// View renders the table.
func (m *Model) View(focused bool) string {
	if len(m.Columns) == 0 {
		return ""
	}

	vc := m.VisibleCols()
	endCol := min(m.ColOffset+vc, len(m.Columns))
	colSep := separatorColor.Render(" │ ")
	hdrSep := separatorColor.Render("─┼─")

	// compute partial column
	usedWidth := 0
	for i := m.ColOffset; i < endCol; i++ {
		if i > m.ColOffset {
			usedWidth += 3
		}
		usedWidth += m.Columns[i].Width
	}
	partialCol := -1
	partialW := 0
	if endCol < len(m.Columns) {
		remaining := m.Width - usedWidth - 3
		if remaining > 0 {
			partialCol = endCol
			partialW = remaining
		}
	}

	var sb strings.Builder

	// header
	for i := m.ColOffset; i < endCol; i++ {
		if i > m.ColOffset {
			sb.WriteString(colSep)
		}
		text := padCell(m.Columns[i].Title, m.Columns[i].Width)
		if m.Visual != VisualNone && !m.isColSelected(i) {
			sb.WriteString(dimStyle.Render(text))
		} else {
			sb.WriteString(headerStyle.Render(text))
		}
	}
	if partialCol >= 0 {
		sb.WriteString(colSep)
		text := padCell(m.Columns[partialCol].Title, partialW)
		sb.WriteString(dimStyle.Render(text))
	}
	sb.WriteByte('\n')

	// separator
	for i := m.ColOffset; i < endCol; i++ {
		if i > m.ColOffset {
			sb.WriteString(hdrSep)
		}
		sb.WriteString(separatorColor.Render(strings.Repeat("─", m.Columns[i].Width)))
	}
	if partialCol >= 0 {
		sb.WriteString(hdrSep)
		sb.WriteString(separatorColor.Render(strings.Repeat("─", partialW)))
	}
	sb.WriteByte('\n')

	// rows
	vh := m.ViewHeight()
	endRow := min(m.RowOffset+vh, len(m.Rows))

	for r := m.RowOffset; r < endRow; r++ {
		for i := m.ColOffset; i < endCol; i++ {
			if i > m.ColOffset {
				sb.WriteString(colSep)
			}
			val := ""
			if i < len(m.Rows[r]) {
				val = m.Rows[r][i]
			}
			isNull := val == NullPlaceholder
			if isNull {
				text := padCell("NULL", m.Columns[i].Width)
				styled := m.styleCell(text, r, i, focused)
				if !m.hasSelectionStyle(r, i, focused) {
					styled = nullStyle.Render(text)
				}
				sb.WriteString(styled)
			} else {
				text := truncateCell(val, m.Columns[i].Width)
				sb.WriteString(m.styleCell(text, r, i, focused))
			}
		}
		if partialCol >= 0 {
			sb.WriteString(colSep)
			val := ""
			if partialCol < len(m.Rows[r]) {
				val = m.Rows[r][partialCol]
			}
			if val == NullPlaceholder {
				sb.WriteString(dimStyle.Render(padCell("NULL", partialW)))
			} else {
				sb.WriteString(dimStyle.Render(truncateCell(val, partialW)))
			}
		}
		if r < endRow-1 {
			sb.WriteByte('\n')
		}
	}

	return sb.String()
}

func (m *Model) styleCell(text string, row, col int, focused bool) string {
	switch m.Visual {
	case VisualLine:
		inRow := m.inRowRange(row)
		inCol := col >= m.LineColStart && col <= m.LineColEnd
		if inRow && inCol {
			return selectStyle.Render(text)
		}
		if inCol {
			return colSelectStyle.Render(text)
		}
		if inRow {
			return dimStyle.Render(text)
		}
		return text

	case VisualBlock:
		if m.inRowRange(row) && m.inBlockColRange(col) {
			return selectStyle.Render(text)
		}
		return text

	default:
		if focused && row == m.CursorRow && col == m.CursorCol {
			return cursorStyle.Render(text)
		}
		if focused && row == m.CursorRow {
			return cursorRowStyle.Render(text)
		}
		return text
	}
}

func (m *Model) inRowRange(row int) bool {
	s, e := m.RowRange()
	return row >= s && row <= e
}

func (m *Model) inBlockColRange(col int) bool {
	s, e := m.BlockColRange()
	return col >= s && col <= e
}

func (m *Model) isColSelected(col int) bool {
	if m.Visual == VisualLine {
		return col >= m.LineColStart && col <= m.LineColEnd
	}
	if m.Visual == VisualBlock {
		return m.inBlockColRange(col)
	}
	return true
}

func (m *Model) hasSelectionStyle(row, col int, focused bool) bool {
	switch m.Visual {
	case VisualLine:
		inRow := m.inRowRange(row)
		inCol := col >= m.LineColStart && col <= m.LineColEnd
		return (inRow && inCol) || inCol
	case VisualBlock:
		return m.inRowRange(row) && m.inBlockColRange(col)
	default:
		return focused && row == m.CursorRow
	}
}

func padCell(val string, width int) string {
	if len(val) > width {
		val = val[:width]
	}
	return fmt.Sprintf("%-*s", width, val)
}

func truncateCell(val string, width int) string {
	if len(val) > width {
		if width > 1 {
			return fmt.Sprintf("%-*s", width, val[:width-1]+"\u2026")
		}
		return val[:width]
	}
	return fmt.Sprintf("%-*s", width, val)
}
