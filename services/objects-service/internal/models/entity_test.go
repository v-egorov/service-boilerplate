package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEntity_ToResponse(t *testing.T) {
	now := time.Now()
	entity := &Entity{
		ID:          1,
		Name:        "Test Entity",
		Description: "Test Description",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	response := entity.ToResponse()

	assert.Equal(t, int64(1), response.ID)
	assert.Equal(t, "Test Entity", response.Name)
	assert.Equal(t, "Test Description", response.Description)
	assert.Equal(t, now.Format(time.RFC3339), response.CreatedAt)
	assert.Equal(t, now.Format(time.RFC3339), response.UpdatedAt)
}

func TestEntity_TableName(t *testing.T) {
	entity := &Entity{}
	assert.Equal(t, "entities", entity.TableName())
}
