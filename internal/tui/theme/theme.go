// Package theme provides the theming system for the TUI.
package theme

import (
	"sync"

	"github.com/charmbracelet/lipgloss"
)

var (
	mu      sync.RWMutex
	current *Theme
)

func init() {
	current = DefaultDark()
}

// Current returns the active theme.
func Current() *Theme {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

// Set changes the active theme.
func Set(t *Theme) {
	mu.Lock()
	defer mu.Unlock()
	current = t
}

// Theme defines all colors and pre-built styles for the TUI.
type Theme struct {
	Name   string `yaml:"name"`
	Colors Colors `yaml:"colors"`
	Styles Styles `yaml:"-"`
}

// Build computes Styles from Colors. Must be called after setting Colors.
func (t *Theme) Build() *Theme {
	c := t.Colors
	t.Styles = Styles{
		// Borders
		BorderFocused:   lipgloss.Color(c.Chrome.BorderFocused),
		BorderUnfocused: lipgloss.Color(c.Chrome.Border),
		BorderVisual:    lipgloss.Color(c.Chrome.BorderVisual),

		// Table
		Header:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(c.UI.Header)),
		Separator:     lipgloss.NewStyle().Foreground(lipgloss.Color(c.UI.Separator)),
		Cursor:        lipgloss.NewStyle().Foreground(lipgloss.Color(c.UI.CursorFG)).Background(lipgloss.Color(c.UI.Cursor)),
		CursorRow:     lipgloss.NewStyle().Foreground(lipgloss.Color(c.UI.CursorRow)),
		Selection:     lipgloss.NewStyle().Foreground(lipgloss.Color(c.UI.SelectionFG)).Background(lipgloss.Color(c.UI.Selection)),
		ColSelection:  lipgloss.NewStyle().Foreground(lipgloss.Color(c.UI.ColSelectionFG)).Background(lipgloss.Color(c.UI.ColSelection)),
		Dim:           lipgloss.NewStyle().Foreground(lipgloss.Color(c.UI.Dim)),
		Null:          lipgloss.NewStyle().Foreground(lipgloss.Color(c.Data.Null)).Italic(true),
		DataBoolTrue:  lipgloss.NewStyle().Foreground(lipgloss.Color(c.Data.BoolTrue)),
		DataBoolFalse: lipgloss.NewStyle().Foreground(lipgloss.Color(c.Data.BoolFalse)),
		DataNumber:    lipgloss.NewStyle().Foreground(lipgloss.Color(c.Data.Number)),
		DataDate:      lipgloss.NewStyle().Foreground(lipgloss.Color(c.Data.Date)),
		DataUUID:      lipgloss.NewStyle().Foreground(lipgloss.Color(c.Data.UUID)),
		DataString:    lipgloss.NewStyle().Foreground(lipgloss.Color(c.Data.String)),

		// Status bar
		StatusBarBG: lipgloss.Color(c.Chrome.StatusBarBG),
		StatusBarFG: lipgloss.NewStyle().Foreground(lipgloss.Color(c.Chrome.StatusBarFG)),
		ModeNormal: lipgloss.NewStyle().Bold(true).Padding(0, 1).
			Background(lipgloss.Color(c.Chrome.ModeNormalBG)).
			Foreground(lipgloss.Color(c.Chrome.ModeNormalFG)),
		ModeInsert: lipgloss.NewStyle().Bold(true).Padding(0, 1).
			Background(lipgloss.Color(c.Chrome.ModeInsertBG)).
			Foreground(lipgloss.Color(c.Chrome.ModeInsertFG)),
		ModeCommand: lipgloss.NewStyle().Bold(true).Padding(0, 1).
			Background(lipgloss.Color(c.Chrome.ModeCommandBG)).
			Foreground(lipgloss.Color(c.Chrome.ModeCommandFG)),
		ConnectedFG:    lipgloss.NewStyle().Foreground(lipgloss.Color(c.Chrome.ConnectedFG)).Padding(0, 1),
		DisconnectedFG: lipgloss.NewStyle().Foreground(lipgloss.Color(c.Chrome.DisconnectedFG)).Padding(0, 1),
		TxFG:           lipgloss.NewStyle().Foreground(lipgloss.Color(c.Chrome.TxFG)).Padding(0, 1),
		CommandPrompt:  lipgloss.NewStyle().Foreground(lipgloss.Color(c.Chrome.PromptFG)).Bold(true),

		// Syntax
		Keyword:  lipgloss.NewStyle().Foreground(lipgloss.Color(c.Syntax.Keyword)).Bold(true),
		String:   lipgloss.NewStyle().Foreground(lipgloss.Color(c.Syntax.String)),
		Number:   lipgloss.NewStyle().Foreground(lipgloss.Color(c.Syntax.Number)),
		Comment:  lipgloss.NewStyle().Foreground(lipgloss.Color(c.Syntax.Comment)).Italic(true),
		Type:     lipgloss.NewStyle().Foreground(lipgloss.Color(c.Syntax.Type)),
		Function: lipgloss.NewStyle().Foreground(lipgloss.Color(c.Syntax.Function)),
		Operator: lipgloss.NewStyle().Foreground(lipgloss.Color(c.Syntax.Operator)),

		// Editor
		InsertCursor: lipgloss.NewStyle().Underline(true).Foreground(lipgloss.Color(c.UI.InsertCursor)),
		NormalCursor: lipgloss.NewStyle().Reverse(true),
		EditorSelect: lipgloss.NewStyle().Background(lipgloss.Color(c.UI.EditorSelect)).Foreground(lipgloss.Color(c.UI.EditorSelectFG)),
		Gutter:       lipgloss.NewStyle().Foreground(lipgloss.Color(c.UI.Gutter)),

		// Data editing
		Modified: lipgloss.NewStyle().Foreground(lipgloss.Color(c.UI.ModifiedFG)).Background(lipgloss.Color(c.UI.Modified)),
		Deleted:  lipgloss.NewStyle().Foreground(lipgloss.Color(c.UI.Deleted)).Strikethrough(true),

		// Misc
		Error:   lipgloss.NewStyle().Foreground(lipgloss.Color(c.UI.Error)),
		Success: lipgloss.NewStyle().Foreground(lipgloss.Color(c.UI.Success)),
		Warning: lipgloss.NewStyle().Foreground(lipgloss.Color(c.UI.Warning)),
	}
	return t
}

