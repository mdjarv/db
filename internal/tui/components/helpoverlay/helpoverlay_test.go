package helpoverlay

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mdjarv/db/internal/tui/core"
	"github.com/mdjarv/db/internal/tui/pane"
)

func TestOpenClose(t *testing.T) {
	m := New()
	if m.IsActive() {
		t.Error("should start inactive")
	}
	m.Open()
	if !m.IsActive() {
		t.Error("should be active after Open")
	}
	m.Close()
	if m.IsActive() {
		t.Error("should be inactive after Close")
	}
}

func TestOpenTopic(t *testing.T) {
	m := New()
	m.OpenTopic("commands")
	if !m.IsActive() {
		t.Error("should be active after OpenTopic")
	}
	if m.topic != "commands" {
		t.Errorf("topic = %q, want %q", m.topic, "commands")
	}
}

func TestDismissKeys(t *testing.T) {
	for _, key := range []string{"?", "esc", "q"} {
		m := New()
		m.Open()
		var msg tea.KeyMsg
		switch key {
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		default:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
		}
		m.Update(msg)
		if m.IsActive() {
			t.Errorf("key %q should dismiss overlay", key)
		}
	}
}

func TestScrollKeys(t *testing.T) {
	m := New()
	m.Open()
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if m.scroll != 1 {
		t.Errorf("scroll after j = %d, want 1", m.scroll)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if m.scroll != 0 {
		t.Errorf("scroll after k = %d, want 0", m.scroll)
	}
	// k at 0 should not go negative
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if m.scroll != 0 {
		t.Errorf("scroll should not go negative, got %d", m.scroll)
	}
	if !m.IsActive() {
		t.Error("scroll keys should not dismiss overlay")
	}
}

func TestContextSections_TableList(t *testing.T) {
	sections := contextSections(pane.TableList, core.ModeNormal)
	found := false
	for _, s := range sections {
		if s.Title == "Table List" {
			found = true
		}
		if s.Title == "Results" {
			t.Error("TableList context should not include Results section")
		}
	}
	if !found {
		t.Error("TableList context should include Table List section")
	}
}

func TestContextSections_ResultView(t *testing.T) {
	sections := contextSections(pane.ResultView, core.ModeNormal)
	found := false
	for _, s := range sections {
		if s.Title == "Results" {
			found = true
		}
		if s.Title == "Table List" {
			t.Error("ResultView context should not include Table List section")
		}
	}
	if !found {
		t.Error("ResultView context should include Results section")
	}
}

func TestContextSections_VisualMode(t *testing.T) {
	sections := contextSections(pane.ResultView, core.ModeVisualLine)
	found := false
	for _, s := range sections {
		if s.Title == "Visual Mode" {
			found = true
		}
	}
	if !found {
		t.Error("ResultView in visual mode should include Visual Mode section")
	}
}

func TestTopicSections_Valid(t *testing.T) {
	sections := topicSections("commands")
	if len(sections) == 0 {
		t.Error("commands topic should have sections")
	}
}

func TestTopicSections_Invalid(t *testing.T) {
	sections := topicSections("nonexistent")
	if len(sections) == 0 {
		t.Error("unknown topic should return fallback section")
	}
	if !strings.Contains(sections[0].Title, "Unknown topic") {
		t.Errorf("expected unknown topic title, got %q", sections[0].Title)
	}
}

func TestTopicNames(t *testing.T) {
	names := TopicNames()
	if len(names) == 0 {
		t.Error("should have at least one topic")
	}
	has := func(name string) bool {
		for _, n := range names {
			if n == name {
				return true
			}
		}
		return false
	}
	for _, want := range []string{"navigation", "modes", "editor", "results", "commands"} {
		if !has(want) {
			t.Errorf("missing topic %q", want)
		}
	}
}

func TestViewRenders(t *testing.T) {
	m := New()
	m.Open()
	out := m.View(pane.TableList, core.ModeNormal, 80, 40)
	if out == "" {
		t.Error("View should produce output when active")
	}
	if !strings.Contains(out, "Keybindings") {
		t.Error("View should contain header")
	}
}

func TestViewInactive(t *testing.T) {
	m := New()
	out := m.View(pane.TableList, core.ModeNormal, 80, 40)
	if out != "" {
		t.Error("View should be empty when inactive")
	}
}
