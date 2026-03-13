package postgres

import (
	"github.com/jackc/pgx/v5"

	"github.com/mdjarv/db/internal/db"
)

func buildResult(rows pgx.Rows) db.Result {
	fds := rows.FieldDescriptions()
	cols := make([]db.Column, len(fds))
	for i, fd := range fds {
		cols[i] = db.Column{
			Name:     fd.Name,
			TypeName: oidToTypeName(fd.DataTypeOID),
			TypeOID:  fd.DataTypeOID,
		}
	}
	return db.Result{
		Columns: cols,
		Rows:    &rowIterator{rows: rows},
	}
}

type rowIterator struct {
	rows pgx.Rows
}

func (r *rowIterator) Next() bool             { return r.rows.Next() }
func (r *rowIterator) Values() ([]any, error) { return r.rows.Values() }
func (r *rowIterator) Err() error             { return r.rows.Err() }
func (r *rowIterator) Close()                 { r.rows.Close() }
