package httpx

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/wwq-2020/go.common/errorsx"
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
	operationName := method + "." + url
	tracingOpts := StartSpanOptionsFromContext(ctx)
	span, ctx := tracing.StartSpanFromContext(ctx, operationName, tracingOpts...)
	defer span.Finish(&err)

	stack := stack.New().
		Set("httpmethod", method).
		Set("url", url)
	span.WithFields(stack)
	var reqBody io.Reader
	var reqData []byte
	if req != nil {
		var err error
		reqData, err = options.codec.Encode(req)
		if err != nil {
			return errorsx.TraceWithFields(err, stack)
		}
		reqDataStr := string(reqData)
		stack.Set("reqData", reqDataStr)
		span.WithField("reqData", reqDataStr)
		reqBody = bytes.NewReader(reqData)
	}
	httpReq, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return errorsx.TraceWithFields(err, stack)
	}
	ctx = log.ContextEnsureTraceID(ctx)
	httpReq = httpReq.WithContext(ctx)
	if options.reqInterceptor != nil {
		if err := options.reqInterceptor(httpReq); err != nil {
			return errorsx.TraceWithFields(err, stack)
		}
	}
	start := time.Now()
	log.WithFields(stack).
		InfoContext(ctx, "start invoke")

	httpResp, err := options.client.Do(httpReq)
	if err != nil {
		return errorsx.TraceWithFields(err, stack)
	}

	respData, respBody, err := DrainBody(httpResp.Body)
	if err != nil {
		return errorsx.TraceWithFields(err, stack)
	}
	elapsed := time.Now().Sub(start).Milliseconds()
	respDataStr := string(respData)
	stack.Set("respData", respDataStr)
	span.WithField("respData", respDataStr).
		WithField("elapsed", elapsed)
	log.WithField("respData", respDataStr).
		WithField("elapsed", elapsed).
		InfoContext(ctx, "invoke finish")
	httpResp.Body = respBody

	if options.respInterceptor != nil {
		if err := options.respInterceptor(httpResp); err != nil {
			return errorsx.TraceWithFields(err, stack)
		}
	}
	if resp != nil {
		if err := options.codec.Decode(respData, resp); err != nil {
			return errorsx.TraceWithFields(err, stack)
		}
	}

	return nil
}
