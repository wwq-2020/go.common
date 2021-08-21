package tracing

import (
	"context"

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
	Finish()
	WithField(string, interface{})
	WithFields(stack.Fields)
}

type span struct {
	span opentracing.Span
}

func (s *span) Finish() {
	s.span.Finish()
}

func (s *span) WithField(key string, value interface{}) {
	s.span.LogFields(opentracinglog.Object(key, value))
}

func (s *span) WithFields(fields stack.Fields) {
	s.span.LogKV(fields.KVsSlice()...)
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
		if opt == nil {
			continue
		}
		opt(&options)
	}

	tracingOptions := make([]opentracing.StartSpanOption, 0, 4)

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
