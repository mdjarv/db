// Package dumpform provides a modal form for configuring database dumps.
package dumpform

import (
	"fmt"
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

	tablePickerHeight = 6
)

var fieldLabels = [fieldCount]string{"Format", "Output", "Schema only", "Tables"}

var formats = []struct {
	name   string
	desc   string
	format dump.Format
}{
	{"custom", "compressed, use pg_restore", dump.Custom},
	{"plain", "SQL text, human-readable", dump.Plain},
	{"directory", "parallel restore, large DBs", dump.Directory},
	{"tar", "portable archive", dump.Tar},
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

	// table picker
	availTables []string
	selected    map[string]bool
	tableCursor int
	tableOffset int
	tableOpen   bool
}

// New creates an inactive dump form.
func New() *Model {
	return &Model{}
}

// IsActive returns whether the form is visible.
func (m *Model) IsActive() bool { return m.active }

// Open opens the dump form pre-filled with connection and table info.
func (m *Model) Open(tableName, dbName, host, port, user, password, sslMode string, tables []string) {
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

	m.fields[fieldOutput].set(dump.DefaultOutputPath(dbName, formats[m.formatIdx].format))

	// table picker
	m.availTables = tables
	m.selected = make(map[string]bool)
	m.tableCursor = 0
	m.tableOffset = 0
	if tableName != "" {
		m.selected[tableName] = true
		// move cursor to the pre-selected table
		for i, t := range m.availTables {
			if t == tableName {
				m.tableCursor = i
				break
			}
		}
	}
}

// OpenSchemaOnly opens the form with schema-only pre-selected.
func (m *Model) OpenSchemaOnly(tableName, dbName, host, port, user, password, sslMode string, tables []string) {
	m.Open(tableName, dbName, host, port, user, password, sslMode, tables)
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

	// table picker open — capture input
	if m.tableOpen {
		return m.updateTablePicker(msg)
	}

	switch msg.String() {
	case "esc":
		m.Close()
		return func() tea.Msg { return CancelMsg{} }
	case "tab":
		m.focus = (m.focus + 1) % focusCount
		return nil
	case "shift+tab":
		m.focus = (m.focus - 1 + focusCount) % focusCount
		return nil
	case "enter":
		if m.focus == fieldTables {
			m.tableOpen = true
			return nil
		}
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
		case "down":
			m.focus++
		}
		return nil
	}

	// Schema only toggle: space/left/right
	if m.focus == fieldSchemaOnly {
		switch msg.String() {
		case " ", "left", "right", "h", "l":
			m.schemaOnly = !m.schemaOnly
		case "down":
			m.focus++
		case "up":
			m.focus--
		}
		return nil
	}

	// Tables field (closed) — arrows only
	if m.focus == fieldTables {
		switch msg.String() {
		case "down":
			m.focus = focusStart
		case "up":
			m.focus--
		}
		return nil
	}

	// Text input for output field
	if m.focus == fieldOutput {
		switch msg.String() {
		case "down":
			m.focus++
			return nil
		case "up":
			m.focus--
			return nil
		}
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

	// buttons
	if m.focus == focusStart || m.focus == focusCancel {
		switch msg.String() {
		case "up", "k":
			m.focus = fieldTables
		case "left", "h":
			if m.focus == focusCancel {
				m.focus = focusStart
			}
		case "right", "l":
			if m.focus == focusStart {
				m.focus = focusCancel
			}
		}
		return nil
	}

	return nil
}

func (m *Model) updateTablePicker(msg tea.KeyMsg) tea.Cmd {
	n := len(m.availTables)

	switch msg.String() {
	case "esc", "enter":
		m.tableOpen = false
		return nil
	case "j", "down":
		if m.tableCursor < n-1 {
			m.tableCursor++
			m.clampTableOffset()
		}
	case "k", "up":
		if m.tableCursor > 0 {
			m.tableCursor--
			m.clampTableOffset()
		}
	case " ":
		if n > 0 {
			t := m.availTables[m.tableCursor]
			if m.selected[t] {
				delete(m.selected, t)
			} else {
				m.selected[t] = true
			}
		}
	case "a":
		if len(m.selected) == n {
			m.selected = make(map[string]bool)
		} else {
			for _, t := range m.availTables {
				m.selected[t] = true
			}
		}
	}
	return nil
}

