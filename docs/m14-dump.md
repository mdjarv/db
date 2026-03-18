# M14: Database Dump (pg_dump Wrapper)

## Goal

Wrap `pg_dump` to provide database dump with progress tracking in CLI and TUI. Includes a reusable context menu component for the table browser and a dump configuration form with sensible defaults.

## Design Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Dump engine | Wrap `pg_dump` | Battle-tested, handles all edge cases |
| Credential passing | `PGPASSWORD` env var | Simplest, widely supported, no temp files |
| Progress tracking | Parse `--verbose` stderr | Coarse (per-object) but reliable |
| pg_restore | Out of scope | Follow-up milestone, keep runner generic |
| Binary resolution | PATH + config override | `exec.LookPath` default, `pg_dump_path` in config |
| Context menu | Reusable component | Pays off for future pane-specific actions |

## UX Design

### Table Browser Panel Footer

The table browser border integrates action hints in the bottom border, adapting to context:

```
╭─ Tables ─────────────────╮
│ users              100   │
│ posts               42   │
│ comments            17   │
│ tags                  5  │
│                          │
╰─ Space:actions  /:filter ╯
```

When filtering is active, footer changes:
```
╰─ /use                    ╯
```

When in detail view:
```
╰─ Esc:back  q:back        ╯
```

The footer is rendered inside the bottom border using `lipgloss.PlaceHorizontal` or by injecting styled text into the border bottom. Follows the same pattern as the result viewer's `rows 1-100 of 2847` status in its bottom border.

### Context Menu (`Space` on table browser)

A popup menu anchored near the selected table, showing actions relevant to the current table:

```
╭─ Tables ──────────╮
│ users          100 │╭──────────────────────╮
│ posts           42 ││ Query (SELECT *)     │
│ comments        17 ││ Describe schema      │
│                    ││ Dump table...        │
│                    ││ Dump schema only...  │
│                    ││ Copy table name      │
╰────────────────────╯│ Refresh schema       │
                      ╰──────────────────────╯
```

- `j/k` navigate, `Enter` selects, `Esc` dismisses
- Actions ending in `...` open a follow-up dialog (dump form)
- Menu items map to existing keybindings (Enter, d, y, R) plus new dump actions
- The component is reusable -- other panes can define their own action lists

### Dump Configuration Form

Opened from context menu "Dump table..." / "Dump schema only...", or via `:dump`:

```
╭─ Dump: users ────────────────────╮
│                                  │
│ Format:       [custom       ▾]   │
│ Output:       mydb_20260318.dump │
│ Schema only:  [ ]                │
│ Tables:       users              │
│                                  │
│         [Start]    [Cancel]      │
╰──────────────────────────────────╯
```

**Defaults:**
- **Format**: `custom` (compressed, supports pg_restore, most useful)
- **Output**: `<dbname>_<YYYYMMDD>.<ext>` in configured dump dir or cwd
  - `.dump` for custom, `.sql` for plain, `.tar` for tar, directory name for directory
- **Schema only**: off
- **Tables**: blank = all tables; pre-filled when triggered from table context menu
- Tab cycles fields, Enter on Start begins dump

### Progress Modal

Replaces the form once dump starts:

```
╭─ Dumping: mydb ──────────────────╮
│                                  │
│  public.users                    │
│  [████████████░░░░░░░░░]  12/47  │
│                                  │
│  Elapsed: 4s                     │
│                                  │
│             [Cancel]             │
╰──────────────────────────────────╯
```

- Updates per-object as pg_dump emits `--verbose` lines
- Esc or Cancel sends context cancellation
- On completion: dismisses, shows success in status bar with path + size
- On error: dismisses, shows error in status bar with pg_dump stderr excerpt

### Full Flow

1. User presses `Space` on table browser → context menu opens
2. User selects "Dump table..." → dump form opens, table pre-filled
3. User adjusts format/output or accepts defaults → presses Enter on Start
4. Progress modal appears with live per-object updates
5. Dump completes → modal dismisses, status bar shows `dumped to mydb_20260318.dump (2.4 MB, 12s)`

Alternative entries:
- `:dump` → form opens for full database (tables field blank)
- `:dump users` → form opens with `users` pre-filled
- `D` on table browser → form opens with selected table pre-filled (shortcut, skips context menu)

## Tasks

