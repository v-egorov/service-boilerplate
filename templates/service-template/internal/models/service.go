package models

import (
	"time"
)

// Service represents a service entity
type Service struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	// Add more fields as needed
}

// TableName returns the table name for the Service model
func (Service) TableName() string {
	return "services"
}

// CreateServiceRequest represents the request payload for creating a service
type CreateServiceRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=500"`
	// Add more fields as needed
}

// UpdateServiceRequest represents the request payload for updating a service
type UpdateServiceRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	// Add more fields as needed
}

// ServiceResponse represents the response payload for service operations
type ServiceResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	// Add more fields as needed
}

// ToResponse converts a Service model to ServiceResponse
func (s *Service) ToResponse() *ServiceResponse {
	return &ServiceResponse{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		CreatedAt:   s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   s.UpdatedAt.Format(time.RFC3339),
	}
}
