package postgres

import (
	"github.com/jackc/pgx/v5"

	"github.com/mdjarv/db/internal/db"
)

func buildResult(rows pgx.Rows, tm *TypeMap) db.Result {
	fds := rows.FieldDescriptions()
	cols := make([]db.Column, len(fds))
	for i, fd := range fds {
		cols[i] = db.Column{
			Name:       fd.Name,
			TypeName:   tm.Resolve(fd.DataTypeOID),
			EnumValues: tm.EnumValues(fd.DataTypeOID),
		}
		if pgFields := tm.CompositeFields(fd.DataTypeOID); pgFields != nil {
			dbFields := make([]db.CompositeField, len(pgFields))
			for j, f := range pgFields {
				dbFields[j] = db.CompositeField{Name: f.Name, TypeName: f.TypeName}
			}
			cols[i].CompositeFields = dbFields
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
