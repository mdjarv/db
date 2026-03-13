// Package dialog provides a reusable confirmation dialog component.
package dialog

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("196")).
			Padding(1, 2)
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	bodyStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	hintStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	confirmStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("28"))
	cancelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// ResultMsg carries the dialog result.
type ResultMsg struct {
	Action    string
	Confirmed bool
}

// Model is the dialog state.
type Model struct {
	active bool
	action string
	title  string
	body   string
	width  int
}

// New creates an inactive dialog.
func New() *Model {
	return &Model{}
}

// IsActive returns whether the dialog is visible.
func (m *Model) IsActive() bool { return m.active }

// Open shows a confirmation dialog.
func (m *Model) Open(action, title, body string) {
	m.active = true
	m.action = action
	m.title = title
	m.body = body
}

// Close dismisses the dialog.
func (m *Model) Close() {
	m.active = false
}

// SetWidth sets the render width.
func (m *Model) SetWidth(w int) { m.width = w }

// Update handles key input.
func (m *Model) Update(msg tea.KeyMsg) tea.Cmd {
	if !m.active {
		return nil
	}
	switch msg.String() {
	case "y", "Y", "enter":
		m.Close()
		action := m.action
		return func() tea.Msg { return ResultMsg{Action: action, Confirmed: true} }
	case "n", "N", "esc", "q":
		m.Close()
		action := m.action
		return func() tea.Msg { return ResultMsg{Action: action, Confirmed: false} }
	}
	return nil
}

// View renders the dialog.
func (m *Model) View(containerW, containerH int) string {
	if !m.active {
		return ""
	}

	w := min(containerW-4, 50)
	if w < 20 {
		w = 20
	}

	var sb strings.Builder
	sb.WriteString(titleStyle.Render(m.title))
	sb.WriteByte('\n')
	if m.body != "" {
		sb.WriteByte('\n')
		sb.WriteString(bodyStyle.Render(m.body))
		sb.WriteByte('\n')
	}
	sb.WriteByte('\n')
	sb.WriteString(confirmStyle.Render("[y]es") + "  " + cancelStyle.Render("[n]o"))
	sb.WriteByte('\n')
	sb.WriteString(hintStyle.Render("Enter to confirm, Esc to cancel"))

	box := borderStyle.Width(w).Render(sb.String())
	return lipgloss.Place(containerW, containerH, lipgloss.Center, lipgloss.Center, box)
}
