package resultview

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	inspectorBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)
	inspectorTitle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("62"))
	inspectorType  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// Inspector shows the full cell value in an overlay.
type Inspector struct {
	active   bool
	column   string
	typeName string
	value    string
	scroll   int
}

// IsActive returns whether the inspector is visible.
func (ins *Inspector) IsActive() bool { return ins.active }

// Open shows the inspector for a cell.
func (ins *Inspector) Open(column, typeName, value string) {
	ins.active = true
	ins.column = column
	ins.typeName = typeName
	ins.value = value
	ins.scroll = 0
}

// Close hides the inspector.
func (ins *Inspector) Close() {
	ins.active = false
}

// Update handles input while the inspector is active.
func (ins *Inspector) Update(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q", "enter":
		ins.Close()
	case "j", "down":
		ins.scroll++
	case "k", "up":
		if ins.scroll > 0 {
			ins.scroll--
		}
	}
	return nil
}

// View renders the inspector overlay.
func (ins *Inspector) View(containerW, containerH int) string {
	w := min(containerW-4, 60)
	h := min(containerH-4, 20)
	if w < 10 || h < 5 {
		return ""
	}

	header := inspectorTitle.Render(ins.column)
	if ins.typeName != "" {
		header += " " + inspectorType.Render(ins.typeName)
	}

	contentW := w - 6 // padding + border
	lines := wrapText(ins.value, contentW)

	viewH := h - 6 // border + padding + header + blank line
	if viewH < 1 {
		viewH = 1
	}

	if ins.scroll > len(lines)-viewH {
		ins.scroll = max(len(lines)-viewH, 0)
	}

	end := min(ins.scroll+viewH, len(lines))
	visible := lines[ins.scroll:end]

	body := header + "\n\n" + strings.Join(visible, "\n")

	style := inspectorBorder.
		Width(w).
		Height(h)

	overlay := style.Render(body)

	return lipgloss.Place(containerW, containerH, lipgloss.Center, lipgloss.Center, overlay)
}

func wrapText(s string, width int) []string {
	if width < 1 {
		width = 1
	}
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		if len(line) == 0 {
			lines = append(lines, "")
			continue
		}
		for len(line) > width {
			lines = append(lines, line[:width])
			line = line[width:]
		}
		lines = append(lines, fmt.Sprintf("%-*s", width, line))
	}
	return lines
}
