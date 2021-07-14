package syncx

import (
	"context"
	"reflect"
)

type callbackFunc struct {
	f        reflect.Value
	hasInput bool
}

// BatchSubscribeChans BatchSubscribeChans
func BatchSubscribeChans(ctx context.Context, items map[interface{}]interface{}) {
	cases := make([]reflect.SelectCase, 0, len(items)+1)
	funcs := make([]*callbackFunc, 0, len(items)+1)
	for k, v := range items {
		kValue := reflect.ValueOf(k)
		if kValue.Kind() != reflect.Chan {
			continue
		}
		vValue := reflect.ValueOf(v)
		if vValue.Kind() != reflect.Func {
			continue
		}
		vType := reflect.TypeOf(v)
		if vType.NumIn() >= 2 {
			continue
		}

		item := reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: kValue,
		}
		if vType.NumIn() > 0 && kValue.Type().Elem() != vType.In(0) {
			continue
		}
		cases = append(cases, item)
		funcs = append(funcs, &callbackFunc{
			f:        vValue,
			hasInput: vType.NumIn() > 0,
		})
	}
	ctxItem := reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(ctx.Done()),
	}
	cases = append(cases, ctxItem)
	go func() {
		for {
			chosen, recv, ok := reflect.Select(cases)
			if !ok {
				return
			}
			callbackFunc := funcs[chosen]
			if callbackFunc.hasInput {
				callbackFunc.f.Call([]reflect.Value{recv})
				continue
			}
			callbackFunc.f.Call(nil)
		}
	}()
}

// SubscribeChans SubscribeChans
func SubscribeChans(ctx context.Context, callbak interface{}, chs ...interface{}) {
	m := make(map[interface{}]interface{}, len(chs))
	for _, ch := range chs {
		m[ch] = callbak
	}
	BatchSubscribeChans(ctx, m)
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
				close(ret)
				return
			}
			ret <- struct{}{}
		}
	}()
	return ret
}
