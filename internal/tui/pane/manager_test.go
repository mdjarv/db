package pane

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type stubPane struct {
	focused bool
	width   int
	height  int
}

func (s *stubPane) Update(_ tea.Msg) (Pane, tea.Cmd) { return s, nil }
func (s *stubPane) View() string                     { return "" }
func (s *stubPane) Focused() bool                    { return s.focused }
func (s *stubPane) SetFocused(f bool)                { s.focused = f }
func (s *stubPane) SetSize(w, h int)                 { s.width = w; s.height = h }

func newTestManager() (*Manager, []*stubPane) {
	m := NewManager()
	panes := make([]*stubPane, 3)
	for i, id := range []ID{TableList, QueryEditor, ResultView} {
		panes[i] = &stubPane{}
		m.Register(id, panes[i])
	}
	m.SetActive(TableList)
	return m, panes
}

func TestCycleForward(t *testing.T) {
	m, panes := newTestManager()

	if m.ActiveID() != TableList {
		t.Fatalf("initial active = %d, want TableList", m.ActiveID())
	}
	if !panes[0].focused {
		t.Error("TableList should be focused initially")
	}

	m.CycleForward()
	if m.ActiveID() != QueryEditor {
		t.Errorf("after CycleForward: active = %d, want QueryEditor", m.ActiveID())
	}
	if panes[0].focused {
		t.Error("TableList should not be focused after cycle")
	}
	if !panes[1].focused {
		t.Error("QueryEditor should be focused after cycle")
	}

	m.CycleForward()
	if m.ActiveID() != ResultView {
		t.Errorf("after 2x CycleForward: active = %d, want ResultView", m.ActiveID())
	}

	m.CycleForward()
	if m.ActiveID() != TableList {
		t.Errorf("after 3x CycleForward: active = %d, want TableList (wrap)", m.ActiveID())
	}
}

func TestCycleBackward(t *testing.T) {
	m, _ := newTestManager()

	m.CycleBackward()
	if m.ActiveID() != ResultView {
		t.Errorf("CycleBackward from TableList = %d, want ResultView", m.ActiveID())
	}
}

func TestFocusDirections(t *testing.T) {
	m, _ := newTestManager()

	m.SetActive(QueryEditor)
	m.FocusLeft()
	if m.ActiveID() != TableList {
		t.Errorf("FocusLeft from QueryEditor = %d, want TableList", m.ActiveID())
	}

	m.FocusRight()
	if m.ActiveID() != QueryEditor {
		t.Errorf("FocusRight from TableList = %d, want QueryEditor", m.ActiveID())
	}

	m.FocusDown()
	if m.ActiveID() != ResultView {
		t.Errorf("FocusDown from QueryEditor = %d, want ResultView", m.ActiveID())
	}

	m.FocusUp()
	if m.ActiveID() != QueryEditor {
		t.Errorf("FocusUp from ResultView = %d, want QueryEditor", m.ActiveID())
	}
}

func TestFocusByNumber(t *testing.T) {
	m, _ := newTestManager()

	m.FocusByNumber(2)
	if m.ActiveID() != QueryEditor {
		t.Errorf("FocusByNumber(2) = %d, want QueryEditor", m.ActiveID())
	}

	m.FocusByNumber(3)
	if m.ActiveID() != ResultView {
		t.Errorf("FocusByNumber(3) = %d, want ResultView", m.ActiveID())
	}

	m.FocusByNumber(1)
	if m.ActiveID() != TableList {
		t.Errorf("FocusByNumber(1) = %d, want TableList", m.ActiveID())
	}

	// Out of range - should not change
	m.FocusByNumber(0)
	if m.ActiveID() != TableList {
		t.Errorf("FocusByNumber(0) should not change, got %d", m.ActiveID())
	}

	m.FocusByNumber(4)
	if m.ActiveID() != TableList {
		t.Errorf("FocusByNumber(4) should not change, got %d", m.ActiveID())
	}
}

func TestSetSize(t *testing.T) {
	m, panes := newTestManager()

	p := m.Get(TableList)
	p.SetSize(20, 30)

	if panes[0].width != 20 || panes[0].height != 30 {
		t.Errorf("SetSize(20,30) = %dx%d, want 20x30", panes[0].width, panes[0].height)
	}
}
