// Package contextmenu provides a reusable context menu overlay component.
package contextmenu

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/tui/theme"
)

// MenuItem defines a single menu entry.
type MenuItem struct {
	Label    string
	ActionID string
	Hint     string // optional, appended as "..." suffix
}

// Model is the context menu state.
type Model struct {
	items  []MenuItem
	cursor int
	active bool
}

// New creates an inactive context menu.
func New() *Model {
	return &Model{}
}

// Open displays the context menu with the given items.
func (m *Model) Open(items []MenuItem) {
	m.items = items
	m.cursor = 0
	m.active = true
}

// Close dismisses the context menu.
func (m *Model) Close() {
	m.active = false
	m.items = nil
	m.cursor = 0
}

// IsActive returns whether the menu is visible.
func (m *Model) IsActive() bool { return m.active }

// Cursor returns the current cursor position (for testing).
func (m *Model) Cursor() int { return m.cursor }

// Update handles key input. Returns (actionID, selected).
func (m *Model) Update(msg tea.KeyMsg) (string, bool) {
	if !m.active || len(m.items) == 0 {
		return "", false
	}
	switch msg.String() {
	case "j", "down":
		m.cursor = (m.cursor + 1) % len(m.items)
	case "k", "up":
		m.cursor = (m.cursor - 1 + len(m.items)) % len(m.items)
	case "enter":
		id := m.items[m.cursor].ActionID
		m.Close()
		return id, true
	case "esc", "q":
		m.Close()
	}
	return "", false
}

// View renders the context menu as a centered overlay box.
func (m *Model) View(containerW, containerH int) string {
	if !m.active || len(m.items) == 0 {
		return ""
	}

	t := theme.Current()
	cursorStyle := t.Styles.Cursor
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	var lines []string
	for i, item := range m.items {
		label := item.Label
		if item.Hint != "" {
			label += hintStyle.Render(item.Hint)
		}
		if i == m.cursor {
			lines = append(lines, cursorStyle.Render(" "+item.Label+" ")+hintSuffix(item, hintStyle))
		} else {
			lines = append(lines, normalStyle.Render(" "+label+" "))
		}
	}

	w := min(containerW-4, 40)
	if w < 20 {
		w = 20
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Styles.BorderFocused).
		Padding(0, 1).
		Width(w).
		Render(strings.Join(lines, "\n"))

	return lipgloss.Place(containerW, containerH, lipgloss.Center, lipgloss.Center, box)
}

func hintSuffix(item MenuItem, style lipgloss.Style) string {
	if item.Hint == "" {
		return ""
	}
	return style.Render(item.Hint)
}
