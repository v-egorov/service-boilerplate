package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestObjectType_Initialization(t *testing.T) {
	name := "TestType"
	desc := "Test description"
	ot := ObjectType{
		ID:          1,
		Name:        name,
		Description: desc,
		IsSealed:    false,
		Metadata:    json.RawMessage(`{"key": "value"}`),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	assert.Equal(t, int64(1), ot.ID)
	assert.Equal(t, "TestType", ot.Name)
	assert.Equal(t, "Test description", ot.Description)
	assert.False(t, ot.IsSealed)
}

func TestObjectType_CanHaveChildren(t *testing.T) {
	ot := ObjectType{
		ID:       1,
		IsSealed: false,
	}
	assert.True(t, ot.CanHaveChildren())

	ot.IsSealed = true
	assert.False(t, ot.CanHaveChildren())
}

func TestObjectType_IsRoot(t *testing.T) {
	ot := ObjectType{
		ID:           1,
		ParentTypeID: nil,
	}
	assert.True(t, ot.IsRoot())

	parentID := int64(10)
	ot.ParentTypeID = &parentID
	assert.False(t, ot.IsRoot())
}

func TestObjectType_GetParentID(t *testing.T) {
	ot := ObjectType{
		ID:           1,
		ParentTypeID: nil,
	}
	assert.Nil(t, ot.GetParentID())

	parentID := int64(10)
	ot.ParentTypeID = &parentID
	assert.Equal(t, &parentID, ot.GetParentID())
}

func TestObjectType_HasConcreteTable(t *testing.T) {
	ot := ObjectType{
		ID:                1,
		ConcreteTableName: nil,
	}
	assert.False(t, ot.HasConcreteTable())

	tableName := "test_table"
	ot.ConcreteTableName = &tableName
	assert.True(t, ot.HasConcreteTable())

	emptyTable := ""
	ot.ConcreteTableName = &emptyTable
	assert.False(t, ot.HasConcreteTable())
}

func TestObjectType_GetMetadataMap(t *testing.T) {
	ot := ObjectType{
		ID:       1,
		Metadata: json.RawMessage(`{"key1": "value1", "key2": 123}`),
	}

	metadata := ot.GetMetadataMap()
	assert.Equal(t, "value1", metadata["key1"])
	assert.Equal(t, float64(123), metadata["key2"])
}

func TestObjectType_GetMetadataMap_Empty(t *testing.T) {
	ot := ObjectType{
		ID:       1,
		Metadata: nil,
	}

	metadata := ot.GetMetadataMap()
	assert.NotNil(t, metadata)
	assert.Equal(t, 0, len(metadata))
}

func TestObjectType_SetMetadataMap(t *testing.T) {
	ot := &ObjectType{
		ID: 1,
	}

	err := ot.SetMetadataMap(map[string]interface{}{"key": "value"})
	assert.NoError(t, err)

	var metadata map[string]interface{}
	err = json.Unmarshal(ot.Metadata, &metadata)
	assert.NoError(t, err)
	assert.Equal(t, "value", metadata["key"])
}

func TestObjectType_SetMetadataMap_Nil(t *testing.T) {
	ot := &ObjectType{
		ID: 1,
	}

	err := ot.SetMetadataMap(nil)
	assert.NoError(t, err)
	assert.Equal(t, json.RawMessage("{}"), ot.Metadata)
}

func TestObjectType_TableName(t *testing.T) {
	ot := ObjectType{
		ID: 1,
	}
	assert.Equal(t, "object_types", ot.TableName())
}

func TestObjectType_WithParent(t *testing.T) {
	parentID := int64(1)
	ot := ObjectType{
		ID:           2,
		Name:         "ChildType",
		ParentTypeID: &parentID,
		IsSealed:     false,
	}

	assert.False(t, ot.IsRoot())
	assert.Equal(t, &parentID, ot.GetParentID())
	assert.True(t, ot.CanHaveChildren())
}

func TestObjectType_Sealed(t *testing.T) {
	ot := ObjectType{
		ID:       1,
		Name:     "SealedType",
		IsSealed: true,
	}

	assert.True(t, ot.IsSealed)
	assert.False(t, ot.CanHaveChildren())
}
