package models

import "fmt"

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
