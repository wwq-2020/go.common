package rpc

import (
	"github.com/wwq-2020/go.common/rpc/interceptor"
)

// ServerOptions ServerOptions
type ServerOptions struct {
	codec         Codec
	interceptors  []interceptor.ServerInterceptor
	routerFactory RouterFactory
}

// ServerOption ServerOption
type ServerOption func(*ServerOptions)

var (
	defaultServerOptions = ServerOptions{
		codec:         JSONCodec(),
		interceptors:  []interceptor.ServerInterceptor{},
		routerFactory: NewRouter,
	}
)

// ServerWithCodec ServerWithCodec
func ServerWithCodec(codec Codec) ServerOption {
	return func(o *ServerOptions) {
		o.codec = codec
	}
}

// ServerWithInterceptors ServerWithInterceptors
func ServerWithInterceptors(interceptors ...interceptor.ServerInterceptor) ServerOption {
	return func(o *ServerOptions) {
		o.interceptors = interceptors
	}
}
