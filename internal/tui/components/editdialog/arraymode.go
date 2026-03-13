package editdialog

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/tui/theme"
)

type arrayMode struct {
	elements     []string
	cursor       int
	editing      bool
	editMode     inputMode
	baseTypeName string
	enumValues   []string
	pending      string
}

func newArrayMode(typeName, value string, enumValues []string) *arrayMode {
	base := strings.TrimSuffix(typeName, "[]")
	elems := parseArray(value)
	if elems == nil {
		elems = []string{}
	}
	return &arrayMode{
		elements:     elems,
		baseTypeName: base,
		enumValues:   enumValues,
	}
}

func (m *arrayMode) Update(msg tea.KeyMsg) tea.Cmd {
	if m.editing {
		return m.updateEditing(msg)
	}
	return m.updateNormal(msg)
}

func (m *arrayMode) updateEditing(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEnter:
		m.elements[m.cursor] = m.editMode.Value()
		m.editing = false
		m.editMode = nil
		return nil
	case tea.KeyEsc:
		m.editing = false
		m.editMode = nil
		return nil
	}
	return m.editMode.Update(msg)
}

func (m *arrayMode) updateNormal(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	// handle pending "dd" sequence
	if m.pending == "d" {
		m.pending = ""
		if key == "d" && len(m.elements) > 0 {
			m.elements = append(m.elements[:m.cursor], m.elements[m.cursor+1:]...)
			if m.cursor >= len(m.elements) && m.cursor > 0 {
				m.cursor--
			}
		}
		return nil
	}

	switch key {
	case "j", "down":
		if m.cursor < len(m.elements)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "g":
		m.cursor = 0
	case "G":
		if len(m.elements) > 0 {
			m.cursor = len(m.elements) - 1
		}
	case "e", "enter":
		if len(m.elements) > 0 {
			m.startEdit(m.elements[m.cursor])
		}
	case "a":
		// insert new empty element after cursor
		idx := m.cursor + 1
		if len(m.elements) == 0 {
			idx = 0
		}
		m.elements = append(m.elements[:idx], append([]string{""}, m.elements[idx:]...)...)
		m.cursor = idx
		m.startEdit("")
	case "J":
		if m.cursor < len(m.elements)-1 {
			m.elements[m.cursor], m.elements[m.cursor+1] = m.elements[m.cursor+1], m.elements[m.cursor]
			m.cursor++
		}
	case "K":
		if m.cursor > 0 {
			m.elements[m.cursor], m.elements[m.cursor-1] = m.elements[m.cursor-1], m.elements[m.cursor]
			m.cursor--
		}
	case "d":
		m.pending = "d"
	}
	return nil
}

func (m *arrayMode) startEdit(value string) {
	m.editing = true
	m.editMode = m.elementEditor(value)
}

func (m *arrayMode) elementEditor(value string) inputMode {
	if m.enumValues != nil {
		return newEnumMode(m.enumValues, value)
	}
	var filter charFilter
	switch m.baseTypeName {
	case "int2", "int4", "int8":
		filter = intFilter
	case "float4", "float8", "numeric", "money":
		filter = floatFilter
	}
	return newTextMode(value, filter, "", false)
}

func (m *arrayMode) View(contentW int, focused bool, t theme.Styles) string {
	inputBg := lipgloss.Color("236")
	if focused {
		inputBg = lipgloss.Color("238")
	}
	inputStyle := lipgloss.NewStyle().
		Background(inputBg).
		Width(contentW).
		Padding(0, 1)

	if len(m.elements) == 0 {
		return inputStyle.Render(lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("(empty array)"))
	}

	var sb strings.Builder
	for i, e := range m.elements {
		if i > 0 {
			sb.WriteByte('\n')
		}

		prefix := fmt.Sprintf("[%d] ", i)

		if m.editing && i == m.cursor {
			if tm, ok := m.editMode.(*textMode); ok {
				sb.WriteString(prefix)
				sb.WriteString(tm.inlineView(focused, t))
			} else {
				// enum element editor — render full picker below the prefix
				sb.WriteString(prefix + "...\n")
				sb.WriteString(m.editMode.View(contentW-4, focused, t))
			}
			continue
		}

		if focused && i == m.cursor {
			line := lipgloss.NewStyle().
				Foreground(lipgloss.Color("232")).
				Background(t.BorderFocused).
				Render(prefix + e)
			sb.WriteString(line)
		} else {
			sb.WriteString(prefix + e)
		}
	}

	return inputStyle.Render(sb.String())
}

func (m *arrayMode) Value() string {
	return formatArray(m.elements)
}

func (m *arrayMode) Hint() string {
	if m.editing {
		return "Enter select | Esc cancel"
	}
	return "j/k nav | J/K move | e edit | a add | dd delete"
}

func (m *arrayMode) SubmitsOnEnter() bool {
	return false
}

// inlineView renders the text cursor without the background wrapper.
func (m *textMode) inlineView(focused bool, t theme.Styles) string {
	line := m.lines[0]
	if !focused {
		return line
	}
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
	return before + cursor + after
}
