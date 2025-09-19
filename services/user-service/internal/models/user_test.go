package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUserModel(t *testing.T) {
	user := &User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email to be 'test@example.com', got '%s'", user.Email)
	}

	if user.FirstName != "John" {
		t.Errorf("Expected first name to be 'John', got '%s'", user.FirstName)
	}
}

func TestCreateUserRequestValidation(t *testing.T) {
	req := &CreateUserRequest{
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}

	if req.Email == "" {
		t.Error("Email should not be empty")
	}

	if req.FirstName == "" {
		t.Error("First name should not be empty")
	}
}
