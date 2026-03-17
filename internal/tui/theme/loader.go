package theme

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/mdjarv/db/internal/config"
)

var hexColorRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// LoadFile loads a custom theme from a YAML file and merges with defaults.
func LoadFile(path string) (*Theme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read theme: %w", err)
	}
	return Parse(data)
}

// Parse parses theme YAML bytes, merges with defaults, validates, and builds.
func Parse(data []byte) (*Theme, error) {
	base := DefaultDark()
	if err := yaml.Unmarshal(data, base); err != nil {
		return nil, fmt.Errorf("parse theme: %w", err)
	}
	if err := Validate(base); err != nil {
		return nil, err
	}
	return base.Build(), nil
}

// LoadCustom loads a custom theme by name from the themes directory.
func LoadCustom(name string) (*Theme, error) {
	dir := config.ThemesDir()
	path := filepath.Join(dir, name+".yaml")
	return LoadFile(path)
}

// Available returns all theme names (built-in + custom from disk).
func Available() []string {
	names := Names()
	dir := config.ThemesDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return names
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if !strings.HasSuffix(n, ".yaml") {
			continue
		}
		n = strings.TrimSuffix(n, ".yaml")
		names = append(names, n)
	}
	return names
}

// Resolve returns a theme by name: checks built-ins first, then custom files.
func Resolve(name string) (*Theme, error) {
	if t := Get(name); t != nil {
		return t, nil
	}
	return LoadCustom(name)
}

// Validate checks that all color values in a theme are valid.
func Validate(t *Theme) error {
	var errs []string
	check := func(field, val string) {
		if val == "" {
			return
		}
		if isValidColor(val) {
			return
		}
		errs = append(errs, fmt.Sprintf("%s: invalid color %q", field, val))
	}

	c := t.Colors
	check("chrome.border", c.Chrome.Border)
	check("chrome.border_focused", c.Chrome.BorderFocused)
	check("chrome.border_visual", c.Chrome.BorderVisual)
	check("chrome.statusbar_bg", c.Chrome.StatusBarBG)
	check("chrome.statusbar_fg", c.Chrome.StatusBarFG)
	check("chrome.mode_normal", c.Chrome.ModeNormalBG)
	check("chrome.mode_normal_fg", c.Chrome.ModeNormalFG)
	check("chrome.mode_insert", c.Chrome.ModeInsertBG)
	check("chrome.mode_insert_fg", c.Chrome.ModeInsertFG)
	check("chrome.mode_command", c.Chrome.ModeCommandBG)
	check("chrome.mode_command_fg", c.Chrome.ModeCommandFG)
	check("chrome.connected", c.Chrome.ConnectedFG)
	check("chrome.disconnected", c.Chrome.DisconnectedFG)
	check("chrome.tx", c.Chrome.TxFG)
	check("chrome.prompt", c.Chrome.PromptFG)

	check("syntax.keyword", c.Syntax.Keyword)
	check("syntax.string", c.Syntax.String)
	check("syntax.number", c.Syntax.Number)
	check("syntax.comment", c.Syntax.Comment)
	check("syntax.type", c.Syntax.Type)
	check("syntax.function", c.Syntax.Function)
	check("syntax.operator", c.Syntax.Operator)

	check("data.null", c.Data.Null)
	check("data.boolean", c.Data.Boolean)
	check("data.bool_true", c.Data.BoolTrue)
	check("data.bool_false", c.Data.BoolFalse)
	check("data.number", c.Data.Number)
	check("data.string", c.Data.String)
	check("data.date", c.Data.Date)
	check("data.uuid", c.Data.UUID)

	check("ui.cursor", c.UI.Cursor)
	check("ui.cursor_fg", c.UI.CursorFG)
	check("ui.cursor_row", c.UI.CursorRow)
	check("ui.selection", c.UI.Selection)
	check("ui.selection_fg", c.UI.SelectionFG)
	check("ui.col_selection", c.UI.ColSelection)
	check("ui.col_selection_fg", c.UI.ColSelectionFG)
	check("ui.dim", c.UI.Dim)
	check("ui.header", c.UI.Header)
	check("ui.separator", c.UI.Separator)
	check("ui.error", c.UI.Error)
	check("ui.warning", c.UI.Warning)
	check("ui.success", c.UI.Success)
	check("ui.insert_cursor", c.UI.InsertCursor)
	check("ui.editor_select", c.UI.EditorSelect)
	check("ui.editor_select_fg", c.UI.EditorSelectFG)
	check("ui.gutter", c.UI.Gutter)
	check("ui.modified", c.UI.Modified)
	check("ui.modified_fg", c.UI.ModifiedFG)
	check("ui.deleted", c.UI.Deleted)

	if len(errs) > 0 {
		return fmt.Errorf("theme validation: %s", strings.Join(errs, "; "))
	}
	return nil
}

// isValidColor checks if a string is a valid ANSI color number or hex color.
func isValidColor(s string) bool {
	if strings.HasPrefix(s, "#") {
		return hexColorRe.MatchString(s)
	}
	// ANSI color number (0-255)
	if len(s) > 3 || len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	n := 0
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	return n >= 0 && n <= 255
}