// Colors holds raw color values grouped by area.
type Colors struct {
	Chrome ChromeColors `yaml:"chrome"`
	Syntax SyntaxColors `yaml:"syntax"`
	Data   DataColors   `yaml:"data"`
	UI     UIColors     `yaml:"ui"`
}

// ChromeColors defines border, status bar, and mode indicator colors.
type ChromeColors struct {
	Border         string `yaml:"border"`
	BorderFocused  string `yaml:"border_focused"`
	BorderVisual   string `yaml:"border_visual"`
	StatusBarBG    string `yaml:"statusbar_bg"`
	StatusBarFG    string `yaml:"statusbar_fg"`
	ModeNormalBG   string `yaml:"mode_normal"`
	ModeNormalFG   string `yaml:"mode_normal_fg"`
	ModeInsertBG   string `yaml:"mode_insert"`
	ModeInsertFG   string `yaml:"mode_insert_fg"`
	ModeCommandBG  string `yaml:"mode_command"`
	ModeCommandFG  string `yaml:"mode_command_fg"`
	ConnectedFG    string `yaml:"connected"`
	DisconnectedFG string `yaml:"disconnected"`
	TxFG           string `yaml:"tx"`
	PromptFG       string `yaml:"prompt"`
}

// SyntaxColors defines SQL syntax highlighting colors.
type SyntaxColors struct {
	Keyword  string `yaml:"keyword"`
	String   string `yaml:"string"`
	Number   string `yaml:"number"`
	Comment  string `yaml:"comment"`
	Type     string `yaml:"type"`
	Function string `yaml:"function"`
	Operator string `yaml:"operator"`
}

