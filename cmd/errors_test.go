package cmd

import (
	"errors"
	"testing"
)

func TestExitCode_Nil(t *testing.T) {
	if code := ExitCode(nil); code != 0 {
		t.Errorf("ExitCode(nil) = %d, want 0", code)
	}
}

func TestExitCode_CLIError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"connection", wrapConnection("conn", errors.New("refused")), ExitConnection},
		{"auth", wrapAuth("auth", errors.New("bad pw")), ExitAuth},
		{"query", wrapQuery("exec", errors.New("syntax")), ExitQuery},
		{"io", wrapIO("read", errors.New("not found")), ExitIO},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if code := ExitCode(tt.err); code != tt.want {
				t.Errorf("ExitCode = %d, want %d", code, tt.want)
			}
		})
	}
}

func TestExitCode_PlainError(t *testing.T) {
	if code := ExitCode(errors.New("generic")); code != ExitGeneral {
		t.Errorf("ExitCode(plain) = %d, want %d", code, ExitGeneral)
	}
}

func TestCLIError_Unwrap(t *testing.T) {
	inner := errors.New("inner")
	err := wrapConnection("outer", inner)
	if !errors.Is(err, inner) {
		t.Error("Unwrap should expose inner error")
	}
}

func TestClassifyConnError_Auth(t *testing.T) {
	err := classifyConnError(errors.New("password authentication failed"))
	if ExitCode(err) != ExitAuth {
		t.Errorf("auth error should have ExitAuth, got %d", ExitCode(err))
	}
}

func TestClassifyConnError_Connection(t *testing.T) {
	err := classifyConnError(errors.New("connection refused"))
	if ExitCode(err) != ExitConnection {
		t.Errorf("conn error should have ExitConnection, got %d", ExitCode(err))
	}
}
