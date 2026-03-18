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
	dimStyle := t.Styles.Dim
	boldStyle := lipgloss.NewStyle().Bold(true)

	// abbreviate types and compute column widths
	types := make([]string, len(m.columns))
	maxNameW := 0
	maxTypeW := 0
	for i, c := range m.columns {
		types[i] = abbreviateType(c.TypeName)
		nameW := len(c.Name)
		if c.Nullable {
			nameW++ // for "?"
		}
		if nameW > maxNameW {
			maxNameW = nameW
		}
		if len(types[i]) > maxTypeW {
			maxTypeW = len(types[i])
		}
	}

	var lines []string
	for i, c := range m.columns {
		pk := " "
		if c.IsPK {
			pk = "*"
		}

		var name string
		if c.Nullable {
			padded := c.Name + dimStyle.Render("?") + strings.Repeat(" ", maxNameW-len(c.Name)-1)
			name = padded
		} else {
			name = fmt.Sprintf("%-*s", maxNameW, c.Name)
		}
		if c.IsPK {
			name = boldStyle.Render(fmt.Sprintf("%-*s", maxNameW, c.Name))
		}

		typeStr := fmt.Sprintf("%-*s", maxTypeW, types[i])
		if r := theme.ForType(c.TypeName); r != nil {
			typeStr = r.RenderType(typeStr)
		}

		parts := []string{"  " + pk + " " + name, typeStr}
		if c.Default != "" {
			parts = append(parts, dimStyle.Render("= "+c.Default))
		}
		lines = append(lines, strings.Join(parts, "  "))
	}
	return lines
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

func abbreviateType(typeName string) string {
	lower := strings.ToLower(typeName)
	switch {
	case lower == "timestamp with time zone":
		return "timestamptz"
	case lower == "timestamp without time zone":
		return "timestamp"
	case lower == "time with time zone":
		return "timetz"
	case lower == "time without time zone":
		return "time"
	case lower == "double precision":
		return "float8"
	case strings.HasPrefix(lower, "character varying"):
		return "varchar" + typeName[len("character varying"):]
	case lower == "character" || strings.HasPrefix(lower, "character("):
		return "char" + typeName[len("character"):]
	}
	return typeName
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
