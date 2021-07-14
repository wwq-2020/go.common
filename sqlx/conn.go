package sqlx

import (
	"context"
	"database/sql"
	"time"

	"github.com/wwq-2020/go.common/errors"
)

// Conn Conn
type Conn interface {
	Stmt
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
	Close() error
	PingContext(ctx context.Context) error
}

type conn struct {
	*sql.Conn
}

func (conn *conn) PrepareContext(ctx context.Context, query string) (PreparedStmt, error) {
	stdStmt, err := conn.Conn.PrepareContext(ctx, query)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &stmt{Stmt: stdStmt}, nil
}

func (conn *conn) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	stdTx, err := conn.Conn.BeginTx(ctx, opts)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &tx{Tx: stdTx}, nil
}

func (conn *conn) Close() error {
	if err := conn.Conn.Close(); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (conn *conn) PingContext(ctx context.Context) error {
	if err := conn.Conn.PingContext(ctx); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (conn *conn) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, time.Second)
		defer func() {
			cancel()
		}()
	}
	result, err := conn.Conn.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return result, nil
}

func (conn *conn) QueryContext(ctx context.Context, query string, args ...interface{}) (r Rows, err error) {
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, time.Second)
		defer func() {
			if err != nil {
				cancel()
			}
		}()
	}
	stdRows, err := conn.Conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &rows{Rows: stdRows, cancel: cancel}, nil
}

func (conn *conn) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, time.Second)
	}
	stdRow := conn.Conn.QueryRowContext(ctx, query, args...)
	return &row{Row: stdRow, cancel: cancel}
}
