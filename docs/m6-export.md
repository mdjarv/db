# M6: Export

## Goal

Export query results to CSV, JSON, and SQL formats. Used by CLI `--format` flag and TUI export action.

## Tasks

### Exporter Interface (`internal/export/exporter.go`)

- [x] `Exporter` interface: `Export(w io.Writer, result *db.Result) error`
- [x] `Format` type: CSV, JSON, SQL, Table
- [x] Factory: `NewExporter(format Format, opts Options) Exporter`
- [x] `Options`: delimiter (CSV), pretty (JSON), table name (SQL), include headers

### CSV Exporter (`internal/export/csv.go`)

- [x] Standard `encoding/csv` writer
- [x] Configurable delimiter (comma, tab, pipe)
- [x] Optional header row
- [x] NULL handling: empty string or configurable placeholder
- [x] Proper quoting/escaping

### JSON Exporter (`internal/export/json.go`)

- [x] Array of objects: `[{"id": 1, "name": "alice"}, ...]`
- [x] Streaming: write objects as rows arrive, don't buffer entire result
- [x] Pretty-print option (indented)
- [x] JSON lines option (one object per line, no array wrapper)
- [x] Type-aware: numbers as numbers, booleans as booleans, nulls as null

### SQL Exporter (`internal/export/sql.go`)

- [x] Generate INSERT statements: `INSERT INTO <table> (cols) VALUES (...);`
- [x] Proper value escaping (strings, NULLs, booleans)
- [x] Batch mode: multi-row INSERT for efficiency
- [x] Table name required (from flag or active table context)

### Table Formatter (`internal/export/table.go`)

- [x] Pretty aligned table for terminal output (default CLI format)
- [x] Auto-detect column widths from data
- [x] Max column width with truncation + ellipsis
- [x] Row count footer
- [x] Unicode box-drawing borders

### Tests

- [x] Unit: CSV with various data types, NULL handling, delimiter options
- [x] Unit: JSON with type preservation, streaming correctness
- [x] Unit: JSON lines format
- [x] Unit: SQL with escaping edge cases (quotes in strings, NULL, booleans)
- [x] Unit: Table formatter with alignment, truncation
- [x] All tests use mock Results with known data — no DB needed

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
