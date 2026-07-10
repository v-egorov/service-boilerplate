// Package models provides re-export aliases for typed error types that are now
// defined in common/errors. These aliases maintain backward compatibility — any code
// importing user-service/models and using ValidationError, ConflictError, etc.
// continues working because Go type aliases have identical struct identity to the
// original types. This allows existing type-switch dispatchers (e.g., in handlers)
// to match on the aliased types without changes.
//
// New code SHOULD import directly from common/errors:
//   errors.NewNotFoundError("user", "email", "...")
package models

import errors "github.com/v-egorov/service-boilerplate/common/errors"

// Re-export typed error types and constructors from common/errors for backward
// compatibility. Type aliases ensure Go struct identity is preserved (type-switches work).
// Constructor function aliases mean existing call sites like models.NewValidationError()
// continue working without import changes.
type ValidationError = errors.ValidationError
type ConflictError = errors.ConflictError
type NotFoundError = errors.NotFoundError
type InternalError = errors.InternalError

var NewValidationError = errors.NewValidationError
var NewConflictError = errors.NewConflictError
var NewNotFoundError = errors.NewNotFoundError
var NewInternalError = errors.NewInternalError
