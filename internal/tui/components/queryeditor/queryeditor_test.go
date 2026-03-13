package queryeditor

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mdjarv/db/internal/tui/core"
)

func keyMsg(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func ctrlMsg(t tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: t}
}

func setup(text string) *Model {
	m := New()
	m.SetFocused(true)
	m.SetSize(80, 24)
	if text != "" {
		m.lines = strings.Split(text, "\n")
	}
	return m
}

func enterInsert(m *Model) {
	m.Update(core.ModeChangedMsg{Mode: core.ModeInsert})
}

func TestCursorMovement(t *testing.T) {
	m := setup("hello world\nfoo bar\nbaz")

	m.Update(keyMsg("j"))
	if m.cursorY != 1 {
		t.Errorf("j: cursorY = %d, want 1", m.cursorY)
	}

	m.Update(keyMsg("k"))
	if m.cursorY != 0 {
		t.Errorf("k: cursorY = %d, want 0", m.cursorY)
	}

	m.Update(keyMsg("l"))
	if m.cursorX != 1 {
		t.Errorf("l: cursorX = %d, want 1", m.cursorX)
	}

	m.Update(keyMsg("h"))
	if m.cursorX != 0 {
		t.Errorf("h: cursorX = %d, want 0", m.cursorX)
	}

	m.Update(keyMsg("$"))
	if m.cursorX != 10 { // "hello world" len=11, last char index=10
		t.Errorf("$: cursorX = %d, want 10", m.cursorX)
	}

	m.Update(keyMsg("0"))
	if m.cursorX != 0 {
		t.Errorf("0: cursorX = %d, want 0", m.cursorX)
	}
}

func TestWordMovement(t *testing.T) {
	m := setup("hello world foo")

	m.Update(keyMsg("w"))
	if m.cursorX != 6 {
		t.Errorf("w: cursorX = %d, want 6", m.cursorX)
	}

	m.Update(keyMsg("w"))
	if m.cursorX != 12 {
		t.Errorf("w2: cursorX = %d, want 12", m.cursorX)
	}

	m.Update(keyMsg("b"))
	if m.cursorX != 6 {
		t.Errorf("b: cursorX = %d, want 6", m.cursorX)
	}

	m.Update(keyMsg("b"))
	if m.cursorX != 0 {
		t.Errorf("b2: cursorX = %d, want 0", m.cursorX)
	}
}

func TestWordMovementCrossLine(t *testing.T) {
	m := setup("hello\nworld")

	m.cursorX = 4
	m.Update(keyMsg("w"))
	if m.cursorY != 1 || m.cursorX != 0 {
		t.Errorf("w cross line: got (%d,%d), want (1,0)", m.cursorY, m.cursorX)
	}

	m.Update(keyMsg("b"))
	if m.cursorY != 0 {
		t.Errorf("b cross line: cursorY = %d, want 0", m.cursorY)
	}
}

func TestInsertText(t *testing.T) {
	m := setup("")
	enterInsert(m)

	m.Update(keyMsg("S"))
	m.Update(keyMsg("E"))
	m.Update(keyMsg("L"))

	if m.Content() != "SEL" {
		t.Errorf("insert: content = %q, want %q", m.Content(), "SEL")
	}
	if m.cursorX != 3 {
		t.Errorf("insert: cursorX = %d, want 3", m.cursorX)
	}
}

func TestBackspace(t *testing.T) {
	m := setup("abc")
	enterInsert(m)
	m.cursorX = 3

	m.Update(ctrlMsg(tea.KeyBackspace))
	if m.Content() != "ab" {
		t.Errorf("backspace: content = %q, want %q", m.Content(), "ab")
	}
}

func TestBackspaceJoinLines(t *testing.T) {
	m := setup("hello\nworld")
	enterInsert(m)
	m.cursorY = 1
	m.cursorX = 0

	m.Update(ctrlMsg(tea.KeyBackspace))
	if m.Content() != "helloworld" {
		t.Errorf("backspace join: content = %q, want %q", m.Content(), "helloworld")
	}
	if m.cursorX != 5 {
		t.Errorf("backspace join: cursorX = %d, want 5", m.cursorX)
	}
}

