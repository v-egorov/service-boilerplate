// Package models provides re-export aliases for typed error types that are now
// defined in common/errors. These aliases maintain backward compatibility — any code
// importing auth-service/models and using ValidationError, ConflictError, etc.
// continues working because Go type aliases have identical struct identity to the
// original types. This allows existing errors.As() dispatchers (e.g., in handlers)
// to match on the aliased types without changes.
//
// New code SHOULD import directly from common/errors:
//   errors.NewNotFoundError("user", "email", "...")
package models

import (
	"fmt"

	errors_pkg "github.com/v-egorov/service-boilerplate/common/errors"
	"github.com/google/uuid"
)

// Re-export typed error types and constructors from common/errors for backward compatibility.
type ValidationError = errors_pkg.ValidationError
type ConflictError = errors_pkg.ConflictError
type NotFoundError = errors_pkg.NotFoundError
type InternalError = errors_pkg.InternalError

var NewValidationError = errors_pkg.NewValidationError
var NewConflictError = errors_pkg.NewConflictError
var NewNotFoundError = errors_pkg.NewNotFoundError
var NewInternalError = errors_pkg.NewInternalError

// PermissionParseError represents a malformed permission name (bad format/structure).
type PermissionParseError struct {
	Name   string // The invalid permission name that was passed
	Reason string // Human-readable explanation of why parsing failed
}

func (e PermissionParseError) Error() string {
	return fmt.Sprintf("invalid permission format %q: %s", e.Name, e.Reason)
}

// ScopedVariantConflictError represents a scope conflict between two permissions in the same role.
type ScopedVariantConflictError struct {
	RoleID      uuid.UUID // Role where the conflict was detected
	Permission1 string    // First conflicting permission name
	Permission2 string    // Second conflicting permission name
}

func (e ScopedVariantConflictError) Error() string {
	return fmt.Sprintf("scoped variant conflict in role %s: %q conflicts with %q", e.RoleID, e.Permission1, e.Permission2)
}
