package models

import (
	"encoding/json"
	"time"
)

// ObjectType represents a hierarchical object type in the taxonomy system
type ObjectType struct {
	ID                int64           `json:"id" db:"id"`
	Name              string          `json:"name" db:"name"`
	ParentTypeID      *int64          `json:"parent_type_id,omitempty" db:"parent_type_id"`
	ConcreteTableName *string         `json:"concrete_table_name,omitempty" db:"concrete_table_name"`
	Description       string          `json:"description,omitempty" db:"description"`
	IsSealed          bool            `json:"is_sealed" db:"is_sealed"`
	Metadata          json.RawMessage `json:"metadata,omitempty" db:"metadata"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" db:"updated_at"`

	// Eager loading relationships
	ParentType *ObjectType  `json:"parent_type,omitempty"`
	Children   []ObjectType `json:"children,omitempty"`
	Objects    []Object     `json:"objects,omitempty"`
}

// TableName returns the table name for ObjectType model
func (ot *ObjectType) TableName() string {
	return "object_types"
}

// CanHaveChildren returns true if this object type can have child types
func (ot *ObjectType) CanHaveChildren() bool {
	return !ot.IsSealed
}

// GetParentID returns the parent type ID (nil for root types)
func (ot *ObjectType) GetParentID() *int64 {
	return ot.ParentTypeID
}

// IsRoot returns true if this is a root object type (no parent)
func (ot *ObjectType) IsRoot() bool {
	return ot.ParentTypeID == nil
}

// HasConcreteTable returns true if this type has a concrete table mapping
func (ot *ObjectType) HasConcreteTable() bool {
	return ot.ConcreteTableName != nil && *ot.ConcreteTableName != ""
}

// GetMetadataMap converts JSONB metadata to map[string]interface{}
func (ot *ObjectType) GetMetadataMap() map[string]interface{} {
	if len(ot.Metadata) == 0 {
		return make(map[string]interface{})
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(ot.Metadata, &metadata); err != nil {
		return make(map[string]interface{})
	}
	return metadata
}

// SetMetadataMap sets metadata from map[string]interface{} to JSONB format
func (ot *ObjectType) SetMetadataMap(metadata map[string]interface{}) error {
	if metadata == nil {
		ot.Metadata = json.RawMessage("{}")
		return nil
	}

	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	ot.Metadata = data
	return nil
}
