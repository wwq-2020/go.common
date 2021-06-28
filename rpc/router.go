package rpc

import (
	"fmt"
	"net/http"
)

// Router Router
type Router interface {
	http.Handler
	Handle(path string, handler http.Handler)
}

// NewRouter NewRouter
func NewRouter() Router {
	return &router{
		m:               make(map[string]http.Handler, 100),
		NotFoundHandler: http.NotFoundHandler(),
	}
}

type router struct {
	m               map[string]http.Handler
	NotFoundHandler http.Handler
}

func (r *router) Handle(path string, handler http.Handler) {
	if _, ok := r.m[path]; ok {
		panic(fmt.Sprintf("duplicate path register: %s", path))
	}
	r.m[path] = handler
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
