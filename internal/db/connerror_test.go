package db

import (
	"errors"
	"fmt"
	"io"
	"net"
	"syscall"
	"testing"
)

func TestIsConnectionError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"generic error", errors.New("syntax error"), false},
		{"EOF", io.EOF, true},
		{"unexpected EOF", io.ErrUnexpectedEOF, true},
		{"ECONNRESET", syscall.ECONNRESET, true},
		{"ECONNREFUSED", syscall.ECONNREFUSED, true},
		{"EPIPE", syscall.EPIPE, true},
		{"wrapped conn closed", fmt.Errorf("postgres: query: %w", errors.New("conn closed")), true},
		{"wrapped broken pipe", fmt.Errorf("postgres: query: broken pipe"), true},
		{"wrapped connection refused", fmt.Errorf("postgres: connect: connection refused"), true},
		{"net OpError", &net.OpError{Op: "read", Err: errors.New("reset")}, true},
		{"closed pool", fmt.Errorf("closed pool"), true},
		{"pg syntax error", errors.New("ERROR: syntax error at or near \"foo\" (SQLSTATE 42601)"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsConnectionError(tt.err)
			if got != tt.want {
				t.Errorf("IsConnectionError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
