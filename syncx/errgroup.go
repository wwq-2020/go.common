package syncx

import (
	"context"

	"github.com/wwq-2020/go.common/errors"
	"golang.org/x/sync/errgroup"
)

// ErrGroup ErrGroup
type ErrGroup struct {
	eg  *errgroup.Group
	ctx context.Context
	ch  chan struct{}
}

// NewErrGroup NewErrGroup
func NewErrGroup(ctx context.Context, size int) *ErrGroup {
	eg, newCtx := errgroup.WithContext(ctx)
	return &ErrGroup{
		eg:  eg,
		ctx: newCtx,
		ch:  make(chan struct{}, size),
	}
}

// Go Go
func (eg *ErrGroup) Go(task func(ctx context.Context) error) {
	select {
	case <-eg.ctx.Done():
		return
	case eg.ch <- struct{}{}:
		eg.eg.Go(func() error {
			defer func() { <-eg.ch }()
			if err := task(eg.ctx); err != nil {
				return errors.Trace(err)
			}
			return nil
		})
	}
}

// Wait Wait
func (eg *ErrGroup) Wait() error {
	if err := eg.eg.Wait(); err != nil {
		return errors.Trace(err)
	}
	return nil
}
