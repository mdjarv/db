package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mdjarv/db/internal/tui/components/commandbar"
	"github.com/mdjarv/db/internal/tui/core"
	"github.com/mdjarv/db/internal/tui/theme"
)

// cmdHandler processes a command-mode command, returning updated model and cmd.
type cmdHandler func(m *Model, args string) (tea.Model, tea.Cmd)

var commandRegistry = map[string]cmdHandler{
	"q":        cmdQuit,
	"quit":     cmdQuit,
	"w":        cmdExecute,
	"clear":    cmdClear,
	"set":      cmdSet,
	"commit":   cmdCommit,
	"rollback": cmdRollback,
	"changes":  cmdChanges,
	"export":   cmdExport,
	"new":      cmdNewBuffer,
	"enew":     cmdNewBuffer,
	"bd":       cmdCloseBuffer,
	"bn":       cmdNextBuffer,
	"bp":       cmdPrevBuffer,
	"b":        cmdSwitchBuffer,
	"ls":       cmdListBuffers,
	"buffers":  cmdListBuffers,
	"theme":    cmdTheme,
}

func (m Model) handleCommand(msg commandbar.ExecuteMsg) (tea.Model, tea.Cmd) {
	m.mode = core.ModeNormal
	m.statusBar.SetMode(m.mode)

	if handler, ok := commandRegistry[msg.Command]; ok {
		return handler(&m, msg.Args)
	}
	m.statusBar.SetMessage("unknown command: " + msg.Command)
	return m, nil
}

func cmdQuit(m *Model, _ string) (tea.Model, tea.Cmd) {
	return *m, tea.Quit
}

func cmdExecute(m *Model, _ string) (tea.Model, tea.Cmd) {
	sql := m.queryEditor.Content()
	return *m, func() tea.Msg {
		return core.QuerySubmittedMsg{SQL: sql}
	}
}

func cmdClear(m *Model, _ string) (tea.Model, tea.Cmd) {
	m.queryEditor.SetContent("")
	m.recalcLayout()
	m.statusBar.SetMessage("buffer cleared")
	return *m, nil
}

func cmdSet(m *Model, args string) (tea.Model, tea.Cmd) {
	m.handleSetCommand(args)
	return *m, nil
}

func cmdCommit(m *Model, _ string) (tea.Model, tea.Cmd) {
	return m.handleCommit()
}

func cmdRollback(m *Model, _ string) (tea.Model, tea.Cmd) {
	return m.handleRollback()
}

func cmdChanges(m *Model, _ string) (tea.Model, tea.Cmd) {
	return m.handleChanges()
}

func cmdExport(m *Model, args string) (tea.Model, tea.Cmd) {
	return *m, m.parseExport(args)
}

func cmdNewBuffer(m *Model, _ string) (tea.Model, tea.Cmd) {
	m.saveBufferState()
	if !m.buffers.New() {
		m.statusBar.SetMessage("max buffers reached")
		return *m, nil
	}
	m.restoreBufferState()
	m.statusBar.SetMessage(fmt.Sprintf("buffer %d", m.buffers.ActiveIndex()))
	return *m, nil
}

func cmdCloseBuffer(m *Model, _ string) (tea.Model, tea.Cmd) {
	if !m.buffers.Close() {
		m.statusBar.SetMessage("cannot close last buffer")
		return *m, nil
	}
	m.restoreBufferState()
	m.statusBar.SetMessage(fmt.Sprintf("buffer %d", m.buffers.ActiveIndex()))
	return *m, nil
}

func cmdNextBuffer(m *Model, _ string) (tea.Model, tea.Cmd) {
	m.saveBufferState()
	m.buffers.Next()
	m.restoreBufferState()
	m.statusBar.SetMessage(fmt.Sprintf("buffer %d", m.buffers.ActiveIndex()))
	return *m, nil
}

func cmdPrevBuffer(m *Model, _ string) (tea.Model, tea.Cmd) {
	m.saveBufferState()
	m.buffers.Prev()
	m.restoreBufferState()
	m.statusBar.SetMessage(fmt.Sprintf("buffer %d", m.buffers.ActiveIndex()))
	return *m, nil
}

func cmdSwitchBuffer(m *Model, args string) (tea.Model, tea.Cmd) {
	n := 0
	if _, err := fmt.Sscanf(args, "%d", &n); err != nil {
		m.statusBar.SetMessage("invalid buffer number")
		return *m, nil
	}
	m.saveBufferState()
	if !m.buffers.SwitchTo(n) {
		m.statusBar.SetMessage("invalid buffer number")
		return *m, nil
	}
	m.restoreBufferState()
	m.statusBar.SetMessage(fmt.Sprintf("buffer %d", m.buffers.ActiveIndex()))
	return *m, nil
}

func cmdListBuffers(m *Model, _ string) (tea.Model, tea.Cmd) {
	m.saveBufferState()
	m.statusBar.SetMessage(m.buffers.List())
	return *m, nil
}

func cmdTheme(m *Model, args string) (tea.Model, tea.Cmd) {
	if args == "" {
		names := theme.Available()
		m.statusBar.SetMessage("themes: " + strings.Join(names, ", "))
	} else {
		t, err := theme.Resolve(args)
		if err != nil {
			m.statusBar.SetMessage("unknown theme: " + args)
		} else {
			theme.Set(t)
			m.statusBar.SetMessage("theme: " + t.Name)
		}
	}
	return *m, nil
}
