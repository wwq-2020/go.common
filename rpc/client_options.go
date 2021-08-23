package rpc

// ClientOptions ClientOptions
type ClientOptions struct {
	codec           Codec
	balancer        Balancer
	resolverFactory ResolverFactory
}

// ClientOption ClientOption
type ClientOption func(*ClientOptions)

var (
	defaultClientOptions = ClientOptions{
		codec:           JSONCodec(),
		balancer:        NewRandomBalancer(),
		resolverFactory: NewK8SResolver,
	}
)

// ClientWithCodec ClientWithCodec
func ClientWithCodec(codec Codec) ClientOption {
	return func(o *ClientOptions) {
		o.codec = codec
	}
}
