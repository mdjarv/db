// Package dialog provides a reusable confirmation dialog component.
package dialog

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/tui/theme"
)

// ResultMsg carries the dialog result.
type ResultMsg struct {
	Action    string
	Confirmed bool
}

const (
	focusConfirm = 0
	focusCancel  = 1
)

// Model is the dialog state.
type Model struct {
	active bool
	action string
	title  string
	body   string
	focus  int
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
	m.focus = focusConfirm
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
	case "tab", "right", "l", "left", "h":
		if m.focus == focusConfirm {
			m.focus = focusCancel
		} else {
			m.focus = focusConfirm
		}
	case "enter":
		m.Close()
		action := m.action
		confirmed := m.focus == focusConfirm
		return func() tea.Msg { return ResultMsg{Action: action, Confirmed: confirmed} }
	case "esc", "q":
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

	t := theme.Current()
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(t.Styles.BorderFocused)
	bodyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

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

	confirmLabel := "[ Confirm ]"
	cancelLabel := "[ Cancel ]"
	if m.focus == focusConfirm {
		confirmLabel = lipgloss.NewStyle().Bold(true).Reverse(true).Render(confirmLabel)
		cancelLabel = hintStyle.Render(cancelLabel)
	} else {
		confirmLabel = hintStyle.Render(confirmLabel)
		cancelLabel = lipgloss.NewStyle().Bold(true).Reverse(true).Render(cancelLabel)
	}
	sb.WriteString(confirmLabel + "  " + cancelLabel)

	sb.WriteByte('\n')
	sb.WriteString(hintStyle.Render("Tab switch  Enter select  Esc cancel"))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Styles.BorderFocused).
		Padding(1, 2).
		Width(w).
		Render(sb.String())

	return lipgloss.Place(containerW, containerH, lipgloss.Center, lipgloss.Center, box)
}
