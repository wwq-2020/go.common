package httpx

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/wwq-2020/go.common/errorsx"
	"github.com/wwq-2020/go.common/httputilx"
	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/stack"
	"github.com/wwq-2020/go.common/tracing"
)

// Get Get
func Get(ctx context.Context, url string, resp interface{}, opts ...Option) error {
	return do(ctx, http.MethodGet, url, nil, resp, opts...)
}

// Post Post
func Post(ctx context.Context, url string, req, resp interface{}, opts ...Option) error {
	return do(ctx, http.MethodPost, url, req, resp, opts...)
}

// Put Put
func Put(ctx context.Context, url string, req, resp interface{}, opts ...Option) error {
	return do(ctx, http.MethodPut, url, req, resp, opts...)
}

// Delete Delete
func Delete(ctx context.Context, url string, req, resp interface{}, opts ...Option) error {
	return do(ctx, http.MethodDelete, url, req, resp, opts...)
}

func do(ctx context.Context, method, url string, req, resp interface{}, opts ...Option) (err error) {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}
	tracingOptions := options.tracingOptions
	operationName := method + "." + url
	if tracingOptions.OperationName != "" {
		operationName = tracingOptions.OperationName
	}
	stack := stack.New()

	span, ctx := tracing.StartSpan(ctx, operationName, append(tracingOptions.StartSpanOptions, tracing.Root(tracingOptions.Root))...)
	defer span.FinishWithFields(&err, stack)
	stack.Set("httpmethod", method).
		Set("url", url)

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
	httpReq = httpReq.WithContext(ctx)
	span.InjectToHTTPReq(httpReq)

	if options.reqInterceptor != nil {
		if err := options.reqInterceptor(httpReq); err != nil {
			return errorsx.TraceWithFields(err, stack)
		}
	}
	start := time.Now()
	stack.Set("invokeStart", start.Format("2006-01-02 15:04:05"))
	log.WithFields(stack).
		InfoContext(ctx, "start invoke")
	httpResp, err := options.client.Do(httpReq)
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
	if options.respInterceptor != nil {
		if err := options.respInterceptor(httpResp); err != nil {
			return errorsx.Trace(err)
		}
	}
	if resp != nil {
		if err := options.codec.Decode(respData, resp); err != nil {
			return errorsx.Trace(err)
		}
	}

	return nil
}
