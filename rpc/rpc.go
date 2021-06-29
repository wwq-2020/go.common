package rpc

import (
	"context"
	"net/http"

	"github.com/wwq-2020/go.common/app"
	"github.com/wwq-2020/go.common/log"
	"google.golang.org/grpc"
)

// Conf Conf
type Conf struct {
	Addr    string
	Handler http.Handler
}

// ListenAndServe ListenAndServe
func ListenAndServe(sd *grpc.ServiceDesc, ss interface{}, opts ...ServerOption) {
	server := NewServer(opts...)
	server.RegisterService(sd, ss)
	app.GoAsync(func() {
		if err := server.ListenAndServe(); err != nil {
			log.WithError(err).
				Fatal("failed to ListenAndServe")
		}
	})
	app.AddShutdownHook(func() {
		server.Stop(context.TODO())
	})
	app.Wait()
}
