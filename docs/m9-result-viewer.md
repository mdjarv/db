# M9: Result Viewer (Right-Bottom Pane)

## Goal

Tabular result display with virtual scrolling for large result sets. Supports navigation, column resizing, and value inspection.

## Tasks

### Virtual Scroll Engine (`internal/tui/components/resultview/scroll.go`)

- [ ] `ScrollState`: total rows (estimated), loaded rows, viewport offset, viewport height
- [ ] Fetch strategy: request page when viewport approaches edge of loaded data
- [ ] Page size: configurable, default 200 rows
- [ ] Bidirectional: support scrolling up through already-fetched data
- [ ] Row cache: keep fetched rows in memory, evict oldest when over limit
- [ ] Cache size limit: configurable, default 10,000 rows

### Result Table Component (`internal/tui/components/resultview/`)

- [x] Column headers with type indicators (blue, bold)
- [ ] Fixed header row (doesn't scroll with data)
- [ ] Aligned columns with auto-detected widths
- [ ] Column resize: manually adjust with keybinding
- [x] Horizontal scrolling for wide results (partial rightmost column shown dimmed as hint)
- [ ] Cell value truncation with ellipsis
- [ ] NULL display: distinct style (e.g., dim italic "NULL")
- [ ] Row cursor: highlighted current row
- [x] Status integrated into bottom border: "rows 1-50 of ~2847 | N cols | duration"

### Cell Inspector

- [ ] `Enter` on cell: popup showing full value (for long text, JSON, etc.)
- [ ] Copy cell value: `y` yanks cell content
- [ ] Copy row: `Y` yanks entire row as tab-separated

### Navigation (Normal Mode)

| Key | Action |
|---|---|
| `j/k` | Move cursor down/up one row |
| `h/l` | Scroll columns left/right |
| `gg/G` | Jump to first/last row |
| `Ctrl-d/Ctrl-u` | Half-page down/up |
| `Ctrl-f/Ctrl-b` | Full page down/up |
| `0/$` | First/last column |
| `Enter` | Inspect cell value |
| `y` | Yank cell value |
| `Y` | Yank row |
| `/` | Search within results |

### Export from TUI

- [ ] `:export csv <file>` — export current results to file
- [ ] `:export json <file>` — export as JSON
- [ ] `:export sql <file>` — export as INSERT statements
- [ ] Uses `internal/export` package

### Result Metadata

- [ ] Query execution time display
- [ ] Row count (exact after full fetch, estimated before)
- [ ] Column count

## Tests

- [ ] Unit: virtual scroll state transitions (page fetching logic)
- [ ] Unit: column width calculation
- [ ] Unit: cell truncation and NULL rendering
- [ ] Unit: row cache eviction
- [ ] teatest: render with mock data, verify layout and scrolling

## Acceptance Criteria

- Large result sets (100k+ rows) load without delay — only visible rows fetched
- Scrolling is smooth with no visible loading stutter for cached data
- Loading indicator appears when fetching next page
- Columns auto-size to content
- Long values truncated with full value available via Enter
- Horizontal scrolling works for wide tables
- Export commands write correct files

## Dependencies

- M2 (TUI shell — pane system)
- M4 (query engine — Result type, streaming)
- M6 (export — for TUI export commands)
