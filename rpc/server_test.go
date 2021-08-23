package rpc_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/wwq-2020/go.common/errorsx"
	"github.com/wwq-2020/go.common/rpc"
	"github.com/wwq-2020/go.common/rpc/interceptor"
	"github.com/wwq-2020/go.common/tracing"
)

type respWrap struct {
	Code int
	Msg  string
	Data *resp
}
type resp struct {
	Data string
}

func (r *resp) Wrap() bool {
	return true
}

func (r *resp) UnWrap() bool {
	return true
}

func TestServer(t *testing.T) {
	s := rpc.NewServer("aa", ":8083")

	type req struct {
		Data string
	}
	cleanup := tracing.MustInitGlobalTracer("aa", "http://127.0.0.1:14268/api/traces?format=jaeger.thrift", 1)
	defer cleanup()

	handler := func(ctx context.Context, req *req) (*resp, error) {
		return &resp{req.Data}, nil
	}

	wrapHandler := func(ctx context.Context, dec func(interface{}) error, interceptor interceptor.ServerInterceptor) (interface{}, error) {
		in := &req{}
		if err := dec(in); err != nil {
			return nil, errorsx.Trace(err)
		}
		handler := func(ctx context.Context, reqObj interface{}) (interface{}, error) {
			resp, err := handler(ctx, reqObj.(*req))
			if err != nil {
				return nil, errorsx.Trace(err)
			}
			return resp, nil
		}
		resp, err := interceptor(ctx, in, handler)
		if err != nil {
			return nil, errorsx.Trace(err)
		}
		return resp, nil
	}
	interceptor := func(ctx context.Context, req interface{}, handler interceptor.ServerHandler) (resp interface{}, err error) {
		ctx = context.WithValue(ctx, "aa", 1)
		return handler(ctx, req)
	}

	s.Handle("/", wrapHandler, interceptor)

	go s.Start()

	time.Sleep(time.Second)
	c := rpc.NewClient("aa", "127.0.0.1:8083")
	respObj := &resp{}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*3)
	defer cancel()
	fmt.Println(c.Invoke(ctx, "/", req{Data: "xx"}, respObj), respObj)
	time.Sleep(time.Second * 2)
}