func TestNewline(t *testing.T) {
	m := setup("hello world")
	enterInsert(m)
	m.cursorX = 5

	m.Update(ctrlMsg(tea.KeyEnter))
	if m.Content() != "hello\n world" {
		t.Errorf("newline: content = %q, want %q", m.Content(), "hello\n world")
	}
	if m.cursorY != 1 || m.cursorX != 0 {
		t.Errorf("newline: cursor = (%d,%d), want (1,0)", m.cursorY, m.cursorX)
	}
}

func TestDeleteLine(t *testing.T) {
	m := setup("line1\nline2\nline3")

	m.Update(keyMsg("d"))
	m.Update(keyMsg("d"))

	if m.Content() != "line2\nline3" {
		t.Errorf("dd: content = %q, want %q", m.Content(), "line2\nline3")
	}
}

func TestDeleteLineSingle(t *testing.T) {
	m := setup("only")

	m.Update(keyMsg("d"))
	m.Update(keyMsg("d"))

	if m.Content() != "" {
		t.Errorf("dd single: content = %q, want empty", m.Content())
	}
}

func TestDeleteToEnd(t *testing.T) {
	m := setup("hello world")
	m.cursorX = 5

	m.Update(keyMsg("D"))
	if m.Content() != "hello" {
		t.Errorf("D: content = %q, want %q", m.Content(), "hello")
	}
}

func TestUndoRedo(t *testing.T) {
	m := setup("original")
	enterInsert(m)
	m.cursorX = 8

	m.Update(keyMsg("!"))

	if m.Content() != "original!" {
		t.Fatalf("after insert: content = %q", m.Content())
	}

	m.Update(core.ModeChangedMsg{Mode: core.ModeNormal})

	m.Update(keyMsg("u"))
	if m.Content() != "original" {
		t.Errorf("undo: content = %q, want %q", m.Content(), "original")
	}

	m.Update(ctrlMsg(tea.KeyCtrlR))
	if m.Content() != "original!" {
		t.Errorf("redo: content = %q, want %q", m.Content(), "original!")
	}
}

func TestGotoTopBottom(t *testing.T) {
	m := setup("a\nb\nc\nd\ne")

	m.Update(keyMsg("G"))
	if m.cursorY != 4 {
		t.Errorf("G: cursorY = %d, want 4", m.cursorY)
	}

	m.Update(keyMsg("g"))
	m.Update(keyMsg("g"))
	if m.cursorY != 0 {
		t.Errorf("gg: cursorY = %d, want 0", m.cursorY)
	}
}

func TestInsertAfter(t *testing.T) {
	m := setup("abc")
	m.cursorX = 1

	cmd := m.Update(keyMsg("a"))
	if m.cursorX != 2 {
		t.Errorf("a: cursorX = %d, want 2", m.cursorX)
	}
	if cmd == nil {
		t.Fatal("a: expected ModeChangedMsg cmd")
	}
}

func TestInsertLineStart(t *testing.T) {
	m := setup("  hello")
	m.cursorX = 5

	cmd := m.Update(keyMsg("I"))
	if m.cursorX != 2 {
		t.Errorf("I: cursorX = %d, want 2", m.cursorX)
	}
	if cmd == nil {
		t.Fatal("I: expected ModeChangedMsg cmd")
	}
}

func TestInsertEndOfLine(t *testing.T) {
	m := setup("hello")
	m.cursorX = 0

	cmd := m.Update(keyMsg("A"))
	if m.cursorX != 5 {
		t.Errorf("A: cursorX = %d, want 5", m.cursorX)
	}
	if cmd == nil {
		t.Fatal("A: expected ModeChangedMsg cmd")
	}
}

