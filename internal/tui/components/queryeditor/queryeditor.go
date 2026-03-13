// Package queryeditor implements the SQL query editor pane.
package queryeditor

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/tui/core"
)

// Model is the query editor state.
type Model struct {
	lines   []string
	cursorX int
	cursorY int
	focused bool
	width   int
	height  int
	offset  int
}

// New creates a query editor with stub content.
func New() *Model {
	return &Model{
		lines: []string{"SELECT * FROM users LIMIT 10;"},
	}
}

// Update handles input messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if !m.focused {
		return nil
	}
	switch msg := msg.(type) {
	case core.ModeChangedMsg:
		_ = msg
	case tea.KeyMsg:
		return m.normalUpdate(msg)
	}
	return nil
}

func (m *Model) normalUpdate(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "j", "down":
		if m.cursorY < len(m.lines)-1 {
			m.cursorY++
			m.clampX()
		}
	case "k", "up":
		if m.cursorY > 0 {
			m.cursorY--
			m.clampX()
		}
	case "h", "left":
		if m.cursorX > 0 {
			m.cursorX--
		}
	case "l", "right":
		if m.cursorY < len(m.lines) && m.cursorX < len(m.lines[m.cursorY]) {
			m.cursorX++
		}
	case "0":
		m.cursorX = 0
	case "$":
		if m.cursorY < len(m.lines) {
			m.cursorX = len(m.lines[m.cursorY])
		}
	}
	return nil
}

func (m *Model) clampX() {
	if m.cursorY < len(m.lines) && m.cursorX > len(m.lines[m.cursorY]) {
		m.cursorX = len(m.lines[m.cursorY])
	}
}

// View renders the query editor.
func (m *Model) View() string {
	var sb strings.Builder
	vh := max(m.height-2, 1)
	end := min(m.offset+vh, len(m.lines))

	for i := m.offset; i < end; i++ {
		sb.WriteString(m.lines[i])
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

// Content returns the full editor text.
func (m *Model) Content() string {
	return strings.Join(m.lines, "\n")
}

// Focused returns whether the pane is focused.
func (m *Model) Focused() bool { return m.focused }

// SetFocused sets the focus state.
func (m *Model) SetFocused(f bool) { m.focused = f }

// SetSize sets the render dimensions.
func (m *Model) SetSize(w, h int) { m.width = w; m.height = h }
