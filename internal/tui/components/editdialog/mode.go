package editdialog

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mdjarv/db/internal/tui/theme"
)

type inputMode interface {
	Update(msg tea.KeyMsg) tea.Cmd
	View(contentW int, focused bool, t theme.Styles) string
	Value() string
	Hint() string
	SubmitsOnEnter() bool
}

type charFilter func(r rune) bool

var intFilter charFilter = func(r rune) bool {
	return (r >= '0' && r <= '9') || r == '-'
}

var floatFilter charFilter = func(r rune) bool {
	return (r >= '0' && r <= '9') || r == '-' || r == '.' || r == 'e' || r == 'E'
}

func resolveMode(opts OpenOpts) inputMode {
	// array check first — enum arrays should use array mode with enum element editors
	if strings.HasSuffix(opts.TypeName, "[]") {
		return newArrayMode(opts.TypeName, opts.Value, opts.EnumValues)
	}

	if opts.EnumValues != nil {
		return newEnumMode(opts.EnumValues, opts.Value)
	}
	if opts.CompositeFields != nil {
		return newCompositeMode(opts.CompositeFields, opts.Value)
	}

	switch opts.TypeName {
	case "bool":
		return newEnumMode([]string{"true", "false"}, opts.Value)
	case "int2", "int4", "int8":
		return newTextMode(opts.Value, intFilter, "", false)
	case "float4", "float8", "numeric", "money":
		return newTextMode(opts.Value, floatFilter, "", false)
	case "date":
		return newTextMode(opts.Value, nil, "YYYY-MM-DD", false)
	case "time", "timetz":
		return newTextMode(opts.Value, nil, "HH:MM:SS[.fff][+/-HH]", false)
	case "timestamp", "timestamptz":
		return newTextMode(opts.Value, nil, "YYYY-MM-DD HH:MM:SS[.fff][+/-HH]", false)
	case "uuid":
		return newTextMode(opts.Value, nil, "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", false)
	case "json", "jsonb":
		return newTextMode(opts.Value, nil, "JSON", true)
	}

	return newTextMode(opts.Value, nil, "Ctrl+J newline", true)
}
