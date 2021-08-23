package httpx

import (
	"net/http"

	"github.com/wwq-2020/go.common/tracing"
)

// Option Option
type Option func(*Options)

// Options Options
type Options struct {
	codec           Codec
	client          *http.Client
	reqInterceptor  ReqInterceptor
	respInterceptor RespInterceptor
	tracingOptions  TracingOptions
}

// TracingOptions TracingOptions
type TracingOptions struct {
	StartSpanOptions []tracing.StartSpanOption
	Root             bool
	OperationName    string
}

var defaultOptions = Options{
	codec:           JSONCodec(),
	client:          DefaultClient(),
	reqInterceptor:  ChainedReqInterceptor(ContentTypeReqInterceptor(ContentTypeJSON)),
	respInterceptor: ChainedRespInterceptor(StatusCodeRespInterceptor(http.StatusOK)),
}

var defaultTracingOptions = TracingOptions{
	Root: true,
}

// WithCodec WithCodec
func WithCodec(codec Codec) Option {
	return func(o *Options) {
		o.codec = codec
	}
}

// WithClient WithClient
func WithClient(client *http.Client) Option {
	return func(o *Options) {
		o.client = client
	}
}

// WithReqInterceptors WithReqInterceptors
func WithReqInterceptors(reqInterceptors ...ReqInterceptor) Option {
	return func(o *Options) {
		o.reqInterceptor = ChainedReqInterceptor(reqInterceptors...)
	}
}

// WithRespInterceptors WithRespInterceptors
func WithRespInterceptors(respInterceptors ...RespInterceptor) Option {
	return func(o *Options) {
		o.respInterceptor = ChainedRespInterceptor(respInterceptors...)
	}
}

// WithTracingOptions WithTracingOptions
func WithTracingOptions(tracingOptions TracingOptions) Option {
	return func(o *Options) {
		o.tracingOptions = tracingOptions
	}
}
