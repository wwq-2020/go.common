package interceptor

import (
	"context"
	"fmt"
	"net/http"

	"github.com/wwq-2020/go.common/errorsx"
	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/stack"
)

// Recover Recover
func Recover(ctx context.Context, req interface{}, handler ServerHandler) (_ interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			if e == http.ErrAbortHandler {
				panic(http.ErrAbortHandler)
			}
			switch v := e.(type) {
			case error:
				err = v
			default:
				err = fmt.Errorf("%+v", v)
			}
			stack := stack.Callers(nil)
			log.WithField("stack", stack).
				ErrorContext(ctx, err)
		}
	}()
	resp, err := handler(ctx, req)
	if err != nil {
		return nil, errorsx.Trace(err)
	}
	return resp, nil
}
