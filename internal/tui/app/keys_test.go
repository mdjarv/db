package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mdjarv/db/internal/tui/core"
	"github.com/mdjarv/db/internal/tui/pane"
)

func keyMsg(s string) tea.KeyMsg {
	switch s {
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case "ctrl+h":
		return tea.KeyMsg{Type: tea.KeyCtrlH}
	case "ctrl+j":
		return tea.KeyMsg{Type: tea.KeyCtrlJ}
	case "ctrl+k":
		return tea.KeyMsg{Type: tea.KeyCtrlK}
	case "ctrl+l":
		return tea.KeyMsg{Type: tea.KeyCtrlL}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func TestMatchGlobal_NormalMode(t *testing.T) {
	tests := []struct {
		key  string
		want Action
	}{
		{"i", ActionModeInsert},
		{":", ActionModeCommand},
		{"?", ActionHelp},
		{"ctrl+h", ActionFocusLeft},
		{"ctrl+j", ActionFocusDown},
		{"ctrl+k", ActionFocusUp},
		{"ctrl+l", ActionFocusRight},
		{"tab", ActionFocusNext},
		{"shift+tab", ActionFocusPrev},
		{"1", ActionFocusPane1},
		{"2", ActionFocusPane2},
		{"3", ActionFocusPane3},
		{"+", ActionResizeGrow},
		{"-", ActionResizeShrink},
	}
	for _, tt := range tests {
		got := MatchGlobal(keyMsg(tt.key), core.ModeNormal)
		if got != tt.want {
			t.Errorf("MatchGlobal(%q, Normal) = %d, want %d", tt.key, got, tt.want)
		}
	}
}

func TestMatchGlobal_InsertMode(t *testing.T) {
	got := MatchGlobal(keyMsg("i"), core.ModeInsert)
	if got != ActionNone {
		t.Errorf("MatchGlobal(i, Insert) = %d, want ActionNone", got)
	}

	got = MatchGlobal(keyMsg("esc"), core.ModeInsert)
	if got != ActionModeNormal {
		t.Errorf("MatchGlobal(esc, Insert) = %d, want ActionModeNormal", got)
	}

	got = MatchGlobal(keyMsg("ctrl+c"), core.ModeInsert)
	if got != ActionQuit {
		t.Errorf("MatchGlobal(ctrl+c, Insert) = %d, want ActionQuit", got)
	}

	// ctrl+hjkl should not work in insert mode
	got = MatchGlobal(keyMsg("ctrl+h"), core.ModeInsert)
	if got != ActionNone {
		t.Errorf("MatchGlobal(ctrl+h, Insert) = %d, want ActionNone", got)
	}
}

func TestHelpText(t *testing.T) {
	// TableList pane should show Table List section
	text := HelpText(pane.TableList)
	if text == "" {
		t.Error("HelpText(TableList) returned empty string")
	}
	for _, want := range []string{"Navigation:", "Modes:", "Table List:"} {
		if !strings.Contains(text, want) {
			t.Errorf("HelpText(TableList) missing section %q", want)
		}
	}
	if strings.Contains(text, "Results:") {
		t.Error("HelpText(TableList) should not contain Results section")
	}

	// ResultView pane should show Results section
	text = HelpText(pane.ResultView)
	for _, want := range []string{"Navigation:", "Results:", "Visual:", "Editing:"} {
		if !strings.Contains(text, want) {
			t.Errorf("HelpText(ResultView) missing section %q", want)
		}
	}
	if strings.Contains(text, "Table List:") {
		t.Error("HelpText(ResultView) should not contain Table List section")
	}

	// QueryEditor pane should show Query Editor section
	text = HelpText(pane.QueryEditor)
	if !strings.Contains(text, "Query Editor:") {
		t.Error("HelpText(QueryEditor) missing Query Editor section")
	}
}
