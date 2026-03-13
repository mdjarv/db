// Package resultview implements the query result display pane.
package resultview

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/tui/core"
)

const defaultSeparator = ","

// Model is the result view state.
type Model struct {
	table     table.Model
	focused   bool
	width     int
	height    int
	visual    bool
	anchorRow int
	colSelect int // -1 = all columns
	separator string
}

// New creates a result view with stub data.
func New() *Model {
	cols := []table.Column{
		{Title: "id", Width: 4},
		{Title: "name", Width: 16},
		{Title: "email", Width: 24},
		{Title: "active", Width: 6},
	}
	rows := []table.Row{
		{"1", "Alice Johnson", "alice@example.com", "true"},
		{"2", "Bob Smith", "bob@example.com", "true"},
		{"3", "Carol White", "carol@example.com", "false"},
		{"4", "Dave Brown", "dave@example.com", "true"},
		{"5", "Eve Davis", "eve@example.com", "true"},
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(false),
		table.WithHeight(5),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57"))
	t.SetStyles(s)

	return &Model{table: t, colSelect: -1, separator: defaultSeparator}
}

// EnterVisual starts visual selection from the current cursor row.
func (m *Model) EnterVisual() {
	m.visual = true
	m.anchorRow = m.table.Cursor()
	m.colSelect = -1
}

// ExitVisual cancels visual selection.
func (m *Model) ExitVisual() {
	m.visual = false
}

// Update handles input messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if !m.focused {
		return nil
	}

	if m.visual {
		return m.updateVisual(msg)
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return cmd
}

func (m *Model) updateVisual(msg tea.Msg) tea.Cmd {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	cols := m.table.Columns()
	switch km.String() {
	case "j", "down":
		m.table.MoveDown(1)
	case "k", "up":
		m.table.MoveUp(1)
	case "h", "left":
		if m.colSelect > 0 {
			m.colSelect--
		}
	case "l", "right":
		if m.colSelect < len(cols)-1 {
			if m.colSelect == -1 {
				m.colSelect = 0
			} else {
				m.colSelect++
			}
		}
	case "H":
		m.colSelect = -1
	case "y":
		content := m.yankContent()
		m.visual = false
		return func() tea.Msg { return core.YankMsg{Content: content} }
	}
	return nil
}

func (m *Model) yankContent() string {
	rows := m.table.Rows()
	cols := m.table.Columns()
	startRow, endRow := m.selectionRange()

	var sb strings.Builder

	if m.colSelect == -1 {
		// header
		for i, col := range cols {
			if i > 0 {
				sb.WriteString(m.separator)
			}
			sb.WriteString(col.Title)
		}
		sb.WriteByte('\n')
		// rows
		for i := startRow; i <= endRow; i++ {
			for j, val := range rows[i] {
				if j > 0 {
					sb.WriteString(m.separator)
				}
				sb.WriteString(m.escapeCSV(val))
			}
			if i < endRow {
				sb.WriteByte('\n')
			}
		}
	} else {
		// single column
		sb.WriteString(cols[m.colSelect].Title)
		sb.WriteByte('\n')
		for i := startRow; i <= endRow; i++ {
			sb.WriteString(m.escapeCSV(rows[i][m.colSelect]))
			if i < endRow {
				sb.WriteByte('\n')
			}
		}
	}
	return sb.String()
}

func (m *Model) escapeCSV(val string) string {
	if strings.ContainsAny(val, m.separator+"\"\n") {
		return "\"" + strings.ReplaceAll(val, "\"", "\"\"") + "\""
	}
	return val
}

func (m *Model) selectionRange() (int, int) {
	cursor := m.table.Cursor()
	if m.anchorRow <= cursor {
		return m.anchorRow, cursor
	}
	return cursor, m.anchorRow
}

// View renders the result table.
func (m *Model) View() string {
	borderColor := lipgloss.Color("240")
	if m.focused {
		borderColor = lipgloss.Color("62")
	}
	if m.visual {
		borderColor = lipgloss.Color("208")
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(m.width - 2).
		Height(m.height - 2)

	if m.visual {
		return style.Render(m.visualView())
	}
	return style.Render(m.table.View())
}

func (m *Model) visualView() string {
	cols := m.table.Columns()
	rows := m.table.Rows()
	startRow, endRow := m.selectionRange()

	highlight := lipgloss.NewStyle().
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("208"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	headerStyle := lipgloss.NewStyle().Bold(true)

	var sb strings.Builder

	// header
	for i, col := range cols {
		if i > 0 {
			sb.WriteString(" | ")
		}
		text := fmt.Sprintf("%-*s", col.Width, col.Title)
		if m.colSelect == -1 || m.colSelect == i {
			sb.WriteString(headerStyle.Render(text))
		} else {
			sb.WriteString(dimStyle.Render(text))
		}
	}
	sb.WriteByte('\n')

	// separator
	for i, col := range cols {
		if i > 0 {
			sb.WriteString("-+-")
		}
		sb.WriteString(strings.Repeat("-", col.Width))
	}
	sb.WriteByte('\n')

	// rows
	vh := max(m.height-6, 1)
	cursor := m.table.Cursor()
	offset := 0
	if cursor >= vh {
		offset = cursor - vh + 1
	}
	end := min(offset+vh, len(rows))

	for i := offset; i < end; i++ {
		inRange := i >= startRow && i <= endRow
		for j, val := range rows[i] {
			if j > 0 {
				sb.WriteString(" | ")
			}
			w := cols[j].Width
			if len(val) > w {
				val = val[:w]
			}
			text := fmt.Sprintf("%-*s", w, val)
			colSelected := m.colSelect == -1 || m.colSelect == j
			if inRange && colSelected {
				sb.WriteString(highlight.Render(text))
			} else if inRange {
				sb.WriteString(dimStyle.Render(text))
			} else {
				sb.WriteString(text)
			}
		}
		if i < end-1 {
			sb.WriteByte('\n')
		}
	}

	return sb.String()
}

// Focused returns whether the pane is focused.
func (m *Model) Focused() bool { return m.focused }

// SetFocused sets the focus state.
func (m *Model) SetFocused(f bool) {
	m.focused = f
	if f {
		m.table.Focus()
	} else {
		m.table.Blur()
	}
}

// SetSize sets the render dimensions.
func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.table.SetWidth(w - 2)
	m.table.SetHeight(max(h-4, 1))
}

// SetSeparator sets the CSV separator.
func (m *Model) SetSeparator(sep string) { m.separator = sep }
