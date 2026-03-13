# M12: Multi-Query Buffers

## Goal

Multiple query buffers against the same connection. Each buffer has its own query text and result set. Switch between buffers like vim buffers.

## Tasks

### Buffer Manager (`internal/tui/app/buffers.go`)

- [x] `BufferManager`: list of query buffers, active buffer index
- [x] `Buffer`: query text, result set, scroll state, execution state
- [x] Create new buffer: `:new` or `:enew`
- [x] Close buffer: `:bd` (buffer delete)
- [x] Switch buffer: `:bn` / `:bp` (next/prev) or `:b <n>` (by number)
- [x] Buffer list: `:ls` or `:buffers`
- [x] Max buffers: configurable, default 10

### TUI Integration

- [x] Query editor and result viewer swap content on buffer switch
- [x] Buffer indicator in status bar: `[2/5]`
- [ ] Buffer list overlay (`:ls` shows popup)
- [ ] Modified indicator: `[+]` for buffers with unsaved queries
- [x] Each buffer maintains independent scroll position in results

### Keybindings

| Key | Action |
|---|---|
| `:new` | Create new query buffer |
| `:bd` | Close current buffer |
| `:bn` / `gt` | Next buffer |
| `:bp` / `gT` | Previous buffer |
| `:b <n>` | Switch to buffer N |
| `:ls` | List buffers |

## Tests

- [x] Unit: buffer creation, deletion, switching
- [x] Unit: buffer manager wrapping (last -> first)
- [x] Unit: buffer state preservation on switch

## Acceptance Criteria

- Can create multiple query buffers
- Switching buffers swaps query text and results
- Each buffer maintains its own state
- Buffer indicator visible in status bar
- `:ls` shows all buffers with indicators

## Dependencies

- M2 (TUI shell)
- M8 (query editor)
- M9 (result viewer)
