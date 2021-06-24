package httpx

import (
	"net/http"

	"github.com/wwq-2020/go.common/errors"
	"github.com/wwq-2020/go.common/log"
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

// LoggingReqInterceptor LoggingReqInterceptor
func LoggingReqInterceptor(httpReq *http.Request) error {
	logger := log.WithField("method", httpReq.Method).
		WithField("url", httpReq.URL.String())
	if httpReq.Body != nil {
		reqData, reqBody, err := drainBody(httpReq.Body)
		if err != nil {
			return errors.Trace(err)
		}
		httpReq.Body = reqBody
		logger = logger.WithField("reqData", string(reqData))
	}
	logger.Info("do http req")
	return nil
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

// LoggingRespInterceptor LoggingRespInterceptor
func LoggingRespInterceptor(httpResp *http.Response) error {
	logger := log.WithField("statuscode", httpResp.StatusCode)
	if httpResp.Body != nil {
		respData, respBody, err := drainBody(httpResp.Body)
		if err != nil {
			return errors.Trace(err)
		}
		logger = logger.WithField("respData", string(respData))
		httpResp.Body = respBody
	}

	logger.Info("got http resp")
	return nil
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
