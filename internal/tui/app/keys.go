package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mdjarv/db/internal/tui/core"
	"github.com/mdjarv/db/internal/tui/pane"
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
	ActionConnSelector
	ActionCommit
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
	{Action: ActionConnSelector, Key: "ctrl+o", Desc: "switch connection", Mode: core.ModeNormal},
	{Action: ActionCommit, Key: "ctrl+s", Desc: "commit changes", Mode: core.ModeNormal},
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

// HelpText returns formatted keybinding help scoped to the active pane.
func HelpText(activePane pane.ID) string {
	var sb strings.Builder
	sb.WriteString("Keybindings\n")

	// Global section — always shown
	sb.WriteString("\nNavigation:\n")
	sb.WriteString("  ctrl+h/j/k/l  focus left/down/up/right\n")
	sb.WriteString("  tab/shift+tab  cycle panes\n")
	sb.WriteString("  1/2/3          jump to pane\n")
	sb.WriteString("  +/-            grow/shrink pane\n")

	sb.WriteString("\nModes:\n")
	sb.WriteString("  i    insert mode\n")
	sb.WriteString("  esc  normal mode\n")
	sb.WriteString("  :    command mode\n")

	sb.WriteString("\nConnection Selector:\n")
	sb.WriteString("  ctrl+o    switch connection\n")
	sb.WriteString("  a         add connection (in selector)\n")
	sb.WriteString("  e         edit connection (in selector)\n")
	sb.WriteString("  d         delete connection (in selector)\n")

	sb.WriteString("\nGeneral:\n")
	sb.WriteString("  :q  :help [topic]\n")

	sb.WriteString("\nQuery:\n")
	sb.WriteString("  :run  :clear  :set <opt>\n")

	sb.WriteString("\nData:\n")
	sb.WriteString("  :commit  :rollback  :changes  :export <fmt> <file>\n")

	sb.WriteString("\nBuffers:\n")
	sb.WriteString("  gt/gT  :new  :bd  :bn/:bp  :b N  :ls\n")

	sb.WriteString("\nConnection:\n")
	sb.WriteString("  :connect  :refresh  :dump [table]  :schema [name]\n")

	sb.WriteString("\nAppearance:\n")
	sb.WriteString("  :theme [name]\n")

	// Pane-specific section
	switch activePane {
	case pane.TableList:
		sb.WriteString("\nTable List:\n")
		sb.WriteString("  j/k    navigate tables\n")
		sb.WriteString("  gg/G   top (header)/bottom\n")
		sb.WriteString("  /      filter tables\n")
		sb.WriteString("  Enter  query selected table\n")
		sb.WriteString("  Space  context menu\n")
		sb.WriteString("  d      describe table\n")
		sb.WriteString("  D      dump table (database on header)\n")
		sb.WriteString("  y      yank table name\n")
		sb.WriteString("  R      refresh schema\n")

	case pane.QueryEditor:
		sb.WriteString("\nQuery Editor:\n")
		sb.WriteString("  j/k/h/l  navigate\n")
		sb.WriteString("  0/$      line start/end\n")
		sb.WriteString("  w/b      word forward/back\n")
		sb.WriteString("  gg/G     top/bottom\n")
		sb.WriteString("  dd       delete line\n")
		sb.WriteString("  D        delete to end\n")
		sb.WriteString("  u        undo\n")
		sb.WriteString("  a/A/I    insert variants\n")
		sb.WriteString("  o/O      open line below/above\n")
		sb.WriteString("  x        delete char\n")
		sb.WriteString("  y        yank, p/P paste\n")
		sb.WriteString("  v/V      visual mode\n")
		sb.WriteString("  Enter    run query\n")

	case pane.ResultView:
		sb.WriteString("\nResults:\n")
		sb.WriteString("  j/k/h/l   navigate\n")
		sb.WriteString("  g/G       top/bottom\n")
		sb.WriteString("  0/$       first/last column\n")
		sb.WriteString("  ctrl+d/u  half page down/up\n")
		sb.WriteString("  ctrl+f/b  full page down/up\n")
		sb.WriteString("  e         edit cell\n")
		sb.WriteString("  y/Y       yank cell/row\n")

		sb.WriteString("\nVisual:\n")
		sb.WriteString("  V/v       line/block select\n")
		sb.WriteString("  j/k/h/l   extend selection\n")
		sb.WriteString("  tab       toggle axis (V mode)\n")
		sb.WriteString("  y         yank selection\n")
		sb.WriteString("  esc       cancel\n")

		sb.WriteString("\nEditing:\n")
		sb.WriteString("  dR        delete row\n")
		sb.WriteString("  oR        insert row\n")
		sb.WriteString("  u         undo last change\n")
		sb.WriteString("  ctrl+s    commit changes\n")
		sb.WriteString("  :commit   apply changes\n")
		sb.WriteString("  :rollback discard changes\n")
	}

	return sb.String()
}
