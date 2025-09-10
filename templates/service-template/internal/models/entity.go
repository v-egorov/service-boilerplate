package models

import (
	"time"
)

// Entity represents a generic entity that can be customized for different object types
type Entity struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	// Add more fields as needed for specific entity types
}

// TableName returns the table name for the Entity model
func (Entity) TableName() string {
	return "entities" // This will be replaced with the actual table name
}

// CreateEntityRequest represents the request payload for creating an entity
type CreateEntityRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=500"`
	// Add more fields as needed for specific entity types
}

// UpdateEntityRequest represents the request payload for updating an entity
type UpdateEntityRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	// Add more fields as needed for specific entity types
}

// EntityResponse represents the response payload for entity operations
type EntityResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	// Add more fields as needed for specific entity types
}

// ToResponse converts an Entity model to EntityResponse
func (e *Entity) ToResponse() *EntityResponse {
	return &EntityResponse{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		CreatedAt:   e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   e.UpdatedAt.Format(time.RFC3339),
	}
}