func (m *Model) clampTableOffset() {
	if m.tableCursor < m.tableOffset {
		m.tableOffset = m.tableCursor
	}
	if m.tableCursor >= m.tableOffset+tablePickerHeight {
		m.tableOffset = m.tableCursor - tablePickerHeight + 1
	}
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
	for _, t := range m.availTables {
		if m.selected[t] {
			tables = append(tables, t)
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

	w := max(min(containerW-4, 68), 34)
	inputW := w - 18

	var lines []string
	lines = append(lines, titleStyle.Render("Dump Database"))
	lines = append(lines, "")

	lines = append(lines, labelStyle.Render(fieldLabels[fieldFormat])+m.renderFormatField(inputW, t))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	lines = append(lines, labelStyle.Render("")+descStyle.Render(formats[m.formatIdx].desc))
	lines = append(lines, labelStyle.Render(fieldLabels[fieldOutput])+m.renderTextField(fieldOutput, inputW, t))
	lines = append(lines, labelStyle.Render(fieldLabels[fieldSchemaOnly])+m.renderToggle(inputW, t))

	// table summary (picker is a separate popup)
	lines = append(lines, labelStyle.Render(fieldLabels[fieldTables])+m.renderTableSummary(inputW))

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
	lines = append(lines, hintStyle.Render("Tab navigate  Space toggle  a all  Esc cancel"))

	content := strings.Join(lines, "\n")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Styles.BorderFocused).
		Padding(1, 2).
		Width(w).
		Render(content)

	form := lipgloss.Place(containerW, containerH, lipgloss.Center, lipgloss.Center, box)

	if m.tableOpen {
		popup := m.renderTablePopup(containerW, containerH, t)
		return popup
	}

	return form
}

func (m *Model) renderTableSummary(width int) string {
	focused := m.focus == fieldTables

	inputBg := lipgloss.Color("236")
	if focused {
		inputBg = lipgloss.Color("238")
	}

	innerW := width - 2 // padding
	var label string
	total := len(m.availTables)
	if total == 0 {
		label = "(no tables)"
	} else if len(m.selected) == 0 {
		label = fmt.Sprintf("all %d tables", total)
	} else {
		// comma-separated selected names, truncated
		var names []string
		for _, t := range m.availTables {
			if m.selected[t] {
				names = append(names, t)
			}
		}
		label = strings.Join(names, ", ")
		if len(label) > innerW {
			label = label[:max(innerW-3, 0)] + "..."
		}
	}

	return lipgloss.NewStyle().
		Background(inputBg).
		Width(width).
		Padding(0, 1).
		Render(label)
}

func (m *Model) renderTablePopup(containerW, containerH int, t *theme.Theme) string {
	n := len(m.availTables)
	visible := min(n, tablePickerHeight)
	end := min(m.tableOffset+visible, n)

	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("232")).
		Background(t.Styles.BorderFocused)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(t.Styles.BorderFocused)

	var lines []string
	lines = append(lines, titleStyle.Render("Select Tables"))
	sel := len(m.selected)
	if sel == 0 {
		lines = append(lines, hintStyle.Render(fmt.Sprintf("none selected (dumps all %d)", n)))
	} else {
		lines = append(lines, hintStyle.Render(fmt.Sprintf("%d of %d selected", sel, n)))
	}
	lines = append(lines, "")

	for i := m.tableOffset; i < end; i++ {
		name := m.availTables[i]
		check := "[ ]"
		if m.selected[name] {
			check = "[x]"
		}
		entry := fmt.Sprintf(" %s %s", check, name)
		if i == m.tableCursor {
			entry = cursorStyle.Render(entry)
		}
		lines = append(lines, entry)
	}

	if n > tablePickerHeight {
		var scrollHint string
		if m.tableOffset > 0 && end < n {
			scrollHint = fmt.Sprintf("... %d above, %d below", m.tableOffset, n-end)
		} else if m.tableOffset > 0 {
			scrollHint = fmt.Sprintf("... %d above", m.tableOffset)
		} else if end < n {
			scrollHint = fmt.Sprintf("... %d below", n-end)
		}
		if scrollHint != "" {
			lines = append(lines, hintStyle.Render(" "+scrollHint))
		}
	}

	lines = append(lines, "")
	lines = append(lines, hintStyle.Render("j/k nav  Space toggle  a all  Enter/Esc done"))

	content := strings.Join(lines, "\n")

	w := max(min(containerW-4, 56), 30)
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
	f := formats[m.formatIdx]

	inputBg := lipgloss.Color("236")
	if focused {
		inputBg = lipgloss.Color("238")
	}

	style := lipgloss.NewStyle().
		Background(inputBg).
		Width(width).
		Padding(0, 1)

	if focused {
		style = style.
			Foreground(lipgloss.Color("232")).
			Background(t.Styles.BorderFocused)
	}

	return style.Render(f.name)
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
