package sqlx

import (
	"context"
	"database/sql"

	"github.com/wwq-2020/go.common/errors"
)

// Rows Rows
type Rows interface {
	Close() error
	Columns() ([]string, error)
	ColumnTypes() ([]*sql.ColumnType, error)
	Err() error
	Next() bool
	NextResultSet() bool
	Scan(dest ...interface{}) error
}

type rows struct {
	*sql.Rows
	cancel context.CancelFunc
}

func (r *rows) Close() error {
	if r.cancel != nil {
		defer r.cancel()
	}
	if err := r.Rows.Close(); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (r *rows) Columns() ([]string, error) {
	columns, err := r.Rows.Columns()
	if err != nil {
		return nil, errors.Trace(err)
	}
	return columns, nil
}

func (r *rows) NextResultSet() bool {
	ok := r.Rows.NextResultSet()
	if !ok && r.cancel != nil {
		r.cancel()
	}
	return ok
}

func (r *rows) ColumnTypes() ([]*sql.ColumnType, error) {
	columnTypes, err := r.Rows.ColumnTypes()
	if err != nil {
		return nil, errors.Trace(err)
	}
	return columnTypes, nil
}

func (r *rows) Err() error {
	if err := r.Rows.Err(); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (r *rows) Next() bool {
	ok := r.Rows.Next()
	if !ok && r.cancel != nil {
		r.cancel()
	}
	return ok
}

func (r *rows) Scan(dest ...interface{}) error {
	if err := r.Rows.Scan(dest...); err != nil {
		return errors.Trace(err)
	}
	return nil
}
