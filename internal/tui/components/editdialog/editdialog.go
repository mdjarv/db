// Package editdialog provides a popup dialog for editing cell values.
package editdialog

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/tui/theme"
)

// SubmitMsg carries the confirmed edit result.
type SubmitMsg struct {
	Row      int
	Col      int
	OldValue string
	NewValue string
	IsNull   bool
}

// CancelMsg signals the edit was cancelled.
type CancelMsg struct{}

type focus int

const (
	focusInput focus = iota
	focusOK
	focusNull
	focusCancel
)

// Model is the edit dialog state.
type Model struct {
	active   bool
	row      int
	col      int
	colName  string
	typeName string
	oldValue string
	nullable bool

	lines     []string
	cursorRow int
	cursorCol int

	focus focus
	width int
}

// New creates an inactive edit dialog.
func New() *Model {
	return &Model{}
}

// IsActive returns whether the dialog is visible.
func (m *Model) IsActive() bool { return m.active }

// Open shows the edit dialog for a cell value.
func (m *Model) Open(row, col int, colName, typeName, value string, nullable bool) {
	m.active = true
	m.row = row
	m.col = col
	m.colName = colName
	m.typeName = typeName
	m.oldValue = value
	m.nullable = nullable
	m.lines = strings.Split(value, "\n")
	m.cursorRow = 0
	m.cursorCol = len(m.lines[0])
	m.focus = focusInput
}

// Close dismisses the dialog.
func (m *Model) Close() {
	m.active = false
}

// SetWidth sets the render width hint.
func (m *Model) SetWidth(w int) { m.width = w }

// Update handles key input.
func (m *Model) Update(msg tea.KeyMsg) tea.Cmd {
	if !m.active {
		return nil
	}

	switch msg.Type {
	case tea.KeyEsc:
		m.Close()
		return func() tea.Msg { return CancelMsg{} }

	case tea.KeyTab:
		m.cycleForward()
		return nil

	case tea.KeyShiftTab:
		m.cycleBackward()
		return nil

	case tea.KeyEnter:
		if m.focus != focusInput {
			return m.activateButton()
		}
		// enter in input = submit value
		return m.submit(false)
	}

	// keys only for input focus
	if m.focus == focusInput {
		return m.updateInput(msg)
	}

	return nil
}

