# M10: Data Editing

## Goal

Edit table data inline: INSERT, UPDATE, DELETE rows via TUI with buffered changes and configurable transaction control.

## Tasks

### Change Buffer (`internal/editor/buffer.go`)

- [ ] `ChangeBuffer`: ordered list of pending changes
- [ ] `Change` types: `Insert`, `Update`, `Delete`
- [ ] `Update` tracks: table, PK values, column -> old/new value
- [ ] `Delete` tracks: table, PK values
- [ ] `Insert` tracks: table, column -> value
- [ ] Add, remove, list pending changes
- [ ] Conflict detection: multiple edits to same cell collapse
- [ ] Unit tests for all change operations

### Change Applier (`internal/editor/applier.go`)

- [ ] `Apply(ctx, conn, changes) error` — execute changes against DB
- [ ] Generate parameterized SQL (prevent injection)
- [ ] In explicit-commit mode: wrap in transaction, return Tx for user commit/rollback
- [ ] In auto-commit mode: execute each change immediately
- [ ] Report per-change success/failure
- [ ] Rollback on first error (explicit mode)
- [ ] Unit tests with mock Conn/Tx

### DML Generator (`internal/editor/dml.go`)

- [ ] `GenerateUpdate(table, pk, changes) (sql, args)`
- [ ] `GenerateInsert(table, values) (sql, args)`
- [ ] `GenerateDelete(table, pk) (sql, args)`
- [ ] Always use parameterized queries (`$1, $2, ...`)
- [ ] Handle NULL values correctly
- [ ] Unit tests for SQL generation

### TUI Integration

- [ ] Edit mode in result viewer: `e` on a cell enters edit mode
- [ ] Cell editing: type new value, `Enter` to confirm, `Esc` to cancel
- [ ] Visual indicator: modified cells highlighted (e.g., yellow background)
- [ ] Pending changes summary in status bar: "3 changes pending"
- [ ] `:commit` — apply and commit pending changes
- [ ] `:rollback` — discard pending changes
- [ ] `:changes` — show list of pending changes
- [ ] `Ctrl-z` — undo last change (remove from buffer)
- [ ] Delete row: `dR` in normal mode on a row (with confirmation)
- [ ] Insert row: `oR` opens blank row for editing

### Transaction Mode Toggle

- [ ] `:set autocommit` / `:set noautocommit`
- [ ] Status bar indicator: `[AUTO]` or `[TXN]`
- [ ] Warning when switching modes with pending changes
- [ ] Default: explicit commit (noautocommit)

## Keybindings (Result Viewer, Normal Mode)

| Key | Action |
|---|---|
| `e` | Edit cell under cursor |
| `dR` | Delete row (with confirmation dialog) |
| `oR` | Insert new row |
| `:commit` | Apply and commit changes |
| `:rollback` | Discard pending changes |
| `:changes` | List pending changes |
| `Ctrl-z` | Undo last change |

## Safety

- PK is required for UPDATE/DELETE. If table has no PK, editing is disabled with a message.
- Confirmation dialog before DELETE.
- Confirmation before `:commit` showing change summary.
- Auto-commit mode shows warning on enable: "Changes will be applied immediately."

## Tests

- [ ] Unit: ChangeBuffer add/remove/collapse
- [ ] Unit: DML generation for INSERT/UPDATE/DELETE
- [ ] Unit: NULL handling in DML
- [ ] Unit: parameterized query correctness
- [ ] Integration: apply changes to real PostgreSQL (testcontainers)
- [ ] Integration: transaction commit/rollback

## Acceptance Criteria

- Can edit a cell value and see it highlighted as pending
- `:commit` applies all changes in a transaction
- `:rollback` discards changes and restores original values
- Auto-commit mode executes each edit immediately
- Tables without PK show "editing disabled" message
- DELETE requires confirmation
- Pending change count visible in status bar

## Dependencies

- M1 (database layer — Conn, Tx)
- M2 (TUI shell)
- M9 (result viewer — editing happens within result view)
- M5 (schema — need PK detection to generate UPDATE/DELETE)
