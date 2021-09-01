package sqlx

import (
	"context"
	"database/sql/driver"

	"github.com/wwq-2020/go.common/errorsx"
	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/tracing"
)

// Tx Tx
type Tx interface {
	Commit() error
	Rollback() error
}

type wrappedTx struct {
	tx     driver.Tx
	ctx    context.Context
	cancel context.CancelFunc
}

// NewTx NewTx
func NewTx(ctx context.Context, tx driver.Tx, cancel context.CancelFunc) driver.Tx {
	return &wrappedTx{
		tx:     tx,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (tx *wrappedTx) Commit() (err error) {
	if tx.cancel != nil {
		defer tx.cancel()
	}
	span, ctx := tracing.StartSpan(tx.ctx, "Commit")
	defer span.Finish(&err)
	log.InfoContext(ctx, "Commit")
	if err = tx.tx.Commit(); err != nil {
		return errorsx.Trace(err)
	}
	return nil
}

func (tx *wrappedTx) Rollback() (err error) {
	if tx.cancel != nil {
		defer tx.cancel()
	}
	span, ctx := tracing.StartSpan(tx.ctx, "Rollback")
	defer span.Finish(&err)
	log.InfoContext(ctx, "Rollback")
	if err = tx.tx.Rollback(); err != nil {
		return errorsx.Trace(err)
	}
	return nil
}
