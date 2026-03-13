# db

TUI database client for PostgreSQL. See [ROADMAP.md](ROADMAP.md) for full context.

## Build / Test / Lint

```
make build      # build binary
make lint       # golangci-lint
make test       # unit tests
make test-integration  # integration tests (requires docker)
```

## Architecture

- `cmd/` — cobra CLI commands
- `internal/db/` — driver abstraction, postgres implementation
- `internal/conn/` — connection management
- `internal/query/` — query execution
- `internal/schema/` — schema introspection
- `internal/export/` — result export (CSV/JSON/SQL)
- `internal/editor/` — data editing buffer
- `internal/tui/` — bubbletea TUI layer
- `internal/config/` — viper config, XDG paths

See [docs/architecture.md](docs/architecture.md) for interfaces and design rules.
