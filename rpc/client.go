package rpc

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/wwq-2020/go.common/errorsx"
	"github.com/wwq-2020/go.common/httpx"
	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/stack"
)

// Client Client
type Client interface {
	Invoke(ctx context.Context, path string, in, out interface{}, opts ...InvokeOption) (err error)
}

type client struct {
	options ClientOptions
}

// NewClient NewClient
func NewClient(addr string, opts ...ClientOption) Client {
	options := defaultClientOptions
	for _, opt := range opts {
		opt(&options)
	}
	target := os.Getenv("TARGET")
	if target != "" {
		addr = target
	}
	resolver := options.resolverFactory(addr)
	resolver.OnAdd(options.balancer.Add)
	resolver.OnDel(options.balancer.Del)
	resolver.Start()
	return &client{
		options: options,
	}
}

// Invoke Invoke
func (c *client) Invoke(ctx context.Context, path string, req, resp interface{}, opts ...InvokeOption) error {
	endpoint, err := c.options.balancer.Pick()
	if err != nil {
		return errorsx.Trace(err)
	}
	traceID := log.GenTraceID()
	ctx = log.ContextWithTraceID(ctx, traceID)
	metadata := NewMetadata()
	givenMetadata := OutgoingMetadataFromContext(ctx)
	metadata.Add("traceID", traceID)
	metadata = metadata.Merge(givenMetadata)
	options := &InvokeOptions{
		metadata:     metadata,
		expectedCode: 0,
	}
	for _, opt := range opts {
		opt(options)
	}
	reqOption := httpx.WithReqInterceptors(func(req *http.Request) error {
		for k, vs := range options.metadata {
			for _, v := range vs {
				req.Header.Add(k, v)
			}
		}
		return nil
	}, httpx.ContentTypeReqInterceptor(httpx.ContentTypeJSON))
	codecOption := httpx.WithCodec(c.options.codec)
	needUnWrap := isRespNeedUnWrap(resp)
	url := fmt.Sprintf("http://%s%s", endpoint, path)
	if needUnWrap {
		gotResp := &respObj{
			Data: resp,
		}
		if err := httpx.Post(ctx, url, req, gotResp, reqOption, codecOption); err != nil {
			return errorsx.Trace(err)
		}
		if gotResp.Code != options.expectedCode {
			stack := stack.New().Set("expectedcode", options.expectedCode).Set("gotcode", gotResp.Code)
			return errorsx.NewWithFields("got unexpected code", stack)
		}
		return nil
	}
	if err := httpx.Post(ctx, url, req, resp, reqOption, codecOption); err != nil {
		return errorsx.Trace(err)
	}
	return nil
}
