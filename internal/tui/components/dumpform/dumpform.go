// Package dumpform provides a modal form for configuring database dumps.
package dumpform

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/dump"
	"github.com/mdjarv/db/internal/tui/theme"
)

// SubmitMsg carries the completed dump configuration.
type SubmitMsg struct {
	Config dump.Config
}

// CancelMsg signals the dump form was dismissed.
type CancelMsg struct{}

const (
	fieldFormat     = 0
	fieldOutput     = 1
	fieldSchemaOnly = 2
	fieldTables     = 3
	fieldCount      = 4
	focusStart      = 4
	focusCancel     = 5
	focusCount      = 6
)

var fieldLabels = [fieldCount]string{"Format", "Output", "Schema only", "Tables"}

var formats = []struct {
	name   string
	format dump.Format
}{
	{"custom", dump.Custom},
	{"plain", dump.Plain},
	{"directory", dump.Directory},
	{"tar", dump.Tar},
}

// field is a single-line text input with cursor.
type field struct {
	value  []rune
	cursor int
}

func (f *field) insertRune(r rune) {
	f.value = append(f.value[:f.cursor], append([]rune{r}, f.value[f.cursor:]...)...)
	f.cursor++
}

func (f *field) backspace() {
	if f.cursor > 0 {
		f.value = append(f.value[:f.cursor-1], f.value[f.cursor:]...)
		f.cursor--
	}
}

func (f *field) delete() {
	if f.cursor < len(f.value) {
		f.value = append(f.value[:f.cursor], f.value[f.cursor+1:]...)
	}
}

func (f *field) moveLeft() {
	if f.cursor > 0 {
		f.cursor--
	}
}

func (f *field) moveRight() {
	if f.cursor < len(f.value) {
		f.cursor++
	}
}

func (f *field) set(s string) {
	f.value = []rune(s)
	f.cursor = len(f.value)
}

func (f *field) text() string {
	return string(f.value)
}

// Model is the dump form state.
type Model struct {
	active     bool
	fields     [fieldCount]field
	formatIdx  int
	schemaOnly bool
	focus      int
	dbName     string
	host       string
	port       string
	user       string
	password   string
	sslMode    string
	err        string
}

// New creates an inactive dump form.
func New() *Model {
	return &Model{}
}

// IsActive returns whether the form is visible.
func (m *Model) IsActive() bool { return m.active }

// Open opens the dump form pre-filled with connection and table info.
func (m *Model) Open(tableName, dbName, host, port, user, password, sslMode string) {
	m.active = true
	m.err = ""
	m.dbName = dbName
	m.host = host
	m.port = port
	m.user = user
	m.password = password
	m.sslMode = sslMode
	m.formatIdx = 0 // custom
	m.schemaOnly = false
	m.focus = fieldFormat

	for i := range m.fields {
		m.fields[i] = field{}
	}

	m.fields[fieldTables].set(tableName)
	m.fields[fieldOutput].set(dump.DefaultOutputPath(dbName, formats[m.formatIdx].format))
}

// OpenSchemaOnly opens the form with schema-only pre-selected.
func (m *Model) OpenSchemaOnly(tableName, dbName, host, port, user, password, sslMode string) {
	m.Open(tableName, dbName, host, port, user, password, sslMode)
	m.schemaOnly = true
}

// Close dismisses the form.
func (m *Model) Close() {
	m.active = false
}

// Update handles key input.
func (m *Model) Update(msg tea.KeyMsg) tea.Cmd {
	if !m.active {
		return nil
	}

	switch msg.String() {
	case "esc":
		m.Close()
		return func() tea.Msg { return CancelMsg{} }
	case "tab", "down":
		m.focus = (m.focus + 1) % focusCount
		return nil
	case "shift+tab", "up":
		m.focus = (m.focus - 1 + focusCount) % focusCount
		return nil
	case "enter":
		if m.focus < fieldCount && m.focus != fieldFormat && m.focus != fieldSchemaOnly {
			m.focus++
			return nil
		}
		if m.focus == focusCancel {
			m.Close()
			return func() tea.Msg { return CancelMsg{} }
		}
		if m.focus == focusStart {
			return m.submit()
		}
		if m.focus == fieldSchemaOnly {
			m.schemaOnly = !m.schemaOnly
			return nil
		}
		return nil
	}

	// Format selector: left/right cycles
	if m.focus == fieldFormat {
		switch msg.String() {
		case "left", "h":
			if m.formatIdx > 0 {
				m.formatIdx--
			} else {
				m.formatIdx = len(formats) - 1
			}
			m.updateOutputExtension()
		case "right", "l":
			m.formatIdx = (m.formatIdx + 1) % len(formats)
			m.updateOutputExtension()
		}
		return nil
	}

	// Schema only toggle: space/left/right
	if m.focus == fieldSchemaOnly {
		switch msg.String() {
		case " ", "left", "right", "h", "l":
			m.schemaOnly = !m.schemaOnly
		}
		return nil
	}

	// Text input for output and tables fields
	if m.focus == fieldOutput || m.focus == fieldTables {
		f := &m.fields[m.focus]
		switch msg.Type {
		case tea.KeyBackspace:
			f.backspace()
		case tea.KeyDelete:
			f.delete()
		case tea.KeyLeft:
			f.moveLeft()
		case tea.KeyRight:
			f.moveRight()
		case tea.KeyHome:
			f.cursor = 0
		case tea.KeyEnd:
			f.cursor = len(f.value)
		case tea.KeySpace:
			f.insertRune(' ')
		case tea.KeyRunes:
			for _, r := range msg.Runes {
				f.insertRune(r)
			}
		}
	}

	return nil
}

