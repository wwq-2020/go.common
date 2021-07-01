package syncx

import (
	"context"
	"reflect"
)

// SubscribeChans SubscribeChans
func SubscribeChans(ctx context.Context, callbak func(), chs ...interface{}) {
	length := len(chs)
	cases := make([]reflect.SelectCase, 0, length+1)
	for _, ch := range chs {
		value := reflect.ValueOf(ch)
		if value.Kind() != reflect.Chan {
			continue
		}
		item := reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: value,
		}
		cases = append(cases, item)
	}
	ctxItem := reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(ctx.Done()),
	}
	cases = append(cases, ctxItem)
	go func() {
		for {
			if _, _, ok := reflect.Select(cases); !ok {
				return
			}
			callbak()
		}
	}()
}
