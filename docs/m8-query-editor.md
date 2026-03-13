# M8: Query Editor (Right-Top Pane)

## Goal

Multi-line SQL editor with syntax highlighting and autocomplete. Vim-aware: only editable in INSERT mode.

## Tasks

### Editor Component (`internal/tui/components/queryeditor/`)

- [x] Multi-line text area with dynamic height (1 line up to 1/4 screen)
- [x] SQL syntax highlighting via chroma
- [x] Line numbers
- [x] Cursor positioning and movement (block cursor normal mode, underline cursor insert mode)
- [x] Text selection (visual mode — v char, V line)

### Vim Integration

- [x] NORMAL mode: hjkl cursor movement, w/b word jump, 0/$ line start/end
- [x] INSERT mode: text input, backspace, enter (newline), arrow keys, space
- [x] `i` enter insert at cursor, `a` after cursor, `I` start of line, `A` end of line
- [x] `o` new line below, `O` new line above
- [x] `dd` delete line, `D` delete to end of line
- [x] `u` undo, `Ctrl-r` redo (undo ring)
- [x] `Esc` return to normal mode

### Autocomplete

- [ ] Trigger: `Ctrl-Space` or automatic after `.` or space after keyword
- [ ] Popup menu with completion candidates
- [ ] `Tab` / `Shift-Tab` to navigate candidates
- [ ] `Enter` to accept, `Esc` to dismiss
- [ ] Sources: SQL keywords, table names, column names (from `query.Completer`)
- [ ] Context-aware: after FROM show tables, after SELECT show columns of contextual table

### Query Execution

- [ ] `:w` or `Ctrl-Enter` to execute query
- [ ] Send `QuerySubmitMsg` with SQL text to app model
- [ ] Show execution indicator (spinner) while running
- [ ] `Ctrl-C` to cancel running query
- [ ] Multi-statement: execute first statement only, or allow `;`-separated execution (configurable)

### Query Buffer

- [x] Current query text preserved across pane switches
- [x] Pre-populated when table browser sends a query
- [x] Clear buffer: `:clear` command

## Keybindings

| Key | Mode | Action |
|---|---|---|
| `i/a/I/A/o/O` | Normal | Enter insert mode |
| `Esc` | Insert | Return to normal mode |
| `hjkl` | Normal | Cursor movement |
| `w/b` | Normal | Word forward/backward |
| `0/$` | Normal | Start/end of line |
| `dd` | Normal | Delete line |
| `D` | Normal | Delete to end of line |
| `x` | Normal | Delete char |
| `u` | Normal | Undo |
| `Ctrl-r` | Normal | Redo |
| `v` | Normal | Visual char mode |
| `V` | Normal | Visual line mode |
| `yy`/`Y` | Normal | Yank current line |
| `p`/`P` | Normal | Paste after/before |
| `y` | Visual | Yank selection + exit visual |
| `Esc` | Visual | Exit visual mode |
| `arrow keys` | Insert | Move cursor |
| `Home`/`End` | Insert | Line start/end |
| `Ctrl-Space` | Insert | Trigger autocomplete |
| `Enter` | Normal | Execute query |
| `:w` | Command | Execute query |

## Tests

- [ ] Unit: cursor movement (hjkl, w/b, 0/$)
- [ ] Unit: text insertion and deletion
- [ ] Unit: undo/redo ring
- [ ] Unit: autocomplete candidate matching
- [ ] Unit: autocomplete context detection (after FROM vs SELECT)
- [ ] teatest: render with sample SQL, verify highlighting

## Acceptance Criteria

- Multi-line SQL editing with vim motions
- Syntax highlighting for SQL keywords, strings, numbers
- Autocomplete shows tables after FROM, columns after SELECT
- `:w` executes query and results appear in result viewer
- Undo/redo works across edit operations
- Line numbers visible

## Dependencies

- M2 (TUI shell — pane system, vim mode)
- M4 (query engine — Executor for running queries, Completer for autocomplete)
