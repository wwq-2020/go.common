package httpx

import (
	"context"
	"crypto/tls"
	"net/http/httptrace"
	"net/textproto"

	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/tracing"
)

type startSpanOptionsKey struct{}

// StartSpanOptionsFromContext StartSpanOptionsFromContext
func StartSpanOptionsFromContext(ctx context.Context) []tracing.StartSpanOption {
	value := ctx.Value(startSpanOptionsKey{})
	if value == nil {
		return nil
	}
	return value.([]tracing.StartSpanOption)
}

// ContextWithStartSpanOptions ContextWithStartSpanOptions
func ContextWithStartSpanOptions(ctx context.Context, opts ...tracing.StartSpanOption) context.Context {
	return context.WithValue(ctx, startSpanOptionsKey{}, opts)
}

// ContextWithClientTrace ContextWithClientTrace
func ContextWithClientTrace(ctx context.Context) context.Context {
	return httptrace.WithClientTrace(ctx, BuildClientTrace(ctx))
}

// BuildClientTrace BuildClientTrace
func BuildClientTrace(ctx context.Context) *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			log.WithField("hostPort", hostPort).
				InfoContext(ctx, "GetConn")
		},
		GotConn: func(gotConnInfo httptrace.GotConnInfo) {
			log.WithField("WasIdle", gotConnInfo.WasIdle).
				WithField("Reused", gotConnInfo.Reused).
				WithField("IdleTime", gotConnInfo.IdleTime.Milliseconds()).
				InfoContext(ctx, "GotConn")
		},
		PutIdleConn: func(err error) {
			if err != nil {
				log.ErrorContext(ctx, err)
			} else {
				log.InfoContext(ctx, "PutIdleConn")
			}
		},
		GotFirstResponseByte: func() {
			log.InfoContext(ctx, "GotFirstResponseByte")
		},
		Got100Continue: func() {
			log.InfoContext(ctx, "Got100Continue")
		},
		Got1xxResponse: func(code int, header textproto.MIMEHeader) error {
			log.WithField("code", code).
				WithField("header", header).
				InfoContext(ctx, "Got1xxResponse")
			return nil
		},
		DNSStart: func(di httptrace.DNSStartInfo) {
			log.WithField("Host", di.Host).
				InfoContext(ctx, "DNSStart")
		},
		DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
			addrs := make([]string, 0, len(dnsInfo.Addrs))
			for _, each := range dnsInfo.Addrs {
				addrs = append(addrs, each.IP.String())
			}
			logger := log.WithField("addrs", addrs).
				WithField("Coalesced", dnsInfo.Coalesced)
			if dnsInfo.Err != nil {
				logger.ErrorContext(ctx, dnsInfo.Err)
			} else {
				logger.InfoContext(ctx, "DNSDone")
			}
		},
		ConnectStart: func(network, addr string) {
			log.WithField("network", network).
				WithField("addr", addr).
				InfoContext(ctx, "ConnectStart")
		},
		ConnectDone: func(network, addr string, err error) {
			logger := log.WithField("network", network).
				WithField("addr", addr)
			if err != nil {
				logger.ErrorContext(ctx, err)
			} else {
				logger.InfoContext(ctx, "ConnectStart")
			}
		},
		TLSHandshakeStart: func() {
			log.InfoContext(ctx, "TLSHandshakeStart")
		},
		TLSHandshakeDone: func(cs tls.ConnectionState, e error) {
			log.InfoContext(ctx, "TLSHandshakeDone")
		},
		WroteHeaderField: func(key string, value []string) {
			log.WithField("key", key).
				WithField("value", value).
				InfoContext(ctx, "WroteHeaderField")
		},
		WroteHeaders: func() {
			log.InfoContext(ctx, "WroteHeaders")
		},
		Wait100Continue: func() {
			log.InfoContext(ctx, "Wait100Continue")
		},
		WroteRequest: func(wri httptrace.WroteRequestInfo) {
			if wri.Err != nil {
				log.ErrorContext(ctx, wri.Err)
			} else {
				log.InfoContext(ctx, "WroteRequest")
			}
		},
	}
}
