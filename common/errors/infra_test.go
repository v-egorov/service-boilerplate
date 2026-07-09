package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestSentinels_AreNonNil(t *testing.T) {
	sentinels := map[string]error{
		"ErrNotFound":      ErrNotFound,
		"ErrAlreadyExists": ErrAlreadyExists,
		"ErrInvalidInput":  ErrInvalidInput,
	}

	for name, sentinel := range sentinels {
		if sentinel == nil {
			t.Errorf("%s is nil — must be a non-nil sentinel", name)
		}
	}
}

func TestSentinels_ErrorString(t *testing.T) {
	tests := []struct {
		sentinel error
		want     string
	}{
		{ErrNotFound, "resource not found"},
		{ErrAlreadyExists, "resource already exists"},
		{ErrInvalidInput, "invalid input"},
	}

	for _, tt := range tests {
		if got := tt.sentinel.Error(); got != tt.want {
			t.Errorf("Error() = %q, want %q", got, tt.want)
		}
	}
}

func TestSentinels_IsMatch(t *testing.T) {
	sentinels := []error{ErrNotFound, ErrAlreadyExists, ErrInvalidInput}

	for _, s := range sentinels {
		if !errors.Is(s, s) {
			t.Errorf("errors.Is(%T, %T) should be true for same sentinel", s, s)
		}
	}
}

func TestSentinels_Wrapped_IsMatch(t *testing.T) {
	tests := []struct {
		wrapped  error
		target   error
	}{
		{fmt.Errorf("repo: %w", ErrNotFound), ErrNotFound},
		{fmt.Errorf("service: %w", ErrAlreadyExists), ErrAlreadyExists},
		{fmt.Errorf("handler: %w", ErrInvalidInput), ErrInvalidInput},
		{fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", ErrNotFound)), ErrNotFound},
	}

	for _, tt := range tests {
		if !errors.Is(tt.wrapped, tt.target) {
			t.Errorf("errors.Is(%q, %v) should be true — sentinel must match through wrapping", tt.wrapped.Error(), tt.target)
		}
	}
}

func TestSentinels_CrossIsNotMatch(t *testing.T) {
	if errors.Is(ErrNotFound, ErrAlreadyExists) {
		t.Error("ErrNotFound should not match ErrAlreadyExists")
	}
	if errors.Is(ErrInvalidInput, ErrNotFound) {
		t.Error("ErrInvalidInput should not match ErrNotFound")
	}
	if errors.Is(ErrAlreadyExists, ErrInvalidInput) {
		t.Error("ErrAlreadyExists should not match ErrInvalidInput")
	}
}
