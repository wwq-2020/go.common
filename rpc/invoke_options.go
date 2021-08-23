package rpc

// InvokeOptions InvokeOptions
type InvokeOptions struct {
	metadata     Metadata
	expectedCode int
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