func (m *Model) updateInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyCtrlJ:
		m.insertNewline()

	case tea.KeyBackspace:
		m.backspace()

	case tea.KeyDelete:
		m.delete()

	case tea.KeyLeft:
		m.moveLeft()

	case tea.KeyRight:
		m.moveRight()

	case tea.KeyUp:
		if m.cursorRow > 0 {
			m.cursorRow--
			m.cursorCol = min(m.cursorCol, len(m.lines[m.cursorRow]))
		}

	case tea.KeyDown:
		if m.cursorRow < len(m.lines)-1 {
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

func (m *Model) submit(isNull bool) tea.Cmd {
	m.Close()
	row, col, oldVal := m.row, m.col, m.oldValue
	newVal := m.value()
	return func() tea.Msg {
		return SubmitMsg{Row: row, Col: col, OldValue: oldVal, NewValue: newVal, IsNull: isNull}
	}
}

func (m *Model) activateButton() tea.Cmd {
	switch m.focus {
	case focusOK:
		return m.submit(false)
	case focusNull:
		if m.nullable {
			return m.submit(true)
		}
		return nil
	case focusCancel:
		m.Close()
		return func() tea.Msg { return CancelMsg{} }
	}
	return nil
}

func (m *Model) cycleForward() {
	switch m.focus {
	case focusInput:
		m.focus = focusOK
	case focusOK:
		if m.nullable {
			m.focus = focusNull
		} else {
			m.focus = focusCancel
		}
	case focusNull:
		m.focus = focusCancel
	case focusCancel:
		m.focus = focusInput
	}
}

func (m *Model) cycleBackward() {
	switch m.focus {
	case focusInput:
		m.focus = focusCancel
	case focusOK:
		m.focus = focusInput
	case focusNull:
		m.focus = focusOK
	case focusCancel:
		if m.nullable {
			m.focus = focusNull
		} else {
			m.focus = focusOK
		}
	}
}

func (m *Model) value() string {
	return strings.Join(m.lines, "\n")
}

func (m *Model) insertRune(r rune) {
	line := m.lines[m.cursorRow]
	m.lines[m.cursorRow] = line[:m.cursorCol] + string(r) + line[m.cursorCol:]
	m.cursorCol++
}

func (m *Model) insertNewline() {
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

func (m *Model) backspace() {
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

func (m *Model) delete() {
	line := m.lines[m.cursorRow]
	if m.cursorCol < len(line) {
		m.lines[m.cursorRow] = line[:m.cursorCol] + line[m.cursorCol+1:]
	} else if m.cursorRow < len(m.lines)-1 {
		m.lines[m.cursorRow] = line + m.lines[m.cursorRow+1]
		m.lines = append(m.lines[:m.cursorRow+1], m.lines[m.cursorRow+2:]...)
	}
}

func (m *Model) moveLeft() {
	if m.cursorCol > 0 {
		m.cursorCol--
	} else if m.cursorRow > 0 {
		m.cursorRow--
		m.cursorCol = len(m.lines[m.cursorRow])
	}
}

func (m *Model) moveRight() {
	line := m.lines[m.cursorRow]
	if m.cursorCol < len(line) {
		m.cursorCol++
	} else if m.cursorRow < len(m.lines)-1 {
		m.cursorRow++
		m.cursorCol = 0
	}
}

// View renders the edit dialog.
func (m *Model) View(containerW, containerH int) string {
	if !m.active {
		return ""
	}

	t := theme.Current().Styles

	w := min(containerW-4, 56)
	if w < 30 {
		w = 30
	}
	contentW := w - 6 // outer border(2) + padding(4)

	var sb strings.Builder

	// title
	title := m.colName
	if m.typeName != "" {
		title += " [" + m.typeName + "]"
	}
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(t.BorderFocused)
	sb.WriteString(titleStyle.Render(title))
	sb.WriteString("\n\n")

	// input area — background highlight, no nested border
	inputBg := lipgloss.Color("236")
	if m.focus == focusInput {
		inputBg = lipgloss.Color("238")
	}
	inputStyle := lipgloss.NewStyle().
		Background(inputBg).
		Width(contentW).
		Padding(0, 1)

	var inputContent strings.Builder
	for i, line := range m.lines {
		if i > 0 {
			inputContent.WriteByte('\n')
		}
		if m.focus == focusInput && i == m.cursorRow {
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
			inputContent.WriteString(before)
			inputContent.WriteString(cursor)
			inputContent.WriteString(after)
		} else {
			inputContent.WriteString(line)
		}
	}

	sb.WriteString(inputStyle.Render(inputContent.String()))
	sb.WriteString("\n\n")

	// hint
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	sb.WriteString(hintStyle.Render("Ctrl+J newline | Tab cycle"))
	sb.WriteString("\n\n")

	// buttons
	sb.WriteString(m.renderButtons(t))

	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.BorderFocused).
		Padding(1, 2).
		Width(w)

	box := border.Render(sb.String())
	return lipgloss.Place(containerW, containerH, lipgloss.Center, lipgloss.Center, box)
}

func (m *Model) renderButtons(t theme.Styles) string {
	btnActive := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("232")).
		Background(t.BorderFocused).
		Padding(0, 2)

	btnNormal := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color("239")).
		Padding(0, 2)

	btnDisabled := lipgloss.NewStyle().
		Foreground(lipgloss.Color("242")).
		Background(lipgloss.Color("236")).
		Padding(0, 2)

	style := func(f focus) lipgloss.Style {
		if f == focusNull && !m.nullable {
			return btnDisabled
		}
		if m.focus == f {
			return btnActive
		}
		return btnNormal
	}

	ok := style(focusOK).Render("OK")
	null := style(focusNull).Render("NULL")
	cancel := style(focusCancel).Render("Cancel")

	return ok + "  " + null + "  " + cancel
}
