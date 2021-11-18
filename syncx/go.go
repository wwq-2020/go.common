package syncx

import (
	"context"
	"fmt"

	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/stack"
)

func SafeGo(f func()) {
	safeFunc := SafeFunc(f)
	go safeFunc()
}

func SafeFunc(f func()) func() {
	return func() {
		defer func() {
			if e := recover(); e != nil {
				stack := stack.Callers(stack.StdFilter)
				var err error
				switch t := e.(type) {
				case error:
					err = t
				default:
					err = fmt.Errorf("%#v", t)
				}
				log.WithField("stack", stack).
					Error(err)
			}
		}()
		f()
	}
}

func SafeLoop(ctx context.Context, f func()) {
	safeFunc := SafeFunc(f)
	for {
		safeFunc()
		select {
		case <-ctx.Done():
		default:
		}
	}
}

func SafeLoopGo(ctx context.Context, f func()) {
	safeFunc := SafeFunc(f)
	go func() {
		for {
			safeFunc()
			select {
			case <-ctx.Done():
			default:
			}
		}
	}()
}

func SafeLoopex(ctx context.Context, f func(), onStart, onStop func()) {
	safeFunc := SafeFunc(f)
	for {
		safeFunc()
		select {
		case <-ctx.Done():
		default:
		}
	}
}

func SafeLoopGoex(ctx context.Context, f func(), onStart, onStop func()) {
	safeFunc := SafeFunc(f)
	go func() {
		for {
			safeFunc()
			select {
			case <-ctx.Done():
			default:
			}
		}
	}()
}
