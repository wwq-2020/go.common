package httpx

import (
	"net/http"

	"github.com/wwq-2020/go.common/errorsx"
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
				return errorsx.Trace(err)
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
		return errorsx.New("statuscode mismatch").
			WithField("expected statuscode", expected).
			WithField("got statuscode", got)
	}
}

// StatusCodeRangeRespInterceptor StatusCodeRangeRespInterceptor
func StatusCodeRangeRespInterceptor(codeStart, codeEnd int) RespInterceptor {
	return func(httpResp *http.Response) error {
		got := httpResp.StatusCode
		if got < codeStart || got > codeEnd {
			return errorsx.New("statuscode mismatch").
				WithField("expected statuscode start", codeStart).
				WithField("expected statuscode end", codeEnd)
		}
		return nil
	}
}

// ChainedRespInterceptor ChainedRespInterceptor
func ChainedRespInterceptor(respInterceptors ...RespInterceptor) RespInterceptor {
	return func(httpResp *http.Response) error {
		for _, respInterceptor := range respInterceptors {
			if err := respInterceptor(httpResp); err != nil {
				return errorsx.Trace(err)
			}
		}
		return nil
	}
}
