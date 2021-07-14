package sqlx

import (
	"context"
	"database/sql"
	"time"

	"github.com/wwq-2020/go.common/errors"
)

// Tx Tx
type Tx interface {
	Stmt
	Commit() error
	Rollback() error
}

type tx struct {
	*sql.Tx
}

func (tx *tx) Commit() error {
	if err := tx.Tx.Commit(); err != nil {
		return errors.Trace(err)
	}
	return nil

}

func (tx *tx) PrepareContext(ctx context.Context, query string) (PreparedStmt, error) {
	stdStmt, err := tx.Tx.PrepareContext(ctx, query)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &stmt{Stmt: stdStmt}, nil
}

func (tx *tx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, time.Second)
		defer func() {
			cancel()
		}()
	}
	result, err := tx.Tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return result, nil
}

func (tx *tx) QueryContext(ctx context.Context, query string, args ...interface{}) (r Rows, err error) {
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, time.Second)
		defer func() {
			if err != nil {
				cancel()
			}
		}()
	}
	stdRows, err := tx.Tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &rows{Rows: stdRows, cancel: cancel}, nil
}

func (tx *tx) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, time.Second)
	}
	stdRow := tx.Tx.QueryRowContext(ctx, query, args...)
	return &row{Row: stdRow, cancel: cancel}
}
