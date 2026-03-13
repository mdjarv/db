package export

import "github.com/mdjarv/db/internal/db"

type sliceIterator struct {
	data [][]any
	pos  int
	err  error
}

func (it *sliceIterator) Next() bool {
	if it.pos < len(it.data) {
		it.pos++
		return true
	}
	return false
}

func (it *sliceIterator) Values() ([]any, error) {
	return it.data[it.pos-1], nil
}

func (it *sliceIterator) Err() error { return it.err }
func (it *sliceIterator) Close()     {}

func mockResult(cols []db.Column, data [][]any) *db.Result {
	return &db.Result{
		Columns: cols,
		Rows:    &sliceIterator{data: data},
	}
}
