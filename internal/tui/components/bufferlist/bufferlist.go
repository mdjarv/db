// Package bufferlist provides a modal overlay listing query buffers.
package bufferlist

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/tui/theme"
)

// BufferInfo holds display data for a single buffer.
type BufferInfo struct {
	Index    int
	Query    string
	Active   bool
	Modified bool
}

// Model holds the buffer list overlay state.
type Model struct {
	active  bool
	buffers []BufferInfo
}

// New creates an inactive buffer list overlay.
func New() *Model {
	return &Model{}
}

// IsActive returns whether the overlay is visible.
func (m *Model) IsActive() bool { return m.active }

// Open shows the buffer list overlay with the given buffer info.
func (m *Model) Open(buffers []BufferInfo) {
	m.active = true
	m.buffers = buffers
}

// Close dismisses the overlay.
func (m *Model) Close() {
	m.active = false
	m.buffers = nil
}

// Update handles key input — any key dismisses.
func (m *Model) Update(_ tea.KeyMsg) tea.Cmd {
	if !m.active {
		return nil
	}
	m.Close()
	return nil
}

// View renders the buffer list overlay centered in the given dimensions.
func (m *Model) View(containerW, containerH int) string {
	if !m.active {
		return ""
	}

	t := theme.Current()
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(t.Styles.BorderFocused)
	numStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("215")).Bold(true)
	queryStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	modStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	var lines []string
	lines = append(lines, titleStyle.Render("Buffers"))
	lines = append(lines, "")

	for _, b := range m.buffers {
		marker := "  "
		if b.Active {
			marker = "> "
		}

		mod := " "
		if b.Modified {
			mod = "+"
		}

		query := b.Query
		if len(query) > 40 {
			query = query[:40] + "..."
		}
		if query == "" {
			query = "[empty]"
		}
		// replace newlines for single-line display
		query = strings.ReplaceAll(query, "\n", " ")

		num := fmt.Sprintf("%d", b.Index)
		var line string
		if b.Active {
			line = activeStyle.Render(marker) +
				activeStyle.Render(num) + " " +
				modStyle.Render(mod) + " " +
				activeStyle.Render(query)
		} else {
			line = marker +
				numStyle.Render(num) + " " +
				modStyle.Render(mod) + " " +
				queryStyle.Render(query)
		}
		lines = append(lines, line)
	}

	lines = append(lines, "")
	lines = append(lines, hintStyle.Render("any key to dismiss"))

	boxW := min(containerW-4, 58)
	if boxW < 30 {
		boxW = 30
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Styles.BorderFocused).
		Padding(1, 2).
		Width(boxW).
		Render(strings.Join(lines, "\n"))

	return lipgloss.Place(containerW, containerH, lipgloss.Center, lipgloss.Center, box)
}
