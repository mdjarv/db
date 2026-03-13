package core

import "testing"

func TestModeString(t *testing.T) {
	tests := []struct {
		mode Mode
		want string
	}{
		{ModeNormal, "NORMAL"},
		{ModeInsert, "INSERT"},
		{ModeCommand, "COMMAND"},
		{ModeVisual, "VISUAL"},
	}
	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.want {
			t.Errorf("Mode(%d).String() = %q, want %q", tt.mode, got, tt.want)
		}
	}
}

func TestModeChecks(t *testing.T) {
	if !ModeNormal.IsNormal() {
		t.Error("ModeNormal.IsNormal() should be true")
	}
	if !ModeInsert.IsInsert() {
		t.Error("ModeInsert.IsInsert() should be true")
	}
	if !ModeCommand.IsCommand() {
		t.Error("ModeCommand.IsCommand() should be true")
	}
	if ModeNormal.IsInsert() {
		t.Error("ModeNormal.IsInsert() should be false")
	}
	if !ModeVisual.IsVisual() {
		t.Error("ModeVisual.IsVisual() should be true")
	}
	if ModeNormal.IsVisual() {
		t.Error("ModeNormal.IsVisual() should be false")
	}
}
