package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CreateObjectRequest represents the request payload for creating an object
type CreateObjectRequest struct {
	ObjectTypeID   int64                  `json:"object_type_id" binding:"required" validate:"required,gt=0"`
	ParentObjectID *int64                 `json:"parent_object_id,omitempty" validate:"omitempty,gt=0"`
	Name           string                 `json:"name" binding:"required" validate:"required,min=1,max=255"`
	Description    string                 `json:"description,omitempty" validate:"max=1000"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	CreatedBy      string                 `json:"-" db:"created_by"`
}

// UpdateObjectRequest represents the request payload for updating an object
type UpdateObjectRequest struct {
	ObjectTypeID   *int64                  `json:"object_type_id,omitempty" validate:"omitempty,gt=0"`
	ParentObjectID *int64                  `json:"parent_object_id,omitempty" validate:"omitempty,gt=0"`
	Name           *string                 `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description    *string                 `json:"description,omitempty" validate:"omitempty,max=1000"`
	Metadata       *map[string]interface{} `json:"metadata,omitempty"`
	Tags           *[]string               `json:"tags,omitempty"`
	Status         *string                 `json:"status,omitempty" validate:"omitempty,oneof=active inactive archived deleted pending"`
	Version        *int64                  `json:"version,omitempty" validate:"omitempty,gt=0"`
	UpdatedBy      string                  `json:"-" db:"updated_by"`
}

// ReplaceObjectRequest represents the request payload for replacing an object
type ReplaceObjectRequest struct {
	ObjectTypeID   int64                  `json:"object_type_id" binding:"required" validate:"required,gt=0"`
	ParentObjectID *int64                 `json:"parent_object_id,omitempty" validate:"omitempty,gt=0"`
	Name           string                 `json:"name" binding:"required" validate:"required,min=1,max=255"`
	Description    string                 `json:"description,omitempty" validate:"max=1000"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	Status         string                 `json:"status,omitempty" validate:"omitempty,oneof=active inactive archived deleted pending"`
}

// ObjectFilter represents query parameters for listing objects
type ObjectFilter struct {
	Name           string     `json:"name,omitempty" form:"name"`
	ObjectTypeID   *int64     `json:"object_type_id,omitempty" form:"object_type_id"`
	ParentObjectID *int64     `json:"parent_object_id,omitempty" form:"parent_object_id"`
	Status         string     `json:"status,omitempty" form:"status"`
	Tags           []string   `json:"tags,omitempty" form:"tags"`
	CreatedAfter   *time.Time `json:"created_after,omitempty" form:"created_after"`
	CreatedBefore  *time.Time `json:"created_before,omitempty" form:"created_before"`
	UpdatedAfter   *time.Time `json:"updated_after,omitempty" form:"updated_after"`
	UpdatedBefore  *time.Time `json:"updated_before,omitempty" form:"updated_before"`
	HasMetadata    bool       `json:"has_metadata,omitempty" form:"has_metadata"`
	MetadataKey    string     `json:"metadata_key,omitempty" form:"metadata_key"`
	MetadataValue  string     `json:"metadata_value,omitempty" form:"metadata_value"`
	Limit          int        `json:"limit,omitempty" form:"limit"`
	Offset         int        `json:"offset,omitempty" form:"offset"`
	SortBy         string     `json:"sort_by,omitempty" form:"sort_by"`
	SortOrder      string     `json:"sort_order,omitempty" form:"sort_order"`
}

// ObjectResponse represents response payload for object operations
type ObjectResponse struct {
	ID             int64                  `json:"id"`
	PublicID       uuid.UUID              `json:"public_id"`
	ObjectTypeID   int64                  `json:"object_type_id"`
	ObjectTypeName string                 `json:"object_type_name,omitempty"`
	ParentObjectID *int64                 `json:"parent_object_id,omitempty"`
	ParentName     *string                `json:"parent_name,omitempty"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	Status         string                 `json:"status"`
	Version        int64                  `json:"version"`
	CreatedBy      string                 `json:"created_by"`
	UpdatedBy      string                 `json:"updated_by"`
	CreatedAt      string                 `json:"created_at"`
	UpdatedAt      string                 `json:"updated_at"`
	DeletedAt      *string                `json:"deleted_at,omitempty"`
	Children       []ObjectResponse       `json:"children,omitempty"`
}

