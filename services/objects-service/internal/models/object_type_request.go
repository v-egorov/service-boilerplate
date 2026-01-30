package models

import (
	"encoding/json"
	"time"
)

// CreateObjectTypeRequest represents the request payload for creating an object type
type CreateObjectTypeRequest struct {
	Name              string                 `json:"name" binding:"required" validate:"required,min=1,max=255"`
	ParentTypeID      *int64                 `json:"parent_type_id,omitempty" validate:"omitempty,gt=0"`
	ConcreteTableName *string                `json:"concrete_table_name,omitempty" validate:"omitempty,min=1,max=255"`
	Description       string                 `json:"description,omitempty" validate:"max=1000"`
	IsSealed          *bool                  `json:"is_sealed,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateObjectTypeRequest represents the request payload for updating an object type
type UpdateObjectTypeRequest struct {
	Name              *string                 `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	ParentTypeID      *int64                  `json:"parent_type_id,omitempty" validate:"omitempty,gt=0"`
	ConcreteTableName *string                 `json:"concrete_table_name,omitempty" validate:"omitempty,min=1,max=255"`
	Description       *string                 `json:"description,omitempty" validate:"omitempty,max=1000"`
	IsSealed          *bool                   `json:"is_sealed,omitempty"`
	Metadata          *map[string]interface{} `json:"metadata,omitempty"`
}

// ReplaceObjectTypeRequest represents the request payload for replacing an object type
type ReplaceObjectTypeRequest struct {
	Name              string                 `json:"name" binding:"required" validate:"required,min=1,max=255"`
	ParentTypeID      *int64                 `json:"parent_type_id,omitempty" validate:"omitempty,gt=0"`
	ConcreteTableName *string                `json:"concrete_table_name,omitempty" validate:"omitempty,min=1,max=255"`
	Description       string                 `json:"description,omitempty" validate:"max=1000"`
	IsSealed          *bool                  `json:"is_sealed,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// ObjectTypeFilter represents query parameters for listing object types
type ObjectTypeFilter struct {
	Name        string `json:"name,omitempty" form:"name"`
	ParentID    *int64 `json:"parent_id,omitempty" form:"parent_id"`
	IsSealed    *bool  `json:"is_sealed,omitempty" form:"is_sealed"`
	HasConcrete *bool  `json:"has_concrete,omitempty" form:"has_concrete"`
	Limit       int    `json:"limit,omitempty" form:"limit"`
	Offset      int    `json:"offset,omitempty" form:"offset"`
	SortBy      string `json:"sort_by,omitempty" form:"sort_by"`
	SortOrder   string `json:"sort_order,omitempty" form:"sort_order"`
}

// ObjectTypeResponse represents the response payload for object type operations
type ObjectTypeResponse struct {
	ID                int64                  `json:"id"`
	Name              string                 `json:"name"`
	ParentTypeID      *int64                 `json:"parent_type_id,omitempty"`
	ParentName        *string                `json:"parent_name,omitempty"`
	ConcreteTableName *string                `json:"concrete_table_name,omitempty"`
	Description       string                 `json:"description,omitempty"`
	IsSealed          bool                   `json:"is_sealed"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt         string                 `json:"created_at"`
	UpdatedAt         string                 `json:"updated_at"`
	Children          []ObjectTypeResponse   `json:"children,omitempty"`
	ObjectCount       *int64                 `json:"object_count,omitempty"`
}

// ObjectTypeListResponse represents the paginated response for object types list
type ObjectTypeListResponse struct {
	Data       []ObjectTypeResponse `json:"data"`
	Pagination PaginationResponse   `json:"pagination"`
}

// ToResponse converts an ObjectType model to ObjectTypeResponse
func (ot *ObjectType) ToResponse() *ObjectTypeResponse {
	// Convert Metadata from json.RawMessage to map[string]interface{}
	var metadata map[string]interface{}
	if len(ot.Metadata) > 0 {
		if err := json.Unmarshal(ot.Metadata, &metadata); err != nil {
			metadata = make(map[string]interface{})
		}
	} else {
		metadata = make(map[string]interface{})
	}

	// Convert parent type name if available
	var parentName *string
	if ot.ParentType != nil {
		parentName = &ot.ParentType.Name
	}

	// Convert children recursively
	var children []ObjectTypeResponse
	for _, child := range ot.Children {
		childResp := child.ToResponse()
		children = append(children, *childResp)
	}

	// Set object count if objects are loaded
	var objectCount *int64
	if len(ot.Objects) > 0 {
		count := int64(len(ot.Objects))
		objectCount = &count
	}

	return &ObjectTypeResponse{
		ID:                ot.ID,
		Name:              ot.Name,
		ParentTypeID:      ot.ParentTypeID,
		ParentName:        parentName,
		ConcreteTableName: ot.ConcreteTableName,
		Description:       ot.Description,
		IsSealed:          ot.IsSealed,
		Metadata:          metadata,
		CreatedAt:         ot.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         ot.UpdatedAt.Format(time.RFC3339),
		Children:          children,
		ObjectCount:       objectCount,
	}
}

// ToMinimalResponse converts an ObjectType model to minimal ObjectTypeResponse (no children)
func (ot *ObjectType) ToMinimalResponse() *ObjectTypeResponse {
	// Convert Metadata from json.RawMessage to map[string]interface{}
	var metadata map[string]interface{}
	if len(ot.Metadata) > 0 {
		if err := json.Unmarshal(ot.Metadata, &metadata); err != nil {
			metadata = make(map[string]interface{})
		}
	} else {
		metadata = make(map[string]interface{})
	}

	// Convert parent type name if available
	var parentName *string
	if ot.ParentType != nil {
		parentName = &ot.ParentType.Name
	}

	return &ObjectTypeResponse{
		ID:                ot.ID,
		Name:              ot.Name,
		ParentTypeID:      ot.ParentTypeID,
		ParentName:        parentName,
		ConcreteTableName: ot.ConcreteTableName,
		Description:       ot.Description,
		IsSealed:          ot.IsSealed,
		Metadata:          metadata,
		CreatedAt:         ot.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         ot.UpdatedAt.Format(time.RFC3339),
	}
}
