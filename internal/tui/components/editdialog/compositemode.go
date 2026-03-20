package editdialog

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/tui/theme"
)

type compositeMode struct {
	fields   []compositeFieldState
	cursor   int
	editing  bool
	editMode inputMode
}

type compositeFieldState struct {
	name     string
	typeName string
	value    string
}

func newCompositeMode(fields []CompositeField, value string) *compositeMode {
	names := make([]string, len(fields))
	for i, f := range fields {
		names[i] = f.Name
	}

	vals := parseComposite(value, names)
	states := make([]compositeFieldState, len(fields))
	for i, f := range fields {
		states[i] = compositeFieldState{
			name:     f.Name,
			typeName: f.TypeName,
			value:    vals[f.Name],
		}
	}
	return &compositeMode{fields: states}
}

func (m *compositeMode) Update(msg tea.KeyMsg) tea.Cmd {
	if m.editing {
		return m.updateEditing(msg)
	}
	return m.updateNormal(msg)
}

func (m *compositeMode) updateEditing(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEnter:
		m.fields[m.cursor].value = m.editMode.Value()
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

func (m *compositeMode) updateNormal(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "j", "down", "tab":
		if m.cursor < len(m.fields)-1 {
			m.cursor++
		}
	case "k", "up", "shift+tab":
		if m.cursor > 0 {
			m.cursor--
		}
	case "e", "enter":
		if len(m.fields) > 0 {
			m.startEdit()
		}
	case "g":
		m.cursor = 0
	case "G":
		if len(m.fields) > 0 {
			m.cursor = len(m.fields) - 1
		}
	}
	return nil
}

func (m *compositeMode) startEdit() {
	f := m.fields[m.cursor]
	m.editing = true
	m.editMode = m.fieldEditor(f.typeName, f.value)
}

func (m *compositeMode) fieldEditor(typeName, value string) inputMode {
	var filter charFilter
	switch typeName {
	case "int2", "int4", "int8":
		filter = intFilter
	case "float4", "float8", "numeric", "money":
		filter = floatFilter
	}
	return newTextMode(value, filter, "", false)
}

func (m *compositeMode) View(contentW int, focused bool, t theme.Styles) string {
	if len(m.fields) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("(no fields)")
	}

	// label width from longest name
	maxName := 0
	for _, f := range m.fields {
		if len(f.name) > maxName {
			maxName = len(f.name)
		}
	}
	labelW := maxName + 2
	inputW := max(contentW-labelW, 10)

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Width(labelW)

	var lines []string
	for i, f := range m.fields {
		isCursor := focused && i == m.cursor

		label := labelStyle.Render(f.name)

		inputBg := lipgloss.Color("236")
		if isCursor {
			inputBg = lipgloss.Color("238")
		}

		var valueStr string
		if m.editing && i == m.cursor {
			valueStr = m.editMode.(*textMode).inlineView(focused, t)
		} else if isCursor {
			valueStr = lipgloss.NewStyle().
				Foreground(lipgloss.Color("232")).
				Background(t.BorderFocused).
				Render(f.value)
		} else {
			valueStr = f.value
		}

		input := lipgloss.NewStyle().
			Background(inputBg).
			Width(inputW).
			Padding(0, 1).
			Render(valueStr)

		lines = append(lines, label+input)
	}

	return strings.Join(lines, "\n")
}

func (m *compositeMode) Value() string {
	fields := make([]CompositeField, len(m.fields))
	vals := make(map[string]string, len(m.fields))
	for i, f := range m.fields {
		fields[i] = CompositeField{Name: f.name, TypeName: f.typeName}
		vals[f.name] = f.value
	}
	return formatComposite(fields, vals)
}

func (m *compositeMode) Hint() string {
	if m.editing {
		return "Enter confirm | Esc cancel"
	}
	return "j/k nav | e edit field"
}

func (m *compositeMode) SubmitsOnEnter() bool {
	return false
}
