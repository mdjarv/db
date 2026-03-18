package cmd

import (
	"errors"
	"fmt"
	"strings"
)

// Exit codes for structured CLI error reporting.
const (
	ExitGeneral    = 1
	ExitConnection = 2
	ExitAuth       = 3
	ExitQuery      = 4
	ExitIO         = 5
)

// CLIError wraps an error with a category-specific exit code.
type CLIError struct {
	Code    int
	Message string
	Err     error
}

func (e *CLIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *CLIError) Unwrap() error { return e.Err }

// ExitCode returns the exit code for an error. Returns ExitGeneral for
// non-CLIError types, 0 for nil.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var cliErr *CLIError
	if errors.As(err, &cliErr) {
		return cliErr.Code
	}
	return ExitGeneral
}

// wrapConnection wraps an error as a connection error.
func wrapConnection(msg string, err error) error {
	return &CLIError{Code: ExitConnection, Message: msg, Err: err}
}

// wrapAuth wraps an error as an authentication error.
func wrapAuth(msg string, err error) error {
	return &CLIError{Code: ExitAuth, Message: msg, Err: err}
}

// wrapQuery wraps an error as a query execution error.
func wrapQuery(msg string, err error) error {
	return &CLIError{Code: ExitQuery, Message: msg, Err: err}
}

// wrapIO wraps an error as an I/O error.
func wrapIO(msg string, err error) error {
	return &CLIError{Code: ExitIO, Message: msg, Err: err}
}

// classifyConnError inspects an error and wraps it with the appropriate
// exit code (connection vs auth).
func classifyConnError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	lower := strings.ToLower(msg)
	if strings.Contains(lower, "password") ||
		strings.Contains(lower, "authentication") ||
		strings.Contains(lower, "auth") {
		return wrapAuth("authentication failed", err)
	}
	return wrapConnection("connection failed", err)
}
