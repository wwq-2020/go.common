package tracing

import (
	"context"
	"net/http"

	"github.com/opentracing/opentracing-go"
	opentracinglog "github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
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
		Gen128Bit: true,
		Headers: &jaeger.HeadersConfig{
			TraceContextHeaderName: "TraceID",
		},
	}
	closer, err := cfg.InitGlobalTracer(
		serviceName,
		jaegercfg.Logger(&jaegerLogger{}),
		// jaegercfg.Metrics(&jaegerMetrics{}),
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
	FinishWithFields(err *error, fields stack.Fields)
	WithField(string, interface{}) Span
	WithFields(stack.Fields) Span
	TraceID() string
	InjectToHTTPReq(*http.Request)
}

type span struct {
	opentracingSpan opentracing.Span
	stack           stack.Fields
}

func newSpan(opentracingSpan opentracing.Span) Span {
	return &span{
		opentracingSpan: opentracingSpan,
		stack:           stack.New(),
	}
}

func (s *span) InjectToHTTPReq(httpReq *http.Request) {
	ctx := httpReq.Context()
	if err := opentracing.GlobalTracer().Inject(s.opentracingSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(httpReq.Header)); err != nil {
		log.ErrorContext(ctx, err)
	}
}

func (s *span) TraceID() string {
	jaegerSpan := s.opentracingSpan.(*jaeger.Span)
	return jaegerSpan.SpanContext().TraceID().String()
}

func (s *span) Finish(err *error) {
	fields := make([]opentracinglog.Field, 0, 2)
	if err != nil && *err != nil {
		fields = append(fields, opentracinglog.String("status", "failed"))
		fields = append(fields, opentracinglog.String("error", (*err).Error()))
		s.opentracingSpan.LogFields(fields...)
		s.opentracingSpan.LogKV(s.stack.KVsSlice()...)
		s.opentracingSpan.Finish()
		return
	}
	fields = append(fields, opentracinglog.String("status", "success"))
	s.opentracingSpan.LogFields(fields...)
	s.opentracingSpan.LogKV(s.stack.KVsSlice()...)
	s.opentracingSpan.Finish()
}

func (s *span) FinishWithFields(err *error, stack stack.Fields) {
	fields := make([]opentracinglog.Field, 0, 2)
	if err != nil && *err != nil {
		fields = append(fields, opentracinglog.String("status", "failed"))
		fields = append(fields, opentracinglog.String("error", (*err).Error()))
		s.opentracingSpan.LogFields(fields...)
		s.opentracingSpan.LogKV(s.stack.Merge(stack).KVsSlice()...)
		s.opentracingSpan.Finish()
		return
	}
	fields = append(fields, opentracinglog.String("status", "success"))
	s.opentracingSpan.LogFields(fields...)
	s.opentracingSpan.LogKV(s.stack.Merge(stack).KVsSlice()...)
	s.opentracingSpan.Finish()
}

func (s *span) WithField(key string, value interface{}) Span {
	s.stack.Set(key, value)
	return s
}

func (s *span) WithFields(stack stack.Fields) Span {
	s.stack.Merge(stack)
	return s
}

var defaultStartSpanOptions = StartSpanOptions{
	spanReferenceType: opentracing.ChildOfRef,
}

// StartSpanOptions StartSpanOptions
type StartSpanOptions struct {
	spanReferenceType opentracing.SpanReferenceType
	root              bool
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

// Root Root
func Root(root bool) StartSpanOption {
	return func(o *StartSpanOptions) {
		o.root = root
	}
}

func startSpan(ctx context.Context, operationName string, opts ...StartSpanOption) (opentracing.Span, context.Context) {
	options := defaultStartSpanOptions
	for _, opt := range opts {
		opt(&options)
	}
	if options.root {
		span := opentracing.GlobalTracer().StartSpan(operationName)
		return span, opentracing.ContextWithSpan(ctx, span)
	}
	tracer := opentracing.GlobalTracer()
	parentSpan := opentracing.SpanFromContext(ctx)
	if parentSpan == nil {
		span := tracer.StartSpan(operationName)
		return span, opentracing.ContextWithSpan(ctx, span)
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

	span := tracer.StartSpan(operationName, tracingOptions...)
	return span, opentracing.ContextWithSpan(ctx, span)
}

// StartSpan StartSpan
func StartSpan(ctx context.Context, operationName string, opts ...StartSpanOption) (Span, context.Context) {
	opentracingSpan, ctx := startSpan(ctx, operationName, opts...)
	return newSpan(opentracingSpan), ctx
}

// HTTPServerStartSpan HTTPServerStartSpan
func HTTPServerStartSpan(operationName string, httpReq *http.Request) (Span, context.Context) {
	ctx := httpReq.Context()
	parentSpanContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(httpReq.Header))
	if err == nil {
		opentracingSpan := opentracing.GlobalTracer().StartSpan(operationName, opentracing.ChildOf(parentSpanContext))
		return newSpan(opentracingSpan), opentracing.ContextWithSpan(ctx, opentracingSpan)
	}
	return StartSpan(ctx, operationName, Root(true))
}