### Context Menu Component (`internal/tui/components/contextmenu/`)

- [ ] `MenuItem` struct: label, action ID, has submenu indicator (`...`)
- [ ] `Model` struct: items, cursor, position (x/y anchor), visible
- [ ] `Open(items []MenuItem, anchorX, anchorY int)` -- show menu at position
- [ ] `Close()` -- dismiss
- [ ] j/k navigate, Enter selects, Esc dismisses
- [ ] `View(containerW, containerH int) string` -- render as overlay
- [ ] Styled with theme: border, cursor highlight, dimmed disabled items
- [ ] Unit tests: navigation, selection, dismiss

### Table Browser Footer (`internal/tui/components/tablelist/`)

- [ ] Render action hints in bottom border: `Space:actions  /:filter`
- [ ] Adapt footer to current state (filtering, detail view, list view)
- [ ] Use dim/muted style for hints, consistent with result viewer footer

### Table Browser Context Menu Integration

- [ ] `Space` opens context menu with table-specific actions
- [ ] Menu items: Query, Describe, Dump table..., Dump schema only..., Copy name, Refresh
- [ ] Selection dispatches existing messages (QueryRequestMsg, YankMsg, RefreshSchemaMsg) or new dump messages
- [ ] Close menu on action dispatch

### Dump Form (`internal/tui/components/dumpform/`)

- [ ] `Model` struct: format selector, output path input, schema-only toggle, tables input
- [ ] `Open(tableName string, dbName string)` -- pre-fill from context
- [ ] Default output path computed from dbname + date + format extension
- [ ] Format field updates output extension on change
- [ ] Tab cycles fields: format → output → schema-only → tables → Start → Cancel
- [ ] Enter on Start emits `DumpStartMsg`
- [ ] Esc or Cancel dismisses
- [ ] Styled consistently with connform

### Library Layer (`internal/dump/`)

#### Binary Detection (`pgdump.go`)

- [ ] `FindPgDump(configPath string) (string, error)` -- LookPath with config override
- [ ] `PgDumpVersion(binary string) (string, error)` -- parse `--version` output
- [ ] Version mismatch warning vs connected server (non-blocking)

#### Config Builder (`config.go`)

- [ ] `DumpConfig` struct: connection params, format, schema-only, table filter, output path
- [ ] `Format` type: `Plain`, `Custom`, `Directory`, `Tar`
- [ ] `BuildArgs(cfg DumpConfig) []string` -- construct pg_dump CLI args
- [ ] Always inject `--verbose` for progress parsing
- [ ] `--no-password` always set (credential via env var)

#### Progress Parser (`progress.go`)

- [ ] `ProgressEvent` struct: `Object string`, `Index int`, `Total int`, `Done bool`, `Err error`
- [ ] `ParseProgress(r io.Reader, total int) <-chan ProgressEvent`
- [ ] Regex: `^pg_dump: dumping contents of table "(.+)"$` and similar patterns
- [ ] Count parsed objects against pre-fetched total for percentage

#### Runner (`runner.go`)

- [ ] `Runner` struct: binary path, builds and executes `exec.Cmd`
- [ ] `Run(ctx context.Context, cfg DumpConfig) (<-chan ProgressEvent, error)`
- [ ] Context cancellation kills process
- [ ] `PGPASSWORD` injected via `cmd.Env`
- [ ] Stderr piped to progress parser
- [ ] Non-zero exit: capture stderr tail as error

### CLI (`cmd/dump.go`)

- [ ] `db dump` command with flags:
  - `--schema-only` / `-s`
  - `--table <name>` / `-t` (repeatable)
  - `--format <plain|custom|directory|tar>` / `-F` (default: `custom`)
  - `-o <path>` (default: `<dbname>_<timestamp>.<ext>`)
  - `--verbose` / `-v` -- pass through raw stderr
- [ ] Pre-flight: find pg_dump, version check, validate flags
- [ ] Progress bar on stderr: `[===>   ] 5/12 tables`
- [ ] On completion: print output path and file size

### TUI Messages (`internal/tui/core/msg.go`)

- [ ] `DumpStartMsg` -- carries `DumpConfig`
- [ ] `DumpProgressMsg` -- carries `ProgressEvent`
- [ ] `DumpCompleteMsg` -- output path, file size, duration, or error
- [ ] `ContextMenuActionMsg` -- carries action ID from menu selection

