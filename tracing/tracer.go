package tracing

import (
	"context"
	"net/http"
	"net/url"

	"github.com/opentracing/opentracing-go"
	opentracinglog "github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/transport"
	"github.com/wwq-2020/go.common/errorsx"
	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/stack"
)

var (
	noopTracer = opentracing.NoopTracer{}
)

// consts
const (
	DefaultSampleRate = 0.1
	TraceIDName       = "traceid"
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
			TraceContextHeaderName: TraceIDName,
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
	InjectToHTTPReq(*http.Request)
	InjectToHTTPResponse(ctx context.Context, w http.ResponseWriter)
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

func (s *span) InjectToHTTPResponse(ctx context.Context, w http.ResponseWriter) {
	tracer := opentracing.GlobalTracer()
	traceID := log.TraceIDFromContext(ctx)
	if tracer != noopTracer {
		traceID = traceIDFromOpentracingSpan(s.opentracingSpan)
	}
	w.Header().Set(TraceIDName, traceID)
}

func (s *span) InjectToHTTPReq(httpReq *http.Request) {
	ctx := httpReq.Context()
	tracer := opentracing.GlobalTracer()
	if tracer == noopTracer {
		traceID := log.TraceIDFromContext(ctx)
		httpReq.Header.Set(TraceIDName, traceID)
		return
	}
	if err := tracer.Inject(s.opentracingSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(httpReq.Header)); err != nil {
		log.ErrorContext(ctx, err)
	}
}

func (s *span) Finish(err *error) {
	fields := make([]opentracinglog.Field, 0, 2)
	if err != nil && *err != nil {
		fields = append(fields, opentracinglog.String("status", "failed"))
		fields = append(fields, opentracinglog.String("error", (*err).Error()))
		s.opentracingSpan.LogFields(fields...)
		s.opentracingSpan.LogKV(s.stack.Merge(errorsx.Fields(*err)).KVsSlice()...)
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
		s.opentracingSpan.LogKV(s.stack.Merge(errorsx.Fields(*err)).Merge(stack).KVsSlice()...)
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
	tracer := opentracing.GlobalTracer()

	if options.root {
		span := tracer.StartSpan(operationName)
		return span, opentracing.ContextWithSpan(ctx, span)
	}
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
	tracer := opentracing.GlobalTracer()
	if tracer == noopTracer {
		traceID := log.TraceIDFromContext(ctx)
		if traceID == "" {
			ctx = log.ContextWithTraceID(ctx, log.GenTraceID())
		}
	}
	if tracer != noopTracer {
		traceID := traceIDFromOpentracingSpan(opentracingSpan)
		ctx = log.ContextWithTraceID(ctx, traceID)
	}
	return newSpan(opentracingSpan), ctx
}

// HTTPServerStartSpan HTTPServerStartSpan
func HTTPServerStartSpan(ctx context.Context, operationName string, httpReq *http.Request, w http.ResponseWriter) (Span, context.Context) {
	tracer := opentracing.GlobalTracer()
	if tracer == noopTracer {
		traceID := parseTraceID(httpReq)
		ctx = log.ContextWithTraceID(ctx, traceID)
	}
	parentSpanContext, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(httpReq.Header))
	if err == nil {
		opentracingSpan := tracer.StartSpan(operationName, opentracing.ChildOf(parentSpanContext))
		if tracer != noopTracer {
			traceID := traceIDFromOpentracingSpan(opentracingSpan)
			ctx = log.ContextWithTraceID(ctx, traceID)
		}
		span := newSpan(opentracingSpan)
		span.InjectToHTTPResponse(ctx, w)
		return span, opentracing.ContextWithSpan(ctx, opentracingSpan)
	}
	span, ctx := StartSpan(ctx, operationName, Root(true))
	span.InjectToHTTPResponse(ctx, w)
	return span, ctx
}

// HTTPClientStartSpan HTTPClientStartSpan
func HTTPClientStartSpan(ctx context.Context, operationName string, httpReq *http.Request, opts ...StartSpanOption) (Span, context.Context) {
	opentracingSpan, ctx := startSpan(ctx, operationName, opts...)
	tracer := opentracing.GlobalTracer()
	if tracer == noopTracer {
		traceID := log.TraceIDFromContext(ctx)
		if traceID == "" {
			traceID = log.GenTraceID()
			ctx = log.ContextWithTraceID(ctx, traceID)
		}
		httpReq.Header.Set(TraceIDName, traceID)
	}
	if err := tracer.Inject(opentracingSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(httpReq.Header)); err != nil {
		log.ErrorContext(ctx, err)
	}
	return newSpan(opentracingSpan), ctx
}

func parseTraceID(httpReq *http.Request) string {
	val := httpReq.Header.Get(TraceIDName)
	if v, err := url.QueryUnescape(val); err == nil {
		val = v
	}
	if val == "" {
		return log.GenTraceID()
	}
	spanContext, err := jaeger.ContextFromString(val)
	if err == nil {
		return spanContext.TraceID().String()
	}
	return val
}

func traceIDFromOpentracingSpan(opentracingSpan opentracing.Span) string {
	jaegerSpan := opentracingSpan.(*jaeger.Span)
	return jaegerSpan.SpanContext().TraceID().String()
}
