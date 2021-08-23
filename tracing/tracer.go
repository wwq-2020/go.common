package tracing

import (
	"context"
	"net/http"

	"github.com/opentracing/opentracing-go"
	opentracinglog "github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-client-go/transport"
	"github.com/wwq-2020/go.common/errorsx"
	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/stack"
)

// consts
const (
	DefaultSampleRate = 0.1
)

// MustInitGlobalTracer MustInitGlobalTracer
func MustInitGlobalTracer(serviceName, endpoint string, sampleRate float64) func() {
	cleanup, err := InitGlobalTracer(serviceName, endpoint, sampleRate)
	if err != nil {
		log.WithError(err).Fatal("failed to InitGlobalTracer")
	}
	return cleanup
}

// InitGlobalTracer InitGlobalTracer
func InitGlobalTracer(serviceName, endpoint string, sampleRate float64) (func(), error) {
	if sampleRate < 0 || sampleRate > 1 {
		return nil, errorsx.NewWithField("invalid sampleRate", "sampleRate", sampleRate)
	}
	if sampleRate == 0 {
		sampleRate = DefaultSampleRate
	}

	cfg := jaegercfg.Configuration{
		ServiceName: serviceName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeProbabilistic,
			Param: sampleRate,
		},
	}
	closer, err := cfg.InitGlobalTracer(
		serviceName,
		jaegercfg.Logger(jaegerlog.StdLogger),
		jaegercfg.Gen128Bit(true),
		jaegercfg.Reporter(jaeger.NewRemoteReporter(transport.NewHTTPTransport(endpoint))),
	)
	if err != nil {
		return nil, errorsx.Trace(err)
	}
	return func() {
		if err := closer.Close(); err != nil {
			log.Error(err)
		}
	}, nil
}

// Span Span
type Span interface {
	Finish(err *error)
	WithField(string, interface{}) Span
	WithFields(stack.Fields) Span
}

type span struct {
	span opentracing.Span
}

func (s *span) Finish(err *error) {
	fields := make([]opentracinglog.Field, 0, 2)
	if err != nil && *err != nil {
		fields = append(fields, opentracinglog.String("status", "failed"))
		fields = append(fields, opentracinglog.String("error", (*err).Error()))
		s.span.LogFields(fields...)
		s.span.Finish()
		return
	}
	fields = append(fields, opentracinglog.String("status", "success"))
	s.span.LogFields(fields...)
	s.span.Finish()
}

func (s *span) WithField(key string, value interface{}) Span {
	s.span.LogFields(opentracinglog.Object(key, value))
	return s
}

func (s *span) WithFields(fields stack.Fields) Span {
	s.span.LogKV(fields.KVsSlice()...)
	return s
}

var defaultStartSpanOptions = StartSpanOptions{
	spanReferenceType: opentracing.ChildOfRef,
}

// StartSpanOptions StartSpanOptions
type StartSpanOptions struct {
	spanReferenceType opentracing.SpanReferenceType
}

// StartSpanOption StartSpanOption
type StartSpanOption func(*StartSpanOptions)

// ChildOf ChildOf
func ChildOf() StartSpanOption {
	return func(o *StartSpanOptions) {
		o.spanReferenceType = opentracing.ChildOfRef
	}
}

// FollowsFrom FollowsFrom
func FollowsFrom() StartSpanOption {
	return func(o *StartSpanOptions) {
		o.spanReferenceType = opentracing.FollowsFromRef
	}
}

// StartSpanFromContext StartSpanFromContext
func StartSpanFromContext(ctx context.Context, operationName string, opts ...StartSpanOption) (Span, context.Context) {
	tracer := opentracing.GlobalTracer()
	parentSpan := opentracing.SpanFromContext(ctx)
	if parentSpan == nil {
		opentracingSpan := tracer.StartSpan(operationName)
		return &span{opentracingSpan}, opentracing.ContextWithSpan(ctx, opentracingSpan)
	}

	options := defaultStartSpanOptions
	for _, opt := range opts {
		opt(&options)
	}
	tracingOptions := make([]opentracing.StartSpanOption, 0, 1)

	switch options.spanReferenceType {
	default:
		fallthrough
	case opentracing.ChildOfRef:
		tracingOptions = append(tracingOptions, opentracing.ChildOf(parentSpan.Context()))
	case opentracing.FollowsFromRef:
		tracingOptions = append(tracingOptions, opentracing.FollowsFrom(parentSpan.Context()))
	}

	opentracingSpan := tracer.StartSpan(operationName, tracingOptions...)
	return &span{opentracingSpan}, opentracing.ContextWithSpan(ctx, opentracingSpan)
}

// StartSpan StartSpan
func StartSpan(ctx context.Context, operationName string) (Span, context.Context) {
	opentracingSpan := opentracing.GlobalTracer().StartSpan(operationName)
	return &span{opentracingSpan}, opentracing.ContextWithSpan(ctx, opentracingSpan)
}

// StartSpanFromHTTPReq StartSpanFromHTTPReq
func StartSpanFromHTTPReq(operationName string, httpReq *http.Request) (Span, context.Context) {
	ctx := httpReq.Context()
	parentSpanContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(httpReq.Header))
	if err == nil {
		opentracingSpan := opentracing.GlobalTracer().StartSpan(operationName, opentracing.ChildOf(parentSpanContext))
		return &span{opentracingSpan}, opentracing.ContextWithSpan(ctx, opentracingSpan)
	}
	return StartSpanFromContext(ctx, operationName)
}
