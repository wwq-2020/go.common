package syncx

import (
	"context"
	"reflect"
)

// SubscribeChans SubscribeChans
func SubscribeChans(ctx context.Context, callbak func(), chs ...interface{}) {
	cases := make([]reflect.SelectCase, 0, len(chs))
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

// AggregateChan AggregateChan
func AggregateChan(chs ...interface{}) <-chan struct{} {
	ret := make(chan struct{})
	cases := make([]reflect.SelectCase, 0, len(chs))
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
	go func() {
		for {
			if _, _, ok := reflect.Select(cases); !ok {
				return
			}
			ret <- struct{}{}
		}
	}()
	return ret
}
