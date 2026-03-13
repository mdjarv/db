# db

A vim-modal TUI database client for PostgreSQL. Fast schema browsing, querying, and data editing without leaving the terminal.

## Features

- **Vim keybindings** — normal/insert/command/visual modes, hjkl navigation, `:w` to run queries
- **Table browser** — schema-aware table list with row counts and sizes
- **Query editor** — SQL syntax highlighting, multi-line editing, query history
- **Result viewer** — virtual scrolling, column resizing, visual-line and visual-block selection
- **Data editing** — inline cell editing with type-aware editors (text, enum, array, composite), change buffer with `:commit`/`:rollback`
- **Type introspection** — automatic enum pickers, composite field editors, enum array support
- **Theming** — built-in themes (default-dark, nord, dracula, etc.)
- **Connection management** — saved connections with OS keyring credential storage
- **Export** — CSV, JSON, SQL output formats

## Install

```
go install github.com/mdjarv/db@latest
```

## Usage

```
db                              # launch TUI (default)
db -c myconn                    # launch TUI with saved connection
db --dsn postgres://user:pass@host:5432/dbname
```

### CLI Commands

```
db ping                         # test connection
db tables                       # list tables
db describe <table>             # show columns, indexes, constraints, foreign keys
db introspect <table>           # show type details (OIDs, enum values, composite fields)
db query "SELECT ..."           # one-shot query
db query -f file.sql            # execute SQL file
db query --format csv "..."     # output as csv, json, sql, or table
db connect add                  # save a connection
db connect list                 # list saved connections
db connect default <name>       # set default connection
```

### Global Flags

```
-c, --connection <name>     named connection from config
    --dsn <url>             full connection URL
-H, --host <host>           database host
-p, --port <port>           database port (default 5432)
-U, --user <user>           database user
-W, --password <pass>       database password
-d, --dbname <name>         database name
    --sslmode <mode>        SSL mode
    --theme <name>          color theme
```

### TUI Keybindings

| Key | Mode | Action |
|-----|------|--------|
| `hjkl` | Normal | Navigate cells |
| `Ctrl+h/j/k/l` | Normal | Switch panes |
| `i`, `a`, `o` | Normal | Enter insert mode |
| `Esc` | Insert | Back to normal |
| `:w` | Command | Run query |
| `:q` | Command | Quit |
| `e`, `Enter` | Normal (results) | Edit cell |
| `V` | Normal (results) | Visual-line select |
| `v` | Normal (results) | Visual-block select |
| `y` | Normal/Visual | Yank cell/selection |
| `:commit` | Command | Apply pending edits |
| `:rollback` | Command | Discard pending edits |
| `+`/`-` | Normal | Resize left pane |

### Edit Dialog

Type-aware cell editor with vim-style controls:

- **Text** — single/multi-line input with type-specific filters and placeholders
- **Enum** — j/k picker from database enum values
- **Array** — j/k navigate, e edit, a add, dd delete, J/K reorder elements
- **Composite** — field-by-field editing with per-field type awareness

## Configuration

Config lives at `~/.config/db/config.yaml` (XDG). Connections stored in `~/.config/db/connections.yaml`, passwords in OS keyring.

## Development

```
make build              # build binary
make lint               # gofmt + golangci-lint
make test               # unit tests
make test-integration   # integration tests (requires docker)
```

See [ROADMAP.md](ROADMAP.md) for project phases and [docs/architecture.md](docs/architecture.md) for design.
