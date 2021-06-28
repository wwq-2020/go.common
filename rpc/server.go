package rpc

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/wwq-2020/go.common/errors"
	"github.com/wwq-2020/go.common/httpx"
	"github.com/wwq-2020/go.common/log"
	"google.golang.org/grpc"
)

// errs
var (
	ErrUnExpectedMessageType = errors.New("unexpected messge type")
)

// ServerOption ServerOption
type ServerOption func(*ServerOptions)

// ServerOptions ServerOptions
type ServerOptions struct {
	httpServer *http.Server
	router     Router
}

var defaultServerOptions = ServerOptions{
	httpServer: httpx.DefaultServer(nil),
	router:     NewRouter(),
}

// WithHTTPServer WithHTTPServer
func WithHTTPServer(httpServer *http.Server) ServerOption {
	return func(o *ServerOptions) {
		o.httpServer = httpServer
	}
}

// Server Server
type Server interface {
	ListenAndServe() error
	Stop(ctx context.Context) error
	RegisterService(sd *grpc.ServiceDesc, ss interface{})
}

type server struct {
	httpServer *http.Server
	router     Router
}

// NewServer NewServer
func NewServer(opts ...ServerOption) Server {
	options := defaultServerOptions
	for _, opt := range opts {
		opt(&options)
	}
	options.httpServer.Handler = options.router
	return &server{
		httpServer: options.httpServer,
		router:     options.router,
	}
}

func (s *server) ListenAndServe() error {
	if err := s.httpServer.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		return errors.Trace(err)
	}
	return nil
}

func (s *server) Stop(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (s *server) RegisterService(sd *grpc.ServiceDesc, ss interface{}) {
	s.registerHTTPService(sd, ss)
}

func (s *server) registerHTTPService(sd *grpc.ServiceDesc, ss interface{}) {
	for _, method := range sd.Methods {
		httpMethod := "/" + sd.ServiceName + "/" + method.MethodName
		s.router.Handle(httpMethod, &httpMethodHandler{
			server:  s,
			service: ss,
			handler: methodHandler(method.Handler),
		})
	}
}

type methodHandler func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error)

type httpMethodHandler struct {
	server  Server
	service interface{}
	handler methodHandler
}

func (h *httpMethodHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
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
	resp, err := h.handler(h.service, ctx, reqDec, nil)
	if err != nil {
		log.ErrorContext(ctx, err)
		w.WriteHeader(errors.Code(err))
		io.WriteString(w, err.Error())
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
