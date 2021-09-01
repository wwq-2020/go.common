package httpx

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/wwq-2020/go.common/bus"
	"github.com/wwq-2020/go.common/errorsx"
	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/stack"
	"github.com/wwq-2020/go.common/tracing"
	"github.com/wwq-2020/go.common/util"
)

// vars
var (
	DefaultMaxRetry                  = 3
	DefaultDialTimeout               = 5 * time.Second
	DefaultDialTimeoutStr            = DefaultDialTimeout.String()
	DefaultKeepAlive                 = 30 * time.Second
	DefaultKeepAliveStr              = DefaultKeepAlive.String()
	DefaultDisableKeepAlives         = false
	DefaultDisableCompression        = false
	DefaultMaxIdleConns              = 100
	DefaultMaxIdleConnsPerHost       = 10
	DefaultMaxConnsPerHost           = 100
	DefaultIdleConnTimeout           = 30 * time.Second
	DefaultIdleConnTimeoutStr        = DefaultIdleConnTimeout.String()
	DefaultResponseHeaderTimeout     = 5 * time.Second
	DefaultResponseHeaderTimeoutStr  = DefaultResponseHeaderTimeout.String()
	DefaultExpectContinueTimeout     = 5 * time.Second
	DefaultExpectContinueTimeoutStr  = DefaultExpectContinueTimeout.String()
	DefaultMaxResponseHeaderBytes    = int64(1 << 10)
	DefaultMaxResponseHeaderBytesStr = util.ToByteStr(DefaultMaxResponseHeaderBytes)
	DefaultWriteBufferSize           = 1 << 12
	DefaultWriteBufferSizeStr        = util.ToByteStr(int64(DefaultWriteBufferSize))
	DefaultReadBufferSize            = 1 << 12
	DefaultReadBufferSizeStr         = util.ToByteStr(int64(DefaultReadBufferSize))
	DefaultForceAttemptHTTP2         = false
)

// DefaultTransport DefaultTransport
func DefaultTransport() http.RoundTripper {
	return MakeTransport(defaultTransportConf)
}

type retriableTransport struct {
	maxRetry   int
	rt         http.RoundTripper
	retryCheck RetryCheck
}

func (rt *retriableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var saved []byte
	var err error
	ctx := req.Context()
	stack := stack.New()
	stack.Set("httpmethod", req.Method).
		Set("url", req.URL.String())
	span, ctx := tracing.StartSpan(ctx, "RoundTrip")
	defer span.FinishWithFields(&err, stack)
	req = req.WithContext(ctx)
	span.InjectToHTTPReq(req)
	if req.Body != nil {
		saved, err = ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, errorsx.Trace(err)
		}
		stack.Set("reqBody", string(saved))
	}
	start := time.Now()
	stack.Set("invokeStart", start.Format("2006-01-02 15:04:05"))

	var resp *http.Response
	for i := 0; i < rt.maxRetry; i++ {
		curTime := time.Now()
		stack.Set("curTime", curTime.Format("2006-01-02 15:04:05"))
		stack.Set("retries", i)
		log.WithFields(stack).
			InfoContext(ctx, "start invoke")
		req.Body = io.NopCloser(bytes.NewBuffer(saved))
		resp, err = rt.rt.RoundTrip(req)
		retry := rt.retryCheck(req, resp, err)
		if retry {
			time.Sleep(time.Millisecond * 100)
			continue
		}
		if err != nil {
			return nil, errorsx.Trace(err)
		}
		goto end
	}
end:
	if resp != nil {
		respData, respBody, err := DrainBody(resp.Body)
		if err != nil {
			return nil, errorsx.Trace(err)
		}
		resp.Body = respBody

		end := time.Now()
		elapsed := end.Sub(start).Milliseconds()
		respDataStr := string(respData)
		stack.Set("respData", respDataStr).
			Set("elapsed", elapsed).
			Set("httpStatusCode", resp.StatusCode).
			Set("invokeFinish", end.Format("2006-01-02 15:04:05"))
		log.WithField("respData", respDataStr).
			WithField("elapsed", elapsed).
			WithField("httpStatusCode", resp.StatusCode).
			WithField("invokeFinish", end.Format("2006-01-02 15:04:05")).
			InfoContext(ctx, "invoke finish")
		return resp, nil
	}
	return nil, errorsx.Trace(err)
}

