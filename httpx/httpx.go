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
	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/stack"
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

	stack := stack.New().
		Set("httpmethod", method).
		Set("url", url)
	var reqBody io.Reader
	var reqData []byte
	if req != nil {
		var err error
		reqData, err = options.codec.Encode(req)
		if err != nil {
			return errors.TraceWithFields(err, stack)
		}
		stack.Set("reqData", string(reqData))
		reqBody = bytes.NewReader(reqData)
	}
	httpReq, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return errors.TraceWithFields(err, stack)
	}
	ctx = log.ContextEnsureTraceID(ctx)
	httpReq = httpReq.WithContext(ctx)
	if options.reqInterceptor != nil {
		if err := options.reqInterceptor(httpReq); err != nil {
			return errors.TraceWithFields(err, stack)
		}
	}
	start := time.Now()
	log.WithFields(stack).
		InfoContext(ctx, "start invoke")

	httpResp, err := options.client.Do(httpReq)
	if err != nil {
		return errors.TraceWithFields(err, stack)
	}

	respData, respBody, err := DrainBody(httpResp.Body)
	if err != nil {
		return errors.TraceWithFields(err, stack)
	}
	elapsed := time.Now().Sub(start).Milliseconds()
	stack.Set("respData", string(respData))
	log.WithField("respData", string(respData)).
		WithField("elapsed", elapsed).
		InfoContext(ctx, "invoke finish")
	httpResp.Body = respBody
	defer httpResp.Body.Close()

	if options.respInterceptor != nil {
		if err := options.respInterceptor(httpResp); err != nil {
			return errors.TraceWithFields(err, stack)
		}
	}
	if resp != nil {
		if err := options.codec.Decode(respData, resp); err != nil {
			return errors.TraceWithFields(err, stack)
		}
	}

	return nil
}

// DrainBody DrainBody
func DrainBody(src io.ReadCloser) ([]byte, io.ReadCloser, error) {
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

var (
	defaultServerAddr              = "127.0.0.1:8080"
	defaultServerReadTimeout       = time.Second * 5
	defaultServerReadHeaderTimeout = time.Second * 2
	defaultServerWriteTimeout      = time.Second * 5
	defaultServerIdleTimeout       = time.Second * 30
	defaultServerMaxHeaderBytes    = 1 << 20
)

// ServerConf ServerConf
type ServerConf struct {
	Addr              string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int
	Handler           http.Handler
}

func (s *ServerConf) fill() {
	if s.Addr == "" {
		s.Addr = defaultServerAddr
	}
	if s.ReadTimeout == 0 {
		s.ReadTimeout = defaultServerReadTimeout
	}
	if s.ReadHeaderTimeout == 0 {
		s.ReadHeaderTimeout = defaultServerReadHeaderTimeout
	}
	if s.WriteTimeout == 0 {
		s.WriteTimeout = defaultServerWriteTimeout
	}
	if s.IdleTimeout == 0 {
		s.IdleTimeout = defaultServerIdleTimeout
	}
	if s.MaxHeaderBytes == 0 {
		s.MaxHeaderBytes = defaultServerMaxHeaderBytes
	}
}

var defaultServerConf = &ServerConf{
	Addr:              defaultServerAddr,
	ReadTimeout:       defaultServerReadTimeout,
	ReadHeaderTimeout: defaultServerReadHeaderTimeout,
	WriteTimeout:      defaultServerWriteTimeout,
	IdleTimeout:       defaultServerIdleTimeout,
	MaxHeaderBytes:    defaultServerMaxHeaderBytes,
}

// HTTPServer HTTPServer
func HTTPServer(conf *ServerConf) *http.Server {
	if conf == nil {
		conf = defaultServerConf
	}
	conf.fill()
	return &http.Server{
		Handler:           conf.Handler,
		Addr:              conf.Addr,
		ReadTimeout:       conf.ReadTimeout,
		ReadHeaderTimeout: conf.ReadHeaderTimeout,
		WriteTimeout:      conf.WriteTimeout,
		IdleTimeout:       conf.IdleTimeout,
		MaxHeaderBytes:    conf.MaxHeaderBytes,
	}
}
