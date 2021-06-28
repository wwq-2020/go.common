package rpc

import (
	"context"

	"github.com/wwq-2020/go.common/app"
	"github.com/wwq-2020/go.common/conf"
	"github.com/wwq-2020/go.common/httpx"
	"github.com/wwq-2020/go.common/log"
	"google.golang.org/grpc"
)

// Conf Conf
type Conf struct {
	Server *ServerConf `toml:"server"`
}

// ServerConf ServerConf
type ServerConf struct {
	Addr string `toml:"addr"`
}

// ListenAndServe ListenAndServe
func ListenAndServe(sd *grpc.ServiceDesc, ss interface{}) {
	config := &Conf{}
	conf.MustLoad(config)
	httpServer := httpx.DefaultServer(nil)
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
