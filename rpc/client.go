package rpc

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/wwq-2020/go.common/errors"
	"github.com/wwq-2020/go.common/httpx"
	"google.golang.org/protobuf/proto"
)

// ClientOptions ClientOptions
type ClientOptions struct {
	httpClient *http.Client
}

// ClientOption ClientOption
type ClientOption func(*ClientOptions)

// WithHTTPClient WithHTTPClient
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(o *ClientOptions) {
		o.httpClient = httpClient
	}
}

var defaultClientOptions = ClientOptions{
	httpClient: httpx.RetriableClient(),
}

// Client Client
type Client interface {
	Invoke(ctx context.Context, method string, in, out proto.Message) (err error)
}

type client struct {
	httpClient *http.Client
	addr       string
}

// NewClient NewClient
func NewClient(addr string, opts ...ClientOption) Client {
	options := defaultClientOptions
	for _, opt := range opts {
		opt(&options)
	}
	return &client{
		addr:       addr,
		httpClient: options.httpClient,
	}
}

func (c *client) Invoke(ctx context.Context, method string, in, out proto.Message) (err error) {
	url := fmt.Sprintf("http://%s%s", c.addr, method)
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
	}
	if err := httpx.Post(ctx, url, in, out, httpx.WithClient(c.httpClient)); err != nil {
		return errors.TraceWithField(err, "method", method)
	}
	return nil
}
