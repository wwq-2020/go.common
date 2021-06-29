package rpc

import (
	"context"
	"net/http"

	"github.com/wwq-2020/go.common/app"
	"github.com/wwq-2020/go.common/httpx"
	"github.com/wwq-2020/go.common/log"
	"google.golang.org/grpc"
)

// Conf Conf
type Conf struct {
	Addr    string
	Handler http.Handler
}

// ListenAndServe ListenAndServe
func ListenAndServe(sd *grpc.ServiceDesc, ss interface{}, conf *Conf) {
	var serverConf *httpx.ServerConf
	if conf != nil {
		serverConf = &httpx.ServerConf{Addr: conf.Addr, Handler: conf.Handler}
	}
	httpServer := httpx.HTTPServer(serverConf)
	server := NewServer(WithHTTPServer(httpServer))
	server.RegisterService(sd, ss)
	app.AddShutdownHook(func() {
		server.Stop(context.TODO())
	})
	if err := server.ListenAndServe(); err != nil {
		log.WithError(err).
			Fatal("failed to ListenAndServe")
	}
}
