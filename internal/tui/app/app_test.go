package app

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
)

func newTestModel(t *testing.T) *teatest.TestModel {
	t.Helper()
	return teatest.NewTestModel(t, New(), teatest.WithInitialTermSize(120, 40))
}

func finalModel(t *testing.T, tm *teatest.TestModel) Model {
	t.Helper()
	m := tm.FinalModel(t, teatest.WithFinalTimeout(3*time.Second))
	return m.(Model)
}

func TestApp_InitialState(t *testing.T) {
	tm := newTestModel(t)
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	m := finalModel(t, tm)

	if !m.ready {
		t.Error("model should be ready after WindowSizeMsg")
	}
	if m.mode.String() != "NORMAL" {
		t.Errorf("initial mode = %s, want NORMAL", m.mode)
	}
}

func TestApp_ModeTransitions(t *testing.T) {
	tm := newTestModel(t)

	// normal -> insert
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	// insert -> normal
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	// normal -> command
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(":")})
	// command -> normal via esc
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	m := finalModel(t, tm)

	if m.mode.String() != "NORMAL" {
		t.Errorf("mode after transitions = %s, want NORMAL", m.mode)
	}
}

func TestApp_PaneFocusCtrlHJKL(t *testing.T) {
	tm := newTestModel(t)

	// start on TableList, ctrl+l -> right (QueryEditor)
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlL})
	// ctrl+j -> down (ResultView)
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlJ})
	// ctrl+h -> left (TableList)
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlH})

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	m := finalModel(t, tm)

	if m.panes == nil {
		t.Fatal("panes manager is nil")
	}
}

func TestApp_QuitCommand(t *testing.T) {
	tm := newTestModel(t)

	// : to enter command mode, type "q", enter to execute
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(":")})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	m := finalModel(t, tm)
	_ = m // program should have quit
}

func TestApp_HelpToggle(t *testing.T) {
	tm := newTestModel(t)

	// ? to show help
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	// any key to dismiss
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	m := finalModel(t, tm)

	if m.showHelp {
		t.Error("help should be dismissed")
	}
}

func TestApp_ViewRenders(t *testing.T) {
	tm := newTestModel(t)
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})

	out := tm.FinalOutput(t, teatest.WithFinalTimeout(3*time.Second))
	buf := make([]byte, 32768)
	n, _ := out.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "NORMAL") {
		t.Error("output should contain mode indicator")
	}
}
