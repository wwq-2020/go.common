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

// Conf Conf
type Conf struct {
	Endpoint   string  `json:"endpoint" toml:"endpoint" yaml:"endpoint"`
	SampleRate float64 `json:"sample_rate" toml:"sample_rate" yaml:"sample_rate"`
}

func (c *Conf) fill() {
	if c.SampleRate < 0 || c.SampleRate > 1 {
		c.SampleRate = DefaultSampleRate
	}
}

// MustInitGlobalTracer MustInitGlobalTracer
func MustInitGlobalTracer(serviceName string, conf *Conf) func() {
	cleanup, err := InitGlobalTracer(serviceName, conf)
	if err != nil {
		log.WithError(err).Fatal("failed to InitGlobalTracer")
	}
	return cleanup
}

// InitGlobalTracer InitGlobalTracer
func InitGlobalTracer(serviceName string, conf *Conf) (func(), error) {
	if serviceName == "" {
		return func() {}, errorsx.New("empty serviceName")
	}
	conf.fill()
	cfg := jaegercfg.Configuration{
		ServiceName: serviceName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeProbabilistic,
			Param: conf.SampleRate,
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
		jaegercfg.Reporter(jaeger.NewRemoteReporter(transport.NewHTTPTransport(conf.Endpoint))),
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

func startSpan(ctx context.Context, operationName string) (opentracing.Span, context.Context) {
	options := OptionsFromContext(ctx)
	tracer := opentracing.GlobalTracer()
	if options != nil && options.Root {
		span := tracer.StartSpan(operationName)
		return span, opentracing.ContextWithSpan(ctx, span)
	}
	parentSpan := opentracing.SpanFromContext(ctx)
	if parentSpan == nil {
		span := tracer.StartSpan(operationName)
		return span, opentracing.ContextWithSpan(ctx, span)
	}

	tracingOptions := make([]opentracing.StartSpanOption, 0, 1)
	if options == nil {
		tracingOptions = append(tracingOptions, opentracing.ChildOf(parentSpan.Context()))
		goto start
	}
	switch options.SpanType {
	default:
		fallthrough
	case ChildOfSpanType:
		tracingOptions = append(tracingOptions, opentracing.ChildOf(parentSpan.Context()))
	case FollowsFromSpanType:
		tracingOptions = append(tracingOptions, opentracing.FollowsFrom(parentSpan.Context()))
	}
start:
	span := tracer.StartSpan(operationName, tracingOptions...)
	return span, opentracing.ContextWithSpan(ctx, span)
}

// StartSpan StartSpan
func StartSpan(ctx context.Context, operationName string) (Span, context.Context) {
	opentracingSpan, ctx := startSpan(ctx, operationName)
	tracer := opentracing.GlobalTracer()
	if tracer == noopTracer {
		ctx := log.ContextEnsureTraceIDWithGen(ctx, log.GenTraceID)
		return newSpan(opentracingSpan), ctx
	}
	traceID := traceIDFromOpentracingSpan(opentracingSpan)
	ctx = log.ContextWithTraceID(ctx, traceID)
	return newSpan(opentracingSpan), ctx
}

// HTTPServerStartSpan HTTPServerStartSpan
func HTTPServerStartSpan(ctx context.Context, operationName string, httpReq *http.Request, w http.ResponseWriter) (Span, context.Context) {
	options := OptionsFromContext(ctx)

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
	if options == nil {
		ctx = ContextWithRootOptions(ctx, true)
	}
	span, ctx := StartSpan(ctx, operationName)
	span.InjectToHTTPResponse(ctx, w)
	return span, ctx
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

// Options Options
type Options struct {
	Root     bool
	SpanType SpanType
}

var (
	defaultOptions = Options{
		Root:     true,
		SpanType: ChildOfSpanType,
	}
)

// SpanType SpanType
type SpanType int

// SpanTypes
const (
	ChildOfSpanType SpanType = iota
	FollowsFromSpanType
)

type optionsKey struct {
}

// ContextWithOptions ContextWithOptions
func ContextWithOptions(ctx context.Context, opts *Options) context.Context {
	return context.WithValue(ctx, optionsKey{}, opts)
}

// OptionsFromContext OptionsFromContext
func OptionsFromContext(ctx context.Context) *Options {
	v := ctx.Value(optionsKey{})
	if v == nil {
		return nil
	}
	options, ok := v.(*Options)
	if !ok {
		return nil
	}
	return options
}

// ContextEnsureOptions ContextEnsureOptions
func ContextEnsureOptions(ctx context.Context) context.Context {
	ctx, _ = ContextEnsureOptionsx(ctx)
	return ctx
}

// ContextEnsureOptionsx ContextEnsureOptionsx
func ContextEnsureOptionsx(ctx context.Context) (context.Context, *Options) {
	options := defaultOptions
	return ContextEnsureOptionsWithOptionsx(ctx, &options)
}

// ContextEnsureOptionsWithOptions ContextEnsureOptionsWithOptionsContext
func ContextEnsureOptionsWithOptions(ctx context.Context, options *Options) context.Context {
	ctx, _ = ContextEnsureOptionsWithOptionsx(ctx, options)
	return ctx
}

// ContextEnsureOptionsWithOptionsx ContextEnsureOptionsWithOptionsx
func ContextEnsureOptionsWithOptionsx(ctx context.Context, options *Options) (context.Context, *Options) {
	curOptions := OptionsFromContext(ctx)
	if curOptions == nil {
		return ContextWithOptions(ctx, options), options
	}
	return ctx, curOptions
}

// ContextWithRootOptions ContextWithRootOptions
func ContextWithRootOptions(ctx context.Context, root bool) context.Context {
	ctx, _ = ContextWithRootOptionsx(ctx, root)
	return ctx
}

// ContextWithRootOptionsx ContextWithRootOptionsx
func ContextWithRootOptionsx(ctx context.Context, root bool) (context.Context, *Options) {
	ctx, options := ContextEnsureOptionsx(ctx)
	options.Root = root
	return ctx, options
}

// ContextWithSpanTypeOptions ContextWithSpanTypeOptions
func ContextWithSpanTypeOptions(ctx context.Context, spanType SpanType) context.Context {
	ctx, _ = ContextWithSpanTypeOptionsx(ctx, spanType)
	return ctx
}

// ContextWithSpanTypeOptionsx ContextWithSpanTypeOptionsx
func ContextWithSpanTypeOptionsx(ctx context.Context, spanType SpanType) (context.Context, *Options) {
	ctx, options := ContextEnsureOptionsx(ctx)
	options.SpanType = spanType
	return ctx, options
}
