package sqlx

import (
	"context"
	"database/sql/driver"
	"time"

	"github.com/wwq-2020/go.common/errorsx"
	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/tracing"
)

// Stmt Stmt
type Stmt interface {
	driver.Stmt
	driver.StmtExecContext
	driver.StmtQueryContext
}

type wrappedStmt struct {
	driver.Stmt
	cancel context.CancelFunc
}

// NewStmt NewStmt
func NewStmt(stmt driver.Stmt, cancel context.CancelFunc) Stmt {
	return &wrappedStmt{
		Stmt:   stmt,
		cancel: cancel,
	}
}

func (stmt *wrappedStmt) Close() error {
	if stmt.cancel != nil {
		defer stmt.cancel()
	}
	if err := stmt.Stmt.Close(); err != nil {
		return errorsx.Trace(err)
	}
	return nil
}

func (stmt *wrappedStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (_ driver.Result, err error) {
	span, ctx := tracing.StartSpan(ctx, "ExecContext")
	defer span.WithField("args", args).Finish(&err)
	log.WithField("args", args).
		InfoContext(ctx, "ExecContext")
	rstmt, ok := stmt.Stmt.(driver.StmtExecContext)
	if !ok {
		result, err := stmt.Stmt.Exec(namedValueToValue(args))
		if err != nil {
			return nil, errorsx.Trace(err)
		}
		return result, nil
	}
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, time.Second)
		defer cancel()
	}
	result, err := rstmt.ExecContext(ctx, args)
	if err != nil {
		return nil, errorsx.Trace(err)
	}
	return result, nil
}

func (stmt *wrappedStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (_ driver.Rows, err error) {
	span, ctx := tracing.StartSpan(ctx, "QueryContext")
	defer span.WithField("args", args).Finish(&err)
	log.WithField("args", args).
		InfoContext(ctx, "QueryContext")
	rstmt, ok := stmt.Stmt.(driver.StmtQueryContext)
	if !ok {
		rows, err := stmt.Stmt.Query(namedValueToValue(args))
		if err != nil {
			return nil, errorsx.Trace(err)
		}
		return rows, nil
	}
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, time.Second)
		defer func() {
			if err != nil {
				cancel()
			}
		}()
	}
	rows, err := rstmt.QueryContext(ctx, args)
	if err != nil {
		return nil, errorsx.Trace(err)
	}
	return NewRows(rows, cancel, nil), nil
}

func namedValueToValue(named []driver.NamedValue) []driver.Value {
	dargs := make([]driver.Value, len(named))
	for n, param := range named {

		dargs[n] = param.Value
	}
	return dargs
}
