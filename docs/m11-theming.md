# M11: Theming

## Goal

Built-in theme engine with shipped themes. Selectable via config or `:theme` command.

## Tasks

### Theme Type (`internal/tui/theme/theme.go`)

- [x] `Theme` struct: name + color palette
- [x] Color palette groups:
  - `Chrome`: borders, pane backgrounds, status bar, mode indicator
  - `Syntax`: SQL keywords, strings, numbers, comments, operators
  - `Data`: NULL, boolean, number, string, date cell colors
  - `UI`: cursor, selection, focused border, unfocused border, error, warning, success
- [x] Colors as lipgloss hex strings
- [x] `ComputeStyles(colors)` — derive all lipgloss styles from theme colors

### Built-in Themes (`internal/tui/theme/builtin.go`)

- [x] **Default Dark** — dark background, muted blues and greens
- [x] **Default Light** — light background, darker tones
- [x] **Solarized Dark** — classic solarized
- [x] **Solarized Light** — classic solarized light
- [x] **Nord** — Nord color palette
- [x] **Dracula** — Dracula color palette

### Theme Loader (`internal/tui/theme/loader.go`)

- [x] Load custom themes from `~/.config/db/themes/<name>.yaml`
- [x] YAML format mapping color groups to hex values
- [x] Merge with defaults: custom theme only needs to override specific colors
- [x] Validation: reject invalid hex colors, warn on missing groups

### Integration

- [x] `:theme <name>` command to switch at runtime
- [x] `:theme` with no args lists available themes
- [x] `theme` key in config file sets default
- [x] `--theme` CLI flag overrides config
- [x] Theme change applies immediately to all components

### Theme Config Format

```yaml
name: my-custom
colors:
  chrome:
    border: "#585858"
    border_focused: "#61afef"
    statusbar_bg: "#282c34"
    statusbar_fg: "#abb2bf"
    mode_normal: "#98c379"
    mode_insert: "#61afef"
    mode_command: "#e5c07b"
  syntax:
    keyword: "#c678dd"
    string: "#98c379"
    number: "#d19a66"
    comment: "#5c6370"
  data:
    null: "#5c6370"
    boolean: "#56b6c2"
  ui:
    cursor: "#528bff"
    selection: "#3e4451"
    error: "#e06c75"
    success: "#98c379"
```

## Tests

- [x] Unit: theme loading from YAML
- [x] Unit: merge with defaults
- [x] Unit: validation rejects bad colors
- [x] Unit: all built-in themes parse correctly

## Acceptance Criteria

- Default dark theme applied on startup
- `:theme nord` switches to Nord immediately
- Custom themes loadable from config directory
- All UI elements respect theme colors
- Config file `theme: solarized-dark` persists choice

## Dependencies

- M2 (TUI shell — lipgloss styles used everywhere)
- Mostly independent — can be worked on once basic TUI renders
