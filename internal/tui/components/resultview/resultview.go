// Package resultview implements the query result display pane.
package resultview

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type column struct {
	name  string
	width int
}

// Model is the result view state.
type Model struct {
	columns []column
	rows    [][]string
	cursor  int
	focused bool
	width   int
	height  int
	offset  int
}

// New creates a result view with stub data.
func New() *Model {
	cols := []column{
		{name: "id", width: 4},
		{name: "name", width: 16},
		{name: "email", width: 24},
		{name: "active", width: 6},
	}
	rows := [][]string{
		{"1", "Alice Johnson", "alice@example.com", "true"},
		{"2", "Bob Smith", "bob@example.com", "true"},
		{"3", "Carol White", "carol@example.com", "false"},
		{"4", "Dave Brown", "dave@example.com", "true"},
		{"5", "Eve Davis", "eve@example.com", "true"},
	}
	return &Model{columns: cols, rows: rows}
}

// Update handles input messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if !m.focused {
		return nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.rows)-1 {
				m.cursor++
				if m.cursor-m.offset >= m.viewHeight() {
					m.offset++
				}
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.offset {
					m.offset = m.cursor
				}
			}
		case "g":
			m.cursor = 0
			m.offset = 0
		case "G":
			m.cursor = len(m.rows) - 1
			vh := m.viewHeight()
			if len(m.rows) > vh {
				m.offset = len(m.rows) - vh
			}
		}
	}
	return nil
}

func (m *Model) viewHeight() int {
	return max(m.height-4, 1)
}

// View renders the result table.
func (m *Model) View() string {
	var sb strings.Builder

	for i, col := range m.columns {
		if i > 0 {
			sb.WriteString(" | ")
		}
		fmt.Fprintf(&sb, "%-*s", col.width, col.name)
	}
	sb.WriteByte('\n')

	for i, col := range m.columns {
		if i > 0 {
			sb.WriteString("-+-")
		}
		sb.WriteString(strings.Repeat("-", col.width))
	}
	sb.WriteByte('\n')

	vh := m.viewHeight()
	end := min(m.offset+vh, len(m.rows))

	highlight := lipgloss.NewStyle().Reverse(true)

	for i := m.offset; i < end; i++ {
		var line strings.Builder
		for j, val := range m.rows[i] {
			if j > 0 {
				line.WriteString(" | ")
			}
			w := m.columns[j].width
			if len(val) > w {
				val = val[:w]
			}
			fmt.Fprintf(&line, "%-*s", w, val)
		}
		row := line.String()
		if i == m.cursor && m.focused {
			row = highlight.Render(row)
		}
		sb.WriteString(row)
		if i < end-1 {
			sb.WriteByte('\n')
		}
	}

	borderColor := lipgloss.Color("240")
	if m.focused {
		borderColor = lipgloss.Color("62")
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(m.width - 2).
		Height(m.height - 2)

	return style.Render(sb.String())
}

// Focused returns whether the pane is focused.
func (m *Model) Focused() bool { return m.focused }

// SetFocused sets the focus state.
func (m *Model) SetFocused(f bool) { m.focused = f }

// SetSize sets the render dimensions.
func (m *Model) SetSize(w, h int) { m.width = w; m.height = h }
