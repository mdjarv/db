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
	ActionModeVisualLine
	ActionModeVisualBlock
	ActionBufferNext
	ActionBufferPrev
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
	{Action: ActionModeVisualLine, Key: "V", Desc: "visual line mode", Mode: core.ModeNormal},
	{Action: ActionModeVisualBlock, Key: "v", Desc: "visual block mode", Mode: core.ModeNormal},
	{Action: ActionModeNormal, Key: "esc", Desc: "cancel visual", Mode: core.ModeVisualLine},
	{Action: ActionModeNormal, Key: "esc", Desc: "cancel visual", Mode: core.ModeVisualBlock},
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

	sb.WriteString("\nTable List:\n")
	sb.WriteString("  j/k   - navigate tables\n")
	sb.WriteString("  gg/G  - top/bottom\n")
	sb.WriteString("  /     - filter tables\n")
	sb.WriteString("  Enter - query selected table\n")
	sb.WriteString("  d     - describe table\n")
	sb.WriteString("  y     - yank table name\n")
	sb.WriteString("  R     - refresh schema\n")

	sb.WriteString("\nResults:\n")
	sb.WriteString("  j/k       - navigate items\n")
	sb.WriteString("  g/G       - top/bottom\n")
	sb.WriteString("  h/l       - left/right\n")
	sb.WriteString("  0/$       - first/last column\n")
	sb.WriteString("  ctrl+d/u  - half page down/up\n")
	sb.WriteString("  ctrl+f/b  - full page down/up\n")
	sb.WriteString("  enter     - inspect cell\n")
	sb.WriteString("  y         - yank cell\n")
	sb.WriteString("  Y         - yank row\n")

	sb.WriteString("\nVisual (results):\n")
	sb.WriteString("  V     - visual line (rows)\n")
	sb.WriteString("  v     - visual block (rectangle)\n")
	sb.WriteString("  j/k   - extend row selection\n")
	sb.WriteString("  h/l   - extend column selection\n")
	sb.WriteString("  tab   - toggle row/col axis (V mode)\n")
	sb.WriteString("  y     - yank selection (CSV)\n")
	sb.WriteString("  esc   - cancel\n")

	sb.WriteString("\nEditing (results):\n")
	sb.WriteString("  e       - edit cell\n")
	sb.WriteString("  dR      - delete row\n")
	sb.WriteString("  oR      - insert row\n")
	sb.WriteString("  ctrl+z  - undo last change\n")
	sb.WriteString("  :commit   - apply changes\n")
	sb.WriteString("  :rollback - discard changes\n")
	sb.WriteString("  :changes  - list changes\n")

	sb.WriteString("\nBuffers:\n")
	sb.WriteString("  gt/gT       - next/prev buffer\n")
	sb.WriteString("  :new/:enew  - new buffer\n")
	sb.WriteString("  :bd         - close buffer\n")
	sb.WriteString("  :bn/:bp     - next/prev buffer\n")
	sb.WriteString("  :b N        - switch to buffer N\n")
	sb.WriteString("  :ls         - list buffers\n")

	sb.WriteString("\nCommands:\n")
	sb.WriteString("  :q    - quit\n")
	sb.WriteString("  :w    - run query\n")
	sb.WriteString("  :set  - change setting\n")
	sb.WriteString("  :export csv|json|sql <file>\n")
	sb.WriteString("  :theme <name> - switch theme\n")

	sb.WriteString("\n  ?     - toggle this help\n")
	sb.WriteString("  ctrl+c - quit\n")

	return sb.String()
}
