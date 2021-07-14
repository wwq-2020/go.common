package funcx

import (
	"context"
	"reflect"
)

// Call Call
func Call(fn interface{}, args ...interface{}) {
	fnValue := reflect.ValueOf(fn)
	fnType := reflect.TypeOf(fn)
	if fnValue.Kind() != reflect.Func {
		return
	}
	numIn := fnType.NumIn()
	if numIn == 0 {
		fnType.NumOut()
		fnValue.Call(nil)
		return
	}
	if len(args) < numIn {
		return
	}
	values := make([]reflect.Value, 0, len(args))
	for i, arg := range args {
		if i >= numIn {
			break
		}
		argType := reflect.ValueOf(arg)
		if !argType.Type().AssignableTo(fnType.In(i)) {
			return
		}

		values = append(values, reflect.ValueOf(arg))
	}
	fnValue.Call(values)
}

// ArgsSpliter ArgsSpliter
type ArgsSpliter interface {
	SplitArgs() [][]interface{}
}

// sliceArgsSpliter sliceArgsSpliter
type sliceArgsSpliter struct {
	parts [][]interface{}
}

// SplitArgs SplitArgs
func (s *sliceArgsSpliter) SplitArgs() [][]interface{} {
	return s.parts
}

// SliceArgsSpliterFromObj SliceArgsSpliterFromObj
func SliceArgsSpliterFromObj(size int, args interface{}) ArgsSpliter {
	t := reflect.TypeOf(args)
	if t.Kind() != reflect.Slice {
		return &sliceArgsSpliter{}
	}
	v := reflect.ValueOf(args)
	var parts [][]interface{}
	part := make([]interface{}, 0, size)
	length := v.Len()
	for i := 0; i < length; i++ {
		if len(part) >= size {
			parts = append(parts, part)
			part = make([]interface{}, 0, size)
		}
		part = append(part, v.Index(i).Interface())
		if i == length-1 {
			parts = append(parts, part)
		}
	}
	return &sliceArgsSpliter{
		parts: parts,
	}
}

// SliceArgsSpliterFromSlice SliceArgsSpliterFromSlice
func SliceArgsSpliterFromSlice(size int, args ...interface{}) ArgsSpliter {
	var parts [][]interface{}
	part := make([]interface{}, 0, size)
	length := len(args)
	for i := 0; i < length; i++ {
		if len(part) >= size {
			parts = append(parts, part)
			part = make([]interface{}, 0, size)
		}
		part = append(part, args[i])
		if i == length-1 {
			parts = append(parts, part)
		}
	}
	return &sliceArgsSpliter{
		parts: parts,
	}
}

// BatchCall BatchCall
func BatchCall(fn interface{}, argsSpliter ArgsSpliter) {
	args := argsSpliter.SplitArgs()
	for _, each := range args {
		Call(fn, each...)
	}
}

// BatchCallContext BatchCallContext
func BatchCallContext(ctx context.Context, fn interface{}, argsSpliter ArgsSpliter) {
	args := argsSpliter.SplitArgs()
	for _, each := range args {
		Call(fn, append([]interface{}{ctx}, each...)...)
	}
}
