# Architecture

## Layer Diagram

```
┌─────────────────────────────────────────────────┐
│  CLI (cobra)            TUI (bubbletea)         │  ← entry points
├─────────────────────────────────────────────────┤
│  internal/query         internal/schema          │
│  internal/export        internal/editor          │  ← feature layer
├─────────────────────────────────────────────────┤
│  internal/conn                                   │  ← connection management
├─────────────────────────────────────────────────┤
│  internal/db                                     │  ← driver abstraction
│    ├── Driver interface                          │
│    └── postgres/ (pgx implementation)            │
└─────────────────────────────────────────────────┘
```

## Key Interfaces

### Driver

```go
type Driver interface {
    Connect(ctx context.Context, dsn string) (Conn, error)
}

type Conn interface {
    Query(ctx context.Context, sql string, args ...any) (Result, error)
    Exec(ctx context.Context, sql string, args ...any) (ExecResult, error)
    Begin(ctx context.Context) (Tx, error)
    Dialect() Dialect
    Close(ctx context.Context) error
}

type Tx interface {
    Query(ctx context.Context, sql string, args ...any) (Result, error)
    Exec(ctx context.Context, sql string, args ...any) (ExecResult, error)
    Dialect() Dialect
    Commit(ctx context.Context) error
    Rollback(ctx context.Context) error
}
```

Drivers register themselves via `init()`:

```go
db.Register("postgres", &Driver{})
```

Callers open connections by name or by DSN scheme:

```go
c, err := db.Open(ctx, "postgres", dsn)  // explicit driver name
c, err := db.OpenDSN(ctx, "sqlite:///path/to/file.db")  // scheme-dispatched
```

### Dialect

Dialect captures the SQL syntax that varies between drivers so that feature
layers (editor DML generation, identifier quoting) stay driver-agnostic.

```go
type Dialect struct {
    Name            string
    Placeholder     func(n int) string // $1 / ?
    QuoteIdent      func(s string) string
    DefaultSchema   string
    SupportsSchemas bool
}

func (d Dialect) QualifyTable(schema, table string) string
```

`db.PostgresDialect()` and `db.SQLiteDialect()` provide defaults. Each
driver's `Conn`/`Tx` returns its dialect via `Dialect()`.

### Result

```go
type Result struct {
    Columns []Column
    Rows    RowIterator
}

type RowIterator interface {
    Next() bool
    Values() ([]any, error)
    Err() error
    Close()
}

type Column struct {
    Name            string
    TypeName        string
    EnumValues      []string         // optional (nil if unsupported)
    CompositeFields []CompositeField // optional (nil if unsupported)
}

type CompositeField struct {
    Name     string
    TypeName string
}
```

`Column.IsArray()` and `Column.ElemTypeName()` derive array metadata from
the type name (trailing `[]`), so introspection does not need driver-
specific OIDs.

### Inspector

```go
type Inspector interface {
    Tables(ctx context.Context, schema string) ([]Table, error)
    Columns(ctx context.Context, schema string, table string) ([]ColumnInfo, error)
    Indexes(ctx context.Context, schema string, table string) ([]Index, error)
    Constraints(ctx context.Context, schema string, table string) ([]Constraint, error)
    ForeignKeys(ctx context.Context, schema string, table string) ([]ForeignKey, error)
}
```

An empty `schema` argument means "use the driver's default namespace". The
PostgreSQL inspector normalises it to `public`; drivers without schema
namespaces (SQLite) should ignore non-empty values.

Each driver's `*Conn` exposes `Inspector() schema.Inspector`. Call-sites
obtain an inspector driver-agnostically via:

```go
insp, err := schema.NewInspector(conn)
```

### Exporter

```go
type Exporter interface {
    Export(w io.Writer, result *db.Result) error
}
```

## Design Rules

1. **No TUI imports in internal/db, internal/query, internal/schema, internal/export, internal/editor, internal/conn.** These packages must be usable from CLI commands and tests without bubbletea.

2. **All database access goes through the Driver interface.** No direct pgx calls outside `internal/db/postgres/`.

3. **TUI components communicate via bubbletea messages only.** No shared mutable state between panes. The app model routes messages.

4. **Connection resolution order**: CLI flags > named connection (project-local then global store) > environment variables > store default (project-local then global).

