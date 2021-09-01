package rpc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/wwq-2020/go.common/errorsx"
	"github.com/wwq-2020/go.common/httpx"
	"github.com/wwq-2020/go.common/stack"
)

// Client Client
type Client interface {
	Invoke(ctx context.Context, path string, in, out interface{}, opts ...InvokeOption) (err error)
}

type client struct {
	name    string
	options ClientOptions
}

// NewClient NewClient
func NewClient(name, addr string, opts ...ClientOption) Client {
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
		name:    name,
		options: options,
	}
}

// Invoke Invoke
func (c *client) Invoke(ctx context.Context, path string, req, resp interface{}, opts ...InvokeOption) (err error) {
	options := defaultInvokeOptions.Clone()
	for _, opt := range opts {
		opt(&options)
	}

	method := http.MethodPost

	endpoint, err := c.options.balancer.Pick()
	if err != nil {
		return errorsx.Trace(err)
	}
	url := fmt.Sprintf("http://%s%s", endpoint, path)
	var reqBody io.Reader
	if req != nil {
		reqData, err := options.codec.Encode(req)
		if err != nil {
			return errorsx.Trace(err)
		}
		reqBody = bytes.NewReader(reqData)
	}

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Second)
		defer cancel()
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return errorsx.Trace(err)
	}
	givenMetadata := OutgoingMetadataFromContext(ctx)
	metadata := options.metadata.Merge(givenMetadata)
	for k, vs := range metadata {
		for _, v := range vs {
			httpReq.Header.Add(k, v)
		}
	}

	httpReq = httpReq.WithContext(ctx)

	httpResp, err := httpx.DefaultClient().Do(httpReq)
	if err != nil {
		return errorsx.Trace(err)
	}

	respData, respBody, err := httpx.DrainBody(httpResp.Body)
	if err != nil {
		return errorsx.Trace(err)
	}

	httpResp.Body = respBody
	needUnWrap := isRespNeedWrap(resp)
	if needUnWrap {
		gotResp := &respObj{
			Data: resp,
		}
		if err := options.codec.Decode(respData, gotResp); err != nil {
			return errorsx.Trace(err)
		}
		if gotResp.Code != options.expectedCode {
			stack := stack.New().
				Set("expectedcode", options.expectedCode).
				Set("gotcode", gotResp.Code)
			return errorsx.NewWithFields("got unexpected code", stack)
		}
		return nil
	}
	if err := options.codec.Decode(respData, resp); err != nil {
		return errorsx.Trace(err)
	}
	return nil
}
