// Package tablelist implements the table list pane.
package tablelist

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var stubTables = []string{
	"users",
	"orders",
	"products",
	"categories",
	"reviews",
	"inventory",
}

// Model is the table list state.
type Model struct {
	tables  []string
	cursor  int
	focused bool
	width   int
	height  int
	offset  int
}

// New creates a table list with stub data.
func New() *Model {
	return &Model{
		tables: stubTables,
	}
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
			if m.cursor < len(m.tables)-1 {
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
			m.cursor = len(m.tables) - 1
			vh := m.viewHeight()
			if len(m.tables) > vh {
				m.offset = len(m.tables) - vh
			}
		}
	}
	return nil
}

func (m *Model) viewHeight() int {
	return max(m.height-2, 1)
}

// View renders the table list.
func (m *Model) View() string {
	var sb strings.Builder
	vh := m.viewHeight()
	end := min(m.offset+vh, len(m.tables))

	for i := m.offset; i < end; i++ {
		if i == m.cursor {
			fmt.Fprintf(&sb, " > %s", m.tables[i])
		} else {
			fmt.Fprintf(&sb, "   %s", m.tables[i])
		}
		if i < end-1 {
			sb.WriteByte('\n')
		}
	}

	content := sb.String()
	borderColor := lipgloss.Color("240")
	if m.focused {
		borderColor = lipgloss.Color("62")
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(m.width - 2).
		Height(m.height - 2)

	return style.Render(content)
}

// Focused returns whether the pane is focused.
func (m *Model) Focused() bool { return m.focused }

// SetFocused sets the focus state.
func (m *Model) SetFocused(f bool) { m.focused = f }

// SetSize sets the render dimensions.
func (m *Model) SetSize(w, h int) { m.width = w; m.height = h }