func TestOpenLineBelow(t *testing.T) {
	m := setup("line1\nline2")
	m.cursorY = 0

	cmd := m.Update(keyMsg("o"))
	if len(m.lines) != 3 {
		t.Fatalf("o: lines = %d, want 3", len(m.lines))
	}
	if m.cursorY != 1 {
		t.Errorf("o: cursorY = %d, want 1", m.cursorY)
	}
	if m.lines[1] != "" {
		t.Errorf("o: new line = %q, want empty", m.lines[1])
	}
	if cmd == nil {
		t.Fatal("o: expected ModeChangedMsg cmd")
	}
}

func TestOpenLineAbove(t *testing.T) {
	m := setup("line1\nline2")
	m.cursorY = 1

	cmd := m.Update(keyMsg("O"))
	if len(m.lines) != 3 {
		t.Fatalf("O: lines = %d, want 3", len(m.lines))
	}
	if m.cursorY != 1 {
		t.Errorf("O: cursorY = %d, want 1", m.cursorY)
	}
	if m.lines[1] != "" {
		t.Errorf("O: new line = %q, want empty", m.lines[1])
	}
	if cmd == nil {
		t.Fatal("O: expected ModeChangedMsg cmd")
	}
}

func TestLineNumbers(t *testing.T) {
	m := setup("SELECT 1;\nSELECT 2;")
	m.SetSize(40, 10)
	v := m.View()

	if !strings.Contains(v, "1 ") {
		t.Error("view should contain line number 1")
	}
	if !strings.Contains(v, "2 ") {
		t.Error("view should contain line number 2")
	}
}

func TestSetContent(t *testing.T) {
	m := setup("original")
	m.SetContent("SELECT * FROM users;")

	if m.Content() != "SELECT * FROM users;" {
		t.Errorf("SetContent: content = %q", m.Content())
	}
	if m.cursorX != 0 || m.cursorY != 0 {
		t.Errorf("SetContent: cursor = (%d,%d), want (0,0)", m.cursorY, m.cursorX)
	}
}

func TestDeleteX(t *testing.T) {
	m := setup("abc")
	m.cursorX = 1

	m.Update(keyMsg("x"))
	if m.Content() != "ac" {
		t.Errorf("x: content = %q, want %q", m.Content(), "ac")
	}
}

func TestClampXNormalMode(t *testing.T) {
	m := setup("abc")
	m.cursorX = 5
	m.clampX()
	if m.cursorX != 2 {
		t.Errorf("clampX normal: cursorX = %d, want 2", m.cursorX)
	}
}

func TestClampXInsertMode(t *testing.T) {
	m := setup("abc")
	m.mode = core.ModeInsert
	m.cursorX = 5
	m.clampX()
	if m.cursorX != 3 {
		t.Errorf("clampX insert: cursorX = %d, want 3", m.cursorX)
	}
}

func TestScrollToCursor(t *testing.T) {
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = "line"
	}
	m := setup(strings.Join(lines, "\n"))
	m.SetSize(40, 12) // viewHeight = 10

	m.cursorY = 15
	m.scrollToCursor()
	if m.offset > m.cursorY || m.cursorY >= m.offset+m.viewHeight() {
		t.Errorf("scroll down: offset=%d, cursorY=%d, vh=%d", m.offset, m.cursorY, m.viewHeight())
	}

	m.cursorY = 0
	m.scrollToCursor()
	if m.offset != 0 {
		t.Errorf("scroll up: offset = %d, want 0", m.offset)
	}
}

func TestNotFocusedIgnoresInput(t *testing.T) {
	m := setup("hello")
	m.SetFocused(false)

	m.Update(keyMsg("j"))
	if m.cursorY != 0 {
		t.Error("unfocused model should not respond to keys")
	}
}

func TestUndoStackLimit(t *testing.T) {
	m := setup("start")
	enterInsert(m)
	m.cursorX = 5

	for i := 0; i < undoLimit+10; i++ {
		m.Update(keyMsg("x"))
	}
	if len(m.undoStack) > undoLimit {
		t.Errorf("undo stack size = %d, want <= %d", len(m.undoStack), undoLimit)
	}
}
