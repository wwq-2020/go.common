package httpx

import (
	"net/http"

	"github.com/wwq-2020/go.common/errors"
	"github.com/wwq-2020/go.common/stack"
)

// ReqInterceptor ReqInterceptor
type ReqInterceptor func(*http.Request) error

// RespInterceptor RespInterceptor
type RespInterceptor func(*http.Response) error

// ContentTypeReqInterceptor ContentTypeReqInterceptor
func ContentTypeReqInterceptor(contentType string) ReqInterceptor {
	return func(httpReq *http.Request) error {
		httpReq.Header.Set(ContentTypeHeader, contentType)
		return nil
	}
}

// ChainedReqInterceptor ChainedReqInterceptor
func ChainedReqInterceptor(reqInterceptors ...ReqInterceptor) ReqInterceptor {
	return func(httpReq *http.Request) error {
		for _, reqInterceptor := range reqInterceptors {
			if err := reqInterceptor(httpReq); err != nil {
				return errors.Trace(err)
			}
		}
		return nil
	}
}

// StatusCodeRespInterceptor StatusCodeRespInterceptor
func StatusCodeRespInterceptor(expected int) RespInterceptor {
	return func(httpResp *http.Response) error {
		got := httpResp.StatusCode
		if got == expected {
			return nil
		}
		stack := stack.New().Set("expected statuscode", expected).Set("got statuscode", got)
		return errors.NewWithFields("statuscode mismatch", stack)
	}
}

// StatusCodeRangeRespInterceptor StatusCodeRangeRespInterceptor
func StatusCodeRangeRespInterceptor(codeStart, codeEnd int) RespInterceptor {
	return func(httpResp *http.Response) error {
		got := httpResp.StatusCode
		if got < codeStart || got > codeEnd {
			stack := stack.New().Set("expected statuscode start", codeStart).Set("expected statuscode end", codeEnd)
			return errors.NewWithFields("statuscode mismatch", stack)
		}
		return nil
	}
}

// ChainedRespInterceptor ChainedRespInterceptor
func ChainedRespInterceptor(respInterceptors ...RespInterceptor) RespInterceptor {
	return func(httpResp *http.Response) error {
		for _, respInterceptor := range respInterceptors {
			if err := respInterceptor(httpResp); err != nil {
				return errors.Trace(err)
			}
		}
		return nil
	}
}
