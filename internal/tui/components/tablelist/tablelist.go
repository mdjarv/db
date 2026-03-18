// Package tablelist implements the table browser pane.
package tablelist

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/schema"
	"github.com/mdjarv/db/internal/tui/core"
	"github.com/mdjarv/db/internal/tui/theme"
)

type viewMode int

const (
	viewList viewMode = iota
	viewDetail
)

// Model is the table browser state.
type Model struct {
	tables   []schema.Table
	filtered []schema.Table
	cursor   int
	focused  bool
	width    int
	height   int
	offset   int

	filter    string
	filtering bool

	view   viewMode
	detail TableDetail

	lastG bool // for gg sequence
}

// TableDetail holds schema detail for the selected table.
type TableDetail struct {
	Table       schema.Table
	Columns     []schema.ColumnInfo
	Indexes     []schema.Index
	Constraints []schema.Constraint
	ForeignKeys []schema.ForeignKey
	offset      int
}

// New creates an empty table list.
func New() *Model {
	return &Model{}
}

func (m *Model) selected() (schema.Table, bool) {
	if len(m.filtered) == 0 {
		return schema.Table{}, false
	}
	return m.filtered[m.cursor], true
}

func (m *Model) applyFilter() {
	if m.filter == "" {
		m.filtered = m.tables
		return
	}
	lower := strings.ToLower(m.filter)
	m.filtered = nil
	for _, t := range m.tables {
		if strings.Contains(strings.ToLower(t.Name), lower) {
			m.filtered = append(m.filtered, t)
		}
	}
}

func (m *Model) clampCursor() {
	if m.cursor >= len(m.filtered) {
		m.cursor = max(len(m.filtered)-1, 0)
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.clampOffset()
}

func (m *Model) clampOffset() {
	vh := m.listViewHeight()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor-m.offset >= vh {
		m.offset = m.cursor - vh + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

func (m *Model) listViewHeight() int {
	h := m.height - 2 // border
	if m.filtering {
		h-- // filter input line
	}
	return max(h, 1)
}

func (m *Model) emitSelected() tea.Cmd {
	t, ok := m.selected()
	if !ok {
		return nil
	}
	return func() tea.Msg { return core.TableSelectedMsg{Table: t} }
}

// Update handles input messages.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case core.SchemaLoadedMsg:
		if msg.Err != nil {
			return nil
		}
		m.tables = msg.Tables
		m.applyFilter()
		m.clampCursor()
		return m.emitSelected()

	case core.TableDetailMsg:
		m.detail = TableDetail{
			Table:       msg.Table,
			Columns:     msg.Columns,
			Indexes:     msg.Indexes,
			Constraints: msg.Constraints,
			ForeignKeys: msg.ForeignKeys,
		}
		return nil

	case tea.KeyMsg:
		if !m.focused {
			return nil
		}
		if m.filtering {
			return m.updateFilter(msg)
		}
		if m.view == viewDetail {
			return m.updateDetail(msg)
		}
		return m.updateList(msg)
	}
	return nil
}

func (m *Model) updateFilter(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.filtering = false
		m.filter = ""
		m.applyFilter()
		m.clampCursor()
		return m.emitSelected()
	case "enter":
		m.filtering = false
		return nil
	case "backspace":
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
			m.applyFilter()
			m.clampCursor()
			return m.emitSelected()
		}
	default:
		r := msg.Runes
		if len(r) == 1 {
			m.filter += string(r)
			m.applyFilter()
			m.clampCursor()
			return m.emitSelected()
		}
	}
	return nil
}

func (m *Model) updateList(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	switch key {
	case "j", "down":
		m.lastG = false
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
			m.clampOffset()
			return m.emitSelected()
		}
	case "k", "up":
		m.lastG = false
		if m.cursor > 0 {
			m.cursor--
			m.clampOffset()
			return m.emitSelected()
		}
	case "g":
		if m.lastG {
			m.lastG = false
			m.cursor = 0
			m.offset = 0
			return m.emitSelected()
		}
		m.lastG = true
	case "G":
		m.lastG = false
		if len(m.filtered) > 0 {
			m.cursor = len(m.filtered) - 1
			m.clampOffset()
			return m.emitSelected()
		}
	case "/":
		m.lastG = false
		m.filtering = true
		m.filter = ""
	case "enter":
		m.lastG = false
		if t, ok := m.selected(); ok {
			sql := fmt.Sprintf("SELECT * FROM %s LIMIT 100;", quoteIdent(t.Schema, t.Name))
			return func() tea.Msg { return core.QueryRequestMsg{SQL: sql} }
		}
	case "d":
		m.lastG = false
		m.view = viewDetail
		m.detail.offset = 0
	case "D":
		m.lastG = false
		if t, ok := m.selected(); ok {
			name := t.Name
			return func() tea.Msg { return core.DumpTableMsg{Table: name} }
		}
	case "y":
		m.lastG = false
		if t, ok := m.selected(); ok {
			return func() tea.Msg { return core.YankMsg{Content: t.Name} }
		}
	case "R":
		m.lastG = false
		return func() tea.Msg { return core.RefreshSchemaMsg{} }
	default:
		m.lastG = false
	}
	return nil
}

