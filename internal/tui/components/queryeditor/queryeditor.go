// Package queryeditor implements the SQL query editor pane.
package queryeditor

import (
	"fmt"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/tui/core"
)

const undoLimit = 100

type snapshot struct {
	lines   []string
	cursorX int
	cursorY int
}

// Model is the query editor state.
type Model struct {
	lines   []string
	cursorX int
	cursorY int
	focused bool
	mode    core.Mode
	width   int
	height  int
	offset  int

	undoStack []snapshot
	redoStack []snapshot

	pending string // for multi-key sequences like "dd"
}

// New creates a query editor.
func New() *Model {
	return &Model{
		lines: []string{""},
	}
}

func (m *Model) saveUndo() {
	s := snapshot{
		lines:   make([]string, len(m.lines)),
		cursorX: m.cursorX,
		cursorY: m.cursorY,
	}
	copy(s.lines, m.lines)
	m.undoStack = append(m.undoStack, s)
	if len(m.undoStack) > undoLimit {
		m.undoStack = m.undoStack[1:]
	}
	m.redoStack = nil
}

func (m *Model) undo() {
	if len(m.undoStack) == 0 {
		return
	}
	cur := snapshot{
		lines:   make([]string, len(m.lines)),
		cursorX: m.cursorX,
		cursorY: m.cursorY,
	}
	copy(cur.lines, m.lines)
	m.redoStack = append(m.redoStack, cur)

	s := m.undoStack[len(m.undoStack)-1]
	m.undoStack = m.undoStack[:len(m.undoStack)-1]
	m.lines = s.lines
	m.cursorX = s.cursorX
	m.cursorY = s.cursorY
}

func (m *Model) redo() {
	if len(m.redoStack) == 0 {
		return
	}
	cur := snapshot{
		lines:   make([]string, len(m.lines)),
		cursorX: m.cursorX,
		cursorY: m.cursorY,
	}
	copy(cur.lines, m.lines)
	m.undoStack = append(m.undoStack, cur)

	s := m.redoStack[len(m.redoStack)-1]
	m.redoStack = m.redoStack[:len(m.redoStack)-1]
	m.lines = s.lines
	m.cursorX = s.cursorX
	m.cursorY = s.cursorY
}

// Update handles input messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if !m.focused {
		return nil
	}
	switch msg := msg.(type) {
	case core.ModeChangedMsg:
		m.mode = msg.Mode
		m.pending = ""
	case tea.KeyMsg:
		if m.mode == core.ModeInsert {
			return m.insertUpdate(msg)
		}
		return m.normalUpdate(msg)
	}
	return nil
}

func (m *Model) normalUpdate(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	// handle pending multi-key sequences
	if m.pending == "d" {
		m.pending = ""
		if key == "d" {
			return m.deleteLine()
		}
		return nil
	}
	if m.pending == "g" {
		m.pending = ""
		if key == "g" {
			m.cursorY = 0
			m.cursorX = 0
			m.offset = 0
			return nil
		}
		return nil
	}

	switch key {
	case "j", "down":
		if m.cursorY < len(m.lines)-1 {
			m.cursorY++
			m.clampX()
		}
	case "k", "up":
		if m.cursorY > 0 {
			m.cursorY--
			m.clampX()
		}
	case "h", "left":
		if m.cursorX > 0 {
			m.cursorX--
		}
	case "l", "right":
		line := m.lines[m.cursorY]
		maxX := len(line) - 1
		if maxX < 0 {
			maxX = 0
		}
		if m.cursorX < maxX {
			m.cursorX++
		}
	case "0":
		m.cursorX = 0
	case "$":
		line := m.lines[m.cursorY]
		m.cursorX = max(len(line)-1, 0)
	case "w":
		m.wordForward()
	case "b":
		m.wordBackward()
	case "d":
		m.pending = "d"
	case "D":
		m.deleteToEnd()
	case "g":
		m.pending = "g"
	case "G":
		m.cursorY = len(m.lines) - 1
		m.clampX()
		vh := m.viewHeight()
		if len(m.lines) > vh {
			m.offset = len(m.lines) - vh
		}
	case "u":
		m.undo()
	case "ctrl+r":
		m.redo()
	case "a":
		line := m.lines[m.cursorY]
		if len(line) > 0 {
			m.cursorX = min(m.cursorX+1, len(line))
		}
		return m.enterInsert()
	case "A":
		m.cursorX = len(m.lines[m.cursorY])
		return m.enterInsert()
	case "I":
		m.cursorX = firstNonBlank(m.lines[m.cursorY])
		return m.enterInsert()
	case "o":
		m.saveUndo()
		m.cursorY++
		m.lines = insertAt(m.lines, m.cursorY, "")
		m.cursorX = 0
		return m.enterInsert()
	case "O":
		m.saveUndo()
		m.lines = insertAt(m.lines, m.cursorY, "")
		m.cursorX = 0
		return m.enterInsert()
	case "x":
		line := m.lines[m.cursorY]
		if len(line) > 0 && m.cursorX < len(line) {
			m.saveUndo()
			m.lines[m.cursorY] = line[:m.cursorX] + line[m.cursorX+1:]
			m.clampX()
		}
	case "enter":
		sql := m.Content()
		if strings.TrimSpace(sql) != "" {
			return func() tea.Msg { return core.QuerySubmittedMsg{SQL: sql} }
		}
	}
	m.scrollToCursor()
	return nil
}

