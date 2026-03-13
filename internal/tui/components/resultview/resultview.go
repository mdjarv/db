// Package resultview implements the query result display pane.
package resultview

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model is the result view state.
type Model struct {
	table   table.Model
	focused bool
	width   int
	height  int
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

	return &Model{table: t}
}

// Update handles input messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if !m.focused {
		return nil
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return cmd
}

// View renders the result table.
func (m *Model) View() string {
	borderColor := lipgloss.Color("240")
	if m.focused {
		borderColor = lipgloss.Color("62")
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(m.width - 2).
		Height(m.height - 2)

	return style.Render(m.table.View())
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