func (m *Model) updateDetail(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q":
		m.view = viewList
	case "j", "down":
		m.detail.offset++
	case "k", "up":
		if m.detail.offset > 0 {
			m.detail.offset--
		}
	case "g":
		if m.lastG {
			m.lastG = false
			m.detail.offset = 0
			return nil
		}
		m.lastG = true
	case "G":
		m.lastG = false
		// jump to bottom handled by clamping in view
		m.detail.offset = 9999
	default:
		m.lastG = false
	}
	return nil
}

func quoteIdent(schemaName, table string) string {
	if schemaName != "" && schemaName != "public" {
		return schemaName + "." + table
	}
	return table
}

// View renders the table browser.
func (m *Model) View() string {
	var content string
	if m.view == viewDetail {
		content = m.detailView()
	} else {
		content = m.listView()
	}

	s := theme.Current().Styles
	borderColor := s.BorderUnfocused
	if m.focused {
		borderColor = s.BorderFocused
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(m.width - 2).
		Height(m.height - 2)

	return style.Render(content)
}

func typeIcon(t string) string {
	switch t {
	case "view":
		return "V"
	case "materialized view":
		return "M"
	default:
		return "T"
	}
}

func (m *Model) listView() string {
	s := theme.Current().Styles
	var sb strings.Builder
	vh := m.listViewHeight()
	end := min(m.offset+vh, len(m.filtered))

	nameW := max(m.width-10, 8) // room for padding + count

	for i := m.offset; i < end; i++ {
		t := m.filtered[i]
		name := truncate(t.Name, nameW)

		line := fmt.Sprintf(" %-*s %6d", nameW, name, t.RowEstimate)
		if i == m.cursor {
			line = s.Cursor.Render(line)
		}
		sb.WriteString(line)
		if i < end-1 {
			sb.WriteByte('\n')
		}
	}

	if len(m.filtered) == 0 {
		if len(m.tables) == 0 {
			sb.WriteString("  (no tables)")
		} else {
			sb.WriteString("  (no matches)")
		}
	}

	if m.filtering {
		sb.WriteString("\n/" + m.filter)
	}

	return sb.String()
}

func (m *Model) detailView() string {
	d := m.detail
	var lines []string

	lines = append(lines, fmt.Sprintf("%s %s (%s)", typeIcon(d.Table.Type), d.Table.Name, d.Table.Type))
	lines = append(lines, "")

	if len(d.Columns) > 0 {
		lines = append(lines, "Columns:")
		for _, c := range d.Columns {
			pk := " "
			if c.IsPK {
				pk = "*"
			}
			null := "NOT NULL"
			if c.Nullable {
				null = "NULL"
			}
			def := ""
			if c.Default != "" {
				def = " = " + c.Default
			}
			lines = append(lines, fmt.Sprintf("  %s %-16s %-12s %s%s", pk, c.Name, c.TypeName, null, def))
		}
		lines = append(lines, "")
	}

	if len(d.Indexes) > 0 {
		lines = append(lines, "Indexes:")
		for _, idx := range d.Indexes {
			u := ""
			if idx.Unique {
				u = " UNIQUE"
			}
			lines = append(lines, fmt.Sprintf("  %s (%s)%s [%s]", idx.Name, strings.Join(idx.Columns, ", "), u, idx.Type))
		}
		lines = append(lines, "")
	}

	if len(d.ForeignKeys) > 0 {
		lines = append(lines, "Foreign Keys:")
		for _, fk := range d.ForeignKeys {
			ref := fk.ReferencedTable
			if fk.ReferencedSchema != "" && fk.ReferencedSchema != "public" {
				ref = fk.ReferencedSchema + "." + ref
			}
			lines = append(lines, fmt.Sprintf("  %s -> %s (%s -> %s)",
				strings.Join(fk.Columns, ", "), ref,
				strings.Join(fk.Columns, ", "), strings.Join(fk.ReferencedColumns, ", ")))
		}
		lines = append(lines, "")
	}

	if len(d.Constraints) > 0 {
		lines = append(lines, "Constraints:")
		for _, c := range d.Constraints {
			lines = append(lines, fmt.Sprintf("  %s %s (%s)", c.Type, c.Name, strings.Join(c.Columns, ", ")))
		}
	}

	// clamp detail scroll offset
	vh := max(m.height-2, 1)
	if m.detail.offset > max(len(lines)-vh, 0) {
		m.detail.offset = max(len(lines)-vh, 0)
	}

	end := min(m.detail.offset+vh, len(lines))
	visible := lines[m.detail.offset:end]
	return strings.Join(visible, "\n")
}

func truncate(s string, maxW int) string {
	if len(s) <= maxW {
		return s
	}
	if maxW <= 1 {
		return s[:maxW]
	}
	return s[:maxW-1] + "~"
}

// Focused returns whether the pane is focused.
func (m *Model) Focused() bool { return m.focused }

// SetFocused sets the focus state.
func (m *Model) SetFocused(f bool) { m.focused = f }

// SetSize sets the render dimensions.
func (m *Model) SetSize(w, h int) { m.width = w; m.height = h }

// Tables returns the current unfiltered table list (for testing).
func (m *Model) Tables() []schema.Table { return m.tables }

// Filtered returns the current filtered table list (for testing).
func (m *Model) Filtered() []schema.Table { return m.filtered }

// Cursor returns the current cursor position (for testing).
func (m *Model) Cursor() int { return m.cursor }

// IsFiltering returns whether the filter input is active (for testing).
func (m *Model) IsFiltering() bool { return m.filtering }

// InDetailView returns true when showing schema detail.
func (m *Model) InDetailView() bool { return m.view == viewDetail }
