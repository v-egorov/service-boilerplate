package models

import (
	"testing"
	"time"
)

func TestEntity_ToResponse(t *testing.T) {
	now := time.Now()
	entity := &Entity{
		ID:          1,
		Name:        "Test Entity",
		Description: "Test Description",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	response := entity.ToResponse()

	if response.ID != 1 {
		t.Errorf("Expected ID 1, got %d", response.ID)
	}
	if response.Name != "Test Entity" {
		t.Errorf("Expected Name 'Test Entity', got %s", response.Name)
	}
	if response.Description != "Test Description" {
		t.Errorf("Expected Description 'Test Description', got %s", response.Description)
	}
	if response.CreatedAt == "" {
		t.Error("Expected CreatedAt to be set")
	}
	if response.UpdatedAt == "" {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("field", "message")
	if err.Field != "field" {
		t.Errorf("Expected Field 'field', got %s", err.Field)
	}
	if err.Message != "message" {
		t.Errorf("Expected Message 'message', got %s", err.Message)
	}
	expectedMsg := "validation error on field 'field': message"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestNewConflictError(t *testing.T) {
	err := NewConflictError("resource", "field", "value")
	expectedMsg := "resource with field 'value' already exists"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("resource", "field", "value")
	expectedMsg := "resource with field 'value' not found"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestNewInternalError(t *testing.T) {
	innerErr := NewValidationError("test", "error")
	err := NewInternalError("operation", innerErr)
	expectedMsg := "internal error during operation: validation error on field 'test': error"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
	if err.Unwrap() != innerErr {
		t.Error("Expected Unwrap to return inner error")
	}
}