// updateOutputExtension replaces the file extension based on current format.
func (m *Model) updateOutputExtension() {
	m.fields[fieldOutput].set(dump.DefaultOutputPath(m.dbName, formats[m.formatIdx].format))
}

func (m *Model) submit() tea.Cmd {
	output := strings.TrimSpace(m.fields[fieldOutput].text())
	if output == "" {
		m.err = "output path is required"
		return nil
	}

	var tables []string
	tablesStr := strings.TrimSpace(m.fields[fieldTables].text())
	if tablesStr != "" {
		for _, t := range strings.Split(tablesStr, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				tables = append(tables, t)
			}
		}
	}

	cfg := dump.Config{
		Host:       m.host,
		Port:       m.port,
		User:       m.user,
		Password:   m.password,
		DBName:     m.dbName,
		SSLMode:    m.sslMode,
		Format:     formats[m.formatIdx].format,
		SchemaOnly: m.schemaOnly,
		Tables:     tables,
		OutputPath: output,
	}

	m.Close()
	return func() tea.Msg { return SubmitMsg{Config: cfg} }
}

// View renders the dump form overlay.
func (m *Model) View(containerW, containerH int) string {
	if !m.active {
		return ""
	}

	t := theme.Current()
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(t.Styles.BorderFocused)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Width(14)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	w := min(containerW-4, 56)
	if w < 34 {
		w = 34
	}
	inputW := w - 18

	var lines []string
	lines = append(lines, titleStyle.Render("Dump Database"))
	lines = append(lines, "")

	lines = append(lines, labelStyle.Render(fieldLabels[fieldFormat])+m.renderFormatField(inputW, t))
	lines = append(lines, labelStyle.Render(fieldLabels[fieldOutput])+m.renderTextField(fieldOutput, inputW, t))
	lines = append(lines, labelStyle.Render(fieldLabels[fieldSchemaOnly])+m.renderToggle(inputW, t))
	lines = append(lines, labelStyle.Render(fieldLabels[fieldTables])+m.renderTextField(fieldTables, inputW, t))

	lines = append(lines, "")

	startLabel := "[ Start ]"
	cancelLabel := "[ Cancel ]"
	if m.focus == focusStart {
		startLabel = lipgloss.NewStyle().Bold(true).Reverse(true).Render(startLabel)
	} else {
		startLabel = hintStyle.Render(startLabel)
	}
	if m.focus == focusCancel {
		cancelLabel = lipgloss.NewStyle().Bold(true).Reverse(true).Render(cancelLabel)
	} else {
		cancelLabel = hintStyle.Render(cancelLabel)
	}
	lines = append(lines, "              "+startLabel+"  "+cancelLabel)

	if m.err != "" {
		lines = append(lines, "")
		lines = append(lines, errStyle.Render(m.err))
	}

	lines = append(lines, "")
	lines = append(lines, hintStyle.Render("Tab/arrows navigate  Enter advance  Esc cancel"))

	content := strings.Join(lines, "\n")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Styles.BorderFocused).
		Padding(1, 2).
		Width(w).
		Render(content)

	return lipgloss.Place(containerW, containerH, lipgloss.Center, lipgloss.Center, box)
}

func (m *Model) renderTextField(idx, width int, t *theme.Theme) string {
	f := &m.fields[idx]
	focused := m.focus == idx

	inputBg := lipgloss.Color("236")
	if focused {
		inputBg = lipgloss.Color("238")
	}

	var rendered string
	if focused {
		col := min(f.cursor, len(f.value))
		before := string(f.value[:col])
		cursorChar := " "
		after := ""
		if col < len(f.value) {
			cursorChar = string(f.value[col])
			after = string(f.value[col+1:])
		}
		cursor := lipgloss.NewStyle().
			Foreground(lipgloss.Color("232")).
			Background(t.Styles.BorderFocused).
			Render(cursorChar)
		rendered = before + cursor + after
	} else {
		rendered = string(f.value)
	}

	return lipgloss.NewStyle().
		Background(inputBg).
		Width(width).
		Padding(0, 1).
		Render(rendered)
}

func (m *Model) renderFormatField(width int, t *theme.Theme) string {
	focused := m.focus == fieldFormat
	value := formats[m.formatIdx].name

	inputBg := lipgloss.Color("236")
	if focused {
		inputBg = lipgloss.Color("238")
	}

	var rendered string
	if focused {
		arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		valStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("232")).
			Background(t.Styles.BorderFocused)
		rendered = arrowStyle.Render("<") + " " + valStyle.Render(value) + " " + arrowStyle.Render(">")
	} else {
		rendered = value
	}

	return lipgloss.NewStyle().
		Background(inputBg).
		Width(width).
		Padding(0, 1).
		Render(rendered)
}

func (m *Model) renderToggle(width int, t *theme.Theme) string {
	focused := m.focus == fieldSchemaOnly

	inputBg := lipgloss.Color("236")
	if focused {
		inputBg = lipgloss.Color("238")
	}

	label := "[ ] No"
	if m.schemaOnly {
		label = "[x] Yes"
	}

	var rendered string
	if focused {
		valStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("232")).
			Background(t.Styles.BorderFocused)
		rendered = valStyle.Render(label)
	} else {
		rendered = label
	}

	return lipgloss.NewStyle().
		Background(inputBg).
		Width(width).
		Padding(0, 1).
		Render(rendered)
}
