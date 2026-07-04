package client

import (
	"context"
	"net/http"
)

// identityCtxKey is a private type for the context key to avoid collisions.
type identityCtxKey struct{}

// WithIdentity stores an allow-listed set of HTTP headers in the context.
// It is nil-safe: if hdr is nil, it stores nil and IdentityFromContext returns nil.
func WithIdentity(ctx context.Context, hdr http.Header) context.Context {
	return context.WithValue(ctx, identityCtxKey{}, hdr)
}

// IdentityFromContext retrieves stored identity headers from the context.
// Returns nil when no identity was set (e.g., health checks).
func IdentityFromContext(ctx context.Context) http.Header {
	hdr := ctx.Value(identityCtxKey{})
	if hdr == nil {
		return nil
	}
	h, ok := hdr.(http.Header)
	if !ok {
		return nil
	}
	return h
}
