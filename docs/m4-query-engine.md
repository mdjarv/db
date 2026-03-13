# M4: Query Engine

## Goal

Query execution engine with transaction management. CLI `db query` command for one-shot queries.

## Tasks

### Executor (`internal/query/executor.go`)

- [ ] `Executor` struct: wraps `db.Conn`, manages transaction state
- [ ] `Execute(ctx, sql) -> (*db.Result, error)` — auto-detect SELECT vs DML
- [ ] `ExecDML(ctx, sql) -> (*db.ExecResult, error)` — explicit DML execution
- [ ] Transaction mode: auto-commit vs explicit
- [ ] `Begin()`, `Commit()`, `Rollback()` — delegate to `db.Tx`
- [ ] Query cancellation via context (Ctrl-C in TUI sends cancel)
- [ ] Query timing: measure and return execution duration
- [ ] Error classification: syntax error, connection lost, permission denied, timeout

### SQL Autocomplete (`internal/query/completer.go`)

- [ ] `Completer` struct: holds schema metadata for completion
- [ ] SQL keyword completion (SELECT, FROM, WHERE, JOIN, etc.)
- [ ] Table name completion (from cached schema)
- [ ] Column name completion (context-aware: after SELECT or WHERE, use current table)
- [ ] Schema refresh: reload table/column list on demand
- [ ] Completion ranking: recently used items first
- [ ] Unit tests with mock schema data

### CLI Command (`cmd/query.go`)

- [ ] `db query "SELECT * FROM users LIMIT 10"`
- [ ] `--format` flag: `table` (default), `csv`, `json`, `sql`
- [ ] `--no-header` flag: omit column headers
- [ ] Stdin support: `echo "SELECT 1" | db query`
- [ ] File support: `db query -f queries/report.sql`
- [ ] Connection flags inherited from root command
- [ ] Pretty table output by default (aligned columns, borders)
- [ ] Exit code: 0 success, 1 query error, 2 connection error

### Tests

- [ ] Integration: execute SELECT, INSERT, UPDATE, DELETE
- [ ] Integration: transaction begin/commit/rollback
- [ ] Integration: query cancellation
- [ ] Unit: autocomplete keyword matching
- [ ] Unit: autocomplete table/column matching
- [ ] Unit: SQL type detection (SELECT vs DML)
- [ ] Integration: CLI `db query` end-to-end with testcontainers

## Design Notes

- The executor does NOT parse SQL. It uses simple heuristics (first keyword) to distinguish SELECT from DML. PostgreSQL does the real parsing.
- Auto-complete metadata is loaded lazily on first completion request, then cached. Schema changes require explicit refresh (`:refresh` in TUI or re-query).
- The `--format` flag reuses exporters from M6. If M6 isn't done yet, start with table format only.

## Acceptance Criteria

- `db query "SELECT 1"` prints result
- `db query "INSERT INTO ..."` prints affected rows
- `db query -f file.sql` reads and executes file
- `echo "SELECT 1" | db query` works
- `db query --format csv "SELECT ..."` outputs CSV
- Transaction mode toggle works
- Query timing is reported
- Connection errors produce clear messages

## Dependencies

- M1 (database layer)
- M6 (export) — soft dependency for `--format` flag, can ship without

## Can Be Parallelized With

- M3, M5, M6 — all Phase 2 milestones
