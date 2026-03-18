package statusbar

import (
	"testing"
	"time"

	"github.com/mdjarv/db/internal/tui/core"
)

func TestSetError_ReturnsClearCmd(t *testing.T) {
	m := New()
	m.SetErrorTimeout(10 * time.Millisecond)

	cmd := m.SetError("boom")
	if cmd == nil {
		t.Fatal("SetError should return a non-nil Cmd")
	}
	if m.message != "boom" {
		t.Errorf("message = %q, want %q", m.message, "boom")
	}
	if m.msgLevel != MsgError {
		t.Errorf("level = %d, want MsgError (%d)", m.msgLevel, MsgError)
	}
}

func TestHandleClearError_MatchingID(t *testing.T) {
	m := New()
	m.SetError("err1")
	id := m.errorID

	cleared := m.HandleClearError(core.ClearErrorMsg{ID: id})
	if !cleared {
		t.Error("HandleClearError should return true for matching ID")
	}
	if m.message != "" {
		t.Errorf("message should be empty, got %q", m.message)
	}
	if m.msgLevel != MsgInfo {
		t.Errorf("level should be MsgInfo after clear")
	}
}

func TestHandleClearError_StaleID(t *testing.T) {
	m := New()
	m.SetError("err1")
	staleID := m.errorID

	// second error replaces the first
	m.SetError("err2")

	cleared := m.HandleClearError(core.ClearErrorMsg{ID: staleID})
	if cleared {
		t.Error("stale ID should not clear newer error")
	}
	if m.message != "err2" {
		t.Errorf("message should still be %q, got %q", "err2", m.message)
	}
}

func TestSetSuccess_ClearsError(t *testing.T) {
	m := New()
	m.SetError("err")
	m.SetSuccess("ok")

	if m.message != "ok" {
		t.Errorf("message = %q, want %q", m.message, "ok")
	}
	if m.msgLevel != MsgInfo {
		t.Errorf("level should be MsgInfo after SetSuccess")
	}
}

func TestSetMessage_DoesNotOverwriteError(t *testing.T) {
	m := New()
	m.SetError("err")
	m.SetMessage("info")

	if m.message != "err" {
		t.Errorf("error should not be overwritten, got %q", m.message)
	}
}
