package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mdjarv/db/internal/tui/core"
)

// Action identifies a keybinding action.
type Action int

// Keybinding actions.
const (
	ActionNone Action = iota
	ActionQuit
	ActionModeNormal
	ActionModeInsert
	ActionModeCommand
	ActionFocusLeft
	ActionFocusRight
	ActionFocusUp
	ActionFocusDown
	ActionFocusNext
	ActionFocusPrev
	ActionFocusPane1
	ActionFocusPane2
	ActionFocusPane3
	ActionHelp
	ActionResizeGrow
	ActionResizeShrink
)

// Binding maps a key to an action in a specific mode.
type Binding struct {
	Action Action
	Key    string
	Desc   string
	Mode   core.Mode
}

var globalBindings = []Binding{
	{Action: ActionModeNormal, Key: "esc", Desc: "normal mode", Mode: -1},
	{Action: ActionModeInsert, Key: "i", Desc: "insert mode", Mode: core.ModeNormal},
	{Action: ActionModeCommand, Key: ":", Desc: "command mode", Mode: core.ModeNormal},
	{Action: ActionHelp, Key: "?", Desc: "show help", Mode: core.ModeNormal},
	{Action: ActionQuit, Key: "ctrl+c", Desc: "quit", Mode: -1},
	{Action: ActionFocusLeft, Key: "ctrl+h", Desc: "focus left", Mode: core.ModeNormal},
	{Action: ActionFocusDown, Key: "ctrl+j", Desc: "focus down", Mode: core.ModeNormal},
	{Action: ActionFocusUp, Key: "ctrl+k", Desc: "focus up", Mode: core.ModeNormal},
	{Action: ActionFocusRight, Key: "ctrl+l", Desc: "focus right", Mode: core.ModeNormal},
	{Action: ActionFocusNext, Key: "tab", Desc: "next pane", Mode: core.ModeNormal},
	{Action: ActionFocusPrev, Key: "shift+tab", Desc: "prev pane", Mode: core.ModeNormal},
	{Action: ActionFocusPane1, Key: "1", Desc: "pane 1", Mode: core.ModeNormal},
	{Action: ActionFocusPane2, Key: "2", Desc: "pane 2", Mode: core.ModeNormal},
	{Action: ActionFocusPane3, Key: "3", Desc: "pane 3", Mode: core.ModeNormal},
	{Action: ActionResizeGrow, Key: "+", Desc: "grow pane", Mode: core.ModeNormal},
	{Action: ActionResizeShrink, Key: "-", Desc: "shrink pane", Mode: core.ModeNormal},
}

// MatchGlobal returns the action matching a key in the given mode.
func MatchGlobal(msg tea.KeyMsg, mode core.Mode) Action {
	key := msg.String()
	for _, b := range globalBindings {
		if b.Key == key && (b.Mode == -1 || b.Mode == mode) {
			return b.Action
		}
	}
	return ActionNone
}

// HelpText returns formatted keybinding help.
func HelpText() string {
	var sb strings.Builder
	sb.WriteString("Keybindings:\n\n")

	sb.WriteString("Navigation:\n")
	sb.WriteString("  ctrl+h/j/k/l - focus left/down/up/right\n")
	sb.WriteString("  tab/shift+tab - cycle panes\n")
	sb.WriteString("  1/2/3         - jump to pane\n")
	sb.WriteString("  +/-           - grow/shrink pane\n")

	sb.WriteString("\nModes:\n")
	sb.WriteString("  i     - insert mode\n")
	sb.WriteString("  esc   - normal mode\n")
	sb.WriteString("  :     - command mode\n")

	sb.WriteString("\nPanes:\n")
	sb.WriteString("  j/k   - navigate items\n")
	sb.WriteString("  g/G   - top/bottom\n")
	sb.WriteString("  h/l   - left/right (editor)\n")

	sb.WriteString("\nCommands:\n")
	sb.WriteString("  :q    - quit\n")
	sb.WriteString("  :w    - run query\n")
	sb.WriteString("  :set  - change setting\n")

	sb.WriteString("\n  ?     - toggle this help\n")
	sb.WriteString("  ctrl+c - quit\n")

	return sb.String()
}
