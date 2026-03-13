# M7: Table Browser (Left Pane)

## Goal

Left pane: browsable table list with schema details for the selected table. Replaces M2 stub.

## Tasks

### Table List Component (`internal/tui/components/tablelist/`)

- [ ] Filterable list of tables from `schema.Inspector`
- [ ] Show table type icon/indicator (table, view, materialized view)
- [ ] Show row count estimate next to table name
- [ ] `/` to filter/search tables (fuzzy or prefix)
- [ ] `j/k` to navigate, `Enter` to select
- [ ] `gg` / `G` to jump to top/bottom
- [ ] `R` to refresh schema

### Schema Detail View

- [ ] Below table list (or toggle): show columns of selected table
- [ ] Column name, type, nullable, default, PK indicator
- [ ] Index list with type and columns
- [ ] FK relationships with referenced table
- [ ] Constraints summary
- [ ] Scroll independently from table list

### Table Context Actions

- [ ] `Enter` on table: `SELECT * FROM <table> LIMIT 100` into query editor + run
- [ ] `d` on table: switch to `db describe` view (full schema detail)
- [ ] `y` on table: yank table name to query editor at cursor

### Integration with App

- [ ] Table list loads on connection via `schema.Inspector`
- [ ] Selected table broadcasts `TableSelectedMsg` for schema detail update
- [ ] `Enter` sends `QueryRequestMsg` to query editor pane

## Keybindings (Normal Mode)

| Key | Action |
|---|---|
| `j/k` | Navigate table list |
| `gg/G` | Top/bottom of list |
| `/` | Filter tables (enters insert mode for filter input) |
| `Enter` | Quick-query selected table |
| `d` | Describe selected table |
| `y` | Yank table name |
| `R` | Refresh schema |

## Tests

- [ ] Unit: table list filtering
- [ ] Unit: table selection message dispatch
- [ ] teatest: render with mock schema data, verify layout

## Acceptance Criteria

- Table list shows all tables from connected database
- Filtering narrows list as user types
- Schema detail updates when table selection changes
- Enter triggers query in query editor pane
- Table type indicators visible (table vs view)

## Dependencies

- M2 (TUI shell — pane system)
- M5 (schema inspection — Inspector interface)