func (m *Model) enterInsert() tea.Cmd {
	return func() tea.Msg {
		return core.ModeChangedMsg{Mode: core.ModeInsert}
	}
}

func (m *Model) insertUpdate(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyBackspace:
		m.backspace()
	case tea.KeyEnter:
		m.newline()
	case tea.KeyTab:
		m.saveUndo()
		line := m.lines[m.cursorY]
		m.lines[m.cursorY] = line[:m.cursorX] + "    " + line[m.cursorX:]
		m.cursorX += 4
	case tea.KeyRunes:
		m.saveUndo()
		line := m.lines[m.cursorY]
		text := string(msg.Runes)
		m.lines[m.cursorY] = line[:m.cursorX] + text + line[m.cursorX:]
		m.cursorX += len(text)
	}
	m.scrollToCursor()
	return nil
}

func (m *Model) backspace() {
	if m.cursorX > 0 {
		m.saveUndo()
		line := m.lines[m.cursorY]
		m.lines[m.cursorY] = line[:m.cursorX-1] + line[m.cursorX:]
		m.cursorX--
	} else if m.cursorY > 0 {
		m.saveUndo()
		prevLine := m.lines[m.cursorY-1]
		m.cursorX = len(prevLine)
		m.lines[m.cursorY-1] = prevLine + m.lines[m.cursorY]
		m.lines = removeAt(m.lines, m.cursorY)
		m.cursorY--
	}
}

func (m *Model) newline() {
	m.saveUndo()
	line := m.lines[m.cursorY]
	before := line[:m.cursorX]
	after := line[m.cursorX:]
	m.lines[m.cursorY] = before
	m.cursorY++
	m.lines = insertAt(m.lines, m.cursorY, after)
	m.cursorX = 0
}

func (m *Model) deleteLine() tea.Cmd {
	m.saveUndo()
	if len(m.lines) == 1 {
		m.lines[0] = ""
		m.cursorX = 0
		return nil
	}
	m.lines = removeAt(m.lines, m.cursorY)
	if m.cursorY >= len(m.lines) {
		m.cursorY = len(m.lines) - 1
	}
	m.clampX()
	return nil
}

func (m *Model) deleteToEnd() {
	m.saveUndo()
	line := m.lines[m.cursorY]
	m.lines[m.cursorY] = line[:m.cursorX]
	m.clampX()
}

func (m *Model) wordForward() {
	line := m.lines[m.cursorY]
	runes := []rune(line)
	x := m.cursorX

	if x >= len(runes) {
		if m.cursorY < len(m.lines)-1 {
			m.cursorY++
			m.cursorX = 0
		}
		return
	}

	// skip current word chars
	for x < len(runes) && !unicode.IsSpace(runes[x]) {
		x++
	}
	// skip whitespace
	for x < len(runes) && unicode.IsSpace(runes[x]) {
		x++
	}

	if x >= len(runes) && m.cursorY < len(m.lines)-1 {
		m.cursorY++
		m.cursorX = 0
	} else {
		m.cursorX = x
	}
}

