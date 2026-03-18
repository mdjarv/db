# db - TUI Database Client Roadmap

A vim-modal TUI for PostgreSQL, built for backend developers who want fast schema browsing, querying, and data editing without leaving the terminal. Designed as a pgAdmin replacement.

## Design Principles

- **Separation of concerns**: all database operations live in a library layer usable without TUI
- **CLI-first testing**: every feature has a one-shot CLI subcommand (`db query`, `db tables`, etc.)
- **Driver abstraction**: PostgreSQL now, architecture supports adding SQLite/others later
- **Testability**: interfaces everywhere, testcontainers for integration tests, mock-friendly

## Tech Stack

| Library | Purpose | Why |
|---|---|---|
| `cobra` | CLI framework | Subcommand routing, flag parsing, help generation |
| `viper` | Configuration | Config file (YAML), env vars, flag binding, theme config |
| `bubbletea` | TUI framework | Elm architecture, composable, testable message-passing |
| `bubbles` | TUI components | Table, text input, viewport, list — standard building blocks |
| `lipgloss` | TUI styling | Declarative styling, adaptive colors, border rendering |
| `pgx` | PostgreSQL driver | Pure Go, better perf than lib/pq, COPY support, rich type system |
| `zalando/go-keyring` | OS keyring | Secure credential storage (GNOME Keyring, KWallet, macOS Keychain) |
| `chroma` | Syntax highlighting | SQL highlighting in query editor and result display |
| `testcontainers-go` | Integration tests | Spin up real PostgreSQL in tests, no mocks for DB layer |

## Architecture

```
cmd/                    # cobra commands (CLI entry points)
  root.go               # root command, global flags
  connect.go            # manage saved connections
  ping.go               # test connection
  query.go              # one-shot query execution
  tables.go             # list tables
  describe.go           # describe table schema
  tui.go                # launch TUI (default command)

internal/
  db/                   # database abstraction
    driver.go           # Driver interface
    result.go           # Result, Row, Column types
    postgres/           # pgx-based PostgreSQL implementation
      driver.go
      schema.go
      query.go

  conn/                 # connection management
    config.go           # connection config types
    store.go            # load/save from config file
    keyring.go          # credential storage via OS keyring
    resolve.go          # resolve connection (CLI flags > config > interactive)

  query/                # query execution engine
    executor.go         # execute queries, manage transactions
    history.go          # query history (future)
    completer.go        # SQL autocomplete (tables, columns, keywords)

  schema/               # schema introspection
    inspector.go        # Inspector interface
    types.go            # Table, Column, Index, Constraint, FK types
    postgres.go         # pg_catalog queries

  export/               # result export
    exporter.go         # Exporter interface
    csv.go
    json.go
    sql.go

  editor/               # data editing
    buffer.go           # pending edit buffer (INSERT/UPDATE/DELETE)
    dml.go              # DML generation (parameterized SQL)
    applier.go          # apply changes to DB (with transaction control)

  tui/                  # TUI layer
    app/                # main app model, layout, focus management
      model.go
      layout.go
      keys.go           # global keybinding definitions
      mode.go           # vim mode state machine (normal/insert/command)

    pane/               # pane abstraction
      pane.go           # Pane interface (focusable, resizable)
      manager.go        # pane layout manager (left/right-top/right-bottom)

    components/         # reusable TUI components
      tablelist/        # left pane: table browser
      queryeditor/      # right-top: SQL editor with syntax highlighting
      resultview/       # right-bottom: result table with virtual scrolling
      statusbar/        # bottom: mode indicator, connection info, messages
      commandbar/       # vim : command input
      dialog/           # modal dialogs (confirm, connect, etc.)

    theme/              # theming engine
      theme.go          # Theme type, color palette
      builtin.go        # built-in themes
      loader.go         # load custom themes from config

  config/               # app configuration
    config.go           # viper setup, defaults
    paths.go            # XDG config/data paths
```

See [docs/architecture.md](docs/architecture.md) for detailed design.

## Phases & Milestones

Work is organized into phases. Within each phase, milestones can be worked on **in parallel** by different developers.

### Phase 1: Foundation (complete)

| Milestone | Status | Doc | Description |
|---|---|---|---|
| M0 | done | [Project Skeleton](docs/m0-skeleton.md) | Module init, directory structure, CI, linting |
| M1 | done | [Database Layer](docs/m1-database-layer.md) | Driver interface + PostgreSQL implementation |
| M2 | done | [TUI Shell](docs/m2-tui-shell.md) | App shell, pane layout, vim mode system |

### Phase 2: Core Features (complete)

| Milestone | Status | Doc | Description |
|---|---|---|---|
| M3 | done | [Connection Management](docs/m3-connections.md) | Config file, keyring, CLI `connect`/`ping` |
| M4 | done | [Query Engine](docs/m4-query-engine.md) | Execute queries, transactions, CLI `query` |
| M5 | done | [Schema Inspection](docs/m5-schema.md) | Introspect tables/columns/indexes/FKs, CLI `tables`/`describe` |
| M6 | done | [Export](docs/m6-export.md) | CSV/JSON/SQL export, `--format` flag on `query` |

