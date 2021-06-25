package syncx

import (
	"context"
	"sync"
	"sync/atomic"
)

// Once Once
type Once struct {
	done uint32
	m    sync.Mutex
	err  error
}

// Do Do
func (o *Once) Do(ctx context.Context, f func(ctx context.Context) error) error {
	if atomic.LoadUint32(&o.done) == 0 {
		o.doSlow(ctx, f)
	}
	return o.err
}

func (o *Once) doSlow(ctx context.Context, f func(ctx context.Context) error) {
	o.m.Lock()
	defer o.m.Unlock()
	if o.done == 0 {
		defer atomic.StoreUint32(&o.done, 1)
		o.err = f(ctx)
	}
}
