# M1: Database Abstraction Layer

## Goal

Define the driver interface and implement PostgreSQL via pgx. This is the foundation everything else builds on.

## Tasks

### Interfaces (`internal/db/`)

- [ ] Define `Driver` interface: `Connect(ctx, dsn) (Conn, error)`
- [ ] Define `Conn` interface: `Query`, `Exec`, `Begin`, `Close`
- [ ] Define `Tx` interface: `Query`, `Exec`, `Commit`, `Rollback`
- [ ] Define `Result` struct: `Columns []Column`, `Rows RowIterator`
- [ ] Define `RowIterator` interface: `Next`, `Values`, `Err`, `Close`
- [ ] Define `ExecResult`: `RowsAffected int64`
- [ ] Define `Column`: `Name`, `TypeName`, `TypeOID`
- [ ] Driver registry: `Register(name, driver)`, `Open(name, dsn)`

### PostgreSQL Implementation (`internal/db/postgres/`)

- [ ] Implement `Driver` using `pgx/v5`
- [ ] Implement `Conn` wrapping `pgxpool.Pool`
- [ ] Implement `Tx` wrapping `pgx.Tx`
- [ ] Implement `RowIterator` wrapping `pgx.Rows`
- [ ] Map pgx types to string type names for display
- [ ] Connection pooling: use `pgxpool` with sensible defaults (max 5 conns)

### Tests

- [ ] Integration tests using `testcontainers-go`
  - Spin up PostgreSQL container
  - Create test schema (users, posts tables with FKs)
  - Test Connect, Query, Exec, Begin/Commit/Rollback
  - Test RowIterator lifecycle (Next/Values/Close)
  - Test query cancellation via context
  - Test connection error handling (bad DSN, refused, timeout)

## Design Notes

- `RowIterator` is streaming — rows are not loaded into memory all at once. This is critical for virtual scrolling later.
- `pgxpool` handles connection pooling. The `Conn` wrapper holds a pool, not a single connection.
- Type OIDs are preserved for consumers that need rich type info (e.g., data editing needs to know if a column is a bool vs varchar for UI rendering).

## Acceptance Criteria

- Can connect to a real PostgreSQL instance via DSN
- Can execute SELECT and get streaming results
- Can execute INSERT/UPDATE/DELETE and get affected row count
- Can run queries in a transaction with commit/rollback
- All tests pass against testcontainers PostgreSQL
- No pgx types leak outside `internal/db/postgres/`

## Dependencies

- M0 (project skeleton)

## Can Be Parallelized With

- M2 (TUI Shell) — TUI work uses mock data, no DB dependency
