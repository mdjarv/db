// Package statusbar implements the bottom status bar.
package statusbar

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/tui/core"
	"github.com/mdjarv/db/internal/tui/theme"
)

// DefaultErrorTimeout is the default auto-dismiss duration for error messages.
const DefaultErrorTimeout = 5 * time.Second

// MsgLevel controls message priority. Higher-level messages are not
// overwritten by lower-level ones.
type MsgLevel int

// Message priority levels.
const (
	MsgInfo MsgLevel = iota
	MsgWarn
	MsgError
)

// Model is the status bar state.
type Model struct {
	mode         core.Mode
	connStr      string
	message      string
	msgLevel     MsgLevel
	txMode       string
	bufIdx       int
	bufCnt       int
	bufModified  bool
	width        int
	errorID      int
	errorTimeout time.Duration
}

// New creates a status bar.
func New() *Model {
	return &Model{
		txMode:       "txn",
		errorTimeout: DefaultErrorTimeout,
	}
}

// SetErrorTimeout sets the auto-dismiss duration for error messages.
func (m *Model) SetErrorTimeout(d time.Duration) { m.errorTimeout = d }

// SetMode updates the displayed vim mode.
func (m *Model) SetMode(mode core.Mode) { m.mode = mode }

// SetMessage sets an info-level status message. Does not overwrite
// error-level messages — use ClearError or SetError to replace those.
func (m *Model) SetMessage(msg string) {
	if m.msgLevel <= MsgInfo {
		m.message = msg
		m.msgLevel = MsgInfo
	}
}

// SetError sets an error message that auto-dismisses after the configured
// timeout. Returns a tea.Cmd that schedules the dismiss.
func (m *Model) SetError(msg string) tea.Cmd {
	m.errorID++
	m.message = msg
	m.msgLevel = MsgError
	id := m.errorID
	timeout := m.errorTimeout
	return tea.Tick(timeout, func(_ time.Time) tea.Msg {
		return core.ClearErrorMsg{ID: id}
	})
}

// HandleClearError processes a ClearErrorMsg; returns true if the error was
// cleared (i.e. the ID matched the current error).
func (m *Model) HandleClearError(msg core.ClearErrorMsg) bool {
	if msg.ID == m.errorID && m.msgLevel >= MsgError {
		m.message = ""
		m.msgLevel = MsgInfo
		return true
	}
	return false
}

// SetSuccess sets a success message, clearing any error.
func (m *Model) SetSuccess(msg string) {
	m.message = msg
	m.msgLevel = MsgInfo
}

// ClearError clears any sticky error message.
func (m *Model) ClearError() {
	if m.msgLevel >= MsgError {
		m.message = ""
		m.msgLevel = MsgInfo
	}
}

// SetWidth sets the render width.
func (m *Model) SetWidth(w int) { m.width = w }

// SetConn sets the connection string display.
func (m *Model) SetConn(s string) { m.connStr = s }

// SetTxMode sets the transaction mode display.
func (m *Model) SetTxMode(s string) { m.txMode = s }

// SetBuffer sets the buffer indicator display.
func (m *Model) SetBuffer(idx, count int) { m.bufIdx = idx; m.bufCnt = count }

// SetBufferModified sets whether the active buffer has been modified.
func (m *Model) SetBufferModified(modified bool) { m.bufModified = modified }

// View renders the status bar.
func (m *Model) View() string {
	s := theme.Current().Styles

	var modeStyle lipgloss.Style
	switch m.mode {
	case core.ModeInsert:
		modeStyle = s.ModeInsert
	case core.ModeCommand:
		modeStyle = s.ModeCommand
	default:
		modeStyle = s.ModeNormal
	}

	modeStr := modeStyle.Render(m.mode.String())

	var connStr string
	if m.connStr != "" {
		connStr = s.ConnectedFG.Render("\uf1c0 " + m.connStr)
	} else {
		connStr = s.DisconnectedFG.Render("\uf1c0 disconnected")
	}

	txStr := s.TxFG.Render(fmt.Sprintf("tx:%s", m.txMode))

	var bufStr string
	if m.bufCnt > 1 || m.bufModified {
		bufStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Padding(0, 1)
		mod := ""
		if m.bufModified {
			mod = "[+]"
		}
		bufStr = bufStyle.Render(fmt.Sprintf("[%d/%d]%s", m.bufIdx, m.bufCnt, mod))
	}

	left := modeStr + connStr + bufStr
	right := txStr

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	msgW := max(m.width-leftW-rightW, 0)

	msgStyle := s.StatusBarFG.Width(msgW)
	if m.msgLevel >= MsgError {
		msgStyle = s.Error.Width(msgW)
	}
	msg := msgStyle.Render(m.message)

	bar := lipgloss.NewStyle().
		Width(m.width).
		Background(s.StatusBarBG)

	return bar.Render(left + msg + right)
}
