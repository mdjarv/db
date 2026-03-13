# M3: Connection Management

## Goal

Saved connections with keyring-backed credentials, config file, and CLI commands. Users can add, list, remove connections and connect via flags or config.

## Tasks

### Config Types (`internal/conn/config.go`)

- [ ] `ConnectionConfig`: name, host, port, user, dbname, sslmode, options
- [ ] Password stored separately in keyring, not config file
- [ ] DSN builder from config fields
- [ ] DSN parser to extract fields

### Config Store (`internal/conn/store.go`)

- [ ] Load/save connections from `~/.config/db/connections.yaml`
- [ ] CRUD operations: Add, Get, List, Remove
- [ ] Default connection setting
- [ ] XDG path resolution (`internal/config/paths.go`)
- [ ] Unit tests with temp config files

### Keyring Integration (`internal/conn/keyring.go`)

- [ ] Store password: `SetPassword(connectionName, password)`
- [ ] Retrieve password: `GetPassword(connectionName)`
- [ ] Delete password: `DeletePassword(connectionName)`
- [ ] Keyring interface for testability (mock in tests)
- [ ] Fallback behavior when keyring unavailable (prompt every time)

### Connection Resolver (`internal/conn/resolve.go`)

- [ ] Resolution order: CLI flags > env vars > named connection from config
- [ ] `Resolve(flags, envPrefix) -> ConnectionConfig`
- [ ] Env vars: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_DSN`
- [ ] If password missing from all sources: prompt (CLI) or dialog (TUI)

### CLI Commands

- [ ] `db connect add` — interactive: prompt for host, port, user, password, dbname, name
- [ ] `db connect list` — table of saved connections (password masked)
- [ ] `db connect remove <name>` — remove from config + keyring
- [ ] `db ping` — resolve connection, attempt connect, report success/failure with timing

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
