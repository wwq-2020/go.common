package rpc_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/wwq-2020/go.common/errorsx"
	"github.com/wwq-2020/go.common/rpc"
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
	s := rpc.NewServer(":8083")

	type req struct {
		Data string
	}

	handler := func(ctx context.Context, req *req) (*resp, error) {
		return &resp{req.Data}, nil
	}

	wrapHandler := func(ctx context.Context, dec func(interface{}) error, interceptor rpc.ServerInterceptor) (interface{}, error) {
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
	interceptor := func(ctx context.Context, req interface{}, handler rpc.ServerHandler) (resp interface{}, err error) {
		ctx = context.WithValue(ctx, "aa", 1)
		return handler(ctx, req)
	}

	s.Handle("/", wrapHandler, interceptor)

	go s.Start()

	time.Sleep(time.Second)
	c := rpc.NewClient("oms-fe")
	respObj := &resp{}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*3)
	defer cancel()
	fmt.Println(c.Invoke(ctx, "/a", req{Data: "xx"}, respObj), respObj)
}
