package rpc

import (
	"github.com/wwq-2020/go.common/errcode"
)

// InvokeOptions InvokeOptions
type InvokeOptions struct {
	metadata     Metadata
	expectedCode errcode.ErrCode
	codec        Codec
}

// Clone Clone
func (o *InvokeOptions) Clone() InvokeOptions {
	return InvokeOptions{
		metadata:     o.metadata.Clone(),
		expectedCode: o.expectedCode,
		codec:        o.codec,
	}
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
func InvokeWithExpectedCode(expectedCode errcode.ErrCode) InvokeOption {
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

var (
	defaultInvokeOptions = InvokeOptions{
		metadata:     NewMetadata(),
		expectedCode: errcode.ErrCode_Ok,
		codec:        JSONCodec(),
	}
)
