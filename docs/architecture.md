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
    Close(ctx context.Context) error
    // Schema inspection is separate — see Inspector
}

type Tx interface {
    Query(ctx context.Context, sql string, args ...any) (Result, error)
    Exec(ctx context.Context, sql string, args ...any) (ExecResult, error)
    Commit(ctx context.Context) error
    Rollback(ctx context.Context) error
}
```

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
    TypeOID         uint32
    EnumValues      []string         // non-nil for enum types
    CompositeFields []CompositeField // non-nil for composite types
}

type CompositeField struct {
    Name     string
    TypeName string
}

// Optional — implemented by postgres driver for `db introspect`
type TypeIntrospector interface {
    TypeDetail(oid uint32) TypeDetail
}

type TypeDetail struct {
    OID             uint32
    Name            string
    IsArray         bool
    ElemOID         uint32
    ElemTypeName    string
    EnumValues      []string
    CompositeFields []CompositeField
}
```

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

4. **Connection resolution order**: CLI flags > environment variables > config file > interactive prompt.

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

1. Create `internal/db/sqlite/` (or mysql, etc.)
2. Implement `Driver`, `Conn`, `Tx` interfaces
3. Create `internal/schema/sqlite.go` implementing `Inspector`
4. Register driver in a factory: `db.Register("sqlite", &sqlite.Driver{})`
5. All feature layers (query, export, editor) work unchanged
