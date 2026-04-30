package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestObject_Initialization(t *testing.T) {
	publicID := uuid.New()
	obj := Object{
		ID:           1,
		PublicID:     publicID,
		ObjectTypeID: 100,
		Name:         "TestObject",
		Description:  stringPtr("Test description"),
		Status:       StatusActive,
		Tags:         []string{"tag1", "tag2"},
		Metadata:     json.RawMessage(`{"key": "value"}`),
		Version:      1,
		CreatedBy:    "system",
		UpdatedBy:    "system",
	}

	assert.Equal(t, int64(1), obj.ID)
	assert.Equal(t, publicID, obj.PublicID)
	assert.Equal(t, int64(100), obj.ObjectTypeID)
	assert.Equal(t, "TestObject", obj.Name)
	assert.Equal(t, "Test description", toString(obj.Description))
	assert.Equal(t, StatusActive, obj.Status)
	assert.Equal(t, []string{"tag1", "tag2"}, obj.Tags)
	assert.Equal(t, int64(1), obj.Version)
}

func TestObject_IsActive(t *testing.T) {
	obj := Object{
		ID:     1,
		Status: StatusActive,
	}
	assert.True(t, obj.IsActive())

	obj.Status = StatusInactive
	assert.False(t, obj.IsActive())

	obj.Status = StatusDeleted
	assert.False(t, obj.IsActive())
}

func TestObject_IsSoftDeleted(t *testing.T) {
	obj := Object{
		ID:        1,
		Status:    StatusActive,
		DeletedAt: nil,
	}
	assert.False(t, obj.IsSoftDeleted())

	now := time.Now()
	obj.DeletedAt = &now
	assert.True(t, obj.IsSoftDeleted())
}

func TestObject_IsRoot(t *testing.T) {
	obj := Object{
		ID:             1,
		ParentObjectID: nil,
	}
	assert.True(t, obj.IsRoot())

	parentID := int64(10)
	obj.ParentObjectID = &parentID
	assert.False(t, obj.IsRoot())
}

func TestObject_HasParent(t *testing.T) {
	obj := Object{
		ID:             1,
		ParentObjectID: nil,
	}
	assert.False(t, obj.HasParent())

	parentID := int64(10)
	obj.ParentObjectID = &parentID
	assert.True(t, obj.HasParent())
}

func TestObject_GetParentID(t *testing.T) {
	obj := Object{
		ID:             1,
		ParentObjectID: nil,
	}
	assert.Nil(t, obj.GetParentID())

	parentID := int64(10)
	obj.ParentObjectID = &parentID
	assert.Equal(t, &parentID, obj.GetParentID())
}

func TestObject_IsValidStatus(t *testing.T) {
	obj := Object{
		ID:     1,
		Status: StatusActive,
	}
	assert.True(t, obj.IsValidStatus())

	obj.Status = "invalid"
	assert.False(t, obj.IsValidStatus())
}

func TestObject_GetMetadataMap(t *testing.T) {
	obj := Object{
		ID:       1,
		Metadata: json.RawMessage(`{"key1": "value1", "key2": 123}`),
	}

	metadata := obj.GetMetadataMap()
	assert.Equal(t, "value1", metadata["key1"])
	assert.Equal(t, float64(123), metadata["key2"])
}

func TestObject_GetMetadataMap_Empty(t *testing.T) {
	obj := Object{
		ID:       1,
		Metadata: nil,
	}

	metadata := obj.GetMetadataMap()
	assert.NotNil(t, metadata)
	assert.Equal(t, 0, len(metadata))
}

func TestObject_SetMetadataMap(t *testing.T) {
	obj := &Object{
		ID: 1,
	}

	err := obj.SetMetadataMap(map[string]interface{}{"key": "value"})
	assert.NoError(t, err)

	var metadata map[string]interface{}
	err = json.Unmarshal(obj.Metadata, &metadata)
	assert.NoError(t, err)
	assert.Equal(t, "value", metadata["key"])
}

func TestObject_SetMetadataMap_Nil(t *testing.T) {
	obj := &Object{
		ID: 1,
	}

	err := obj.SetMetadataMap(nil)
	assert.NoError(t, err)
	assert.Equal(t, json.RawMessage("{}"), obj.Metadata)
}

func TestObject_GetTags(t *testing.T) {
	obj := &Object{
		ID:   1,
		Tags: []string{"tag1", "tag2"},
	}

	tags := obj.GetTags()
	assert.Equal(t, []string{"tag1", "tag2"}, tags)

	tags[0] = "modified"
	assert.Equal(t, "tag1", obj.Tags[0])
}

func TestObject_GetTags_Empty(t *testing.T) {
	obj := &Object{
		ID:   1,
		Tags: nil,
	}

	tags := obj.GetTags()
	assert.NotNil(t, tags)
	assert.Equal(t, 0, len(tags))
}

func TestObject_HasTag(t *testing.T) {
	obj := &Object{
		ID:   1,
		Tags: []string{"tag1", "tag2"},
	}

	assert.True(t, obj.HasTag("tag1"))
	assert.True(t, obj.HasTag("tag2"))
	assert.False(t, obj.HasTag("tag3"))
}

func TestObject_AddTag(t *testing.T) {
	obj := &Object{
		ID:   1,
		Tags: []string{"tag1"},
	}

	obj.AddTag("tag1")
	assert.Equal(t, []string{"tag1"}, obj.Tags)

	obj.AddTag("tag2")
	assert.Equal(t, []string{"tag1", "tag2"}, obj.Tags)
}

func TestObject_RemoveTag(t *testing.T) {
	obj := &Object{
		ID:   1,
		Tags: []string{"tag1", "tag2", "tag3"},
	}

	obj.RemoveTag("tag2")
	assert.Equal(t, []string{"tag1", "tag3"}, obj.Tags)

	obj.RemoveTag("tag4")
	assert.Equal(t, []string{"tag1", "tag3"}, obj.Tags)

	obj.RemoveTag("tag1")
	assert.Equal(t, []string{"tag3"}, obj.Tags)
}

func TestObject_TableName(t *testing.T) {
	obj := Object{
		ID: 1,
	}
	assert.Equal(t, "objects", obj.TableName())
}

func TestObject_IsArchived(t *testing.T) {
	obj := Object{
		ID:     1,
		Status: StatusArchived,
	}
	assert.True(t, obj.IsArchived())

	obj.Status = StatusActive
	assert.False(t, obj.IsArchived())
}

func TestObject_IsPending(t *testing.T) {
	obj := Object{
		ID:     1,
		Status: StatusPending,
	}
	assert.True(t, obj.IsPending())

	obj.Status = StatusActive
	assert.False(t, obj.IsPending())
}
