package models

import (
	"time"
)

// Entity represents a legacy entity model for backward compatibility
// This will be replaced with ObjectType and Object models in Phase 3+
type Entity struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// TableName returns table name for Entity model
func (Entity) TableName() string {
	return "entities"
}

// CreateEntityRequest represents the legacy request payload for creating an entity
type CreateEntityRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=500"`
}

// ReplaceEntityRequest represents the legacy request payload for replacing an entity
type ReplaceEntityRequest struct {
	Name        string `json:"name" binding:"required" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=500"`
}

// UpdateEntityRequest represents the legacy request payload for updating an entity
type UpdateEntityRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
}

// EntityResponse represents the legacy response payload for entity operations
type EntityResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
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