### Phase 3: TUI Integration (complete)

| Milestone | Status | Doc | Description |
|---|---|---|---|
| M7 | done | [Table Browser](docs/m7-table-browser.md) | Left pane: table list, schema display |
| M8 | done | [Query Editor](docs/m8-query-editor.md) | Right-top: SQL editor, syntax highlighting |
| M9 | done | [Result Viewer](docs/m9-result-viewer.md) | Right-bottom: result table, virtual scrolling |

### Phase 4: Advanced Features (complete)

| Milestone | Status | Doc | Description |
|---|---|---|---|
| M10 | done | [Data Editing](docs/m10-data-editing.md) | Inline edit, change buffer, commit/rollback |
| M11 | done | [Theming](docs/m11-theming.md) | Theme engine, built-in themes |
| M12 | mostly done | [Multi-Query Buffers](docs/m12-multi-query.md) | Multiple query/result pairs, buffer switching |

### Phase 5: Polish (in progress)

| Milestone | Status | Doc | Description |
|---|---|---|---|
| M13 | in progress | [Integration & Polish](docs/m13-polish.md) | Integration tests, error handling, help system, keybinding cheatsheet |

## Remaining Work

### M12 (minor)
- Buffer list overlay popup (`:ls`)
- Modified indicator `[+]` for unsaved buffers

### M13 (polish)
- TUI-level integration tests via `teatest`
- Full flow TUI tests (connect -> browse -> query -> export)
- Permission denied error display
- TUI rendering profiling (<16ms)
- Connection pool tuning
- Lazy schema loading
- README with screenshots, install instructions
- CHANGELOG.md

### Not planned
- SQL autocomplete (M4/M8) — deferred
- Schema listing (`Schemas()`) — deferred
- Homebrew/AUR packages — future

## Key Decisions

| Decision | Choice | Rationale |
|---|---|---|
| DB engines | PostgreSQL only | Focus. Driver interface allows adding others later. |
| Vim model | Full modal (normal/insert/command) | Matches target audience (backend devs using vim) |
| Query editor | Built-in with autocomplete | No external process spawning, tighter integration |
| Pagination | Virtual scrolling | Best UX despite complexity, core differentiator |
| Multi-connection | Single conn, multi-query | Covers dev workflows without tab complexity |
| Transactions | Configurable (default explicit) | Safe for prod, convenient for local |
| Theming | Built-in themes | Consistent look, good for demos and adoption |
| Credentials | OS keyring | Security requirement for saved connections |
| Go version | 1.26 | Latest features, project is greenfield |
| Module path | github.com/mdjarv/db | Standard GitHub path |

## CLI Command Overview

```
db                          # launch TUI (default)
db tui                      # launch TUI (explicit)
db ping                     # test connection
db connect add              # add/save a connection
db connect list             # list saved connections
db connect remove <name>    # remove saved connection
db connect default <name>   # set default connection
db connect rename <old> <new> # rename saved connection
db connect edit <name>      # edit saved connection interactively
db query "SELECT ..."       # one-shot query, results to stdout
db query -f file.sql        # execute SQL from file
db query --format csv "..." # query with format (csv, json, sql, table)
db tables                   # list tables
db tables --schema <name>   # list tables in schema
db describe <table>         # show table schema
db introspect <table>       # show type introspection details (OIDs, enum values, composite fields)
```

Global flags: `--connection <name>`, `--dsn <url>`, `--host`, `--port`, `--user`, `--password`, `--dbname`, `--sslmode`

## TUI Layout

```
╭───────────────────╮╭──────────────────────────────────────────╮
│ users          100 ││1 SELECT * FROM users                     │
│ posts           42 │╰──────────────────────────────────────────╯
│ comments        17 │╭──────────────────────────────────────────╮
│ tags             5 ││ id   │ name  │ email    │ active │ crea… │
│ categories       3 ││──────┼───────┼──────────┼────────┼─────  │
│                    ││ 1    │ alice │ a@e.com  │ true   │ 2024  │
│                    ││ 2    │ bob   │ b@e.com  │ true   │ 2024  │
│                    ││ 3    │ carol │ c@e.com  │ false  │ 2024  │
│                    ││ ..   │ ..    │ ..       │ ..     │ ..    │
│                    │╰─ rows 1-100 of 2847 | 5 cols | 12ms ───╯
╰────────────────────╯
 NORMAL | connected: myapp@localhost:5432/mydb   | rows: 2847
```

- Selected table highlighted with background color (no chevron)
- Query pane auto-sizes: 1 line up to 1/4 screen
- Result status integrated into bottom border
- Partial rightmost column (dimmed) hints more columns exist
