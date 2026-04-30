package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Relationship struct {
	ObjectID             int64           `json:"object_id" db:"object_id"`
	PublicID             uuid.UUID       `json:"public_id" db:"public_id"`
	SourceObjectID       int64           `json:"source_object_id" db:"source_object_id"`
	SourceObjectPublicID uuid.UUID       `json:"source_object_public_id" db:"source_object_public_id"`
	TargetObjectID       int64           `json:"target_object_id" db:"target_object_id"`
	TargetObjectPublicID uuid.UUID       `json:"target_object_public_id" db:"target_object_public_id"`
	RelationshipTypeID   int64           `json:"relationship_type_id" db:"relationship_type_id"`
	RelationshipTypeKey  string          `json:"relationship_type_key" db:"relationship_type_key"`
	Status               string          `json:"status" db:"status"`
	RelationshipMetadata json.RawMessage `json:"relationship_metadata" db:"relationship_metadata"`
	CreatedBy            *string         `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy            *string         `json:"updated_by,omitempty" db:"updated_by"`
	CreatedAt            time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at" db:"updated_at"`
}

func (r *Relationship) TableName() string {
	return "objects_relationships"
}

func (r *Relationship) GetMetadataMap() map[string]interface{} {
	if len(r.RelationshipMetadata) == 0 {
		return make(map[string]interface{})
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(r.RelationshipMetadata, &metadata); err != nil {
		return make(map[string]interface{})
	}
	return metadata
}

func (r *Relationship) SetMetadataMap(metadata map[string]interface{}) error {
	if metadata == nil {
		r.RelationshipMetadata = json.RawMessage("{}")
		return nil
	}

	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	r.RelationshipMetadata = data
	return nil
}

type CreateRelationshipRequest struct {
	SourceObjectPublicID string          `json:"source_object_id" binding:"required" validate:"required"`
	TargetObjectPublicID string          `json:"target_object_id" binding:"required" validate:"required"`
	RelationshipTypeKey  string          `json:"type_key" binding:"required" validate:"required,min=1,max=100"`
	Status               string          `json:"status"`
	RelationshipMetadata json.RawMessage `json:"metadata"`
	CreatedBy            string          `json:"-"`
}

func (r *CreateRelationshipRequest) SetDefaults() {
	if r.Status == "" {
		r.Status = StatusActive
	}
	if r.RelationshipMetadata == nil {
		r.RelationshipMetadata = json.RawMessage("{}")
	}
}

func (r *CreateRelationshipRequest) Validate() error {
	if r.SourceObjectPublicID == "" {
		return fmt.Errorf("source_object_id is required")
	}
	if r.TargetObjectPublicID == "" {
		return fmt.Errorf("target_object_id is required")
	}
	if r.RelationshipTypeKey == "" {
		return fmt.Errorf("type_key is required")
	}
	return nil
}

type UpdateRelationshipRequest struct {
	Status               *string         `json:"status,omitempty" validate:"omitempty,oneof=active inactive deprecated"`
	RelationshipMetadata json.RawMessage `json:"metadata,omitempty"`
	UpdatedBy            string          `json:"-"`
}

type RelationshipFilter struct {
	SourceObjectPublicID *string `json:"source_object_id,omitempty" form:"source_object_id"`
	TargetObjectPublicID *string `json:"target_object_id,omitempty" form:"target_object_id"`
	RelationshipTypeKey  *string `json:"type_key,omitempty" form:"type_key"`
	Status               *string `json:"status,omitempty" form:"status"`
	UserID               *int64  `json:"user_id,omitempty"`
	Page                 int     `json:"page,omitempty" form:"page"`
	PageSize             int     `json:"page_size,omitempty" form:"page_size"`
	SortBy               string  `json:"sort_by,omitempty" form:"sort_by"`
	SortOrder            string  `json:"sort_order,omitempty" form:"sort_order"`
}

type RelationshipResponse struct {
	ObjectID             int64                  `json:"object_id"`
	PublicID             uuid.UUID              `json:"public_id"`
	SourceObjectID       int64                  `json:"source_object_id"`
	SourceObjectPublicID string                 `json:"source_object_public_id"`
	TargetObjectID       int64                  `json:"target_object_id"`
	TargetObjectPublicID string                 `json:"target_object_public_id"`
	RelationshipTypeID   int64                  `json:"relationship_type_id"`
	RelationshipTypeKey  string                 `json:"relationship_type_key"`
	Status               string                 `json:"status"`
	RelationshipMetadata map[string]interface{} `json:"metadata"`
	CreatedAt            string                 `json:"created_at"`
	UpdatedAt            string                 `json:"updated_at"`
	CreatedBy            string                 `json:"created_by,omitempty"`
	UpdatedBy            string                 `json:"updated_by,omitempty"`
}

func (r *Relationship) ToResponse() *RelationshipResponse {
	var metadata map[string]interface{}
	if len(r.RelationshipMetadata) > 0 {
		if err := json.Unmarshal(r.RelationshipMetadata, &metadata); err != nil {
			metadata = make(map[string]interface{})
		}
	} else {
		metadata = make(map[string]interface{})
	}

	var createdBy, updatedBy string
	if r.CreatedBy != nil {
		createdBy = *r.CreatedBy
	}
	if r.UpdatedBy != nil {
		updatedBy = *r.UpdatedBy
	}

	return &RelationshipResponse{
		ObjectID:             r.ObjectID,
		PublicID:             r.PublicID,
		SourceObjectID:       r.SourceObjectID,
		SourceObjectPublicID: r.SourceObjectPublicID.String(),
		TargetObjectID:       r.TargetObjectID,
		TargetObjectPublicID: r.TargetObjectPublicID.String(),
		RelationshipTypeID:   r.RelationshipTypeID,
		RelationshipTypeKey:  r.RelationshipTypeKey,
		Status:               r.Status,
		RelationshipMetadata: metadata,
		CreatedAt:            r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:            r.UpdatedAt.Format(time.RFC3339),
		CreatedBy:            createdBy,
		UpdatedBy:            updatedBy,
	}
}

type RelationshipListResponse struct {
	Data       []RelationshipResponse `json:"data"`
	Pagination PaginationResponse     `json:"pagination"`
}

type RelationshipFilterForType struct {
	SourceObjectPublicID *string `json:"source_object_id,omitempty" form:"source_object_id"`
	TargetObjectPublicID *string `json:"target_object_id,omitempty" form:"target_object_id"`
	Status               *string `json:"status,omitempty" form:"status"`
	Page                 int     `json:"page,omitempty" form:"page"`
	PageSize             int     `json:"page_size,omitempty" form:"page_size"`
}
