package httpx

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/wwq-2020/go.common/errors"
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

func do(ctx context.Context, method, url string, req, resp interface{}, opts ...Option) error {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}

	var reqBody io.Reader
	var reqData []byte
	if req != nil {
		var err error
		reqData, err = options.codec.Encode(req)
		if err != nil {
			return errors.Trace(err)
		}
		reqBody = bytes.NewReader(reqData)
	}
	httpReq, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return errors.Trace(err)
	}
	httpReq = httpReq.WithContext(ctx)
	if options.reqInterceptor != nil {
		if err := options.reqInterceptor(httpReq); err != nil {
			return errors.Trace(err)
		}
	}

	httpResp, err := options.client.Do(httpReq)
	if err != nil {
		return errors.Trace(err)
	}

	respData, respBody, err := drainBody(httpResp.Body)
	if err != nil {
		return errors.Trace(err)
	}
	defer httpResp.Body.Close()

	httpResp.Body = respBody

	if options.respInterceptor != nil {
		if err := options.respInterceptor(httpResp); err != nil {
			return errors.Trace(err)
		}
	}
	if resp != nil {
		if err := options.codec.Decode(respData, resp); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

func drainBody(src io.ReadCloser) ([]byte, io.ReadCloser, error) {
	defer src.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(src); err != nil {
		return nil, nil, errors.Trace(err)
	}
	return buf.Bytes(), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

// Client Client
func Client() *http.Client {
	return &http.Client{
		Transport: Transport(),
	}
}

// Transport Transport
func Transport() http.RoundTripper {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

type retriableTransport struct {
	maxRetry int
	rt       http.RoundTripper
	options  *RetriableTransportOptions
}

func (rt *retriableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var saved []byte
	var err error
	if rt.maxRetry > 0 && req.Body != nil {
		saved, err = ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, errors.Trace(err)
		}
	}
	var resp *http.Response
	for i := 0; i < rt.maxRetry; i++ {
		req.Body = io.NopCloser(bytes.NewBuffer(saved))
		resp, err = rt.rt.RoundTrip(req)
		retry := rt.options.retryCheck(resp, err)
		if retry {
			time.Sleep(time.Millisecond * 100)
			continue
		}
		if err != nil {
			return nil, errors.Trace(err)
		}
		return resp, nil
	}
	if resp != nil {
		return resp, nil
	}
	return nil, errors.Trace(err)
}

// WithRetryCheck WithRetryCheck
func WithRetryCheck(retryCheck RetryCheck) RetriableTransportOption {
	return func(o *RetriableTransportOptions) {
		if retryCheck == nil {
			retryCheck = DefaultRetryCheck
		}
		o.retryCheck = retryCheck
	}
}

// RetriableTransportOption RetriableTransportOption
type RetriableTransportOption func(*RetriableTransportOptions)

// DefaultRetryCheck DefaultRetryCheck
func DefaultRetryCheck(resp *http.Response, err error) bool {
	return err != nil || (resp != nil && resp.StatusCode >= http.StatusInternalServerError)
}

// RetryCheck RetryCheck
type RetryCheck func(resp *http.Response, err error) bool

// RetriableTransportOptions RetriableTransportOptions
type RetriableTransportOptions struct {
	retryCheck RetryCheck
}

var defaultRetriableTransportOptions = RetriableTransportOptions{
	retryCheck: DefaultRetryCheck,
}

// RetriableTransport RetriableTransport
func RetriableTransport(maxRetry int, rt http.RoundTripper, opts ...RetriableTransportOption) http.RoundTripper {
	options := defaultRetriableTransportOptions
	for _, opt := range opts {
		opt(&options)
	}
	return &retriableTransport{
		rt:       rt,
		maxRetry: maxRetry,
		options:  &options,
	}
}

// RetriableClient RetriableClient
func RetriableClient() *http.Client {
	return &http.Client{
		Transport: RetriableTransport(maxRetry, Transport()),
	}
}
