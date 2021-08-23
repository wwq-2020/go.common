package rpc

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"github.com/wwq-2020/go.common/errorsx"
	"github.com/wwq-2020/go.common/httputilx"
	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/rpc/interceptor"
	"github.com/wwq-2020/go.common/stack"
	"github.com/wwq-2020/go.common/tracing"
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
	name    string
	addr    string
	server  *http.Server
	options ServerOptions
	router  Router
}

// NewServer NewServer
func NewServer(name, addr string, opts ...ServerOption) Server {
	options := defaultServerOptions
	for _, opt := range opts {
		opt(&options)
	}
	router := options.routerFactory(name)
	return &server{
		name:   name,
		addr:   addr,
		router: router,
		server: &http.Server{
			Addr:    addr,
			Handler: router,
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
	s.router.Handle(path, wrappedHandler)
}

func (s *server) wrapHandler(h interceptor.MethodHandler, interceptors ...interceptor.ServerInterceptor) http.HandlerFunc {
	interceptor := interceptor.ChainServerInerceptor(append(s.options.interceptors, interceptors...)...)
	return func(w http.ResponseWriter, r *http.Request) {
		span, ctx := tracing.HTTPServerStartSpan(r.Context(), s.name+"-serve", r, w)
		stack := stack.New().
			Set("httpmethod", r.Method).
			Set("path", r.URL.Path)
		var err error
		defer span.FinishWithFields(&err, stack)
		reqData, reqBody, err := httputilx.DrainBody(r.Body)
		if err != nil {
			log.WithFields(stack).
				ErrorContext(ctx, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		start := time.Now()
		stack.Set("reqData", string(reqData)).
			Set("handleStart", start.Format("2006-01-02 15:04:05"))
		log.WithFields(stack).
			InfoContext(ctx, "recv req")
		r.Body = reqBody
		rw := &responseWriter{ResponseWriter: w, buffer: bytes.NewBuffer(nil)}
		codec := serverCodecFactory(ctx, r.Body, rw, s.options.codec)
		ctx = ContextWithIncomingMetadata(ctx, Metadata(r.Header))
		err = s.handle(ctx, codec, h, interceptor)
		if err != nil {
			log.ErrorContext(ctx, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		end := time.Now()
		stack.Set("respData", string(rw.buffer.Bytes())).
			Set("handleEnd", end.Format("2006-01-02 15:04:05"))
		log.WithField("respData", string(rw.buffer.Bytes())).
			WithField("handleEnd", end.Format("2006-01-02 15:04:05")).
			InfoContext(ctx, "finish req")
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
		router:  s.router,
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

type responseWriter struct {
	http.ResponseWriter
	buffer *bytes.Buffer
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(data)
	if err != nil {
		return 0, errorsx.Trace(err)
	}
	rw.buffer.Write(data[:n])
	return n, nil
}
