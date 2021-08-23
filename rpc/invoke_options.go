package rpc

import "github.com/wwq-2020/go.common/tracing"

// InvokeOptions InvokeOptions
type InvokeOptions struct {
	metadata       Metadata
	expectedCode   int
	codec          Codec
	tracingOptions TracingOptions
}

// Clone Clone
func (o *InvokeOptions) Clone() InvokeOptions {
	return InvokeOptions{
		metadata:       o.metadata.Clone(),
		expectedCode:   o.expectedCode,
		codec:          o.codec,
		tracingOptions: o.tracingOptions,
	}
}

// TracingOptions TracingOptions
type TracingOptions struct {
	StartSpanOptions []tracing.StartSpanOption
	Root             bool
	OperationName    string
}

// InvokeOption InvokeOption
type InvokeOption func(*InvokeOptions)

// InvokeWithMetadata InvokeWithMetadata
func InvokeWithMetadata(metadata Metadata) InvokeOption {
	return func(o *InvokeOptions) {
		o.metadata = metadata
	}
}

// InvokeWithExpectedCode InvokeWithExpectedCode
func InvokeWithExpectedCode(expectedCode int) InvokeOption {
	return func(o *InvokeOptions) {
		o.expectedCode = expectedCode
	}
}

// InvokeWithCodec InvokeWithCodec
func InvokeWithCodec(codec Codec) InvokeOption {
	return func(o *InvokeOptions) {
		o.codec = codec
	}
}

// InvokeWithTracingOptions InvokeWithTracingOptions
func InvokeWithTracingOptions(tracingOptions TracingOptions) InvokeOption {
	return func(o *InvokeOptions) {
		o.tracingOptions = tracingOptions
	}
}

var (
	defaultInvokeOptions = InvokeOptions{
		metadata:     NewMetadata(),
		expectedCode: 0,
		codec:        JSONCodec(),
		tracingOptions: TracingOptions{
			Root:             true,
			StartSpanOptions: []tracing.StartSpanOption{tracing.ChildOf()},
		},
	}
)
