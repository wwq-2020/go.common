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
	"github.com/wwq-2020/go.common/httputilx"
	"github.com/wwq-2020/go.common/httpx"
	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/stack"
	"github.com/wwq-2020/go.common/tracing"
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
	tracingOptions := options.tracingOptions
	stack := stack.New()
	span, ctx := tracing.StartSpan(ctx, c.name+"-invoke", append(tracingOptions.StartSpanOptions, tracing.Root(tracingOptions.Root))...)
	defer span.FinishWithFields(&err, stack)
	stack.Set("httpmethod", method)
	endpoint, err := c.options.balancer.Pick()
	if err != nil {
		return errorsx.Trace(err)
	}
	url := fmt.Sprintf("http://%s%s", endpoint, path)
	stack.Set("url", url)
	var reqBody io.Reader
	if req != nil {
		reqData, err := options.codec.Encode(req)
		if err != nil {
			return errorsx.TraceWithFields(err, stack)
		}
		reqDataStr := string(reqData)
		stack.Set("reqData", reqDataStr)
		reqBody = bytes.NewReader(reqData)
	}
	httpReq, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return errorsx.TraceWithFields(err, stack)
	}
	givenMetadata := OutgoingMetadataFromContext(ctx)
	metadata := options.metadata.Merge(givenMetadata)
	for k, vs := range metadata {
		for _, v := range vs {
			httpReq.Header.Add(k, v)
		}
	}

	httpReq = httpReq.WithContext(ctx)
	span.InjectToHTTPReq(httpReq)

	start := time.Now()
	stack.Set("invokeStart", start.Format("2006-01-02 15:04:05"))
	log.WithFields(stack).
		InfoContext(ctx, "start invoke")
	httpResp, err := httpx.DefaultClient().Do(httpReq)
	if err != nil {
		return errorsx.Trace(err)
	}

	respData, respBody, err := httputilx.DrainBody(httpResp.Body)
	if err != nil {
		return errorsx.Trace(err)
	}
	end := time.Now()
	elapsed := end.Sub(start).Milliseconds()
	respDataStr := string(respData)
	stack.Set("respData", respDataStr).
		Set("elapsed", elapsed).
		Set("invokeFinish", end.Format("2006-01-02 15:04:05"))
	log.WithField("respData", respDataStr).
		WithField("elapsed", elapsed).
		WithField("invokeFinish", end.Format("2006-01-02 15:04:05")).
		InfoContext(ctx, "invoke finish")
	httpResp.Body = respBody
	needUnWrap := isRespNeedUnWrap(resp)
	if needUnWrap {
		gotResp := &respObj{
			Data: resp,
		}
		if err := options.codec.Decode(respData, gotResp); err != nil {
			return errorsx.Trace(err)
		}
		if gotResp.Code != options.expectedCode {
			stack.Set("expectedcode", options.expectedCode).Set("gotcode", gotResp.Code)
			return errorsx.NewWithFields("got unexpected code", stack)
		}
		return nil
	}
	if err := options.codec.Decode(respData, resp); err != nil {
		return errorsx.Trace(err)
	}
	return nil
}
