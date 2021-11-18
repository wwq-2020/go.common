package bus

import (
	"context"
	"reflect"

	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/syncx"
)

type callbackFunc struct {
	f        reflect.Value
	hasInput bool
	hasCtx   bool
}

// BatchSubscribeChansContext BatchSubscribeChansContext
func BatchSubscribeChansContext(ctx context.Context, chMap map[interface{}]interface{}) {
	ctxValue := reflect.ValueOf(ctx)
	ctxType := ctxValue.Type()
	cases := make([]reflect.SelectCase, 0, len(chMap)+1)
	funcs := make([]*callbackFunc, 0, len(chMap)+1)
	for k, v := range chMap {
		kValue := reflect.ValueOf(k)
		kKind := kValue.Kind()
		if kKind != reflect.Chan {
			log.WithField("kind", kKind).
				WithField("ch", k).
				Errorf("unexpected ch type")
			continue
		}
		vValue := reflect.ValueOf(v)
		vKind := vValue.Kind()
		if vKind != reflect.Func {
			log.WithField("kind", vKind).
				WithField("callback", v).
				Errorf("unexpected callback type")
			continue
		}
		vType := reflect.TypeOf(v)
		numin := vType.NumIn()
		if numin > 2 {
			log.WithField("numin", numin).
				WithField("ch", k).
				Errorf("unexpected numin")
			continue
		}
		item := reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: kValue,
		}
		if numin == 2 {
			wantType1 := vType.In(0)
			wantType2 := vType.In(1)
			if !ctxType.AssignableTo(wantType1) {
				log.WithField("wantType", wantType1).
					WithField("givenType", ctxType).
					WithField("callback", v).
					Errorf("unexpected callback type")
				continue
			}
			givenType2 := kValue.Type().Elem()
			if !givenType2.AssignableTo(wantType2) {
				log.WithField("wantType", wantType2).
					WithField("givenType", givenType2).
					WithField("callback", v).
					Errorf("unexpected callback type")
				continue
			}
			cases = append(cases, item)
			funcs = append(funcs, &callbackFunc{
				f:        vValue,
				hasInput: vType.NumIn() > 0,
				hasCtx:   true,
			})
			continue
		}

		wantType := vType.In(0)
		givenType := kValue.Type().Elem()
		if vType.NumIn() > 0 && !givenType.AssignableTo(wantType) {
			log.WithField("wantType", wantType).
				WithField("callback", v).
				WithField("givenType", givenType).
				Errorf("unexpected callback type")
			continue
		}
		cases = append(cases, item)
		funcs = append(funcs, &callbackFunc{
			f:        vValue,
			hasInput: vType.NumIn() > 0,
		})
	}
	if len(cases) == 0 {
		log.Warn("no cases")
	}
	ctxCh := ctx.Done()
	ctxChValue := reflect.ValueOf(ctxCh)
	item := reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: ctxChValue,
	}
	ctxFunc := func() {
	}
	ctxFuncValue := reflect.ValueOf(ctxFunc)
	cases = append(cases, item)
	funcs = append(funcs, &callbackFunc{
		f:        ctxFuncValue,
		hasInput: false,
	})
	newCtx, cancel := context.WithCancel(ctx)
	syncx.SafeLoopGo(newCtx, func() {
		for {
			chosen, recv, ok := reflect.Select(cases)
			if !ok {
				cancel()
				log.WithField("chosen", chosen).
					Warn("ch closed")
				return
			}
			callbackFunc := funcs[chosen]
			if callbackFunc.hasInput {
				if callbackFunc.hasCtx {
					callbackFunc.f.Call([]reflect.Value{ctxValue, recv})
					continue
				}
				callbackFunc.f.Call([]reflect.Value{recv})
				continue
			}
			callbackFunc.f.Call(nil)
		}
	})

}

// BatchSubscribeChans BatchSubscribeChans
func BatchSubscribeChans(chMap map[interface{}]interface{}) {
	BatchSubscribeChansContext(context.TODO(), chMap)
}

// SubscribeChans SubscribeChans
func SubscribeChans(callbak interface{}, chs ...interface{}) {
	SubscribeChansContext(context.TODO(), callbak, chs...)
}

// SubscribeChansContext SubscribeChansContext
func SubscribeChansContext(ctx context.Context, callbak interface{}, chs ...interface{}) {
	chMap := make(map[interface{}]interface{}, len(chs))
	for _, ch := range chs {
		chMap[ch] = callbak
	}
	BatchSubscribeChansContext(ctx, chMap)
}

// AggregateChan AggregateChan
func AggregateChan(chs ...interface{}) <-chan struct{} {
	return AggregateChanContext(context.TODO(), chs...)
}

// AggregateChanContext AggregateChanContext
func AggregateChanContext(ctx context.Context, chs ...interface{}) <-chan struct{} {
	ret := make(chan struct{})
	cases := make([]reflect.SelectCase, 0, len(chs))
	for idx, ch := range chs {
		value := reflect.ValueOf(ch)
		kind := value.Kind()
		if kind != reflect.Chan {
			log.WithField("idx", idx).
				WithField("kind", kind).
				Errorf("unexpected ch arg kind")
			continue
		}
		item := reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: value,
		}
		cases = append(cases, item)
	}
	if len(cases) == 0 {
		log.Warn("no cases")
	}
	ctxCh := ctx.Done()
	ctxChValue := reflect.ValueOf(ctxCh)
	item := reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: ctxChValue,
	}
	cases = append(cases, item)
	syncx.SafeGo(func() {
		for {
			if chosen, _, ok := reflect.Select(cases); !ok {
				log.WithField("chosen", chosen).
					Warn("ch closed")
				close(ret)
				return
			}
			ret <- struct{}{}
		}
	})

	return ret
}

// BroadcastChan BroadcastChan
func BroadcastChan(revCh interface{}, sendChs ...interface{}) {
	revChValue := reflect.ValueOf(revCh)
	revChkind := revChValue.Kind()
	if revChkind != reflect.Chan {
		log.WithField("kind", revChkind).
			Errorf("unexpected revCh type")
		return
	}

	values := make([]reflect.Value, 0, len(sendChs))
	for idx, ch := range sendChs {
		value := reflect.ValueOf(ch)
		kind := value.Kind()
		if kind != reflect.Chan {
			log.WithField("idx", idx).
				WithField("kind", kind).
				Errorf("unexpected ch arg kind")
			continue
		}
		values = append(values, value)
	}
	syncx.SafeGo(func() {
		for {
			_, ok := revChValue.Recv()
			if !ok {
				log.Warn("ch closed")
				return
			}
			for _, value := range values {
				v := reflect.New(value.Type().Elem()).Elem()
				value.Send(v)
			}
		}
	})
}
