package rpc

import (
	"context"
	"net/http"
)

// Metadata Metadata
type Metadata map[string][]string

// Clone Clone
func (m Metadata) Clone() Metadata {
	return Metadata(http.Header(m).Clone())
}

// Add Add
func (m Metadata) Add(k, v string) Metadata {
	http.Header(m).Add(k, v)
	return m
}

// Get Get
func (m Metadata) Get(k string) string {
	return http.Header(m).Get(k)
}

// Merge Merge
func (m Metadata) Merge(givenMetadata Metadata) Metadata {
	newMetadata := NewMetadata()
	for k, vs := range m {
		for _, v := range vs {
			newMetadata.Add(k, v)
		}
	}
	for k, vs := range givenMetadata {
		for _, v := range vs {
			newMetadata.Add(k, v)
		}
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

// consts
const (
	LdapKey  = "ldap"
	TokenKey = "token"
)

// LdapFromIncomingContext LdapFromIncomingContext
func LdapFromIncomingContext(ctx context.Context) string {
	metadata := IncomingMetadataFromContext(ctx)
	return metadata.Get(LdapKey)
}

// TokenFromIncomingContext TokenFromIncomingContext
func TokenFromIncomingContext(ctx context.Context) string {
	metadata := IncomingMetadataFromContext(ctx)
	return metadata.Get(TokenKey)
}

// IncomingContextWithToken IncomingContextWithToken
func IncomingContextWithToken(ctx context.Context, token string) context.Context {
	metadata := NewMetadata().Add(TokenKey, token)
	return ContextWithIncomingMetadata(ctx, metadata)
}

// IncomingContextWithLdap IncomingContextWithLdap
func IncomingContextWithLdap(ctx context.Context, ldap string) context.Context {
	metadata := NewMetadata().Add(LdapKey, ldap)
	return ContextWithIncomingMetadata(ctx, metadata)
}

// LdapFromOutgoingContext LdapFromOutgoingContext
func LdapFromOutgoingContext(ctx context.Context) string {
	metadata := OutgoingMetadataFromContext(ctx)
	return metadata.Get(LdapKey)
}

// TokenFromOutgoingContext TokenFromOutgoingContext
func TokenFromOutgoingContext(ctx context.Context) string {
	metadata := OutgoingMetadataFromContext(ctx)
	return metadata.Get(TokenKey)
}

// OutgoingContextWithToken OutgoingContextWithToken
func OutgoingContextWithToken(ctx context.Context, token string) context.Context {
	metadata := NewMetadata().Add(TokenKey, token)
	return ContextWithOutgoingMetadata(ctx, metadata)
}

// OutgoingContextWithLdap OutgoingContextWithLdap
func OutgoingContextWithLdap(ctx context.Context, ldap string) context.Context {
	metadata := NewMetadata().Add(LdapKey, ldap)
	return ContextWithOutgoingMetadata(ctx, metadata)
}
