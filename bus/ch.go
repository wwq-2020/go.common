package bus

import (
	"context"
	"reflect"
)

// ChanSpliter ChanSpliter
type ChanSpliter interface {
	SplitChan() map[interface{}]interface{}
}

type mapChanSpliter struct {
	m map[interface{}]interface{}
}

func (m *mapChanSpliter) SplitChan() map[interface{}]interface{} {
	return m.m
}

// MapChanSpliterFromMap MapChanSpliterFromMap
func MapChanSpliterFromMap(m map[interface{}]interface{}) ChanSpliter {
	return &mapChanSpliter{
		m: m,
	}
}

// MapChanSpliterFromObj MapChanSpliterFromObj
func MapChanSpliterFromObj(m interface{}) ChanSpliter {
	v := reflect.ValueOf(m)
	if v.Kind() != reflect.Map {
		return &mapChanSpliter{}
	}
	items := make(map[interface{}]interface{})
	iter := v.MapRange()
	for iter.Next() {
		items[iter.Key().Interface()] = iter.Value().Interface()
	}
	return &mapChanSpliter{
		m: items,
	}
}

type callbackFunc struct {
	f        reflect.Value
	hasInput bool
}

// BatchSubscribeChans BatchSubscribeChans
func BatchSubscribeChans(ctx context.Context, chanSpliter ChanSpliter) {
	items := chanSpliter.SplitChan()
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
	BatchSubscribeChans(ctx, MapChanSpliterFromMap(m))
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
