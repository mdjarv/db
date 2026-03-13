# M2: TUI Shell

## Goal

Build the TUI application shell: pane layout, vim mode system, focus management, status bar. All using mock/stub data so this can be developed in parallel with the database layer.

## Tasks

### Vim Mode System (`internal/tui/app/mode.go`)

- [ ] Define `Mode` type: `Normal`, `Insert`, `Command`
- [ ] Mode state machine with transitions (see architecture.md)
- [ ] Mode-aware key dispatch: keys route differently per mode
- [ ] Unit tests for mode transitions

### Keybinding System (`internal/tui/app/keys.go`)

- [ ] Define keybinding registry: action name -> key sequence
- [ ] Global bindings (quit, pane switch, mode switch)
- [ ] Per-pane bindings (scroll, select, edit)
- [ ] `?` to show keybinding help overlay
- [ ] Unit tests for key matching

### Pane System (`internal/tui/pane/`)

- [ ] `Pane` interface: `Update`, `View`, `Focused`, `SetSize`
- [ ] `Manager`: tracks panes, active pane, handles focus switching
- [ ] Three-pane layout: left (table list), right-top (query), right-bottom (results)
- [ ] Pane resizing: `Ctrl-w +/-` to grow/shrink active pane
- [ ] Focus indicators: highlighted border on active pane
- [ ] Unit tests for focus cycling and layout calculation

### Status Bar (`internal/tui/components/statusbar/`)

- [ ] Current mode indicator (NORMAL/INSERT/COMMAND)
- [ ] Connection info display (placeholder for now)
- [ ] Message area for feedback (query time, row count, errors)
- [ ] Transaction mode indicator (auto-commit vs explicit)

### Command Bar (`internal/tui/components/commandbar/`)

- [ ] `:` activates command mode
- [ ] Text input with command parsing
- [ ] Command registry: `:q` quit, `:w` run query, `:set` settings
- [ ] Command history (up/down arrows)
- [ ] Esc or Enter exits command mode
- [ ] Tab completion for command names

### App Model (`internal/tui/app/model.go`)

- [ ] Main bubbletea `Model` composing all panes
- [ ] Message routing: global keys handled by app, rest delegated to active pane
- [ ] Window resize handling: recalculate pane sizes
- [ ] Startup: show connection dialog or connect from flags
- [ ] Graceful shutdown: close DB connection, save state

### Stub/Mock Panes

- [ ] Stub table list pane showing hardcoded table names
- [ ] Stub query editor with basic text input
- [ ] Stub result view showing hardcoded data rows
- [ ] These stubs will be replaced by real components in Phase 3

## Design Notes

- The app model is the **only** component that knows about all panes. Panes don't reference each other.
- Pane communication happens via bubbletea messages routed through the app model. Example: query editor sends `QuerySubmittedMsg`, app model forwards to query engine, result comes back as `QueryResultMsg` routed to result viewer.
- The vim mode is global state owned by the app model. Panes receive the current mode and behave accordingly (e.g., query editor only accepts text input in INSERT mode).

## Acceptance Criteria

- TUI launches with three-pane layout
- Can switch focus between panes with Ctrl-w + hjkl and Tab
- Mode indicator shows NORMAL/INSERT/COMMAND
- `:q` quits the application
- `?` shows keybinding overlay
- Pane borders highlight on focus
- Window resize recalculates layout correctly
- All mode transitions work: Esc -> Normal, i -> Insert, : -> Command

## Dependencies

- M0 (project skeleton)
- Does NOT need M1 — uses stub data

## Can Be Parallelized With

- M1 (Database Layer) — completely independent work
