package interceptor

import "context"

// MethodHandler MethodHandler
type MethodHandler func(ctx context.Context, dec func(interface{}) error, interceptor ServerInterceptor) (resp interface{}, err error)

// ServerHandler ServerHandler
type ServerHandler func(ctx context.Context, req interface{}) (resp interface{}, err error)

// ServerInterceptor ServerInterceptor
type ServerInterceptor func(ctx context.Context, req interface{}, handler ServerHandler) (resp interface{}, err error)

// ChainServerInerceptor ChainServerInerceptor
func ChainServerInerceptor(interceptors ...ServerInterceptor) ServerInterceptor {
	n := len(interceptors)
	return func(ctx context.Context, req interface{}, handler ServerHandler) (interface{}, error) {
		chainer := func(currentInterceptor ServerInterceptor, currentHandler ServerHandler) ServerHandler {
			return func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
				return currentInterceptor(currentCtx, currentReq, currentHandler)
			}
		}
		chainedHandler := handler
		for i := n - 1; i >= 0; i-- {
			chainedHandler = chainer(interceptors[i], chainedHandler)
		}
		return chainedHandler(ctx, req)
	}
}
