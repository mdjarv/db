// Package statusbar implements the bottom status bar.
package statusbar

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/tui/core"
	"github.com/mdjarv/db/internal/tui/theme"
)

// Model is the status bar state.
type Model struct {
	mode    core.Mode
	connStr string
	message string
	txMode  string
	bufIdx  int
	bufCnt  int
	width   int
}

// New creates a status bar.
func New() *Model {
	return &Model{
		txMode: "txn",
	}
}

// SetMode updates the displayed vim mode.
func (m *Model) SetMode(mode core.Mode) { m.mode = mode }

// SetMessage sets the status message.
func (m *Model) SetMessage(msg string) { m.message = msg }

// SetWidth sets the render width.
func (m *Model) SetWidth(w int) { m.width = w }

// SetConn sets the connection string display.
func (m *Model) SetConn(s string) { m.connStr = s }

// SetTxMode sets the transaction mode display.
func (m *Model) SetTxMode(s string) { m.txMode = s }

// SetBuffer sets the buffer indicator display.
func (m *Model) SetBuffer(idx, count int) { m.bufIdx = idx; m.bufCnt = count }

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
	if m.bufCnt > 1 {
		bufStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Padding(0, 1)
		bufStr = bufStyle.Render(fmt.Sprintf("[%d/%d]", m.bufIdx, m.bufCnt))
	}

	left := modeStr + connStr + bufStr
	right := txStr

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	msgW := max(m.width-leftW-rightW, 0)

	msg := s.StatusBarFG.Width(msgW).Render(m.message)

	bar := lipgloss.NewStyle().
		Width(m.width).
		Background(s.StatusBarBG)

	return bar.Render(left + msg + right)
}
