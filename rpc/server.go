package rpc

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/wwq-2020/go.common/errcode"
	"github.com/wwq-2020/go.common/errorsx"
	"github.com/wwq-2020/go.common/httpx"
	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/rpc/interceptor"
	"github.com/wwq-2020/go.common/stack"
	"github.com/wwq-2020/go.common/tracing"
	"google.golang.org/grpc"
)

// consts
const (
	StatusCodeHeader string = "statuscode"
	StatusMsgHeader  string = "statusmsg"
)

// Interceptor Interceptor
type Interceptor interface {
	Interceptor(ctx context.Context, req interface{}, handler interceptor.ServerHandler) (resp interface{}, err error)
}

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

// ServerConf ServerConf
type ServerConf struct {
	Addr string `toml:"addr" yaml:"addr" json:"addr"`
}

func (c *ServerConf) fill() {

}

var defaultServerConf = &ServerConf{
	Addr: "127.0.0.1:8080",
}

// NewServer NewServer
func NewServer(conf *ServerConf, opts ...ServerOption) Server {
	if conf == nil {
		conf = defaultServerConf
	}
	conf.fill()
	options := defaultServerOptions
	for _, opt := range opts {
		opt(&options)
	}
	wrappedHandler := wrapHTTPHandler(options.router)
	return &server{
		addr:   conf.Addr,
		router: options.router,
		server: &http.Server{
			Addr:    conf.Addr,
			Handler: wrappedHandler,
		},
		options: options,
	}
}

// Start Start
func (s *server) Start() error {
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

		svcInterceptor, ok := ss.(Interceptor)
		if !ok {
			s.Handle(path, handler, interceptors...)
			return
		}
		s.Handle(path, handler, append([]interceptor.ServerInterceptor{svcInterceptor.Interceptor}, interceptors...)...)
	}
}

// Handle Handle
func (s *server) Handle(path string, handler interceptor.MethodHandler, interceptors ...interceptor.ServerInterceptor) {
	wrappedHandler := s.wrapHandler(handler, interceptors...)
	s.router.Handle(path, wrappedHandler)
}

func (s *server) HandleNotFound(handler http.Handler) {
	s.router.HandleNotFound(wrapHTTPHandler(handler))
}

func (s *server) wrapHandler(h interceptor.MethodHandler, interceptors ...interceptor.ServerInterceptor) http.HandlerFunc {
	interceptor := interceptor.ChainServerInerceptor(append(s.options.interceptors, interceptors...)...)
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		codec := serverCodecFactory(req.Body, w, s.options.codec)
		code := errcode.ErrCode_Ok
		msg := "success"
		gotResp, err := h(ctx, codec.Decode, interceptor)
		needWrap := isRespNeedWrap(gotResp)
		if err != nil {
			log.ErrorContext(ctx, err)
			if strings.Contains(err.Error(), "unexpected end of JSON input") {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			code = errorsx.Code(err)
			msg = err.Error()
			if !needWrap {
				w.Header().Set(StatusCodeHeader, strconv.Itoa(int(code)))
				w.Header().Set(StatusMsgHeader, msg)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			gotResp = respObj{
				Code: code,
				Msg:  msg,
			}
			goto ret
		}
		if needWrap {
			gotResp = respObj{
				Code: code,
				Msg:  msg,
				Data: gotResp,
			}
		}
		w.Header().Set(StatusCodeHeader, strconv.Itoa(int(code)))
		w.Header().Set(StatusMsgHeader, msg)
	ret:
		if err := codec.Encode(gotResp); err != nil {
			log.ErrorContext(ctx, err)
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
			Code: errcode.ErrCode_Ok,
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
			return resp, errorsx.Trace(err)
		}
		resp, err := method.Handler(ss, ctx, dec, grpcInterceptor)
		return resp, errorsx.Trace(err)
	}
}

type responseWriter struct {
	http.ResponseWriter
	buffer     *bytes.Buffer
	statusCode int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.ResponseWriter.WriteHeader(statusCode)
	rw.statusCode = statusCode
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(data)
	if err != nil {
		return 0, errorsx.Trace(err)
	}
	rw.buffer.Write(data[:n])
	return n, nil
}

// WrapHTTPHandler WrapHTTPHandler
func WrapHTTPHandler(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		wrappedHandler := wrapHTTPHandler(handler)
		if wrappedHandler != nil {
			wrappedHandler.ServeHTTP(w, req)
		}
	}
}
func wrapHTTPHandler(handler http.Handler) http.Handler {
	middleware := defaultChainedHTTPMiddleware()
	wrappedHandler := middleware(handler)
	return wrappedHandler
}

