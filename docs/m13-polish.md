# M13: Integration & Polish

## Goal

End-to-end integration testing, error handling hardening, help system, and release readiness.

## Tasks

### Integration Tests

- [ ] Full flow: connect -> browse tables -> query -> view results -> export
- [ ] Full flow: connect -> browse tables -> edit cell -> commit
- [x] Full flow: CLI `db connect add` -> `db tables` -> `db query` -> `db query --format csv`
- [ ] TUI tests via `teatest`: launch app, navigate panes, execute query
- [x] Connection failure scenarios: refused, timeout, auth failure, SSL mismatch
- [x] Large result sets: 100k+ rows, verify virtual scroll doesn't OOM
- [x] Concurrent query buffers: run query in one, switch to another

### Error Handling

- [x] Consistent error display in TUI: red status bar message, auto-dismiss after timeout
- [x] Connection lost: detect, show reconnect dialog, auto-reconnect option
- [x] Query timeout: configurable timeout, cancel on timeout, clear message
- [ ] Permission denied: show which permission is missing
- [x] CLI errors: structured error messages with exit codes

### Help System

- [x] `?` overlay: context-sensitive keybinding cheatsheet
- [x] Different help for each pane (table browser, query editor, result viewer)
- [x] Different help per mode (normal vs insert)
- [x] `:help <topic>` for detailed help
- [x] `--help` on all CLI commands (cobra auto-generates)

### Performance

- [ ] Profile TUI rendering — ensure <16ms frame time
- [ ] Connection pool tuning based on usage patterns
- [x] Schema cache: avoid re-querying schema on every table select
- [ ] Lazy loading: don't fetch schema until pane is focused

### Release

- [x] `goreleaser` config for cross-platform builds
- [ ] Homebrew formula (future)
- [ ] AUR package (future)
- [ ] README.md with screenshots, install instructions, keybinding reference
- [ ] CHANGELOG.md

## Acceptance Criteria

- All integration tests pass against testcontainers PostgreSQL
- Connection loss is handled gracefully (no panic, clear message)
- Help overlay shows correct bindings for current context
- TUI renders at 60fps on standard terminal
- Binary builds for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64

## Dependencies

- All previous milestones substantially complete
