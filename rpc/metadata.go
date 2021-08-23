package rpc

import (
	"context"
	"net/http"
)

// Metadata Metadata
type Metadata map[string][]string

// Add Add
func (m Metadata) Add(k, v string) Metadata {
	http.Header(m).Add(k, v)
	return m
}

// Merge Merge
func (m Metadata) Merge(givenMetadata Metadata) Metadata {
	newMetadata := make(Metadata)
	for k, v := range m {
		newMetadata[k] = v
	}
	for k, v := range givenMetadata {
		newMetadata[k] = v
	}
	return newMetadata
}

// NewMetadata NewMetadata
func NewMetadata() Metadata {
	return make(Metadata)
}

type incomingMetadataKey struct{}

type outgoingMetadataKey struct{}

// IncomingMetadataFromContext IncomingMetadataFromContext
func IncomingMetadataFromContext(ctx context.Context) Metadata {
	val := ctx.Value(incomingMetadataKey{})
	if val == nil {
		return NewMetadata()
	}
	return val.(Metadata)
}

// ContextWithIncomingMetadata ContextWithIncomingMetadata
func ContextWithIncomingMetadata(ctx context.Context, givenMetadata Metadata) context.Context {
	curMetadata := IncomingMetadataFromContext(ctx)
	newMetadata := curMetadata.Merge(givenMetadata)
	return context.WithValue(ctx, incomingMetadataKey{}, newMetadata)
}

// OutgoingMetadataFromContext OutgoingMetadataFromContext
func OutgoingMetadataFromContext(ctx context.Context) Metadata {
	val := ctx.Value(outgoingMetadataKey{})
	if val == nil {
		return NewMetadata()
	}
	return val.(Metadata)
}

// ContextWithOutgoingMetadata ContextWithOutgoingMetadata
func ContextWithOutgoingMetadata(ctx context.Context, givenMetadata Metadata) context.Context {
	curMetadata := OutgoingMetadataFromContext(ctx)
	newMetadata := curMetadata.Merge(givenMetadata)
	return context.WithValue(ctx, outgoingMetadataKey{}, newMetadata)
}
