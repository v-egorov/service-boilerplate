package models

import (
	"fmt"

	"github.com/google/uuid"
)

// Custom error types for better error handling
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

type ConflictError struct {
	Resource string
	Field    string
	Value    string
}

func (e ConflictError) Error() string {
	return fmt.Sprintf("%s with %s '%s' already exists", e.Resource, e.Field, e.Value)
}

type NotFoundError struct {
	Resource string
	Field    string
	Value    string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("%s with %s '%s' not found", e.Resource, e.Field, e.Value)
}

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

// Helper functions to create errors
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

// PermissionParseError represents a malformed permission name (bad format/structure).
type PermissionParseError struct {
	Name string // The invalid permission name that was passed
	Reason  string // Human-readable explanation of why parsing failed
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