func defaultChainedHTTPMiddleware() httpMiddleware {
	return chainedHTTPMiddleware(trace, metrics, recovery)
}

func chainedHTTPMiddleware(middlewares ...httpMiddleware) httpMiddleware {
	return func(handler http.Handler) http.Handler {
		chainedHandler := handler
		for i := len(middlewares) - 1; i >= 0; i-- {
			chainedHandler = middlewares[i](chainedHandler)
		}
		return chainedHandler
	}
}

type httpMiddleware func(http.Handler) http.Handler

func trace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		ctx, cancel := context.WithTimeout(ctx, time.Second*30)
		defer cancel()
		span, ctx := tracing.HTTPServerStartSpan(ctx, "serve", req, w)
		ctx = ContextWithIncomingMetadata(ctx, Metadata(req.Header))
		ldap := LdapFromIncomingContext(ctx)
		token := TokenFromIncomingContext(ctx)
		rw := &responseWriter{ResponseWriter: w, buffer: bytes.NewBuffer(nil)}
		stack := stack.New().
			Set("httpmethod", req.Method).
			Set("path", req.URL.Path).
			Set("ldap", ldap).
			Set("token", token)
		var err error
		defer span.FinishWithFields(&err, stack)
		reqData, reqBody, err := httpx.DrainBody(req.Body)
		if err != nil {
			stack.Set("httpStatusCode", rw.statusCode)
			log.WithFields(stack).
				ErrorContext(ctx, err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		start := time.Now()
		stack.Set("reqData", string(reqData)).
			Set("handleStart", start.Format("2006-01-02 15:04:05"))
		log.WithFields(stack).
			InfoContext(ctx, "recv req")
		req = req.WithContext(ctx)
		req.Body = reqBody
		if next != nil {
			next.ServeHTTP(rw, req)
			if rw.statusCode == 0 {
				rw.statusCode = http.StatusOK
			}
		}
		statusCode := rw.Header().Get(StatusCodeHeader)
		statusMsg := rw.Header().Get(StatusMsgHeader)
		end := time.Now()
		stack.Set("respData", string(rw.buffer.Bytes())).
			Set("httpStatusCode", rw.statusCode).
			Set("statusCode", statusCode).
			Set("statusMsg", statusMsg).
			Set("handleEnd", end.Format("2006-01-02 15:04:05"))
		log.WithField("respData", string(rw.buffer.Bytes())).
			WithField("httpStatusCode", rw.statusCode).
			WithField("statusCode", statusCode).
			WithField("statusMsg", statusMsg).
			WithField("handleEnd", end.Format("2006-01-02 15:04:05")).
			InfoContext(ctx, "finish req")
	})
}

func metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// to do
		if next != nil {
			next.ServeHTTP(w, req)
		}
	})
}

func recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			ctx := req.Context()
			if e := recover(); e != nil {
				if e == http.ErrAbortHandler {
					panic(http.ErrAbortHandler)
				}
				var err error
				switch v := e.(type) {
				case error:
					err = v
				default:
					err = fmt.Errorf("%+v", v)
				}
				stack := stack.Callers(stack.StdFilter)
				log.WithField("stack", stack).
					ErrorContext(ctx, err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		if next != nil {
			next.ServeHTTP(w, req)
		}
	})
}
