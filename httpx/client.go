package httpx

import "net/http"

// DefaultClient DefaultClient
func DefaultClient() *http.Client {
	return &http.Client{
		Transport: DefaultTransport(),
	}
}

// ClientConf ClientConf
type ClientConf struct {
	TransportConf *TransportConf
	MaxRetry      int
}

var (
	defaultClientConf = &ClientConf{}
)

func (c *ClientConf) fill() {
	if c.TransportConf == nil {
		c.TransportConf = defaultTransportConf
	}
}

// Client Client
func Client(clientConf *ClientConf) *http.Client {
	if clientConf == nil {
		clientConf = defaultClientConf
	}
	clientConf.fill()
	return &http.Client{
		Transport: Transport(clientConf.TransportConf),
	}
}
