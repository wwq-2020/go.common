package httpx

import (
	"net/http"
	"time"

	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/util"
)

// vars
var (
	DefaultServerAddr                 = "127.0.0.1:8080"
	DefaultServerReadTimeout          = time.Second * 5
	DefaultServerReadTimeoutStr       = DefaultServerReadTimeout.String()
	DefaultServerReadHeaderTimeout    = time.Second * 2
	DefaultServerReadHeaderTimeoutStr = DefaultServerReadHeaderTimeout.String()
	DefaultServerWriteTimeout         = time.Second * 5
	DefaultServerWriteTimeoutStr      = DefaultServerWriteTimeout.String()
	DefaultServerIdleTimeout          = time.Second * 30
	DefaultServerIdleTimeoutStr       = DefaultServerIdleTimeout.String()
	DefaultServerMaxHeaderBytes       = 1 << 20
	DefaultServerMaxHeaderBytesStr    = util.ToByteStr(int64(DefaultServerMaxHeaderBytes))
)

// ServerConf ServerConf
type ServerConf struct {
	Addr              string  `toml:"addr" yaml:"addr" json:"addr"`
	ReadTimeout       *string `toml:"read_timeout" yaml:"read_timeout" json:"read_timeout"`
	ReadHeaderTimeout *string `toml:"read_header_timeout" yaml:"read_header_timeout" json:"read_header_timeout"`
	WriteTimeout      *string `toml:"write_timeout" yaml:"write_timeout" json:"write_timeout"`
	IdleTimeout       *string `toml:"idle_timeout" yaml:"idle_timeout" json:"idle_timeout"`
	MaxHeaderBytes    *string `toml:"max_header_bytes" yaml:"max_header_bytes" json:"max_header_bytes"`
	Handler           http.Handler
}

func (s *ServerConf) fill() {
	if s.Addr == "" {
		s.Addr = DefaultServerAddr
	}
	if s.ReadTimeout == nil || *s.ReadTimeout == "" {
		s.ReadTimeout = &DefaultServerReadTimeoutStr
	}
	if s.ReadHeaderTimeout == nil || *s.ReadHeaderTimeout == "" {
		s.ReadHeaderTimeout = &DefaultServerReadHeaderTimeoutStr
	}
	if s.WriteTimeout == nil || *s.WriteTimeout == "" {
		s.WriteTimeout = &DefaultServerWriteTimeoutStr
	}
	if s.IdleTimeout == nil || *s.IdleTimeout == "" {
		s.IdleTimeout = &DefaultServerIdleTimeoutStr
	}
	if s.MaxHeaderBytes == nil || *s.MaxHeaderBytes == "" {
		s.MaxHeaderBytes = &DefaultServerMaxHeaderBytesStr
	}
}

var defaultServerConf = &ServerConf{
	Addr:              DefaultServerAddr,
	ReadTimeout:       &DefaultServerReadTimeoutStr,
	ReadHeaderTimeout: &DefaultServerReadHeaderTimeoutStr,
	WriteTimeout:      &DefaultServerWriteTimeoutStr,
	IdleTimeout:       &DefaultServerIdleTimeoutStr,
	MaxHeaderBytes:    &DefaultServerMaxHeaderBytesStr,
}

// Server Server
func Server(conf *ServerConf) *http.Server {
	if conf == nil {
		conf = defaultServerConf
	}
	conf.fill()
	server := &http.Server{
		Handler:           conf.Handler,
		Addr:              conf.Addr,
		ReadTimeout:       DefaultServerReadTimeout,
		ReadHeaderTimeout: DefaultServerReadHeaderTimeout,
		WriteTimeout:      DefaultServerWriteTimeout,
		IdleTimeout:       DefaultServerIdleTimeout,
		MaxHeaderBytes:    DefaultServerMaxHeaderBytes,
	}
	readTimeout, err := time.ParseDuration(*conf.ReadTimeout)
	if err == nil && readTimeout != 0 {
		server.ReadTimeout = readTimeout
	}
	if err != nil {
		log.WithField("read_timeout", conf.ReadTimeout).
			Error(err)
	}
	readHeaderTimeout, err := time.ParseDuration(*conf.ReadHeaderTimeout)
	if err == nil && readHeaderTimeout != 0 {
		server.ReadHeaderTimeout = readHeaderTimeout
	}
	if err != nil {
		log.WithField("read_header_timeout", conf.ReadHeaderTimeout).
			Error(err)
	}
	writeTimeout, err := time.ParseDuration(*conf.WriteTimeout)
	if err == nil && writeTimeout != 0 {
		server.WriteTimeout = writeTimeout
	}
	if err != nil {
		log.WithField("write_timeout", conf.WriteTimeout).
			Error(err)
	}
	idleTimeout, err := time.ParseDuration(*conf.IdleTimeout)
	if err == nil && idleTimeout != 0 {
		server.IdleTimeout = idleTimeout
	}
	if err != nil {
		log.WithField("idle_timeout", conf.IdleTimeout).
			Error(err)
	}

	maxHeaderBytes, err := util.ParseByteStr(*conf.MaxHeaderBytes)
	if err == nil && maxHeaderBytes != 0 {
		server.MaxHeaderBytes = int(maxHeaderBytes)
	}
	if err != nil {
		log.WithField("max_header_bytes", conf.MaxHeaderBytes).
			Error(err)
	}
	return server
}

// DefaultServer DefaultServer
func DefaultServer() *http.Server {
	return Server(defaultServerConf)
}
