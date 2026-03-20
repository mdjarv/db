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

// OpenOpts configures the edit dialog when opening.
type OpenOpts struct {
	Row, Col        int
	ColName         string
	TypeName        string
	Value           string
	Nullable        bool
	EnumValues      []string
	CompositeFields []CompositeField
}

// CompositeField describes a field within a composite type.
type CompositeField struct {
	Name     string
	TypeName string
}

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

	mode  inputMode
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
func (m *Model) Open(opts OpenOpts) {
	m.active = true
	m.row = opts.Row
	m.col = opts.Col
	m.colName = opts.ColName
	m.typeName = opts.TypeName
	m.oldValue = opts.Value
	m.nullable = opts.Nullable
	m.mode = resolveMode(opts)
	m.focus = focusInput
}

// Close dismisses the dialog.
func (m *Model) Close() {
	m.active = false
	m.mode = nil
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
		if m.mode != nil && m.mode.SubmitsOnEnter() {
			return m.submit(false)
		}
		// delegate enter to mode
		if m.mode != nil {
			return m.mode.Update(msg)
		}
		return nil
	}

	if m.focus == focusInput && m.mode != nil {
		return m.mode.Update(msg)
	}

	// button focus — arrow navigation
	if m.focus == focusOK || m.focus == focusNull || m.focus == focusCancel {
		switch msg.String() {
		case "up", "k":
			m.focus = focusInput
		case "left", "h":
			m.cycleBackward()
		case "right", "l":
			m.cycleForward()
		}
		return nil
	}

	return nil
}

func (m *Model) submit(isNull bool) tea.Cmd {
	row, col, oldVal := m.row, m.col, m.oldValue
	newVal := ""
	if m.mode != nil {
		newVal = m.mode.Value()
	}
	m.Close()
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

// View renders the edit dialog.
func (m *Model) View(containerW, containerH int) string {
	if !m.active {
		return ""
	}

	t := theme.Current().Styles

	w := min(containerW-4, 68)
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

	// input area via mode
	if m.mode != nil {
		sb.WriteString(m.mode.View(contentW, m.focus == focusInput, t))
	}
	sb.WriteString("\n\n")

	// hint
	hint := ""
	if m.mode != nil {
		hint = m.mode.Hint()
	}
	if hint != "" {
		hint += " | Tab cycle"
	} else {
		hint = "Tab cycle"
	}
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	sb.WriteString(hintStyle.Render(hint))
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

	style := func(f focus) lipgloss.Style {
		if m.focus == f {
			return btnActive
		}
		return btnNormal
	}

	ok := style(focusOK).Render("OK")
	cancel := style(focusCancel).Render("Cancel")

	if m.nullable {
		null := style(focusNull).Render("NULL")
		return ok + "  " + null + "  " + cancel
	}
	return ok + "  " + cancel
}