func (m *Model) wordBackward() {
	x := m.cursorX

	if x == 0 {
		if m.cursorY > 0 {
			m.cursorY--
			m.cursorX = max(len(m.lines[m.cursorY])-1, 0)
		}
		return
	}

	line := m.lines[m.cursorY]
	runes := []rune(line)

	x--
	// skip whitespace backward
	for x > 0 && unicode.IsSpace(runes[x]) {
		x--
	}
	// skip word chars backward
	for x > 0 && !unicode.IsSpace(runes[x-1]) {
		x--
	}
	m.cursorX = x
}

func (m *Model) clampX() {
	line := m.lines[m.cursorY]
	maxX := len(line)
	if m.mode != core.ModeInsert && maxX > 0 {
		maxX = len(line) - 1
	}
	if maxX < 0 {
		maxX = 0
	}
	if m.cursorX > maxX {
		m.cursorX = maxX
	}
}

func (m *Model) viewHeight() int {
	return max(m.height-2, 1)
}

func (m *Model) scrollToCursor() {
	vh := m.viewHeight()
	if m.cursorY < m.offset {
		m.offset = m.cursorY
	}
	if m.cursorY >= m.offset+vh {
		m.offset = m.cursorY - vh + 1
	}
}

// gutterWidth returns the width of the line number gutter.
func (m *Model) gutterWidth() int {
	n := len(m.lines)
	w := 1
	for n >= 10 {
		w++
		n /= 10
	}
	return w + 1 // digit width + 1 space
}

// View renders the query editor.
func (m *Model) View() string {
	var sb strings.Builder
	vh := m.viewHeight()
	end := min(m.offset+vh, len(m.lines))
	gw := m.gutterWidth()

	gutterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle := lipgloss.NewStyle().Reverse(true)
	contentWidth := max(m.width-2-gw, 1) // border takes 2

	for i := m.offset; i < end; i++ {
		gutter := gutterStyle.Render(fmt.Sprintf("%*d ", gw-1, i+1))
		line := m.lines[i]

		var rendered string
		if m.focused && i == m.cursorY {
			rendered = m.renderLineWithCursor(line, cursorStyle)
		} else {
			rendered = m.highlightLine(line)
			if len(rendered) > contentWidth {
				rendered = rendered[:contentWidth]
			}
		}

		sb.WriteString(gutter)
		sb.WriteString(rendered)
		if i < end-1 {
			sb.WriteByte('\n')
		}
	}

	borderColor := lipgloss.Color("240")
	if m.focused {
		borderColor = lipgloss.Color("62")
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(m.width - 2).
		Height(m.height - 2)

	return style.Render(sb.String())
}

func (m *Model) renderLineWithCursor(line string, cursorStyle lipgloss.Style) string {
	if m.cursorX >= len(line) {
		hl := m.highlightLine(line)
		return hl + cursorStyle.Render(" ")
	}
	before := line[:m.cursorX]
	ch := string(line[m.cursorX])
	after := line[m.cursorX+1:]
	return m.highlightLine(before) + cursorStyle.Render(ch) + m.highlightLine(after)
}

// highlightLine applies SQL syntax highlighting to a single line.
func (m *Model) highlightLine(line string) string {
	return highlightSQL(line)
}

// Content returns the full editor text.
func (m *Model) Content() string {
	return strings.Join(m.lines, "\n")
}

// SetContent replaces the editor text.
func (m *Model) SetContent(s string) {
	m.saveUndo()
	if s == "" {
		m.lines = []string{""}
	} else {
		m.lines = strings.Split(s, "\n")
	}
	m.cursorX = 0
	m.cursorY = 0
	m.offset = 0
}

// Focused returns whether the pane is focused.
func (m *Model) Focused() bool { return m.focused }

// SetFocused sets the focus state.
func (m *Model) SetFocused(f bool) { m.focused = f }

// SetSize sets the render dimensions.
func (m *Model) SetSize(w, h int) { m.width = w; m.height = h }

func insertAt(lines []string, idx int, s string) []string {
	lines = append(lines, "")
	copy(lines[idx+1:], lines[idx:])
	lines[idx] = s
	return lines
}

func removeAt(lines []string, idx int) []string {
	return append(lines[:idx], lines[idx+1:]...)
}

func firstNonBlank(line string) int {
	for i, r := range line {
		if !unicode.IsSpace(r) {
			return i
		}
	}
	return 0
}
