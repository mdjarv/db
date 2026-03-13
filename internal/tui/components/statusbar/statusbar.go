// Package statusbar implements the bottom status bar.
package statusbar

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/tui/core"
)

// Model is the status bar state.
type Model struct {
	mode    core.Mode
	connStr string
	message string
	txMode  string
	width   int
}

// New creates a status bar.
func New() *Model {
	return &Model{
		txMode: "auto",
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

// View renders the status bar.
func (m *Model) View() string {
	modeStyle := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1)

	switch m.mode {
	case core.ModeInsert:
		modeStyle = modeStyle.
			Background(lipgloss.Color("28")).
			Foreground(lipgloss.Color("15"))
	case core.ModeCommand:
		modeStyle = modeStyle.
			Background(lipgloss.Color("166")).
			Foreground(lipgloss.Color("15"))
	default:
		modeStyle = modeStyle.
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("15"))
	}

	modeStr := modeStyle.Render(m.mode.String())

	var connStr string
	if m.connStr != "" {
		connStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("78")).
			Padding(0, 1)
		connStr = connStyle.Render("\uf1c0 " + m.connStr)
	} else {
		connStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Padding(0, 1)
		connStr = connStyle.Render("\uf1c0 disconnected")
	}

	txStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Padding(0, 1)
	txStr := txStyle.Render(fmt.Sprintf("tx:%s", m.txMode))

	left := modeStr + connStr
	right := txStr

	msgStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	msgW := max(m.width-leftW-rightW, 0)

	msg := msgStyle.Width(msgW).Render(m.message)

	bar := lipgloss.NewStyle().
		Width(m.width).
		Background(lipgloss.Color("236"))

	return bar.Render(left + msg + right)
}
