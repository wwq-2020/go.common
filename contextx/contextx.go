package contextx

import (
	"context"
	"time"

	"github.com/wwq-2020/go.common/bus"
)

type multi struct {
	ctxs []context.Context
	ch   <-chan struct{}
}

// Multi Multi
func Multi(ctxs ...context.Context) context.Context {
	chs := make([]<-chan struct{}, 0, len(ctxs))
	for _, ctx := range ctxs {
		chs = append(chs, ctx.Done())
	}
	ch := bus.AggregateChan(chs)
	return &multi{
		ch:   ch,
		ctxs: ctxs,
	}
}

func (m *multi) Deadline() (deadline time.Time, ok bool) {
	var d time.Time
	for _, ctx := range m.ctxs {
		if cur, ok := ctx.Deadline(); ok && cur.Sub(d) < 0 {
			d = cur
		}
	}
	return d, !d.IsZero()
}

func (m *multi) Done() <-chan struct{} {
	return m.ch
}

func (m *multi) Err() error {
	for _, ctx := range m.ctxs {
		err := ctx.Err()
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *multi) Value(key interface{}) interface{} {
	for _, ctx := range m.ctxs {
		value := ctx.Value(key)
		if value != nil {
			return value
		}
	}
	return nil
}