// DefaultRetryCheck DefaultRetryCheck
func DefaultRetryCheck(req *http.Request, resp *http.Response, err error) bool {
	return err != nil || (resp != nil && resp.StatusCode >= http.StatusInternalServerError)
}

// RetryCheck RetryCheck
type RetryCheck func(req *http.Request, resp *http.Response, err error) bool

// TransportOptions TransportOptions
type TransportOptions struct {
	retryCheck RetryCheck
}

var defaultTransportOptions = TransportOptions{
	retryCheck: DefaultRetryCheck,
}

// TransportConf TransportConf
type TransportConf struct {
	MaxRetry               *int    `toml:"max_retry" yaml:"max_retry" json:"max_retry"`
	DialTimeout            *string `toml:"dial_timeout" yaml:"dial_timeout" json:"dial_timeout"`
	KeepAlive              *string `toml:"keepalive" yaml:"keepalive" json:"keepalive"`
	DisableKeepAlives      *bool   `toml:"disable_keep_alives" yaml:"disable_keep_alives" json:"disable_keep_alives"`
	DisableCompression     *bool   `toml:"disable_compression" yaml:"disable_compression" json:"disable_compression"`
	MaxIdleConns           *int    `toml:"max_idle_conns" yaml:"max_idle_conns" json:"max_idle_conns"`
	MaxIdleConnsPerHost    *int    `toml:"max_idle_conns_per_host" yaml:"max_idle_conns_per_host" json:"max_idle_conns_per_host"`
	MaxConnsPerHost        *int    `toml:"max_conns_per_host" yaml:"max_conns_per_host" json:"max_conns_per_host"`
	IdleConnTimeout        *string `toml:"idle_conn_timeout" yaml:"idle_conn_timeout" json:"idle_conn_timeout"`
	ResponseHeaderTimeout  *string `toml:"response_header_timeout" yaml:"response_header_timeout" json:"response_header_timeout"`
	ExpectContinueTimeout  *string `toml:"expect_continue_timeout" yaml:"expect_continue_timeout" json:"expect_continue_timeout"`
	MaxResponseHeaderBytes *string `toml:"max_response_header_bytes" yaml:"max_response_header_bytes" json:"max_response_header_bytes"`
	WriteBufferSize        *string `toml:"write_buffer_size" yaml:"write_buffer_size" json:"write_buffer_size"`
	ReadBufferSize         *string `toml:"read_buffer_size" yaml:"read_buffer_size" json:"read_buffer_size"`
	ForceAttemptHTTP2      *bool   `toml:"force_attempt_http2" yaml:"force_attempt_http2" json:"force_attempt_http2"`
	RetryCheck             RetryCheck
}

func (c *TransportConf) fill() {
	if c.MaxRetry == nil || *c.MaxRetry < 0 {
		c.MaxRetry = &DefaultMaxRetry
	}
	if c.DialTimeout == nil || *c.DialTimeout == "" {
		c.DialTimeout = &DefaultDialTimeoutStr
	}
	if c.KeepAlive == nil || *c.KeepAlive == "" {
		c.KeepAlive = &DefaultKeepAliveStr
	}
	if c.DisableKeepAlives == nil {
		c.DisableKeepAlives = &DefaultDisableKeepAlives
	}
	if c.DisableCompression == nil {
		c.DisableCompression = &DefaultDisableCompression
	}
	if c.MaxIdleConns == nil || *c.MaxIdleConns <= 0 {
		c.MaxIdleConns = &DefaultMaxIdleConns
	}
	if c.MaxIdleConnsPerHost == nil || *c.MaxIdleConnsPerHost <= 0 {
		c.MaxIdleConnsPerHost = &DefaultMaxIdleConnsPerHost
	}
	if c.MaxConnsPerHost == nil || *c.MaxConnsPerHost <= 0 {
		c.MaxConnsPerHost = &DefaultMaxConnsPerHost
	}
	if c.IdleConnTimeout == nil || *c.IdleConnTimeout == "" {
		c.IdleConnTimeout = &DefaultIdleConnTimeoutStr
	}
	if c.ResponseHeaderTimeout == nil || *c.ResponseHeaderTimeout == "" {
		c.ResponseHeaderTimeout = &DefaultResponseHeaderTimeoutStr
	}
	if c.ExpectContinueTimeout == nil || *c.ExpectContinueTimeout == "" {
		c.ExpectContinueTimeout = &DefaultExpectContinueTimeoutStr
	}
	if c.MaxResponseHeaderBytes == nil || *c.MaxResponseHeaderBytes == "" {
		c.MaxResponseHeaderBytes = &DefaultMaxResponseHeaderBytesStr
	}
	if c.WriteBufferSize == nil || *c.WriteBufferSize == "" {
		c.WriteBufferSize = &DefaultWriteBufferSizeStr
	}
	if c.ReadBufferSize == nil || *c.ReadBufferSize == "" {
		c.ReadBufferSize = &DefaultReadBufferSizeStr
	}
	if c.ForceAttemptHTTP2 == nil {
		c.ForceAttemptHTTP2 = &DefaultForceAttemptHTTP2
	}
	if c.RetryCheck == nil {
		c.RetryCheck = DefaultRetryCheck
	}
}

