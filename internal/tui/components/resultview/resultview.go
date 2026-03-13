// Package resultview wraps the table component for the result display pane.
package resultview

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/tui/components/table"
	"github.com/mdjarv/db/internal/tui/core"
)

const defaultSeparator = ","

// Model is the result view state.
type Model struct {
	table     table.Model
	focused   bool
	width     int
	height    int
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
	rows := [][]string{
		{"1", "Alice Johnson", "alice@example.com", "true"},
		{"2", "Bob Smith", "bob@example.com", "true"},
		{"3", "Carol White", "carol@example.com", "false"},
		{"4", "Dave Brown", "dave@example.com", "true"},
		{"5", "Eve Davis", "eve@example.com", "true"},
	}
	return &Model{
		table:     table.New(cols, rows),
		separator: defaultSeparator,
	}
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
	case "y":
		content := m.table.YankCell()
		return func() tea.Msg { return core.YankMsg{Content: content} }
	case "Y":
		content := m.table.YankRow(m.separator)
		return func() tea.Msg { return core.YankMsg{Content: content} }
	}
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

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(m.width - 2).
		Height(m.height - 2)

	return style.Render(m.table.View(m.focused))
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
