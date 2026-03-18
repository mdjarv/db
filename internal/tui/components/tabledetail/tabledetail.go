// Package tabledetail provides a modal overlay showing table structure.
package tabledetail

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/schema"
	"github.com/mdjarv/db/internal/tui/theme"
)

// Model holds the table detail overlay state.
type Model struct {
	table       schema.Table
	columns     []schema.ColumnInfo
	indexes     []schema.Index
	constraints []schema.Constraint
	foreignKeys []schema.ForeignKey
	scroll      int
	active      bool
}

// New creates an inactive table detail overlay.
func New() *Model {
	return &Model{}
}

// Open populates and shows the overlay.
func (m *Model) Open(table schema.Table, cols []schema.ColumnInfo, idxs []schema.Index, cons []schema.Constraint, fks []schema.ForeignKey) {
	m.table = table
	m.columns = cols
	m.indexes = idxs
	m.constraints = cons
	m.foreignKeys = fks
	m.scroll = 0
	m.active = true
}

// Close dismisses the overlay.
func (m *Model) Close() {
	m.active = false
	m.scroll = 0
}

// IsActive returns whether the overlay is visible.
func (m *Model) IsActive() bool { return m.active }

// Update handles key input when active.
func (m *Model) Update(msg tea.KeyMsg) tea.Cmd {
	if !m.active {
		return nil
	}
	switch msg.String() {
	case "j", "down":
		m.scroll++
	case "k", "up":
		if m.scroll > 0 {
			m.scroll--
		}
	case "g":
		m.scroll = 0
	case "G":
		m.scroll = 9999
	case "ctrl+d":
		m.scroll += 10
	case "ctrl+u":
		m.scroll -= 10
		if m.scroll < 0 {
			m.scroll = 0
		}
	default:
		m.Close()
	}
	return nil
}

// View renders the overlay centered in the given dimensions.
func (m *Model) View(containerW, containerH int) string {
	if !m.active {
		return ""
	}

	t := theme.Current()
	lines := m.buildLines(t)

	// viewport sizing
	boxW := min(containerW-4, 72)
	if boxW < 36 {
		boxW = 36
	}
	maxLines := containerH - 6
	if maxLines < 5 {
		maxLines = 5
	}

	totalLines := len(lines)
	maxScroll := totalLines - maxLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scroll > maxScroll {
		m.scroll = maxScroll
	}

	start := m.scroll
	end := start + maxLines
	if end > totalLines {
		end = totalLines
	}
	visible := lines[start:end]

	// footer with scroll hints
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	var footer string
	if totalLines > maxLines {
		arrows := ""
		if start > 0 {
			arrows += "▲ "
		}
		if end < totalLines {
			arrows += "▼ "
		}
		footer = hintStyle.Render(fmt.Sprintf("%sj/k scroll  d/Esc close  (%d/%d)", arrows, start+1, totalLines))
	} else {
		footer = hintStyle.Render("d / Esc close")
	}
	visible = append(visible, "")
	visible = append(visible, footer)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Styles.BorderFocused).
		Padding(1, 2).
		Width(boxW).
		Render(strings.Join(visible, "\n"))

	return lipgloss.Place(containerW, containerH, lipgloss.Center, lipgloss.Center, box)
}

func (m *Model) buildLines(t *theme.Theme) []string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(t.Styles.BorderFocused)
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(t.Styles.BorderFocused)
	dimStyle := t.Styles.Dim

	var lines []string

	// header: schema.name (type)  rows
	header := m.table.Name
	if m.table.Schema != "" && m.table.Schema != "public" {
		header = m.table.Schema + "." + header
	}
	rowInfo := ""
	if m.table.RowEstimate > 0 {
		rowInfo = fmt.Sprintf("  %s rows", formatCount(m.table.RowEstimate))
	}
	lines = append(lines, titleStyle.Render(fmt.Sprintf("%s (%s)", header, m.table.Type))+dimStyle.Render(rowInfo))
	lines = append(lines, "")

	// columns
	if len(m.columns) > 0 {
		lines = append(lines, sectionStyle.Render("Columns"))
		lines = append(lines, m.renderColumns(t)...)
		lines = append(lines, "")
	}

	// indexes
	if len(m.indexes) > 0 {
		lines = append(lines, sectionStyle.Render("Indexes"))
		for _, idx := range m.indexes {
			u := ""
			if idx.Unique {
				u = " UNIQUE"
			}
			line := fmt.Sprintf("  %s (%s)%s %s", idx.Name, strings.Join(idx.Columns, ", "), u, idx.Type)
			lines = append(lines, line)
		}
		lines = append(lines, "")
	}

	// foreign keys
	if len(m.foreignKeys) > 0 {
		lines = append(lines, sectionStyle.Render("Foreign Keys"))
		for _, fk := range m.foreignKeys {
			ref := fk.ReferencedTable
			if fk.ReferencedSchema != "" && fk.ReferencedSchema != "public" {
				ref = fk.ReferencedSchema + "." + ref
			}
			line := fmt.Sprintf("  %s -> %s(%s)",
				strings.Join(fk.Columns, ", "),
				ref,
				strings.Join(fk.ReferencedColumns, ", "))
			lines = append(lines, line)
		}
		lines = append(lines, "")
	}

	// constraints (skip PK/FK since they're shown elsewhere)
	filtered := filterConstraints(m.constraints)
	if len(filtered) > 0 {
		lines = append(lines, sectionStyle.Render("Constraints"))
		for _, c := range filtered {
			line := fmt.Sprintf("  %s %s (%s)", c.Type, c.Name, strings.Join(c.Columns, ", "))
			if c.Definition != "" {
				line += " " + dimStyle.Render(c.Definition)
			}
			lines = append(lines, line)
		}
		lines = append(lines, "")
	}

	return lines
}

