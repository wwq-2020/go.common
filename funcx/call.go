package funcx

import (
	"context"
	"reflect"

	"github.com/wwq-2020/go.common/log"
)

// Call Call
func Call(fn interface{}, args ...interface{}) {
	fnValue := reflect.ValueOf(fn)
	fnType := reflect.TypeOf(fn)
	fnKind := fnValue.Kind()
	if fnKind != reflect.Func {
		log.WithField("kind", fnKind).
			Errorf("unexpected fn type")
		return
	}
	numIn := fnType.NumIn()
	if len(args) != numIn {
		log.WithField("expected", numIn).
			WithField("got", len(args)).
			Errorf("unexpected arg num")
		return
	}
	values := make([]reflect.Value, 0, len(args))
	for i, arg := range args {
		argValue := reflect.ValueOf(arg)
		gotType := argValue.Type()
		expectedType := fnType.In(i)
		if !gotType.AssignableTo(expectedType) {
			log.WithField("expected", expectedType).
				WithField("got", gotType).
				Errorf("unexpected arg type")
			return
		}
		values = append(values, reflect.ValueOf(arg))
	}
	fnValue.Call(values)
}

// BatchCall BatchCall
func BatchCall(fn interface{}, args [][]interface{}) {
	for _, each := range args {
		Call(fn, each...)
	}
}

// BatchCallContext BatchCallContext
func BatchCallContext(ctx context.Context, fn interface{}, args [][]interface{}) {
	for _, each := range args {
		Call(fn, append([]interface{}{ctx}, each...)...)
	}
}
