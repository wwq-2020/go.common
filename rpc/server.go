package rpc

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/wwq-2020/go.common/errorsx"
	"github.com/wwq-2020/go.common/httpx"
	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/rpc/middleware"
	"google.golang.org/grpc"
)

// errs
var (
	ErrUnExpectedMessageType = errorsx.New("unexpected messge type")
)

// ServerOption ServerOption
type ServerOption func(*ServerOptions)

// ServerOptions ServerOptions
type ServerOptions struct {
	httpServer  *http.Server
	router      Router
	interceptor grpc.UnaryServerInterceptor
}

var defaultServerOptions = ServerOptions{
	httpServer: httpx.Server(nil),
	router:     NewRouter(),
}

// WithHTTPServer WithHTTPServer
func WithHTTPServer(httpServer *http.Server) ServerOption {
	return func(o *ServerOptions) {
		o.httpServer = httpServer
	}
}

// WithServerInterceptors WithServerInterceptors
func WithServerInterceptors(interceptors ...grpc.UnaryServerInterceptor) ServerOption {
	return func(o *ServerOptions) {
		o.interceptor = middleware.ChainServerInerceptor(interceptors...)
	}
}

// Server Server
type Server interface {
	ListenAndServe() error
	Stop(ctx context.Context) error
	RegisterService(sd *grpc.ServiceDesc, ss interface{})
}

type server struct {
	options *ServerOptions
}

// NewServer NewServer
func NewServer(opts ...ServerOption) Server {
	options := defaultServerOptions
	for _, opt := range opts {
		opt(&options)
	}
	options.httpServer.Handler = options.router.HandleNotFound(options.httpServer.Handler)
	return &server{
		options: &options,
	}
}

func (s *server) ListenAndServe() error {
	s.options.httpServer.Handler = s.options.router
	log.WithField("addr", s.options.httpServer.Addr).
		Info("start serving")
	if err := s.options.httpServer.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		return errorsx.Trace(err)
	}
	return nil
}

func (s *server) Stop(ctx context.Context) error {
	log.WithField("addr", s.options.httpServer.Addr).
		Info("stop serving")
	if err := s.options.httpServer.Shutdown(ctx); err != nil {
		return errorsx.Trace(err)
	}
	return nil
}

func (s *server) RegisterService(sd *grpc.ServiceDesc, ss interface{}) {
	s.registerHTTPService(sd, ss)
}

func (s *server) registerHTTPService(sd *grpc.ServiceDesc, ss interface{}) {
	for _, method := range sd.Methods {
		httpMethod := "/" + sd.ServiceName + "/" + method.MethodName
		s.options.router.Handle(httpMethod, &httpMethodHandler{
			server:  s,
			service: ss,
			handler: methodHandler(method.Handler),
		})
	}
}

type methodHandler func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error)

type httpMethodHandler struct {
	server  *server
	service interface{}
	handler methodHandler
}

func (h *httpMethodHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	traceID := log.GenTraceID()
	ctx = log.ContextWithTraceID(ctx, traceID)
	req = req.WithContext(ctx)
	w.Header().Set("traceID", traceID)
	reqData, reqBody, err := httpx.DrainBody(req.Body)
	if err != nil {
		log.ErrorContext(ctx, err)
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "failed to read reqbody")
		return
	}
	log.WithField("reqData", string(reqData)).
		InfoContext(ctx, "recv req")

	req.Body = reqBody
	start := time.Now()

	reqDec := requestDecoder(ctx, req.Body)
	resp, err := h.handler(h.service, ctx, reqDec, h.server.options.interceptor)
	if err != nil {
		log.ErrorContext(ctx, err)
		code := errorsx.Code(err)
		w.Header().Add("rpccode", strconv.Itoa(code))
		w.Header().Add("rpcmsg", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	elapsed := time.Now().Sub(start).Milliseconds()
	respData, err := json.Marshal(resp)
	if err != nil {
		log.ErrorContext(ctx, err)
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "failed to read reqbody")
		return
	}
	log.WithField("respData", string(respData)).
		WithField("elapsed", elapsed).
		InfoContext(ctx, "finish req")
	if _, err := w.Write(respData); err != nil {
		log.ErrorContext(ctx, err)
		return
	}
}

func requestDecoder(ctx context.Context, r io.Reader) func(interface{}) error {
	return func(v interface{}) error {
		if err := json.NewDecoder(r).Decode(v); err != nil {
			return err
		}
		return nil
	}
}
