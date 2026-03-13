// Package core defines types shared across TUI packages.
package core

// Mode represents the vim mode state.
type Mode int

// Vim modes.
const (
	ModeNormal Mode = iota
	ModeInsert
	ModeCommand
	ModeVisualLine
	ModeVisualBlock
)

// String returns the display name of the mode.
func (m Mode) String() string {
	switch m {
	case ModeInsert:
		return "INSERT"
	case ModeCommand:
		return "COMMAND"
	case ModeVisualLine:
		return "V-LINE"
	case ModeVisualBlock:
		return "V-BLOCK"
	default:
		return "NORMAL"
	}
}

// IsInsert returns true if mode is Insert.
func (m Mode) IsInsert() bool { return m == ModeInsert }

// IsCommand returns true if mode is Command.
func (m Mode) IsCommand() bool { return m == ModeCommand }

// IsNormal returns true if mode is Normal.
func (m Mode) IsNormal() bool { return m == ModeNormal }

// IsVisual returns true if mode is any visual mode.
func (m Mode) IsVisual() bool { return m == ModeVisualLine || m == ModeVisualBlock }
