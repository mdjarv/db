package theme

import (
	"testing"
)

func TestAllBuiltinsParse(t *testing.T) {
	for _, name := range Names() {
		th := Get(name)
		if th == nil {
			t.Errorf("Get(%q) returned nil", name)
			continue
		}
		if th.Name != name {
			t.Errorf("Get(%q).Name = %q", name, th.Name)
		}
	}
}

func TestDefaultDarkCatppuccinMocha(t *testing.T) {
	th := DefaultDark()
	if th.Colors.Chrome.Border != "#585b70" {
		t.Errorf("border = %q, want #585b70 (surface2)", th.Colors.Chrome.Border)
	}
	if th.Colors.UI.Cursor != "#89b4fa" {
		t.Errorf("cursor = %q, want #89b4fa (blue)", th.Colors.UI.Cursor)
	}
}

func TestValidateRejectsInvalidHex(t *testing.T) {
	th := DefaultDark()
	th.Colors.Chrome.Border = "#ZZZZZZ"
	if err := Validate(th); err == nil {
		t.Error("expected validation error for invalid hex")
	}
}

func TestValidateAcceptsAnsiNumbers(t *testing.T) {
	th := DefaultDark()
	if err := Validate(th); err != nil {
		t.Errorf("default dark should validate: %v", err)
	}
}

func TestValidateAcceptsHexColors(t *testing.T) {
	th := SolarizedDark()
	if err := Validate(th); err != nil {
		t.Errorf("solarized dark should validate: %v", err)
	}
}

func TestValidateRejectsOutOfRangeAnsi(t *testing.T) {
	th := DefaultDark()
	th.Colors.Chrome.Border = "999"
	if err := Validate(th); err == nil {
		t.Error("expected validation error for ANSI > 255")
	}
}

func TestParseYAML(t *testing.T) {
	yaml := `
name: test-theme
colors:
  chrome:
    border: "#586e75"
    border_focused: "#268bd2"
`
	th, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if th.Name != "test-theme" {
		t.Errorf("name = %q, want test-theme", th.Name)
	}
	if th.Colors.Chrome.Border != "#586e75" {
		t.Errorf("border = %q, want #586e75", th.Colors.Chrome.Border)
	}
	if th.Colors.Chrome.BorderFocused != "#268bd2" {
		t.Errorf("border_focused = %q, want #268bd2", th.Colors.Chrome.BorderFocused)
	}
}

func TestParseMergesDefaults(t *testing.T) {
	yaml := `
name: partial
colors:
  chrome:
    border: "#111111"
`
	th, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	// overridden field
	if th.Colors.Chrome.Border != "#111111" {
		t.Errorf("border = %q, want #111111", th.Colors.Chrome.Border)
	}
	// default field preserved (catppuccin mocha values)
	if th.Colors.Chrome.BorderFocused != "#89b4fa" {
		t.Errorf("border_focused = %q, want #89b4fa (default)", th.Colors.Chrome.BorderFocused)
	}
	if th.Colors.UI.Cursor != "#89b4fa" {
		t.Errorf("cursor = %q, want #89b4fa (default)", th.Colors.UI.Cursor)
	}
}

func TestParseRejectsInvalidYAML(t *testing.T) {
	_, err := Parse([]byte(`{{{`))
	if err == nil {
		t.Error("expected parse error")
	}
}

func TestParseRejectsInvalidColors(t *testing.T) {
	yaml := `
name: bad
colors:
  chrome:
    border: "#GGGGGG"
`
	_, err := Parse([]byte(yaml))
	if err == nil {
		t.Error("expected validation error")
	}
}

func TestGetUnknownReturnsNil(t *testing.T) {
	if Get("nonexistent") != nil {
		t.Error("expected nil for unknown theme")
	}
}

func TestCurrentDefault(t *testing.T) {
	th := Current()
	if th == nil {
		t.Fatal("Current() should not be nil")
	}
	if th.Name != "default-dark" {
		t.Errorf("default theme = %q, want default-dark", th.Name)
	}
}

func TestSetAndCurrent(t *testing.T) {
	orig := Current()
	defer Set(orig)

	Set(Nord())
	if Current().Name != "nord" {
		t.Errorf("after Set(Nord), Current().Name = %q", Current().Name)
	}
}

func TestNamesMatchGet(t *testing.T) {
	for _, name := range Names() {
		if Get(name) == nil {
			t.Errorf("Names() includes %q but Get() returns nil", name)
		}
	}
}

func TestIsValidColor(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"0", true},
		{"255", true},
		{"256", false},
		{"15", true},
		{"#ff5555", true},
		{"#FF5555", true},
		{"#ggg", false},
		{"abc", false},
		{"", false},
		{"#12345", false},
		{"#1234567", false},
	}
	for _, tt := range tests {
		got := isValidColor(tt.input)
		if got != tt.valid {
			t.Errorf("isValidColor(%q) = %v, want %v", tt.input, got, tt.valid)
		}
	}
}

func TestBuildProducesStyles(t *testing.T) {
	for _, name := range Names() {
		th := Get(name)
		if th.Styles.Header.GetBold() != true {
			t.Errorf("%s: Header style should be bold", name)
		}
		if th.Styles.Comment.GetItalic() != true {
			t.Errorf("%s: Comment style should be italic", name)
		}
	}
}

func TestResolveBuiltin(t *testing.T) {
	th, err := Resolve("nord")
	if err != nil {
		t.Fatalf("resolve nord: %v", err)
	}
	if th.Name != "nord" {
		t.Errorf("name = %q, want nord", th.Name)
	}
}

func TestResolveUnknown(t *testing.T) {
	_, err := Resolve("nonexistent-theme-xyz")
	if err == nil {
		t.Error("expected error for nonexistent theme")
	}
}

func TestValidateEmptyColorsOK(t *testing.T) {
	th := DefaultDark()
	th.Colors.Data.String = ""
	th.Colors.Data.Date = ""
	if err := Validate(th); err != nil {
		t.Errorf("empty optional colors should pass: %v", err)
	}
}