var (
	defaultTransportConf = &TransportConf{
		MaxRetry:               &DefaultMaxRetry,
		DialTimeout:            &DefaultDialTimeoutStr,
		KeepAlive:              &DefaultKeepAliveStr,
		DisableKeepAlives:      &DefaultDisableKeepAlives,
		DisableCompression:     &DefaultDisableCompression,
		MaxIdleConns:           &DefaultMaxIdleConns,
		MaxIdleConnsPerHost:    &DefaultMaxIdleConnsPerHost,
		MaxConnsPerHost:        &DefaultMaxConnsPerHost,
		IdleConnTimeout:        &DefaultIdleConnTimeoutStr,
		ResponseHeaderTimeout:  &DefaultResponseHeaderTimeoutStr,
		ExpectContinueTimeout:  &DefaultExpectContinueTimeoutStr,
		MaxResponseHeaderBytes: &DefaultMaxResponseHeaderBytesStr,
		WriteBufferSize:        &DefaultWriteBufferSizeStr,
		ReadBufferSize:         &DefaultReadBufferSizeStr,
		ForceAttemptHTTP2:      &DefaultForceAttemptHTTP2,
		RetryCheck:             DefaultRetryCheck,
	}
)

// MakeTransport MakeTransport
func MakeTransport(transportConf *TransportConf) http.RoundTripper {
	if transportConf == nil {
		transportConf = defaultTransportConf
	}
	transportConf.fill()
	dialer := &net.Dialer{
		Timeout:   DefaultDialTimeout,
		KeepAlive: DefaultKeepAlive,
		DualStack: true,
	}
	rt := &http.Transport{
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			log.WithField("network", network).
				WithField("address", address).
				DebugContext(ctx, "dial")
			conn, err := dialer.DialContext(ctx, network, address)
			if err != nil {
				return nil, errorsx.Trace(err)
			}
			return conn, nil
		},
		DisableKeepAlives:      DefaultDisableKeepAlives,
		DisableCompression:     DefaultDisableCompression,
		MaxIdleConns:           DefaultMaxIdleConns,
		MaxIdleConnsPerHost:    DefaultMaxIdleConnsPerHost,
		MaxConnsPerHost:        DefaultMaxConnsPerHost,
		IdleConnTimeout:        DefaultIdleConnTimeout,
		ResponseHeaderTimeout:  DefaultResponseHeaderTimeout,
		ExpectContinueTimeout:  DefaultExpectContinueTimeout,
		MaxResponseHeaderBytes: DefaultMaxResponseHeaderBytes,
		WriteBufferSize:        DefaultWriteBufferSize,
		ReadBufferSize:         DefaultReadBufferSize,
		ForceAttemptHTTP2:      DefaultForceAttemptHTTP2,
	}

	dialTimeout, err := time.ParseDuration(*transportConf.DialTimeout)
	if err == nil && dialTimeout != 0 {
		dialer.Timeout = dialTimeout
	}
	if err != nil {
		log.WithField("dial_timeout", transportConf.DialTimeout).
			Error(err)
	}
	keepAlive, err := time.ParseDuration(*transportConf.KeepAlive)
	if err == nil && keepAlive != 0 {
		dialer.KeepAlive = keepAlive
	}
	if err != nil {
		log.WithField("keepalive", transportConf.KeepAlive).
			Error(err)
	}
	rt.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		log.WithField("network", network).
			WithField("address", address).
			DebugContext(ctx, "dial")
		conn, err := dialer.DialContext(ctx, network, address)
		if err != nil {
			return nil, errorsx.Trace(err)
		}
		return conn, nil
	}

	if transportConf.DisableKeepAlives != nil {
		rt.DisableKeepAlives = *transportConf.DisableKeepAlives
	}
	if transportConf.DisableCompression != nil {
		rt.DisableCompression = *transportConf.DisableCompression
	}
	if transportConf.MaxIdleConns != nil && *transportConf.MaxIdleConns >= 0 {
		rt.MaxIdleConns = *transportConf.MaxIdleConns
	}
	if transportConf.MaxIdleConnsPerHost != nil && *transportConf.MaxIdleConnsPerHost >= 0 {
		rt.MaxIdleConnsPerHost = *transportConf.MaxIdleConnsPerHost
	}
	if transportConf.MaxConnsPerHost != nil && *transportConf.MaxConnsPerHost >= 0 {
		rt.MaxConnsPerHost = *transportConf.MaxConnsPerHost
	}
	idleConnTimeout, err := time.ParseDuration(*transportConf.IdleConnTimeout)
	if err == nil && idleConnTimeout != 0 {
		rt.IdleConnTimeout = idleConnTimeout
	}
	if err != nil {
		log.WithField("idle_conn_timeout", transportConf.IdleConnTimeout).
			Error(err)
	}
	responseHeaderTimeout, err := time.ParseDuration(*transportConf.ResponseHeaderTimeout)
	if err == nil && responseHeaderTimeout != 0 {
		rt.ResponseHeaderTimeout = responseHeaderTimeout
	}
	if err != nil {
		log.WithField("response_header_timeout", transportConf.ResponseHeaderTimeout).
			Error(err)
	}
	expectContinueTimeout, err := time.ParseDuration(*transportConf.ExpectContinueTimeout)
	if err == nil && expectContinueTimeout != 0 {
		rt.ExpectContinueTimeout = expectContinueTimeout
	}
	if err != nil {
		log.WithField("expect_continue_timeout", transportConf.ExpectContinueTimeout).
			Error(err)
	}
	maxResponseHeaderBytes, err := util.ParseByteStr(*transportConf.MaxResponseHeaderBytes)
	if err == nil && maxResponseHeaderBytes != 0 {
		rt.MaxResponseHeaderBytes = maxResponseHeaderBytes
	}
	if err != nil {
		log.WithField("max_response_header_bytes", transportConf.MaxResponseHeaderBytes).
			Error(err)
	}
	writeBufferSize, err := util.ParseByteStr(*transportConf.WriteBufferSize)
	if err == nil && writeBufferSize != 0 {
		rt.WriteBufferSize = int(writeBufferSize)
	}
	if err != nil {
		log.WithField("write_buffer_size", transportConf.WriteBufferSize).
			Error(err)
	}

	readBufferSize, err := util.ParseByteStr(*transportConf.ReadBufferSize)
	if err == nil && readBufferSize != 0 {
		rt.ReadBufferSize = int(readBufferSize)
	}
	if err != nil {
		log.WithField("read_buffer_size", transportConf.ReadBufferSize).
			Error(err)
	}
	if transportConf.ForceAttemptHTTP2 != nil {
		rt.ForceAttemptHTTP2 = *transportConf.ForceAttemptHTTP2
	}

	return &retriableTransport{
		rt:         rt,
		maxRetry:   *transportConf.MaxRetry,
		retryCheck: transportConf.RetryCheck,
	}
}

type changableTransport struct {
	rt atomic.Value
}

func (ct *changableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := ct.rt.Load().(http.RoundTripper).RoundTrip(req)
	if err != nil {
		return nil, errorsx.Trace(err)
	}
	return resp, nil
}

// MakeChangableTransport MakeChangableTransport
func MakeChangableTransport(ctx context.Context, ch <-chan *TransportConf) http.RoundTripper {
	ct := &changableTransport{}
	callback := func(transportConf *TransportConf) {
		transport := MakeTransport(transportConf)
		ct.rt.Store(transport)
	}
	transportConf := <-ch
	callback(transportConf)
	bus.SubscribeChans(ctx, ch, callback)
	return ct
}

// ChangableTransport ChangableTransport
func ChangableTransport(ctx context.Context, ch <-chan http.RoundTripper) http.RoundTripper {
	ct := &changableTransport{}
	callback := func(transport http.RoundTripper) {
		ct.rt.Store(transport)
	}
	transport := <-ch
	callback(transport)
	bus.SubscribeChans(ctx, ch, callback)
	return ct
}