5. **Config follows XDG**: `~/.config/db/config.yaml` for settings, `~/.local/share/db/` for data (history, etc.).

## Vim Mode State Machine

```
                 ┌──────────┐
        Esc      │          │    i, a, o, I, A, O
    ┌───────────>│  NORMAL  │──────────────────────┐
    │            │  h/l col │                       │
    │            └┬───┬───┬┬┘                       v
    │             │:  │V  ││v (on results)     ┌──────────┐
    │             v   v   │v                   │  INSERT  │
    │     ┌────────┐┌─────┴─┐┌────────┐        │          │
    │     │COMMAND ││V-LINE ││V-BLOCK │        └──────────┘
    └─────│        ││j/k row││h/j/k/l │
 Esc/Enter└────────┘│Tab col││y yank  │
    │     └────────││y yank │└────────┘
    │     │        │└───────┘
    │     │ Enter/e│ (on results)
    │     v        │
    │  ┌────────┐  │
    └──│  EDIT  │──┘
 Esc   │ dialog │
       └────────┘
```

- **NORMAL**: hjkl cell cursor (row + column), Ctrl+hjkl pane switching, `y` yank cell, `Y` yank row
- **INSERT**: text input in query editor, search filter, command bar
- **COMMAND**: `:` prefix commands (`:w` run query, `:q` quit, `:set` config, `:theme`, `:commit`, `:rollback`)
- **EDIT**: popup dialog with type-aware input modes (text, enum, array, composite); arrays support J/K reorder, enum arrays use enum picker per element; Tab cycles OK/[NULL]/Cancel (NULL hidden when not nullable)
- **V-LINE**: `V` on results — row selection, Tab toggles row/column axis, `y` yanks CSV
- **V-BLOCK**: `v` on results — rectangular selection via h/j/k/l, `y` yanks CSV

## Pane Focus

```
Ctrl+h/l      → move focus left/right
Ctrl+j/k      → move focus up/down (within right panes)
Tab           → cycle focus forward
Shift-Tab     → cycle focus backward
1/2/3         → jump to pane by number (in normal mode)
+/-           → grow/shrink left pane (in normal mode)
```

## Testing Strategy

| Layer | Method | Tools |
|---|---|---|
| `internal/db` | Integration tests | testcontainers-go (real PostgreSQL) |
| `internal/conn` | Unit tests | Mock keyring, temp config files |
| `internal/query` | Integration tests | testcontainers-go |
| `internal/schema` | Integration tests | testcontainers-go with known schema |
| `internal/export` | Unit tests | Mock Result with known data |
| `internal/editor` | Unit tests | Mock Conn/Tx |
| `internal/tui` | Unit tests | bubbletea teatest package |
| CLI commands | Integration tests | testcontainers-go + cobra test helpers |
| End-to-end | Manual + scripted | teatest for basic flows |

## Adding a New Database Driver

The `db.Conn` interface is the only contract a driver must satisfy. Drivers
are selected by name (registry) or by DSN scheme (`db.OpenDSN`).

1. Create `internal/db/<driver>/` (e.g. `internal/db/sqlite/`).
2. Implement `db.Driver`, `db.Conn`, `db.Tx`. `Conn` and `Tx` both return a
   `db.Dialect` describing placeholder syntax, identifier quoting, and
   whether the driver supports schema namespaces. Two dialects ship in
   `internal/db`: `PostgresDialect()` and `SQLiteDialect()`.
3. Create `internal/schema/<driver>.go` implementing `schema.Inspector`.
   Drivers without schema namespaces should accept an empty schema argument
   and ignore non-default values.
4. Expose `(*Conn).Inspector() schema.Inspector` so `schema.NewInspector`
   can locate it via type assertion.
5. Register the driver in `init()`: `db.Register("sqlite", &Driver{})`.
6. Add the scheme mapping to `conn.DriverFromScheme` and the corresponding
   `DSN`/`ParseDSN` arm in `internal/conn/config.go`.
7. Blank-import the driver from `cmd/flags.go` to register it at startup.

All feature layers (query, export, editor) are driver-agnostic: they use
`Dialect` for placeholder/quoting and `Inspector` for metadata, never raw
driver types.
