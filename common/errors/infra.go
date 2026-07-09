// Package errors provides shared sentinel error variables for cross-service
// error handling. All services SHOULD import this package and use these
// sentinels as the base of their error chains via fmt.Errorf("...: %w", err).
//
// Handlers match sentinels using errors.Is() which traverses wrapped errors,
// regardless of intermediate custom types. This enables uniform dispatch
// across services while allowing each service to layer domain-specific typed
// errors on top (e.g., user-service's ValidationError struct).
//
// Usage pattern:
//
//	// Repository layer — check for missing row, return sentinel
//	if errors.Is(err, sql.ErrNoRows) {
//	    return nil, fmt.Errorf("failed to get resource: %w", ErrNotFound)
//	}
//
//	// Service layer — wrap with domain context
//	if err != nil {
//	    return nil, fmt.Errorf("operation failed: %w", err)
//	}
package errors

import "fmt"

var (
	ErrNotFound        = fmt.Errorf("resource not found")
	ErrAlreadyExists   = fmt.Errorf("resource already exists")
	ErrInvalidInput    = fmt.Errorf("invalid input")
)