// ObjectListResponse represents paginated response for objects list
type ObjectListResponse struct {
	Data       []ObjectResponse   `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

// PaginationResponse represents pagination information for list responses
type PaginationResponse struct {
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
	Total  int64 `json:"total,omitempty"` // Repository will populate this
}

// ToResponse converts an Object model to ObjectResponse
func (o *Object) ToResponse() *ObjectResponse {
	// Convert Metadata from json.RawMessage to map[string]interface{}
	var metadata map[string]interface{}
	if len(o.Metadata) > 0 {
		if err := json.Unmarshal(o.Metadata, &metadata); err != nil {
			metadata = make(map[string]interface{})
		}
	} else {
		metadata = make(map[string]interface{})
	}

	// Convert deleted_at to string pointer
	var deletedAt *string
	if o.DeletedAt != nil {
		deletedAtStr := o.DeletedAt.Format(time.RFC3339)
		deletedAt = &deletedAtStr
	}

	// Get object type name if available
	var objectTypeName string
	if o.ObjectType != nil {
		objectTypeName = o.ObjectType.Name
	}

	// Get parent name if available (would need eager loading in future)
	var parentName *string
	// TODO: This will be populated when parent object eager loading is implemented
	// if o.ParentObject != nil {
	//     parentName = &o.ParentObject.Name
	// }

	return &ObjectResponse{
		ID:             o.ID,
		PublicID:       o.PublicID,
		ObjectTypeID:   o.ObjectTypeID,
		ObjectTypeName: objectTypeName,
		ParentObjectID: o.ParentObjectID,
		ParentName:     parentName,
		Name:           o.Name,
		Description:    o.Description,
		Metadata:       metadata,
		Tags:           o.Tags,
		Status:         o.Status,
		Version:        o.Version,
		CreatedBy:      o.CreatedBy,
		UpdatedBy:      o.UpdatedBy,
		CreatedAt:      o.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      o.UpdatedAt.Format(time.RFC3339),
		DeletedAt:      deletedAt,
	}
}

// ToMinimalResponse converts an Object model to minimal ObjectResponse (no children)
func (o *Object) ToMinimalResponse() *ObjectResponse {
	// Convert Metadata from json.RawMessage to map[string]interface{}
	var metadata map[string]interface{}
	if len(o.Metadata) > 0 {
		if err := json.Unmarshal(o.Metadata, &metadata); err != nil {
			metadata = make(map[string]interface{})
		}
	} else {
		metadata = make(map[string]interface{})
	}

	// Convert deleted_at to string pointer
	var deletedAt *string
	if o.DeletedAt != nil {
		deletedAtStr := o.DeletedAt.Format(time.RFC3339)
		deletedAt = &deletedAtStr
	}

	// Get object type name if available
	var objectTypeName string
	if o.ObjectType != nil {
		objectTypeName = o.ObjectType.Name
	}

	return &ObjectResponse{
		ID:             o.ID,
		PublicID:       o.PublicID,
		ObjectTypeID:   o.ObjectTypeID,
		ObjectTypeName: objectTypeName,
		ParentObjectID: o.ParentObjectID,
		Name:           o.Name,
		Description:    o.Description,
		Metadata:       metadata,
		Tags:           o.Tags,
		Status:         o.Status,
		Version:        o.Version,
		CreatedBy:      o.CreatedBy,
		UpdatedBy:      o.UpdatedBy,
		CreatedAt:      o.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      o.UpdatedAt.Format(time.RFC3339),
		DeletedAt:      deletedAt,
	}
}
