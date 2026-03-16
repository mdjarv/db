package export

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/mdjarv/db/internal/db"
)

var numericTypes = map[string]bool{
	"int2":    true,
	"int4":    true,
	"int8":    true,
	"float4":  true,
	"float8":  true,
	"numeric": true,
}

type tableExporter struct {
	opts Options
}

func (e *tableExporter) Export(w io.Writer, result *db.Result) error {
	defer result.Rows.Close()
	cols := result.Columns
	ncols := len(cols)

	// Buffer all rows to compute column widths.
	var rows [][]string
	for result.Rows.Next() {
		vals, err := result.Rows.Values()
		if err != nil {
			return err
		}
		row := make([]string, len(vals))
		for i, v := range vals {
			row[i] = formatValue(v, e.opts.NullString)
		}
		rows = append(rows, row)
	}
	if err := result.Rows.Err(); err != nil {
		return err
	}

	// Compute column widths from headers and data.
	widths := make([]int, ncols)
	if !e.opts.NoHeader {
		for i, col := range cols {
			widths[i] = utf8.RuneCountInString(col.Name)
		}
	}
	for _, row := range rows {
		for i, cell := range row {
			n := utf8.RuneCountInString(cell)
			if n > widths[i] {
				widths[i] = n
			}
		}
	}

	// Apply max column width.
	if e.opts.MaxColWidth > 0 {
		for i := range widths {
			if widths[i] > e.opts.MaxColWidth {
				widths[i] = e.opts.MaxColWidth
			}
		}
	}

	// Determine right-alignment per column.
	rightAlign := make([]bool, ncols)
	for i, col := range cols {
		rightAlign[i] = numericTypes[col.TypeName]
	}

	// Draw.
	topBorder := boxLine("┌", "┬", "┐", widths)
	midBorder := boxLine("├", "┼", "┤", widths)
	botBorder := boxLine("└", "┴", "┘", widths)

	if _, err := fmt.Fprintln(w, topBorder); err != nil {
		return err
	}

	if !e.opts.NoHeader {
		headerCells := make([]string, ncols)
		for i, col := range cols {
			headerCells[i] = padCell(col.Name, widths[i], false)
		}
		if _, err := fmt.Fprintf(w, "│ %s │\n", strings.Join(headerCells, " │ ")); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, midBorder); err != nil {
			return err
		}
	}

	for _, row := range rows {
		cells := make([]string, ncols)
		for i, cell := range row {
			cells[i] = padCell(truncate(cell, widths[i]), widths[i], rightAlign[i])
		}
		if _, err := fmt.Fprintf(w, "│ %s │\n", strings.Join(cells, " │ ")); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(w, botBorder); err != nil {
		return err
	}

	_, err := fmt.Fprintf(w, "(%d rows)\n", len(rows))
	return err
}

func boxLine(left, mid, right string, widths []int) string {
	parts := make([]string, len(widths))
	for i, w := range widths {
		parts[i] = strings.Repeat("─", w+2)
	}
	return left + strings.Join(parts, mid) + right
}

func padCell(s string, width int, rightAlign bool) string {
	n := utf8.RuneCountInString(s)
	if n >= width {
		return s
	}
	pad := strings.Repeat(" ", width-n)
	if rightAlign {
		return pad + s
	}
	return s + pad
}

func truncate(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return s
	}
	n := utf8.RuneCountInString(s)
	if n <= maxWidth {
		return s
	}
	if maxWidth <= 3 {
		return string([]rune(s)[:maxWidth])
	}
	return string([]rune(s)[:maxWidth-3]) + "..."
}
