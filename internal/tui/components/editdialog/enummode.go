package editdialog

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/tui/theme"
)

type enumMode struct {
	values []string
	cursor int
}

func newEnumMode(values []string, currentValue string) *enumMode {
	cursor := 0
	for i, v := range values {
		if v == currentValue {
			cursor = i
			break
		}
	}
	return &enumMode{
		values: values,
		cursor: cursor,
	}
}

func (m *enumMode) Update(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "j", "down":
		if m.cursor < len(m.values)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "g":
		m.cursor = 0
	case "G":
		m.cursor = len(m.values) - 1
	}
	return nil
}

func (m *enumMode) View(contentW int, focused bool, t theme.Styles) string {
	inputBg := lipgloss.Color("236")
	if focused {
		inputBg = lipgloss.Color("238")
	}
	inputStyle := lipgloss.NewStyle().
		Background(inputBg).
		Width(contentW).
		Padding(0, 1)

	var sb strings.Builder
	for i, v := range m.values {
		if i > 0 {
			sb.WriteByte('\n')
		}
		if focused && i == m.cursor {
			cursor := lipgloss.NewStyle().
				Foreground(lipgloss.Color("232")).
				Background(t.BorderFocused).
				Render(v)
			sb.WriteString(cursor)
		} else {
			sb.WriteString(v)
		}
	}

	return inputStyle.Render(sb.String())
}

func (m *enumMode) Value() string {
	if len(m.values) == 0 {
		return ""
	}
	return m.values[m.cursor]
}

func (m *enumMode) Hint() string {
	return "j/k navigate"
}

func (m *enumMode) SubmitsOnEnter() bool {
	return true
}
