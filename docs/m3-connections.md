# M3: Connection Management

## Goal

Saved connections with keyring-backed credentials, config file, and CLI commands. Users can add, list, remove connections and connect via flags or config.

## Tasks

### Config Types (`internal/conn/config.go`)

- [x] `ConnectionConfig`: name, host, port, user, dbname, sslmode, options
- [x] Password stored separately in keyring, not config file
- [x] DSN builder from config fields
- [x] DSN parser to extract fields

### Config Store (`internal/conn/store.go`)

- [x] Load/save connections from `~/.config/db/connections.yaml`
- [x] CRUD operations: Add, Get, List, Remove
- [x] Default connection setting
- [x] XDG path resolution (`internal/config/paths.go`)
- [x] Unit tests with temp config files

### Keyring Integration (`internal/conn/keyring.go`)

- [x] Store password: `SetPassword(connectionName, password)`
- [x] Retrieve password: `GetPassword(connectionName)`
- [x] Delete password: `DeletePassword(connectionName)`
- [x] Keyring interface for testability (mock in tests)
- [ ] Fallback behavior when keyring unavailable (prompt every time)

### Connection Resolver (`internal/conn/resolve.go`)

- [x] Resolution order: CLI flags > env vars > named connection from config
- [x] `Resolve(flags, envPrefix) -> ConnectionConfig`
- [x] Env vars: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_DSN`
- [ ] If password missing from all sources: prompt (CLI) or dialog (TUI)

### CLI Commands

- [x] `db connect add` — interactive: prompt for host, port, user, password, dbname, name
- [x] `db connect list` — table of saved connections (password masked)
- [x] `db connect remove <name>` — remove from config + keyring
- [x] `db ping` — resolve connection, attempt connect, report success/failure with timing

### TUI Connection Dialog (`internal/tui/components/dialog/`)

- [ ] Modal dialog for connection selection on TUI launch (when no connection specified)
- [ ] List saved connections, allow selection
- [ ] "Add new" option within dialog

## Config File Format

```yaml
default: myapp
connections:
  myapp:
    host: localhost
    port: 5432
    user: myapp
    dbname: myapp_dev
    sslmode: prefer
  staging:
    host: staging-db.internal
    port: 5432
    user: readonly
    dbname: myapp
    sslmode: require
```

## Acceptance Criteria

- `db connect add` saves connection to config + password to keyring
- `db connect list` shows saved connections
- `db connect remove` cleans up config + keyring
- `db ping` connects and reports latency
- `db ping --connection myapp` uses saved connection
- `db ping --dsn "postgres://..."` uses one-off DSN
- `DB_DSN=... db ping` uses env var
- Password prompt appears when password not in keyring or flags

## Dependencies

- M0 (project skeleton)
- M1 (database layer — needs Driver.Connect)

## Can Be Parallelized With

- M4 (Query Engine), M5 (Schema), M6 (Export) — all Phase 2 milestones are parallel
