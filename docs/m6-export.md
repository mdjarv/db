# M6: Export

## Goal

Export query results to CSV, JSON, and SQL formats. Used by CLI `--format` flag and TUI export action.

## Tasks

### Exporter Interface (`internal/export/exporter.go`)

- [ ] `Exporter` interface: `Export(w io.Writer, result *db.Result) error`
- [ ] `Format` type: CSV, JSON, SQL, Table
- [ ] Factory: `NewExporter(format Format, opts Options) Exporter`
- [ ] `Options`: delimiter (CSV), pretty (JSON), table name (SQL), include headers

### CSV Exporter (`internal/export/csv.go`)

- [ ] Standard `encoding/csv` writer
- [ ] Configurable delimiter (comma, tab, pipe)
- [ ] Optional header row
- [ ] NULL handling: empty string or configurable placeholder
- [ ] Proper quoting/escaping

### JSON Exporter (`internal/export/json.go`)

- [ ] Array of objects: `[{"id": 1, "name": "alice"}, ...]`
- [ ] Streaming: write objects as rows arrive, don't buffer entire result
- [ ] Pretty-print option (indented)
- [ ] JSON lines option (one object per line, no array wrapper)
- [ ] Type-aware: numbers as numbers, booleans as booleans, nulls as null

### SQL Exporter (`internal/export/sql.go`)

- [ ] Generate INSERT statements: `INSERT INTO <table> (cols) VALUES (...);`
- [ ] Proper value escaping (strings, NULLs, booleans)
- [ ] Batch mode: multi-row INSERT for efficiency
- [ ] Table name required (from flag or active table context)

### Table Formatter (`internal/export/table.go`)

- [ ] Pretty aligned table for terminal output (default CLI format)
- [ ] Auto-detect column widths from data
- [ ] Max column width with truncation + ellipsis
- [ ] Row count footer
- [ ] Unicode box-drawing borders

### Tests

- [ ] Unit: CSV with various data types, NULL handling, delimiter options
- [ ] Unit: JSON with type preservation, streaming correctness
- [ ] Unit: JSON lines format
- [ ] Unit: SQL with escaping edge cases (quotes in strings, NULL, booleans)
- [ ] Unit: Table formatter with alignment, truncation
- [ ] All tests use mock Results with known data — no DB needed

## Acceptance Criteria

- `db query --format csv "SELECT ..."` outputs valid CSV
- `db query --format json "SELECT ..."` outputs valid JSON with correct types
- `db query --format sql --table users "SELECT ..."` outputs INSERT statements
- Default table format is aligned and readable
- NULL values handled correctly in all formats
- Large results stream without OOM (no full buffering)

## Dependencies

- M1 (database layer — needs Result/Column types)
- No DB connection needed for unit tests (mock Results)

## Can Be Parallelized With

- M3, M4, M5 — all Phase 2 milestones
