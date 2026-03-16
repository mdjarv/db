package export

import (
	"encoding/csv"
	"io"

	"github.com/mdjarv/db/internal/db"
)

type csvExporter struct {
	opts Options
}

func (e *csvExporter) Export(w io.Writer, result *db.Result) error {
	defer result.Rows.Close()
	cw := csv.NewWriter(w)
	if e.opts.Delimiter != 0 {
		cw.Comma = e.opts.Delimiter
	}
	defer cw.Flush()

	if !e.opts.NoHeader {
		header := make([]string, len(result.Columns))
		for i, col := range result.Columns {
			header[i] = col.Name
		}
		if err := cw.Write(header); err != nil {
			return err
		}
	}

	for result.Rows.Next() {
		vals, err := result.Rows.Values()
		if err != nil {
			return err
		}
		record := make([]string, len(vals))
		for i, v := range vals {
			record[i] = formatValue(v, e.opts.NullString)
		}
		if err := cw.Write(record); err != nil {
			return err
		}
	}
	if err := result.Rows.Err(); err != nil {
		return err
	}

	cw.Flush()
	return cw.Error()
}
