package sqlx

import (
	"context"
	"database/sql/driver"
	"time"

	"github.com/wwq-2020/go.common/errorsx"
	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/stack"
	"github.com/wwq-2020/go.common/tracing"
)

// Conn Conn
type Conn interface {
	driver.QueryerContext
	driver.ExecerContext
	driver.ConnPrepareContext
	driver.ConnBeginTx
	driver.Conn
}

type wrappedConn struct {
	driver.Conn
}

// NewConn NewConn
func NewConn(conn driver.Conn) Conn {
	return &wrappedConn{
		Conn: conn,
	}
}

func (c *wrappedConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (_ driver.Result, err error) {
	span, ctx := tracing.StartSpan(ctx, "ExecContext")
	stack := stack.New().Set("query", query).Set("args", args)
	defer span.FinishWithFields(&err, stack)
	log.WithFields(stack).
		InfoContext(ctx, "ExecContext")
	conn := c.Conn
	queryerCtx, ok := conn.(driver.ExecerContext)
	if ok {
		var cancel context.CancelFunc
		if _, ok := ctx.Deadline(); !ok {
			ctx, cancel = context.WithTimeout(ctx, time.Second)
			defer cancel()
		}
		result, err := queryerCtx.ExecContext(ctx, query, args)
		if err != nil {
			return nil, errorsx.Trace(err)
		}
		return result, nil
	}
	queryer, ok := conn.(driver.Execer)
	if ok {
		result, err := queryer.Exec(query, namedValueToValue(args))
		if err != nil {
			return nil, errorsx.Trace(err)
		}
		return result, nil
	}
	stmt, err := c.PrepareContext(ctx, query)
	if err != nil {
		return nil, errorsx.Trace(err)
	}
	defer stmt.Close()
	result, err := stmt.Exec(namedValueToValue(args))
	if err != nil {
		return nil, errorsx.Trace(err)
	}
	return result, nil
}

func (c *wrappedConn) PrepareContext(ctx context.Context, query string) (_ driver.Stmt, err error) {
	span, ctx := tracing.StartSpan(ctx, "PrepareContext")
	defer span.WithField("query", query).Finish(&err)
	log.WithField("query", query).
		InfoContext(ctx, "PrepareContext")
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, time.Second)
		defer func() {
			if err != nil {
				cancel()
			}
		}()
	}
	conn := c.Conn
	rc, ok := conn.(driver.ConnPrepareContext)
	if !ok {
		stmt, err := conn.Prepare(query)
		if err != nil {
			return nil, errorsx.Trace(err)
		}
		return NewStmt(stmt, cancel), nil
	}
	stmt, err := rc.PrepareContext(ctx, query)
	if err != nil {
		return nil, errorsx.Trace(err)
	}
	return NewStmt(stmt, cancel), nil
}

func (c *wrappedConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (_ driver.Rows, err error) {
	span, ctx := tracing.StartSpan(ctx, "QueryContext")
	stack := stack.New().Set("query", query).Set("args", args)
	defer span.FinishWithFields(&err, stack)
	log.WithFields(stack).
		InfoContext(ctx, "QueryContext")
	conn := c.Conn
	queryerCtx, ok := conn.(driver.QueryerContext)
	if ok {
		var cancel context.CancelFunc
		if _, ok := ctx.Deadline(); !ok {
			ctx, cancel = context.WithTimeout(ctx, time.Second)
			defer func() {
				if err != nil {
					cancel()
				}
			}()
		}
		rows, err := queryerCtx.QueryContext(ctx, query, args)
		if err != nil {
			return nil, errorsx.Trace(err)
		}
		return NewRows(rows, cancel, nil), nil
	}
	queryer, ok := conn.(driver.Queryer)
	if ok {
		rows, err := queryer.Query(query, namedValueToValue(args))
		if err != nil {
			return nil, errorsx.Trace(err)
		}
		return NewRows(rows, nil, nil), nil
	}
	stmt, err := c.PrepareContext(ctx, query)
	if err != nil {
		return nil, errorsx.Trace(err)
	}
	defer stmt.Close()
	rows, err := stmt.Query(namedValueToValue(args))
	if err != nil {
		return nil, errorsx.Trace(err)
	}
	return NewRows(rows, nil, stmt), nil
}

func (c *wrappedConn) BeginTx(ctx context.Context, opts driver.TxOptions) (_ driver.Tx, err error) {
	span, ctx := tracing.StartSpan(ctx, "BeginTx")
	defer span.WithField("opts", opts).Finish(&err)
	log.WithField("opts", opts).
		InfoContext(ctx, "BeginTx")
	rc, ok := c.Conn.(driver.ConnBeginTx)
	if !ok {
		tx, err := c.Begin()
		if err != nil {
			return nil, errorsx.Trace(err)
		}
		return NewTx(ctx, tx, nil), nil
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
	tx, err := rc.BeginTx(ctx, opts)
	if err != nil {
		return nil, errorsx.Trace(err)
	}
	return NewTx(ctx, tx, cancel), nil
}
