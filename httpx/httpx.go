package httpx

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/wwq-2020/go.common/errorsx"
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

	if options.reqInterceptor != nil {
		if err := options.reqInterceptor(httpReq); err != nil {
			return errorsx.Trace(err)
		}
	}

	httpResp, err := options.client.Do(httpReq)
	if err != nil {
		return errorsx.Trace(err)
	}

	respData, respBody, err := DrainBody(httpResp.Body)
	if err != nil {
		return errorsx.Trace(err)
	}

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
