package sqlx

import (
	"context"
	"database/sql/driver"

	"github.com/wwq-2020/go.common/errorsx"
)

// Rows Rows
type Rows interface {
	driver.Rows
}

type wrappedRows struct {
	driver.Rows
	cancel context.CancelFunc
	stmt   driver.Stmt
}

// NewRows NewRows
func NewRows(rows driver.Rows, cancel context.CancelFunc, stmt driver.Stmt) Rows {
	return &wrappedRows{
		Rows:   rows,
		cancel: cancel,
		stmt:   stmt,
	}
}

func (r *wrappedRows) Close() error {
	if r.cancel != nil {
		defer r.cancel()
	}
	if r.stmt != nil {
		defer r.stmt.Close()
	}
	if err := r.Rows.Close(); err != nil {
		return errorsx.Trace(err)
	}
	return nil
}
