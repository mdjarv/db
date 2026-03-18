package dialog

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/tui/theme"
)

// ProgressCancelMsg signals the user cancelled the progress operation.
type ProgressCancelMsg struct{}

// ProgressModel is a non-interactive modal showing operation progress.
type ProgressModel struct {
	active  bool
	title   string
	object  string
	current int
	total   int
	started time.Time
}

// NewProgress creates an inactive progress modal.
func NewProgress() *ProgressModel {
	return &ProgressModel{}
}

// IsActive returns whether the progress modal is visible.
func (m *ProgressModel) IsActive() bool { return m.active }

// Open shows the progress modal with a title.
func (m *ProgressModel) Open(title string) {
	m.active = true
	m.title = title
	m.object = ""
	m.current = 0
	m.total = 0
	m.started = time.Now()
}

// Close dismisses the progress modal.
func (m *ProgressModel) Close() {
	m.active = false
}

// SetProgress updates the displayed progress.
func (m *ProgressModel) SetProgress(object string, current, total int) {
	m.object = object
	m.current = current
	m.total = total
}

// Update handles key input (only Esc for cancellation).
func (m *ProgressModel) Update(msg tea.KeyMsg) tea.Cmd {
	if !m.active {
		return nil
	}
	if msg.String() == "esc" {
		m.Close()
		return func() tea.Msg { return ProgressCancelMsg{} }
	}
	return nil
}

// View renders the progress modal.
func (m *ProgressModel) View(containerW, containerH int) string {
	if !m.active {
		return ""
	}

	t := theme.Current()
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(t.Styles.BorderFocused)
	bodyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	w := min(containerW-4, 50)
	if w < 30 {
		w = 30
	}

	barWidth := w - 12 // room for count label
	if barWidth < 10 {
		barWidth = 10
	}

	var sb strings.Builder
	sb.WriteString(titleStyle.Render(m.title))
	sb.WriteByte('\n')
	sb.WriteByte('\n')

	if m.object != "" {
		sb.WriteString(bodyStyle.Render(m.object))
		sb.WriteByte('\n')
	}

	// progress bar
	filled := 0
	if m.total > 0 {
		filled = m.current * barWidth / m.total
		if filled > barWidth {
			filled = barWidth
		}
	}
	bar := strings.Repeat("\u2588", filled) + strings.Repeat("\u2591", barWidth-filled)
	countLabel := ""
	if m.total > 0 {
		countLabel = fmt.Sprintf(" %d/%d", m.current, m.total)
	} else if m.current > 0 {
		countLabel = fmt.Sprintf(" %d", m.current)
	}
	sb.WriteString(bodyStyle.Render("[" + bar + "]" + countLabel))
	sb.WriteByte('\n')

	// elapsed time
	elapsed := time.Since(m.started).Truncate(time.Second)
	sb.WriteString(hintStyle.Render(fmt.Sprintf("elapsed: %s", elapsed)))
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteString(hintStyle.Render("Esc cancel"))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Styles.BorderFocused).
		Padding(1, 2).
		Width(w).
		Render(sb.String())

	return lipgloss.Place(containerW, containerH, lipgloss.Center, lipgloss.Center, box)
}
