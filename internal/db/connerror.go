package db

import (
	"errors"
	"io"
	"net"
	"strings"
	"syscall"
)

// IsConnectionError returns true if the error indicates a lost or broken
// database connection rather than a query-level error.
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}

	// io.EOF / unexpected EOF — server closed connection
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}

	// POSIX socket errors
	if errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.EPIPE) ||
		errors.Is(err, syscall.ECONNABORTED) {
		return true
	}

	// net.OpError (covers dial, read, write failures)
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		return true
	}

	// string-based fallback for wrapped errors
	msg := err.Error()
	connPatterns := []string{
		"conn closed",
		"conn busy",
		"closed pool",
		"broken pipe",
		"connection refused",
		"connection reset",
		"connection timed out",
		"no such host",
		"server closed the connection unexpectedly",
	}
	for _, p := range connPatterns {
		if strings.Contains(msg, p) {
			return true
		}
	}

	return false
}