### Progress Modal (`internal/tui/components/dialog/progress.go`)

- [ ] `ProgressModel`: title, current object, count/total, elapsed time, cancel button
- [ ] Non-interactive -- Esc/Enter on Cancel to cancel
- [ ] Renders: title, current object, progress bar, elapsed time
- [ ] Styled with theme

### TUI App Integration (`internal/tui/app/`)

- [ ] Register `:dump` and `:dump <table>` in `commandRegistry`
- [ ] `D` keybinding on table browser -- open dump form for selected table
- [ ] `Space` keybinding on table browser -- open context menu
- [ ] Handle `ContextMenuActionMsg` -- route to appropriate handler
- [ ] Handle `DumpStartMsg` -- find pg_dump, validate, open progress modal, start runner
- [ ] Handle `DumpProgressMsg` -- update progress modal
- [ ] Handle `DumpCompleteMsg` -- dismiss modal, status bar message with path + size
- [ ] Context cancellation wired to Esc in progress modal

### Config (`internal/config/`)

- [ ] `PgDumpPath string` -- override pg_dump binary path
- [ ] `DumpDir string` -- default output directory (defaults to cwd)

## Keybindings

| Context | Key | Action |
|---|---|---|
| Table browser (normal) | `Space` | Open context menu for selected table |
| Table browser (normal) | `D` | Open dump form for selected table (shortcut) |
| Context menu | `j/k` | Navigate menu items |
| Context menu | `Enter` | Select action |
| Context menu | `Esc` | Dismiss menu |
| Dump form | `Tab` | Cycle fields |
| Dump form | `Enter` | Start dump (on Start button) |
| Dump form | `Esc` | Cancel |
| Progress modal | `Esc` | Cancel dump |
| Command mode | `:dump` | Open dump form (full database) |
| Command mode | `:dump <table>` | Open dump form (pre-filled table) |

## Tests

### Unit

- [ ] Context menu: navigation, selection, dismiss, render at anchor position
- [ ] Dump form: default values, format changes extension, pre-fill from table
- [ ] `BuildArgs`: correct flags for each config combination
- [ ] `ParseProgress`: parse verbose lines into events with correct counts
- [ ] `ParseProgress`: handle unexpected lines, EOF produces Done event
- [ ] `FindPgDump`: config path precedence, clear error when not found

### Integration

- [ ] Round-trip: dump test DB, verify output exists and is non-empty
- [ ] Schema-only: contains CREATE TABLE but no INSERT/COPY
- [ ] Single table: output contains only specified table
- [ ] Format variants: plain = SQL text, custom = binary
- [ ] Context cancellation: dump cancelled mid-stream
- [ ] Missing pg_dump: clear error message
- [ ] CLI `db dump` against testcontainers produces valid file

### TUI

- [ ] `Space` opens context menu with correct items
- [ ] Context menu action dispatches correct message
- [ ] `D` opens dump form with selected table pre-filled
- [ ] `:dump` opens dump form for full database
- [ ] Progress modal renders with live updates
- [ ] Esc cancels in-progress dump
- [ ] Completion dismisses modal, shows status

## Implementation Order

Parallelizable tracks:

```
Track 1 (library):  config.go → pgdump.go → progress.go → runner.go → cmd/dump.go
Track 2 (TUI):      contextmenu/ → tablelist footer → dumpform/ → progress modal → app wiring
```

Track 1 has no TUI dependency. Track 2 can start with context menu (useful standalone) while Track 1 builds the dump library. They converge at app wiring.

## Acceptance Criteria

- Table browser shows footer hints (`Space:actions  /:filter`)
- `Space` opens context menu with table-specific actions
- Context menu "Dump table..." opens dump form with table pre-filled
- Dump form shows sensible defaults (custom format, auto-generated output path)
- Starting dump shows progress modal with per-object updates
- `db dump` CLI produces valid pg_dump output that pg_restore can read
- `db dump --schema-only` contains only DDL
- Esc cancels in-progress dump (both CLI and TUI)
- Missing pg_dump: clear actionable error
- Unit tests pass without pg_dump installed (mock exec)

## Dependencies

- M3 (connections -- `ConnectionConfig` for credential access)
- M5 (schema -- `Inspector.Tables()` for pre-fetching table count)
- `pg_dump` runtime dependency (not build dependency)
