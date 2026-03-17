// Package connform provides a modal form for adding/editing connections.
package connform

import (
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/conn"
	"github.com/mdjarv/db/internal/tui/theme"
)

// SubmitMsg carries the form result.
type SubmitMsg struct {
	Config   conn.ConnectionConfig
	Password string
	IsEdit   bool
	OldName  string
	Source   conn.Source
}

// CancelMsg signals the form was dismissed.
type CancelMsg struct{}

const (
	fieldName     = 0
	fieldHost     = 1
	fieldPort     = 2
	fieldUser     = 3
	fieldPassword = 4
	fieldDBName   = 5
	fieldSSLMode  = 6
	fieldCount    = 7
	focusSave     = 7
	focusCancel   = 8
	focusCount    = 9
)

var fieldLabels = [fieldCount]string{"Name", "Host", "Port", "User", "Password", "Database", "SSL Mode"}

var sslModes = []string{"disable", "allow", "prefer", "require", "verify-ca", "verify-full"}

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

// Model is the connection form state.
type Model struct {
	active  bool
	fields  [fieldCount]field
	sslIdx  int
	focus   int
	isEdit  bool
	oldName string
	source  conn.Source
	err     string
}

// New creates an inactive form.
func New() *Model {
	return &Model{}
}

// IsActive returns whether the form is visible.
func (m *Model) IsActive() bool { return m.active }

// OpenAdd opens a blank form for adding a new connection.
func (m *Model) OpenAdd(source conn.Source) {
	m.active = true
	m.isEdit = false
	m.oldName = ""
	m.source = source
	m.err = ""
	for i := range m.fields {
		m.fields[i] = field{}
	}
	m.fields[fieldHost].set("localhost")
	m.fields[fieldPort].set("5432")
	m.fields[fieldUser].set(defaultUser())
	m.fields[fieldDBName].set("postgres")
	m.sslIdx = sslModeIndex("prefer")
	m.focus = fieldName
}

// OpenEdit opens the form pre-filled from an existing connection.
func (m *Model) OpenEdit(cfg conn.ConnectionConfig, password string, source conn.Source) {
	m.active = true
	m.isEdit = true
	m.oldName = cfg.Name
	m.source = source
	m.err = ""
	m.fields[fieldName].set(cfg.Name)
	m.fields[fieldHost].set(cfg.Host)
	if cfg.Port != 0 {
		m.fields[fieldPort].set(strconv.Itoa(cfg.Port))
	} else {
		m.fields[fieldPort].set("5432")
	}
	m.fields[fieldUser].set(cfg.User)
	m.fields[fieldPassword].set(password)
	m.fields[fieldDBName].set(cfg.DBName)
	m.sslIdx = sslModeIndex(cfg.SSLMode)
	m.focus = fieldName
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
		if m.focus < fieldCount {
			m.focus++
			return nil
		}
		if m.focus == focusCancel {
			m.Close()
			return func() tea.Msg { return CancelMsg{} }
		}
		if m.focus == focusSave {
			return m.validate()
		}
		return nil
	}

	// SSL Mode field: left/right cycles options
	if m.focus == fieldSSLMode {
		switch msg.String() {
		case "left", "h":
			if m.sslIdx > 0 {
				m.sslIdx--
			} else {
				m.sslIdx = len(sslModes) - 1
			}
		case "right", "l":
			m.sslIdx = (m.sslIdx + 1) % len(sslModes)
		}
		return nil
	}

	// text input for other fields
	if m.focus >= 0 && m.focus < fieldCount {
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

func (m *Model) validate() tea.Cmd {
	name := strings.TrimSpace(m.fields[fieldName].text())
	host := strings.TrimSpace(m.fields[fieldHost].text())
	portStr := strings.TrimSpace(m.fields[fieldPort].text())

	if name == "" {
		m.err = "name is required"
		return nil
	}
	if host == "" {
		m.err = "host is required"
		return nil
	}

	port := 5432
	if portStr != "" {
		var err error
		port, err = strconv.Atoi(portStr)
		if err != nil {
			m.err = "port must be a number"
			return nil
		}
	}

	cfg := conn.ConnectionConfig{
		Name:    name,
		Host:    host,
		Port:    port,
		User:    strings.TrimSpace(m.fields[fieldUser].text()),
		DBName:  strings.TrimSpace(m.fields[fieldDBName].text()),
		SSLMode: sslModes[m.sslIdx],
	}
	password := m.fields[fieldPassword].text()
	isEdit := m.isEdit
	oldName := m.oldName
	source := m.source

	m.Close()
	return func() tea.Msg {
		return SubmitMsg{
			Config:   cfg,
			Password: password,
			IsEdit:   isEdit,
			OldName:  oldName,
			Source:   source,
		}
	}
}

// View renders the form overlay.
func (m *Model) View(containerW, containerH int) string {
	if !m.active {
		return ""
	}

	t := theme.Current()
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(t.Styles.BorderFocused)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Width(10)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	w := min(containerW-4, 50)
	if w < 30 {
		w = 30
	}
	inputW := w - 14

	var lines []string
	title := "Add Connection"
	if m.isEdit {
		title = "Edit Connection"
	}
	lines = append(lines, titleStyle.Render(title))
	lines = append(lines, "")

	for i := 0; i < fieldCount; i++ {
		label := labelStyle.Render(fieldLabels[i])
		if i == fieldSSLMode {
			lines = append(lines, label+m.renderSSLField(inputW, t))
		} else {
			lines = append(lines, label+m.renderTextField(i, inputW, t))
		}
	}

	lines = append(lines, "")

	saveLabel := "[ Save ]"
	cancelLabel := "[ Cancel ]"
	if m.focus == focusSave {
		saveLabel = lipgloss.NewStyle().Bold(true).Reverse(true).Render(saveLabel)
	} else {
		saveLabel = hintStyle.Render(saveLabel)
	}
	if m.focus == focusCancel {
		cancelLabel = lipgloss.NewStyle().Bold(true).Reverse(true).Render(cancelLabel)
	} else {
		cancelLabel = hintStyle.Render(cancelLabel)
	}
	lines = append(lines, "          "+saveLabel+"  "+cancelLabel)

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
	isPassword := idx == fieldPassword

	inputBg := lipgloss.Color("236")
	if focused {
		inputBg = lipgloss.Color("238")
	}

	var rendered string
	if focused {
		display := f.value
		if isPassword {
			display = make([]rune, len(f.value))
			for i := range display {
				display[i] = '*'
			}
		}
		col := min(f.cursor, len(display))
		before := string(display[:col])
		cursorChar := " "
		after := ""
		if col < len(display) {
			cursorChar = string(display[col])
			after = string(display[col+1:])
		}
		cursor := lipgloss.NewStyle().
			Foreground(lipgloss.Color("232")).
			Background(t.Styles.BorderFocused).
			Render(cursorChar)
		rendered = before + cursor + after
	} else {
		if isPassword {
			rendered = strings.Repeat("*", len(f.value))
		} else {
			rendered = string(f.value)
		}
	}

	return lipgloss.NewStyle().
		Background(inputBg).
		Width(width).
		Padding(0, 1).
		Render(rendered)
}

func (m *Model) renderSSLField(width int, t *theme.Theme) string {
	focused := m.focus == fieldSSLMode
	value := sslModes[m.sslIdx]

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

func defaultUser() string {
	if u := os.Getenv("USER"); u != "" {
		return u
	}
	return ""
}

func sslModeIndex(mode string) int {
	for i, m := range sslModes {
		if m == mode {
			return i
		}
	}
	return 2 // "prefer"
}
