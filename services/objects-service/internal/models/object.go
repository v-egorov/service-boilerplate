package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Status constants for Object
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusArchived = "archived"
	StatusDeleted  = "deleted"
	StatusPending  = "pending"
)

// ValidStatuses contains all valid status values
var ValidStatuses = map[string]bool{
	StatusActive:   true,
	StatusInactive: true,
	StatusArchived: true,
	StatusDeleted:  true,
	StatusPending:  true,
}

// Object represents an instance of an object type in the taxonomy system
type Object struct {
	ID             int64           `json:"id" db:"id"`
	PublicID       uuid.UUID       `json:"public_id" db:"public_id"`
	ObjectTypeID   int64           `json:"object_type_id" db:"object_type_id"`
	ParentObjectID *int64          `json:"parent_object_id,omitempty" db:"parent_object_id"`
	Name           string          `json:"name" db:"name"`
	Description    *string         `json:"description,omitempty" db:"description"`
	Metadata       json.RawMessage `json:"metadata,omitempty" db:"metadata"`
	Tags           []string        `json:"tags,omitempty" db:"tags"`
	Status         string          `json:"status" db:"status"`
	Version        int64           `json:"version" db:"version"`
	CreatedBy      string          `json:"created_by" db:"created_by"`
	UpdatedBy      string          `json:"updated_by" db:"updated_by"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time      `json:"deleted_at,omitempty" db:"deleted_at"`

	// Eager loading relationship (only ObjectType, as planned)
	ObjectType *ObjectType `json:"object_type,omitempty"`
}

// TableName returns the table name for Object model
func (o *Object) TableName() string {
	return "objects"
}

// IsSoftDeleted returns true if the object is soft deleted
func (o *Object) IsSoftDeleted() bool {
	return o.DeletedAt != nil
}

// IsActive returns true if the object is active and not soft deleted
func (o *Object) IsActive() bool {
	return o.Status == StatusActive && o.DeletedAt == nil
}

// IsArchived returns true if the object is archived
func (o *Object) IsArchived() bool {
	return o.Status == StatusArchived
}

// IsPending returns true if the object is pending
func (o *Object) IsPending() bool {
	return o.Status == StatusPending
}

// IsRoot returns true if this is a root object (no parent)
func (o *Object) IsRoot() bool {
	return o.ParentObjectID == nil
}

// HasParent returns true if this object has a parent
func (o *Object) HasParent() bool {
	return o.ParentObjectID != nil
}

// GetParentID returns the parent object ID (nil for root objects)
func (o *Object) GetParentID() *int64 {
	return o.ParentObjectID
}

// IsValidStatus returns true if the status is valid
func (o *Object) IsValidStatus() bool {
	return ValidStatuses[o.Status]
}

// GetMetadataMap converts JSONB metadata to map[string]interface{}
func (o *Object) GetMetadataMap() map[string]interface{} {
	if len(o.Metadata) == 0 {
		return make(map[string]interface{})
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(o.Metadata, &metadata); err != nil {
		return make(map[string]interface{})
	}
	return metadata
}

// SetMetadataMap sets metadata from map[string]interface{} to JSONB format
func (o *Object) SetMetadataMap(metadata map[string]interface{}) error {
	if metadata == nil {
		o.Metadata = json.RawMessage("{}")
		return nil
	}

	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	o.Metadata = data
	return nil
}

// GetTags returns a copy of the tags slice
func (o *Object) GetTags() []string {
	if o.Tags == nil {
		return []string{}
	}
	tags := make([]string, len(o.Tags))
	copy(tags, o.Tags)
	return tags
}

// HasTag returns true if the object has the specified tag
func (o *Object) HasTag(tag string) bool {
	for _, t := range o.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// AddTag adds a tag to the object if it doesn't already exist
func (o *Object) AddTag(tag string) {
	if !o.HasTag(tag) {
		o.Tags = append(o.Tags, tag)
	}
}

// RemoveTag removes a tag from the object if it exists
func (o *Object) RemoveTag(tag string) {
	for i, t := range o.Tags {
		if t == tag {
			o.Tags = append(o.Tags[:i], o.Tags[i+1:]...)
			return
		}
	}
}
