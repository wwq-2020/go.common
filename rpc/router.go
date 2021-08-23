package rpc

import (
	"fmt"
	"net/http"

	"github.com/wwq-2020/go.common/errorsx"
	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/stack"
	"github.com/wwq-2020/go.common/tracing"
)

// Router Router
type Router interface {
	http.Handler
	Handle(path string, handler http.Handler) Router
	HandleNotFound(handler http.Handler) Router
}

// RouterFactory RouterFactory
type RouterFactory func(name string) Router

// NewRouter NewRouter
func NewRouter(name string) Router {
	return &router{
		m:               make(map[string]http.Handler, 100),
		NotFoundHandler: makeNotFoundHandler(name),
	}
}

type router struct {
	m               map[string]http.Handler
	NotFoundHandler http.Handler
}

func (r *router) Handle(path string, handler http.Handler) Router {
	if _, ok := r.m[path]; ok {
		panic(fmt.Sprintf("duplicate path register: %s", path))
	}
	r.m[path] = handler
	return r
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	if req.Method != http.MethodPost {
		r.NotFoundHandler.ServeHTTP(w, req)
		return
	}
	handler, ok := r.m[path]
	if !ok {
		r.NotFoundHandler.ServeHTTP(w, req)
		return
	}
	handler.ServeHTTP(w, req)
}

func (r *router) HandleNotFound(handler http.Handler) Router {
	if handler == nil {
		return r
	}
	r.NotFoundHandler = handler
	return r
}

// WrapHandler WrapHandler
func makeNotFoundHandler(name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		span, ctx := tracing.HTTPServerStartSpan(r.Context(), name+"-serve", r, w)
		stack := stack.New().
			Set("method", r.Method).
			Set("path", r.URL.Path)

		err := errorsx.New("not found")
		log.WithFields(stack).
			ErrorContext(ctx, err)
		span.FinishWithFields(&err, stack)
	})
}
