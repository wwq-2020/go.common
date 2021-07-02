package httpx

import (
	"net/http"
)

// Option Option
type Option func(*Options)

// Options Options
type Options struct {
	codec           Codec
	client          *http.Client
	reqInterceptor  ReqInterceptor
	respInterceptor RespInterceptor
}

var defaultOptions = Options{
	codec:           JSONCodec(),
	client:          DefaultClient(),
	reqInterceptor:  ChainedReqInterceptor(ContentTypeReqInterceptor(ContentTypeJSON)),
	respInterceptor: ChainedRespInterceptor(StatusCodeRespInterceptor(http.StatusOK)),
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