func (m *Model) renderColumns(t *theme.Theme) []string {
	// compute column widths for alignment
	maxNameW := 0
	maxTypeW := 0
	for _, c := range m.columns {
		if len(c.Name) > maxNameW {
			maxNameW = len(c.Name)
		}
		if len(c.TypeName) > maxTypeW {
			maxTypeW = len(c.TypeName)
		}
	}

	dimStyle := t.Styles.Dim
	boldStyle := lipgloss.NewStyle().Bold(true)
	var lines []string

	for _, c := range m.columns {
		// PK indicator
		pk := " "
		if c.IsPK {
			pk = "*"
		}

		// name (bold if PK)
		name := fmt.Sprintf("%-*s", maxNameW, c.Name)
		if c.IsPK {
			name = boldStyle.Render(name)
		}

		// type with color
		typeStr := fmt.Sprintf("%-*s", maxTypeW, c.TypeName)
		typeStr = TypeStyle(t, c.TypeName).Render(typeStr)

		// nullability
		nullStr := ""
		if !c.Nullable {
			nullStr = "NOT NULL"
		}

		// default
		defStr := ""
		if c.Default != "" {
			def := c.Default
			if len(def) > 24 {
				def = def[:21] + "..."
			}
			defStr = dimStyle.Render("DEFAULT " + def)
		}

		parts := []string{"  " + pk + " " + name, typeStr}
		if nullStr != "" {
			parts = append(parts, nullStr)
		}
		if defStr != "" {
			parts = append(parts, defStr)
		}
		lines = append(lines, strings.Join(parts, "  "))
	}
	return lines
}

// TypeCategory classifies a PostgreSQL type for coloring.
type TypeCategory int

// Type categories for PostgreSQL types.
const (
	TypeCategoryOther TypeCategory = iota
	TypeCategoryNumeric
	TypeCategoryString
	TypeCategoryBoolean
	TypeCategoryDateTime
	TypeCategoryJSON
	TypeCategoryArray
)

// CategorizeType returns the category for a PostgreSQL type name.
func CategorizeType(typeName string) TypeCategory {
	lower := strings.ToLower(typeName)

	if strings.HasSuffix(lower, "[]") {
		return TypeCategoryArray
	}

	// strip size specifiers for matching: varchar(100) -> varchar
	base := lower
	if idx := strings.IndexByte(base, '('); idx != -1 {
		base = base[:idx]
	}

	switch base {
	case "serial", "bigserial", "smallserial",
		"integer", "int", "int2", "int4", "int8",
		"bigint", "smallint",
		"numeric", "decimal",
		"real", "float4",
		"double precision", "float8",
		"money", "oid":
		return TypeCategoryNumeric

	case "varchar", "character varying",
		"char", "character",
		"text", "name", "uuid",
		"citext", "bpchar":
		return TypeCategoryString

	case "boolean", "bool":
		return TypeCategoryBoolean

	case "timestamp", "timestamptz",
		"timestamp without time zone",
		"timestamp with time zone",
		"date", "time", "timetz",
		"time without time zone",
		"time with time zone",
		"interval":
		return TypeCategoryDateTime

	case "json", "jsonb":
		return TypeCategoryJSON
	}

	return TypeCategoryOther
}

// TypeStyle returns the lipgloss style for a type name.
func TypeStyle(t *theme.Theme, typeName string) lipgloss.Style {
	cat := CategorizeType(typeName)
	switch cat {
	case TypeCategoryNumeric:
		return t.Styles.DataNumber
	case TypeCategoryString:
		return t.Styles.DataString
	case TypeCategoryBoolean:
		return t.Styles.DataBoolTrue
	case TypeCategoryDateTime:
		return t.Styles.DataDate
	case TypeCategoryJSON:
		return t.Styles.Keyword
	case TypeCategoryArray:
		// base type color but dimmed
		base := strings.TrimSuffix(typeName, "[]")
		return TypeStyle(t, base).Faint(true)
	default:
		return lipgloss.NewStyle()
	}
}

func filterConstraints(constraints []schema.Constraint) []schema.Constraint {
	var out []schema.Constraint
	for _, c := range constraints {
		if c.Type == "PRIMARY KEY" || c.Type == "FOREIGN KEY" {
			continue
		}
		out = append(out, c)
	}
	return out
}

func formatCount(n int64) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}
