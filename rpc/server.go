package rpc

import (
	"context"
	"net/http"

	"github.com/wwq-2020/go.common/errorsx"
	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/rpc/interceptor"
	"google.golang.org/grpc"
)

// Server Server
type Server interface {
	Start() error
	Stop(ctx context.Context) error
	RegisterGRPC(sd *grpc.ServiceDesc, ss interface{}, interceptors ...interceptor.ServerInterceptor)
	Handle(path string, handler interceptor.MethodHandler, interceptors ...interceptor.ServerInterceptor)
	WithCodec(codec Codec) Server // in case of partial codec
}

type server struct {
	addr    string
	server  *http.Server
	options ServerOptions
}

// NewServer NewServer
func NewServer(addr string, opts ...ServerOption) Server {
	options := defaultServerOptions
	for _, opt := range opts {
		opt(&options)
	}
	return &server{
		addr: addr,
		server: &http.Server{
			Addr:    addr,
			Handler: options.router,
		},
		options: options,
	}
}

// Start Start
func (s *server) Start() error {
	if err := s.server.ListenAndServe(); err != nil {
		return errorsx.Trace(err)
	}
	return nil
}

// Stop Stop
func (s *server) Stop(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return errorsx.Trace(err)
	}
	return nil
}

// RegisterGRPC RegisterGRPC
func (s *server) RegisterGRPC(sd *grpc.ServiceDesc, ss interface{}, interceptors ...interceptor.ServerInterceptor) {
	for _, method := range sd.Methods {
		path := "/" + sd.ServiceName + "/" + method.MethodName
		handler := gprcMethodDescToMethodHandler(method, ss)
		s.Handle(path, handler, interceptors...)
	}
}

// Handle Handle
func (s *server) Handle(path string, handler interceptor.MethodHandler, interceptors ...interceptor.ServerInterceptor) {
	wrappedHandler := s.wrapHandler(handler, interceptors...)
	s.options.router.Handle(path, wrappedHandler)
}

func (s *server) wrapHandler(h interceptor.MethodHandler, interceptors ...interceptor.ServerInterceptor) http.HandlerFunc {
	interceptor := interceptor.ChainServerInerceptor(append(s.options.interceptors, interceptors...)...)
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		traceID := r.Header.Get("traceID")
		ctx = log.ContextWithTraceID(ctx, traceID)
		codec := serverCodecFactory(ctx, r.Body, w, s.options.codec)
		ctx = ContextWithIncomingMetadata(ctx, Metadata(r.Header))
		if err := s.handle(ctx, codec, h, interceptor); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func (s *server) handle(ctx context.Context, codec ServerCodec, h interceptor.MethodHandler, interceptor interceptor.ServerInterceptor) error {
	gotResp, err := h(ctx, codec.Decode, interceptor)
	needWrap := isRespNeedWrap(gotResp)
	if err != nil {
		log.ErrorContext(ctx, err)
		if !needWrap {
			return errorsx.Trace(err)
		}
		code := errorsx.Code(err)
		msg := err.Error()
		gotResp = respObj{
			Code: code,
			Msg:  msg,
		}
		goto ret
	}
	if needWrap {
		gotResp = respObj{
			Code: 0,
			Msg:  "success",
			Data: gotResp,
		}
	}
ret:
	if err := codec.Encode(gotResp); err != nil {
		return errorsx.Trace(err)
	}
	return nil
}

// WithCodec WithCodec
func (s *server) WithCodec(codec Codec) Server {
	options := s.options
	options.codec = codec
	return &server{
		addr:    s.addr,
		server:  s.server,
		options: options,
	}
}

func gprcMethodDescToMethodHandler(method grpc.MethodDesc, ss interface{}) interceptor.MethodHandler {
	return func(ctx context.Context, dec func(interface{}) error, intr interceptor.ServerInterceptor) (interface{}, error) {
		grpcInterceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			resp, err := intr(ctx, req, interceptor.ServerHandler(handler))
			if err != nil {
				return nil, errorsx.Trace(err)
			}
			return resp, nil
		}
		resp, err := method.Handler(ss, ctx, dec, grpcInterceptor)
		if err != nil {
			return nil, errorsx.Trace(err)
		}
		return resp, nil
	}
}
