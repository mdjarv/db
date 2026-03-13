// Package pane defines the pane interface and focus manager.
package pane

import tea "github.com/charmbracelet/bubbletea"

// Pane is the interface for TUI panes.
type Pane interface {
	Update(msg tea.Msg) (Pane, tea.Cmd)
	View() string
	Focused() bool
	SetFocused(focused bool)
	SetSize(width, height int)
}
