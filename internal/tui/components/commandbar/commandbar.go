// Package commandbar implements the : command input bar.
package commandbar

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/tui/theme"
)

// ExecuteMsg is sent when a command is submitted.
type ExecuteMsg struct {
	Command string
	Args    string
}

// CancelMsg is sent when command mode is cancelled.
type CancelMsg struct{}

// Command defines a registered command.
type Command struct {
	Name string
	Desc string
}

var defaultCommands = []Command{
	{Name: "q", Desc: "quit"},
	{Name: "quit", Desc: "quit"},
	{Name: "w", Desc: "run query"},
	{Name: "set", Desc: "change setting"},
	{Name: "export", Desc: "export results (csv|json|sql) <file>"},
	{Name: "new", Desc: "new buffer"},
	{Name: "enew", Desc: "new buffer"},
	{Name: "bd", Desc: "close buffer"},
	{Name: "bn", Desc: "next buffer"},
	{Name: "bp", Desc: "prev buffer"},
	{Name: "b", Desc: "switch to buffer N"},
	{Name: "ls", Desc: "list buffers"},
	{Name: "buffers", Desc: "list buffers"},
	{Name: "commit", Desc: "apply and commit pending changes"},
	{Name: "rollback", Desc: "discard pending changes"},
	{Name: "changes", Desc: "list pending changes"},
	{Name: "theme", Desc: "set or list themes"},
}

// Model is the command bar state.
type Model struct {
	input    string
	active   bool
	width    int
	history  []string
	histIdx  int
	commands []Command
}

// New creates a command bar.
func New() *Model {
	return &Model{
		commands: defaultCommands,
		histIdx:  -1,
	}
}

// Active returns whether the command bar is active.
func (m *Model) Active() bool { return m.active }

// Activate enters command mode.
func (m *Model) Activate() {
	m.active = true
	m.input = ""
	m.histIdx = -1
}

// Deactivate exits command mode.
func (m *Model) Deactivate() {
	m.active = false
	m.input = ""
	m.histIdx = -1
}

// SetWidth sets the render width.
func (m *Model) SetWidth(w int) { m.width = w }

// Update handles key input.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if !m.active {
		return nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			m.Deactivate()
			return func() tea.Msg { return CancelMsg{} }
		case tea.KeyEnter:
			cmd := m.input
			if cmd != "" {
				m.history = append(m.history, cmd)
			}
			m.Deactivate()
			return m.parseCommand(cmd)
		case tea.KeyBackspace:
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		case tea.KeyTab:
			m.tabComplete()
		case tea.KeyUp:
			if len(m.history) > 0 {
				if m.histIdx < 0 {
					m.histIdx = len(m.history) - 1
				} else if m.histIdx > 0 {
					m.histIdx--
				}
				m.input = m.history[m.histIdx]
			}
		case tea.KeyDown:
			if m.histIdx >= 0 {
				if m.histIdx < len(m.history)-1 {
					m.histIdx++
					m.input = m.history[m.histIdx]
				} else {
					m.histIdx = -1
					m.input = ""
				}
			}
		case tea.KeyRunes:
			m.input += string(msg.Runes)
		}
	}
	return nil
}

func (m *Model) tabComplete() {
	if m.input == "" {
		return
	}
	var matches []string
	for _, c := range m.commands {
		if strings.HasPrefix(c.Name, m.input) {
			matches = append(matches, c.Name)
		}
	}
	if len(matches) == 1 {
		m.input = matches[0]
	}
}

func (m *Model) parseCommand(input string) tea.Cmd {
	parts := strings.SplitN(strings.TrimSpace(input), " ", 2)
	if len(parts) == 0 || parts[0] == "" {
		return nil
	}
	cmd := parts[0]
	args := ""
	if len(parts) > 1 {
		args = parts[1]
	}
	return func() tea.Msg {
		return ExecuteMsg{Command: cmd, Args: args}
	}
}

// View renders the command bar.
func (m *Model) View() string {
	if !m.active {
		return ""
	}

	prompt := theme.Current().Styles.CommandPrompt.Render(":")

	input := lipgloss.NewStyle().
		Width(m.width - 1).
		Render(m.input + "\u2588")

	return prompt + input
}
