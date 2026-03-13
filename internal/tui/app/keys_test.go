package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mdjarv/db/internal/tui/core"
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
	case "ctrl+w":
		return tea.KeyMsg{Type: tea.KeyCtrlW}
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
		{"tab", ActionFocusNext},
		{"shift+tab", ActionFocusPrev},
		{"1", ActionFocusPane1},
		{"2", ActionFocusPane2},
		{"3", ActionFocusPane3},
	}
	for _, tt := range tests {
		got := MatchGlobal(keyMsg(tt.key), core.ModeNormal)
		if got != tt.want {
			t.Errorf("MatchGlobal(%q, Normal) = %d, want %d", tt.key, got, tt.want)
		}
	}
}

func TestMatchGlobal_InsertMode(t *testing.T) {
	// i should not trigger insert mode when already in insert
	got := MatchGlobal(keyMsg("i"), core.ModeInsert)
	if got != ActionNone {
		t.Errorf("MatchGlobal(i, Insert) = %d, want ActionNone", got)
	}

	// esc works in any mode
	got = MatchGlobal(keyMsg("esc"), core.ModeInsert)
	if got != ActionModeNormal {
		t.Errorf("MatchGlobal(esc, Insert) = %d, want ActionModeNormal", got)
	}

	// ctrl+c works in any mode
	got = MatchGlobal(keyMsg("ctrl+c"), core.ModeInsert)
	if got != ActionQuit {
		t.Errorf("MatchGlobal(ctrl+c, Insert) = %d, want ActionQuit", got)
	}
}

func TestMatchCtrlW(t *testing.T) {
	tests := []struct {
		key  string
		want Action
	}{
		{"h", ActionFocusLeft},
		{"j", ActionFocusDown},
		{"k", ActionFocusUp},
		{"l", ActionFocusRight},
		{"+", ActionResizeGrow},
		{"-", ActionResizeShrink},
		{"x", ActionNone},
	}
	for _, tt := range tests {
		got := MatchCtrlW(keyMsg(tt.key))
		if got != tt.want {
			t.Errorf("MatchCtrlW(%q) = %d, want %d", tt.key, got, tt.want)
		}
	}
}

func TestHelpText(t *testing.T) {
	text := HelpText()
	if text == "" {
		t.Error("HelpText() returned empty string")
	}
	// Should contain key sections
	for _, want := range []string{"Global:", "Ctrl-w prefix:", "Panes:", "Commands:"} {
		if !contains(text, want) {
			t.Errorf("HelpText() missing section %q", want)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && stringContains(s, substr)
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
