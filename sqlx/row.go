package sqlx

import (
	"context"
	"database/sql"

	"github.com/wwq-2020/go.common/errors"
)

// Row 对应 *sql.Row.
type Row interface {
	Scan(dest ...interface{}) error
}

type row struct {
	*sql.Row
	cancel context.CancelFunc
}

func (r *row) Scan(dest ...interface{}) error {
	if r.cancel != nil {
		defer r.cancel()
	}
	if err := r.Row.Scan(); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (r *row) Err() error {
	if err := r.Row.Err(); err != nil {
		return errors.Trace(err)
	}
	return nil
}
