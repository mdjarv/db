# M10: Data Editing

## Goal

Edit table data via popup dialog: INSERT, UPDATE, DELETE rows with buffered changes and configurable transaction control.

## Tasks

### Change Buffer (`internal/editor/buffer.go`)

- [x] `ChangeBuffer`: ordered list of pending changes
- [x] `Change` types: `Insert`, `Update`, `Delete`
- [x] `Update` tracks: table, PK values, column -> old/new value
- [x] `Delete` tracks: table, PK values
- [x] `Insert` tracks: table, column -> value
- [x] Add, remove, list pending changes
- [x] Conflict detection: multiple edits to same cell collapse
- [x] Unit tests for all change operations

### Change Applier (`internal/editor/applier.go`)

- [x] `Apply(ctx, conn, changes, autocommit) ApplyResult` — execute changes against DB
- [x] Generate parameterized SQL (prevent injection)
- [x] In explicit-commit mode: wrap in transaction, return Tx for user commit/rollback
- [x] In auto-commit mode: execute each change immediately
- [x] Report per-change success/failure
- [x] Rollback on first error (explicit mode)
- [x] Unit tests with mock Conn/Tx

### DML Generator (`internal/editor/dml.go`)

- [x] `GenerateUpdate(table, pk, changes) (sql, args)`
- [x] `GenerateInsert(table, values) (sql, args)`
- [x] `GenerateDelete(table, pk) (sql, args)`
- [x] Always use parameterized queries (`$1, $2, ...`)
- [x] Handle NULL values correctly
- [x] Unit tests for SQL generation

### TUI Integration

- [x] Edit dialog: `Enter` or `e` on a cell opens popup edit dialog
- [x] Edit dialog: text input with cursor, space, multiline (Ctrl+J)
- [x] Edit dialog: OK / NULL / Cancel buttons, Tab cycles focus
- [x] Edit dialog: NULL button greyed out for NOT NULL columns
- [x] Visual indicator: modified cells highlighted (yellow background)
- [x] Visual indicator: deleted rows shown with strikethrough
- [x] Pending changes summary in status bar: "3 changes pending"
- [x] `:commit` — apply and commit pending changes (with confirmation)
- [x] `:rollback` — discard pending changes
- [x] `:changes` — show list of pending changes
- [x] `Ctrl-z` — undo last change (remove from buffer)
- [x] Delete row: `dR` in normal mode on a row (with confirmation)
- [x] Insert row: `oR` opens blank row for editing

### Transaction Mode Toggle

- [x] `:set autocommit` / `:set noautocommit`
- [x] Status bar indicator: `tx:txn` (default) or `tx:auto`
- [x] Warning when switching modes with pending changes
- [x] Default: explicit commit (noautocommit)

## Keybindings (Result Viewer, Normal Mode)

| Key | Action |
|---|---|
| `Enter` / `e` | Open edit dialog for cell under cursor |
| `dR` | Delete row (with confirmation dialog) |
| `oR` | Insert new row |
| `:commit` | Apply and commit changes |
| `:rollback` | Discard pending changes |
| `:changes` | List pending changes |
| `Ctrl-z` | Undo last change |

## Edit Dialog

Popup dialog with:
- Column name and type in title
- Text input area with cursor navigation
- Ctrl+J inserts newline (multiline support)
- Tab cycles focus: input -> OK -> NULL -> Cancel
- NULL button disabled (greyed out) for NOT NULL columns
- Enter on input or OK button submits value
- Esc or Cancel button cancels edit

## Safety

- PK is required for UPDATE/DELETE. If table has no PK, editing is disabled with a message.
- Confirmation dialog before DELETE.
- Confirmation before `:commit` showing change summary.
- Auto-commit mode shows warning on enable: "Changes will be applied immediately."

## Tests

- [x] Unit: ChangeBuffer add/remove/collapse
- [x] Unit: DML generation for INSERT/UPDATE/DELETE
- [x] Unit: NULL handling in DML
- [x] Unit: parameterized query correctness
- [x] Integration: apply changes to real PostgreSQL (testcontainers)
- [x] Integration: transaction commit/rollback

## Acceptance Criteria

- Can edit a cell value via popup dialog and see it highlighted as pending
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
