package sqlx

import (
	"context"
	"database/sql"
	"time"

	"github.com/wwq-2020/go.common/errorsx"
)

// PreparedStmt PreparedStmt
type PreparedStmt interface {
	ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, args ...interface{}) (Rows, error)
	QueryRowContext(ctx context.Context, args ...interface{}) Row
}

type stmt struct {
	*sql.Stmt
}

func (stmt *stmt) ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error) {
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, time.Second)
		defer func() {
			cancel()
		}()
	}
	result, err := stmt.Stmt.ExecContext(ctx, args...)
	if err != nil {
		return nil, errorsx.Trace(err)
	}
	return result, nil
}

func (stmt *stmt) QueryContext(ctx context.Context, args ...interface{}) (r Rows, err error) {
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, time.Second)
		defer func() {
			if err != nil {
				cancel()
			}
		}()
	}
	stdRows, err := stmt.Stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, errorsx.Trace(err)
	}
	return &rows{Rows: stdRows, cancel: cancel}, nil
}

func (stmt *stmt) QueryRowContext(ctx context.Context, args ...interface{}) Row {
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, time.Second)
	}
	stdRow := stmt.Stmt.QueryRowContext(ctx, args...)
	return &row{Row: stdRow, cancel: cancel}
}
