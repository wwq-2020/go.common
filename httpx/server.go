package httpx

import (
	"net/http"
	"time"
)

// consts
const (
	DefaultServerAddr              = "127.0.0.1:8080"
	DefaultServerReadTimeout       = time.Second * 5
	DefaultServerReadHeaderTimeout = time.Second * 2
	DefaultServerWriteTimeout      = time.Second * 5
	DefaultServerIdleTimeout       = time.Second * 30
	DefaultServerMaxHeaderBytes    = 1 << 20
)

// ServerConf ServerConf
type ServerConf struct {
	Addr              string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int
	Handler           http.Handler
}

func (s *ServerConf) fill() {
	if s.Addr == "" {
		s.Addr = DefaultServerAddr
	}
	if s.ReadTimeout == 0 {
		s.ReadTimeout = DefaultServerReadTimeout
	}
	if s.ReadHeaderTimeout == 0 {
		s.ReadHeaderTimeout = DefaultServerReadHeaderTimeout
	}
	if s.WriteTimeout == 0 {
		s.WriteTimeout = DefaultServerWriteTimeout
	}
	if s.IdleTimeout == 0 {
		s.IdleTimeout = DefaultServerIdleTimeout
	}
	if s.MaxHeaderBytes == 0 {
		s.MaxHeaderBytes = DefaultServerMaxHeaderBytes
	}
}

var defaultServerConf = &ServerConf{
	Addr:              DefaultServerAddr,
	ReadTimeout:       DefaultServerReadTimeout,
	ReadHeaderTimeout: DefaultServerReadHeaderTimeout,
	WriteTimeout:      DefaultServerWriteTimeout,
	IdleTimeout:       DefaultServerIdleTimeout,
	MaxHeaderBytes:    DefaultServerMaxHeaderBytes,
}

// Server Server
func Server(conf *ServerConf) *http.Server {
	if conf == nil {
		conf = defaultServerConf
	}
	conf.fill()
	return &http.Server{
		Handler:           conf.Handler,
		Addr:              conf.Addr,
		ReadTimeout:       conf.ReadTimeout,
		ReadHeaderTimeout: conf.ReadHeaderTimeout,
		WriteTimeout:      conf.WriteTimeout,
		IdleTimeout:       conf.IdleTimeout,
		MaxHeaderBytes:    conf.MaxHeaderBytes,
	}
}

// DefaultServer DefaultServer
func DefaultServer() *http.Server {
	return Server(defaultServerConf)
}
