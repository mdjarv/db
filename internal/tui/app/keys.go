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
	{Action: ActionFocusNext, Key: "tab", Desc: "next pane", Mode: core.ModeNormal},
	{Action: ActionFocusPrev, Key: "shift+tab", Desc: "prev pane", Mode: core.ModeNormal},
	{Action: ActionFocusPane1, Key: "1", Desc: "pane 1", Mode: core.ModeNormal},
	{Action: ActionFocusPane2, Key: "2", Desc: "pane 2", Mode: core.ModeNormal},
	{Action: ActionFocusPane3, Key: "3", Desc: "pane 3", Mode: core.ModeNormal},
}

var ctrlWBindings = map[string]Action{
	"h": ActionFocusLeft,
	"j": ActionFocusDown,
	"k": ActionFocusUp,
	"l": ActionFocusRight,
	"+": ActionResizeGrow,
	"-": ActionResizeShrink,
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

// MatchCtrlW returns the action for a Ctrl-w prefixed key.
func MatchCtrlW(msg tea.KeyMsg) Action {
	key := msg.String()
	if a, ok := ctrlWBindings[key]; ok {
		return a
	}
	return ActionNone
}

// HelpText returns formatted keybinding help.
func HelpText() string {
	var sb strings.Builder
	sb.WriteString("Keybindings:\n\n")

	sb.WriteString("Global:\n")
	for _, b := range globalBindings {
		modeName := "any"
		if b.Mode >= 0 {
			modeName = b.Mode.String()
		}
		sb.WriteString("  " + b.Key + " - " + b.Desc + " (" + modeName + ")\n")
	}

	sb.WriteString("\nCtrl-w prefix:\n")
	for key, action := range ctrlWBindings {
		desc := actionDesc(action)
		sb.WriteString("  Ctrl-w " + key + " - " + desc + "\n")
	}

	sb.WriteString("\nPanes:\n")
	sb.WriteString("  j/k     - navigate items\n")
	sb.WriteString("  g/G     - top/bottom\n")
	sb.WriteString("  h/l     - left/right (editor)\n")
	sb.WriteString("\nCommands:\n")
	sb.WriteString("  :q      - quit\n")
	sb.WriteString("  :w      - run query\n")
	sb.WriteString("  :set    - change setting\n")

	return sb.String()
}

func actionDesc(a Action) string {
	switch a {
	case ActionFocusLeft:
		return "focus left"
	case ActionFocusRight:
		return "focus right"
	case ActionFocusUp:
		return "focus up"
	case ActionFocusDown:
		return "focus down"
	case ActionResizeGrow:
		return "grow pane"
	case ActionResizeShrink:
		return "shrink pane"
	default:
		return ""
	}
}
