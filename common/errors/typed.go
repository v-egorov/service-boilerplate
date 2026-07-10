package errors

import "fmt"

// Typed error structs for cross-service error handling.
// These carry structured metadata (field names, resource identifiers) that
// handlers render into JSON response fields. They are matched via Go's type-switch
// in handler dispatchers, not via errors.Is().
//
// Usage pattern:
//
//	// Service layer — return typed error with field metadata
//	return nil, fmt.Errorf("failed to create user: %w",
//		errors.NewValidationError("email", "already registered"))
//
// Handler type-switch example:
//
//	switch e := err.(type) {
//	case errors.ValidationError:
//	    c.JSON(400, gin.H{"error": e.Error(), "field": e.Field})

// ValidationError represents a validation failure on a specific field.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// ConflictError represents a resource conflict (duplicate).
type ConflictError struct {
	Resource string
	Field    string
	Value    string
}

func (e ConflictError) Error() string {
	return fmt.Sprintf("%s with %s '%s' already exists", e.Resource, e.Field, e.Value)
}

// NotFoundError represents a resource not found with specific identifiers.
type NotFoundError struct {
	Resource string
	Field    string
	Value    string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("%s with %s '%s' not found", e.Resource, e.Field, e.Value)
}

// InternalError represents an internal server error wrapping a cause.
type InternalError struct {
	Operation string
	Err       error
}

func (e InternalError) Error() string {
	return fmt.Sprintf("internal error during %s: %v", e.Operation, e.Err)
}

func (e InternalError) Unwrap() error {
	return e.Err
}

// Constructor functions for typed errors.

func NewValidationError(field, message string) ValidationError {
	return ValidationError{Field: field, Message: message}
}

func NewConflictError(resource, field, value string) ConflictError {
	return ConflictError{Resource: resource, Field: field, Value: value}
}

func NewNotFoundError(resource, field, value string) NotFoundError {
	return NotFoundError{Resource: resource, Field: field, Value: value}
}

func NewInternalError(operation string, err error) InternalError {
	return InternalError{Operation: operation, Err: err}
}