// DataColors defines cell data type colors.
type DataColors struct {
	Null      string `yaml:"null"`
	Boolean   string `yaml:"boolean"`
	BoolTrue  string `yaml:"bool_true"`
	BoolFalse string `yaml:"bool_false"`
	Number    string `yaml:"number"`
	String    string `yaml:"string"`
	Date      string `yaml:"date"`
	UUID      string `yaml:"uuid"`
}

// UIColors defines cursor, selection, and other UI element colors.
type UIColors struct {
	Cursor         string `yaml:"cursor"`
	CursorFG       string `yaml:"cursor_fg"`
	CursorRow      string `yaml:"cursor_row"`
	Selection      string `yaml:"selection"`
	SelectionFG    string `yaml:"selection_fg"`
	ColSelection   string `yaml:"col_selection"`
	ColSelectionFG string `yaml:"col_selection_fg"`
	Dim            string `yaml:"dim"`
	Header         string `yaml:"header"`
	Separator      string `yaml:"separator"`
	Error          string `yaml:"error"`
	Warning        string `yaml:"warning"`
	Success        string `yaml:"success"`
	InsertCursor   string `yaml:"insert_cursor"`
	EditorSelect   string `yaml:"editor_select"`
	EditorSelectFG string `yaml:"editor_select_fg"`
	Gutter         string `yaml:"gutter"`
	Modified       string `yaml:"modified"`
	ModifiedFG     string `yaml:"modified_fg"`
	Deleted        string `yaml:"deleted"`
}

// Styles holds pre-computed lipgloss styles derived from Colors.
type Styles struct {
	// Borders
	BorderFocused   lipgloss.Color
	BorderUnfocused lipgloss.Color
	BorderVisual    lipgloss.Color

	// Table
	Header        lipgloss.Style
	Separator     lipgloss.Style
	Cursor        lipgloss.Style
	CursorRow     lipgloss.Style
	Selection     lipgloss.Style
	ColSelection  lipgloss.Style
	Dim           lipgloss.Style
	Null          lipgloss.Style
	DataBoolTrue  lipgloss.Style
	DataBoolFalse lipgloss.Style
	DataNumber    lipgloss.Style
	DataDate      lipgloss.Style
	DataUUID      lipgloss.Style
	DataString    lipgloss.Style

	// Status bar
	StatusBarBG    lipgloss.Color
	StatusBarFG    lipgloss.Style
	ModeNormal     lipgloss.Style
	ModeInsert     lipgloss.Style
	ModeCommand    lipgloss.Style
	ConnectedFG    lipgloss.Style
	DisconnectedFG lipgloss.Style
	TxFG           lipgloss.Style
	CommandPrompt  lipgloss.Style

	// Syntax
	Keyword  lipgloss.Style
	String   lipgloss.Style
	Number   lipgloss.Style
	Comment  lipgloss.Style
	Type     lipgloss.Style
	Function lipgloss.Style
	Operator lipgloss.Style

	// Editor
	InsertCursor lipgloss.Style
	NormalCursor lipgloss.Style
	EditorSelect lipgloss.Style
	Gutter       lipgloss.Style

	// Data editing
	Modified lipgloss.Style
	Deleted  lipgloss.Style

	// Misc
	Error   lipgloss.Style
	Success lipgloss.Style
	Warning lipgloss.Style
}

// Names returns the names of all built-in themes.
func Names() []string {
	return []string{
		"default-dark",
	}
}

// Get returns a built-in theme by name, or nil if not found.
func Get(name string) *Theme {
	switch name {
	case "default-dark":
		return DefaultDark()
	default:
		return nil
	}
}
