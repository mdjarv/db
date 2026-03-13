package editdialog

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/tui/theme"
)

type textMode struct {
	lines     []string
	cursorRow int
	cursorCol int
	filter    charFilter
	hint      string
	multiline bool
}

func newTextMode(value string, filter charFilter, hint string, multiline bool) *textMode {
	lines := strings.Split(value, "\n")
	return &textMode{
		lines:     lines,
		cursorRow: 0,
		cursorCol: len(lines[0]),
		filter:    filter,
		hint:      hint,
		multiline: multiline,
	}
}

func (m *textMode) Update(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyCtrlJ:
		if m.multiline {
			m.insertNewline()
		}

	case tea.KeyBackspace:
		m.backspace()

	case tea.KeyDelete:
		m.delete()

	case tea.KeyLeft:
		m.moveLeft()

	case tea.KeyRight:
		m.moveRight()

	case tea.KeyUp:
		if m.multiline && m.cursorRow > 0 {
			m.cursorRow--
			m.cursorCol = min(m.cursorCol, len(m.lines[m.cursorRow]))
		}

	case tea.KeyDown:
		if m.multiline && m.cursorRow < len(m.lines)-1 {
			m.cursorRow++
			m.cursorCol = min(m.cursorCol, len(m.lines[m.cursorRow]))
		}

	case tea.KeyHome:
		m.cursorCol = 0

	case tea.KeyEnd:
		m.cursorCol = len(m.lines[m.cursorRow])

	case tea.KeySpace:
		m.insertRune(' ')

	case tea.KeyRunes:
		for _, r := range msg.Runes {
			m.insertRune(r)
		}
	}
	return nil
}

func (m *textMode) View(contentW int, focused bool, t theme.Styles) string {
	inputBg := lipgloss.Color("236")
	if focused {
		inputBg = lipgloss.Color("238")
	}
	inputStyle := lipgloss.NewStyle().
		Background(inputBg).
		Width(contentW).
		Padding(0, 1)

	var sb strings.Builder
	for i, line := range m.lines {
		if i > 0 {
			sb.WriteByte('\n')
		}
		if focused && i == m.cursorRow {
			col := min(m.cursorCol, len(line))
			before := line[:col]
			cursorChar := " "
			after := ""
			if col < len(line) {
				cursorChar = string(line[col])
				after = line[col+1:]
			}
			cursor := lipgloss.NewStyle().
				Foreground(lipgloss.Color("232")).
				Background(t.BorderFocused).
				Render(cursorChar)
			sb.WriteString(before)
			sb.WriteString(cursor)
			sb.WriteString(after)
		} else {
			sb.WriteString(line)
		}
	}

	return inputStyle.Render(sb.String())
}

func (m *textMode) Value() string {
	return strings.Join(m.lines, "\n")
}

func (m *textMode) Hint() string {
	return m.hint
}

func (m *textMode) SubmitsOnEnter() bool {
	return false
}

func (m *textMode) insertRune(r rune) {
	if m.filter != nil && !m.filter(r) {
		return
	}
	line := m.lines[m.cursorRow]
	m.lines[m.cursorRow] = line[:m.cursorCol] + string(r) + line[m.cursorCol:]
	m.cursorCol++
}

func (m *textMode) insertNewline() {
	line := m.lines[m.cursorRow]
	before := line[:m.cursorCol]
	after := line[m.cursorCol:]
	m.lines[m.cursorRow] = before
	rest := make([]string, len(m.lines)-m.cursorRow-1)
	copy(rest, m.lines[m.cursorRow+1:])
	m.lines = append(m.lines[:m.cursorRow+1], after)
	m.lines = append(m.lines, rest...)
	m.cursorRow++
	m.cursorCol = 0
}

func (m *textMode) backspace() {
	if m.cursorCol > 0 {
		line := m.lines[m.cursorRow]
		m.lines[m.cursorRow] = line[:m.cursorCol-1] + line[m.cursorCol:]
		m.cursorCol--
	} else if m.cursorRow > 0 {
		prev := m.lines[m.cursorRow-1]
		m.cursorCol = len(prev)
		m.lines[m.cursorRow-1] = prev + m.lines[m.cursorRow]
		m.lines = append(m.lines[:m.cursorRow], m.lines[m.cursorRow+1:]...)
		m.cursorRow--
	}
}

func (m *textMode) delete() {
	line := m.lines[m.cursorRow]
	if m.cursorCol < len(line) {
		m.lines[m.cursorRow] = line[:m.cursorCol] + line[m.cursorCol+1:]
	} else if m.cursorRow < len(m.lines)-1 {
		m.lines[m.cursorRow] = line + m.lines[m.cursorRow+1]
		m.lines = append(m.lines[:m.cursorRow+1], m.lines[m.cursorRow+2:]...)
	}
}

func (m *textMode) moveLeft() {
	if m.cursorCol > 0 {
		m.cursorCol--
	} else if m.cursorRow > 0 {
		m.cursorRow--
		m.cursorCol = len(m.lines[m.cursorRow])
	}
}

func (m *textMode) moveRight() {
	line := m.lines[m.cursorRow]
	if m.cursorCol < len(line) {
		m.cursorCol++
	} else if m.cursorRow < len(m.lines)-1 {
		m.cursorRow++
		m.cursorCol = 0
	}
}
