package httpx

import "net/http"

// DefaultClient DefaultClient
func DefaultClient() *http.Client {
	return Client(DefaultTransport())
}

// ClientConf ClientConf
type ClientConf struct {
	TransportConf *TransportConf `toml:"transport" yaml:"transport" json:"transport"`
}

var (
	defaultClientConf = &ClientConf{
		TransportConf: defaultTransportConf,
	}
)

func (c *ClientConf) fill() {
	if c.TransportConf == nil {
		c.TransportConf = defaultTransportConf
	}
}

// Client Client
func Client(transport http.RoundTripper) *http.Client {
	return &http.Client{
		Transport: transport,
	}
}

// MakeClient MakeClient
func MakeClient(clientConf *ClientConf) *http.Client {
	if clientConf == nil {
		clientConf = defaultClientConf
	}
	clientConf.fill()
	transport := MakeTransport(clientConf.TransportConf)
	return Client(transport)
}
