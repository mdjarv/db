# M8: Query Editor (Right-Top Pane)

## Goal

Multi-line SQL editor with syntax highlighting and autocomplete. Vim-aware: only editable in INSERT mode.

## Tasks

### Editor Component (`internal/tui/components/queryeditor/`)

- [ ] Multi-line text area (build on `bubbles/textarea` or custom)
- [ ] SQL syntax highlighting via chroma
- [ ] Line numbers
- [ ] Cursor positioning and movement
- [ ] Text selection (visual mode — stretch goal)

### Vim Integration

- [ ] NORMAL mode: hjkl cursor movement, w/b word jump, 0/$ line start/end
- [ ] INSERT mode: text input, backspace, enter (newline)
- [ ] `i` enter insert at cursor, `a` after cursor, `I` start of line, `A` end of line
- [ ] `o` new line below, `O` new line above
- [ ] `dd` delete line, `D` delete to end of line
- [ ] `u` undo, `Ctrl-r` redo (undo ring)
- [ ] `Esc` return to normal mode

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

- [ ] Current query text preserved across pane switches
- [ ] Pre-populated when table browser sends a query
- [ ] Clear buffer: `:clear` command

## Keybindings

| Key | Mode | Action |
|---|---|---|
| `i/a/I/A/o/O` | Normal | Enter insert mode |
| `Esc` | Insert | Return to normal mode |
| `hjkl` | Normal | Cursor movement |
| `w/b` | Normal | Word forward/backward |
| `0/$` | Normal | Start/end of line |
| `dd` | Normal | Delete line |
| `u` | Normal | Undo |
| `Ctrl-r` | Normal | Redo |
| `Ctrl-Space` | Insert | Trigger autocomplete |
| `Ctrl-Enter` | Any | Execute query |
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
