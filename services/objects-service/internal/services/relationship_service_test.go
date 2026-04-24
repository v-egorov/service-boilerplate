package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/repository"
)

func TestRelationshipService_Create_Success(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		existsFunc: func(ctx context.Context, sourceObjectID, targetObjectID int64, typeObjectID int64) (bool, error) {
			return false, nil
		},
		checkCircularFunc: func(ctx context.Context, sourceObjectID, targetObjectID, typeObjectID int64) (bool, error) {
			return false, nil
		},
		countForObjectFunc: func(ctx context.Context, objectID int64, typeKey *string) (int, error) {
			return 0, nil
		},
		createFunc: func(ctx context.Context, input *models.CreateRelationshipRequest) (*models.Relationship, error) {
			srcID, _ := uuid.Parse(input.SourceObjectPublicID)
			tgtID, _ := uuid.Parse(input.TargetObjectPublicID)
			return &models.Relationship{
				ObjectID:             1,
				SourceObjectID:      1,
				SourceObjectPublicID: srcID,
				TargetObjectID:      2,
				TargetObjectPublicID: tgtID,
				RelationshipTypeKey: input.RelationshipTypeKey,
				Status:              input.Status,
				CreatedBy:           &input.CreatedBy,
			}, nil
		},
	}

	mockRelTypeRepo := &mockRelationshipTypeRepository{
		getByTypeKeyFunc: func(ctx context.Context, typeKey string) (*models.RelationshipType, error) {
			return &models.RelationshipType{
				ObjectID:    1,
				TypeKey:     typeKey,
				Cardinality: models.CardinalityOneToMany,
			}, nil
		},
	}

	mockObjRepo := &mockObjectRepository{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Object, error) {
			if publicID == testSourcePublicID {
				return &models.Object{ID: testSourceObjectID, PublicID: publicID}, nil
			}
			return &models.Object{ID: testTargetObjectID, PublicID: publicID}, nil
		},
	}

	service := NewRelationshipService(mockRelRepo, mockRelTypeRepo, mockObjRepo)

	input := &models.CreateRelationshipRequest{
		SourceObjectPublicID: testSourcePublicID.String(),
		TargetObjectPublicID: testTargetPublicID.String(),
		RelationshipTypeKey:  "contains",
	}

	result, err := service.Create(context.Background(), input)

assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestRelationshipService_GetByPublicID_Success(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Relationship, error) {
			return &models.Relationship{
				ObjectID:             1,
				SourceObjectPublicID: testSourcePublicID,
				TargetObjectPublicID: testTargetPublicID,
				RelationshipTypeKey:  "contains",
				Status:              models.StatusActive,
			}, nil
		},
	}

	service := NewRelationshipService(mockRelRepo, &mockRelationshipTypeRepository{}, &mockObjectRepository{})

	result, err := service.GetByPublicID(context.Background(), testSourcePublicID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "contains", result.RelationshipTypeKey)
}

func TestRelationshipService_GetByPublicID_NotFound(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Relationship, error) {
			return nil, repository.ErrNotFound
		},
	}

	service := NewRelationshipService(mockRelRepo, &mockRelationshipTypeRepository{}, &mockObjectRepository{})

	result, err := service.GetByPublicID(context.Background(), testSourcePublicID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
}

func TestRelationshipService_Update_Success(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Relationship, error) {
			return &models.Relationship{ObjectID: 1, SourceObjectPublicID: publicID}, nil
		},
		updateFunc: func(ctx context.Context, objectID int64, input *models.UpdateRelationshipRequest) (*models.Relationship, error) {
			return &models.Relationship{
				ObjectID: objectID,
				Status:  *input.Status,
			}, nil
		},
	}

	service := NewRelationshipService(mockRelRepo, &mockRelationshipTypeRepository{}, &mockObjectRepository{})

	status := models.StatusInactive
	input := &models.UpdateRelationshipRequest{
		Status: &status,
	}

	result, err := service.Update(context.Background(), testSourcePublicID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, models.StatusInactive, result.Status)
}

func TestRelationshipService_Update_NotFound(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Relationship, error) {
			return nil, repository.ErrNotFound
		},
	}

	service := NewRelationshipService(mockRelRepo, &mockRelationshipTypeRepository{}, &mockObjectRepository{})

	status := models.StatusInactive
	input := &models.UpdateRelationshipRequest{
		Status: &status,
	}

	result, err := service.Update(context.Background(), testSourcePublicID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
}

func TestRelationshipService_Update_WithMetadata(t *testing.T) {
	metadata := json.RawMessage(`{"note": "updated"}`)
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Relationship, error) {
			return &models.Relationship{ObjectID: 1}, nil
		},
		updateFunc: func(ctx context.Context, objectID int64, input *models.UpdateRelationshipRequest) (*models.Relationship, error) {
			return &models.Relationship{
				ObjectID:             objectID,
				RelationshipMetadata: input.RelationshipMetadata,
			}, nil
		},
	}

	service := NewRelationshipService(mockRelRepo, &mockRelationshipTypeRepository{}, &mockObjectRepository{})

	input := &models.UpdateRelationshipRequest{
		RelationshipMetadata: metadata,
	}

	result, err := service.Update(context.Background(), testSourcePublicID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestRelationshipService_Delete_Success(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Relationship, error) {
			return &models.Relationship{ObjectID: 1}, nil
		},
		deleteFunc: func(ctx context.Context, objectID int64) error {
			return nil
		},
	}

	service := NewRelationshipService(mockRelRepo, &mockRelationshipTypeRepository{}, &mockObjectRepository{})

	err := service.Delete(context.Background(), testSourcePublicID)

	assert.NoError(t, err)
}

func TestRelationshipService_Delete_NotFound(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Relationship, error) {
			return nil, repository.ErrNotFound
		},
	}

	service := NewRelationshipService(mockRelRepo, &mockRelationshipTypeRepository{}, &mockObjectRepository{})

	err := service.Delete(context.Background(), testSourcePublicID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRelationshipService_Delete_Error(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Relationship, error) {
			return &models.Relationship{ObjectID: 1}, nil
		},
		deleteFunc: func(ctx context.Context, objectID int64) error {
			return errors.New("database error")
		},
	}

	service := NewRelationshipService(mockRelRepo, &mockRelationshipTypeRepository{}, &mockObjectRepository{})

	err := service.Delete(context.Background(), testSourcePublicID)

	assert.Error(t, err)
}

func TestRelationshipService_List_Success(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		listFunc: func(ctx context.Context, filter *models.RelationshipFilter) ([]*models.Relationship, error) {
			return []*models.Relationship{
				{ObjectID: 1, RelationshipTypeKey: "contains"},
				{ObjectID: 2, RelationshipTypeKey: "belongs_to"},
			}, nil
		},
	}

	service := NewRelationshipService(mockRelRepo, &mockRelationshipTypeRepository{}, &mockObjectRepository{})

	result, err := service.List(context.Background(), &models.RelationshipFilter{})

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestRelationshipService_List_WithFilter(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		listFunc: func(ctx context.Context, filter *models.RelationshipFilter) ([]*models.Relationship, error) {
			if filter.RelationshipTypeKey != nil && *filter.RelationshipTypeKey == "contains" {
				return []*models.Relationship{{ObjectID: 1}}, nil
			}
			return []*models.Relationship{}, nil
		},
	}

	service := NewRelationshipService(mockRelRepo, &mockRelationshipTypeRepository{}, &mockObjectRepository{})

	typeKey := "contains"
	result, err := service.List(context.Background(), &models.RelationshipFilter{
		RelationshipTypeKey: &typeKey,
	})

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestRelationshipService_List_WithPagination(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		listFunc: func(ctx context.Context, filter *models.RelationshipFilter) ([]*models.Relationship, error) {
			if filter.Page == 2 && filter.PageSize == 10 {
				return []*models.Relationship{{ObjectID: 11}}, nil
			}
			return []*models.Relationship{}, nil
		},
	}

	service := NewRelationshipService(mockRelRepo, &mockRelationshipTypeRepository{}, &mockObjectRepository{})

	result, err := service.List(context.Background(), &models.RelationshipFilter{
		Page:     2,
		PageSize: 10,
	})

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestRelationshipService_List_NilFilter(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		listFunc: func(ctx context.Context, filter *models.RelationshipFilter) ([]*models.Relationship, error) {
			return []*models.Relationship{{ObjectID: 1}}, nil
		},
	}

	service := NewRelationshipService(mockRelRepo, &mockRelationshipTypeRepository{}, &mockObjectRepository{})

	result, err := service.List(context.Background(), nil)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestRelationshipService_List_Empty(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		listFunc: func(ctx context.Context, filter *models.RelationshipFilter) ([]*models.Relationship, error) {
			return []*models.Relationship{}, nil
		},
	}

	service := NewRelationshipService(mockRelRepo, &mockRelationshipTypeRepository{}, &mockObjectRepository{})

	result, err := service.List(context.Background(), &models.RelationshipFilter{})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

func TestRelationshipService_Create_WithMetadata(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		existsFunc:          func(ctx context.Context, sourceObjectID, targetObjectID int64, typeObjectID int64) (bool, error) { return false, nil },
		checkCircularFunc:   func(ctx context.Context, sourceObjectID, targetObjectID, typeObjectID int64) (bool, error) { return false, nil },
		countForObjectFunc: func(ctx context.Context, objectID int64, typeKey *string) (int, error) { return 0, nil },
		createFunc: func(ctx context.Context, input *models.CreateRelationshipRequest) (*models.Relationship, error) {
			return &models.Relationship{RelationshipTypeKey: input.RelationshipTypeKey}, nil
		},
	}

	mockRelTypeRepo := &mockRelationshipTypeRepository{
		getByTypeKeyFunc: func(ctx context.Context, typeKey string) (*models.RelationshipType, error) {
			return &models.RelationshipType{ObjectID: 1, TypeKey: typeKey, Cardinality: models.CardinalityOneToMany}, nil
		},
	}

	mockObjRepo := &mockObjectRepository{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Object, error) {
			return &models.Object{ID: 1, PublicID: publicID}, nil
		},
	}

	service := NewRelationshipService(mockRelRepo, mockRelTypeRepo, mockObjRepo)

	input := &models.CreateRelationshipRequest{
		SourceObjectPublicID: testSourcePublicID.String(),
		TargetObjectPublicID: testTargetPublicID.String(),
		RelationshipTypeKey:  "contains",
		Status:              "active",
	}

	result, err := service.Create(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestRelationshipService_Create_WithCreatedBy(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		existsFunc:     func(ctx context.Context, sourceObjectID, targetObjectID int64, typeObjectID int64) (bool, error) { return false, nil },
		checkCircularFunc: func(ctx context.Context, sourceObjectID, targetObjectID, typeObjectID int64) (bool, error) { return false, nil },
		countForObjectFunc: func(ctx context.Context, objectID int64, typeKey *string) (int, error) { return 0, nil },
		createFunc: func(ctx context.Context, input *models.CreateRelationshipRequest) (*models.Relationship, error) {
			createdBy := input.CreatedBy
			return &models.Relationship{RelationshipTypeKey: input.RelationshipTypeKey, CreatedBy: &createdBy}, nil
		},
	}

	mockRelTypeRepo := &mockRelationshipTypeRepository{
		getByTypeKeyFunc: func(ctx context.Context, typeKey string) (*models.RelationshipType, error) {
			return &models.RelationshipType{ObjectID: 1, TypeKey: typeKey, Cardinality: models.CardinalityOneToMany}, nil
		},
	}

	mockObjRepo := &mockObjectRepository{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Object, error) {
			return &models.Object{ID: 1, PublicID: publicID}, nil
		},
	}

	service := NewRelationshipService(mockRelRepo, mockRelTypeRepo, mockObjRepo)

	input := &models.CreateRelationshipRequest{
		SourceObjectPublicID: testSourcePublicID.String(),
		TargetObjectPublicID: testTargetPublicID.String(),
		RelationshipTypeKey:  "contains",
		CreatedBy:           "test-user",
	}

	result, err := service.Create(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestRelationshipService_Create_WithDefaultStatus(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		existsFunc:     func(ctx context.Context, sourceObjectID, targetObjectID int64, typeObjectID int64) (bool, error) { return false, nil },
		checkCircularFunc: func(ctx context.Context, sourceObjectID, targetObjectID, typeObjectID int64) (bool, error) { return false, nil },
		countForObjectFunc: func(ctx context.Context, objectID int64, typeKey *string) (int, error) { return 0, nil },
		createFunc: func(ctx context.Context, input *models.CreateRelationshipRequest) (*models.Relationship, error) {
			return &models.Relationship{Status: input.Status}, nil
		},
	}

	mockRelTypeRepo := &mockRelationshipTypeRepository{
		getByTypeKeyFunc: func(ctx context.Context, typeKey string) (*models.RelationshipType, error) {
			return &models.RelationshipType{ObjectID: 1, TypeKey: typeKey, Cardinality: models.CardinalityOneToMany}, nil
		},
	}

	mockObjRepo := &mockObjectRepository{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Object, error) {
			return &models.Object{ID: 1, PublicID: publicID}, nil
		},
	}

	service := NewRelationshipService(mockRelRepo, mockRelTypeRepo, mockObjRepo)

	input := &models.CreateRelationshipRequest{
		SourceObjectPublicID: testSourcePublicID.String(),
		TargetObjectPublicID: testTargetPublicID.String(),
		RelationshipTypeKey:  "contains",
		Status:               "",
	}

	result, err := service.Create(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, models.StatusActive, result.Status)
}

func TestRelationshipService_Create_InvalidSourceUUID(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{}
	mockRelTypeRepo := &mockRelationshipTypeRepository{}
	mockObjRepo := &mockObjectRepository{}

	service := NewRelationshipService(mockRelRepo, mockRelTypeRepo, mockObjRepo)

	input := &models.CreateRelationshipRequest{
		SourceObjectPublicID: "invalid-uuid",
		TargetObjectPublicID: testTargetPublicID.String(),
		RelationshipTypeKey:  "contains",
	}

	result, err := service.Create(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid source_object_id format")
}

func TestRelationshipService_Create_InvalidTargetUUID(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{}
	mockRelTypeRepo := &mockRelationshipTypeRepository{}
	mockObjRepo := &mockObjectRepository{}

	service := NewRelationshipService(mockRelRepo, mockRelTypeRepo, mockObjRepo)

	input := &models.CreateRelationshipRequest{
		SourceObjectPublicID: testSourcePublicID.String(),
		TargetObjectPublicID: "invalid-uuid",
		RelationshipTypeKey:  "contains",
	}

	result, err := service.Create(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid target_object_id format")
}

func TestRelationshipService_Create_SourceObjectNotFound(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{}
	mockRelTypeRepo := &mockRelationshipTypeRepository{}
	mockObjRepo := &mockObjectRepository{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Object, error) {
			return nil, repository.ErrNotFound
		},
	}

	service := NewRelationshipService(mockRelRepo, mockRelTypeRepo, mockObjRepo)

	input := &models.CreateRelationshipRequest{
		SourceObjectPublicID: testSourcePublicID.String(),
		TargetObjectPublicID: testTargetPublicID.String(),
		RelationshipTypeKey:  "contains",
	}

	result, err := service.Create(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "source object not found")
}

func TestRelationshipService_Create_TargetObjectNotFound(t *testing.T) {
	mockObjRepo := &mockObjectRepository{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Object, error) {
			if publicID == testSourcePublicID {
				return &models.Object{ID: 1, PublicID: publicID}, nil
			}
			return nil, repository.ErrNotFound
		},
	}

	service := NewRelationshipService(&mockRelationshipRepositoryForRelationshipService{}, &mockRelationshipTypeRepository{}, mockObjRepo)

	input := &models.CreateRelationshipRequest{
		SourceObjectPublicID: testSourcePublicID.String(),
		TargetObjectPublicID: testTargetPublicID.String(),
		RelationshipTypeKey:  "contains",
	}

	result, err := service.Create(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "target object not found")
}

func TestRelationshipService_Create_RelationshipTypeNotFound(t *testing.T) {
	mockObjRepo := &mockObjectRepository{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Object, error) {
			return &models.Object{ID: 1, PublicID: publicID}, nil
		},
	}

	mockRelTypeRepo := &mockRelationshipTypeRepository{
		getByTypeKeyFunc: func(ctx context.Context, typeKey string) (*models.RelationshipType, error) {
			return nil, repository.ErrNotFound
		},
	}

	service := NewRelationshipService(&mockRelationshipRepositoryForRelationshipService{}, mockRelTypeRepo, mockObjRepo)

	input := &models.CreateRelationshipRequest{
		SourceObjectPublicID: testSourcePublicID.String(),
		TargetObjectPublicID: testTargetPublicID.String(),
		RelationshipTypeKey:  "nonexistent",
	}

	result, err := service.Create(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "type_key")
}

func TestRelationshipService_Create_DuplicateRelationship(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		existsFunc: func(ctx context.Context, sourceObjectID, targetObjectID int64, typeObjectID int64) (bool, error) {
			return true, nil
		},
	}

	mockRelTypeRepo := &mockRelationshipTypeRepository{
		getByTypeKeyFunc: func(ctx context.Context, typeKey string) (*models.RelationshipType, error) {
			return &models.RelationshipType{ObjectID: 1, TypeKey: typeKey}, nil
		},
	}

	mockObjRepo := &mockObjectRepository{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Object, error) {
			return &models.Object{ID: 1, PublicID: publicID}, nil
		},
	}

	service := NewRelationshipService(mockRelRepo, mockRelTypeRepo, mockObjRepo)

	input := &models.CreateRelationshipRequest{
		SourceObjectPublicID: testSourcePublicID.String(),
		TargetObjectPublicID: testTargetPublicID.String(),
		RelationshipTypeKey:  "contains",
	}

	result, err := service.Create(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "already exists")
}

func TestRelationshipService_Create_CircularDetection(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		existsFunc: func(ctx context.Context, sourceObjectID, targetObjectID int64, typeObjectID int64) (bool, error) {
			return false, nil
		},
		checkCircularFunc: func(ctx context.Context, sourceObjectID, targetObjectID, typeObjectID int64) (bool, error) {
			return true, nil
		},
	}

	mockRelTypeRepo := &mockRelationshipTypeRepository{
		getByTypeKeyFunc: func(ctx context.Context, typeKey string) (*models.RelationshipType, error) {
			return &models.RelationshipType{
				ObjectID:    1,
				TypeKey:     typeKey,
				Cardinality: models.CardinalityOneToMany,
			}, nil
		},
	}

	mockObjRepo := &mockObjectRepository{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Object, error) {
			return &models.Object{ID: 1, PublicID: publicID}, nil
		},
	}

	service := NewRelationshipService(mockRelRepo, mockRelTypeRepo, mockObjRepo)

	input := &models.CreateRelationshipRequest{
		SourceObjectPublicID: testSourcePublicID.String(),
		TargetObjectPublicID: testTargetPublicID.String(),
		RelationshipTypeKey:  "contains",
	}

	result, err := service.Create(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cycle")
}

func TestRelationshipService_Create_ManyToMany_NoCircularCheck(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		existsFunc: func(ctx context.Context, sourceObjectID, targetObjectID int64, typeObjectID int64) (bool, error) {
			return false, nil
		},
		countForObjectFunc: func(ctx context.Context, objectID int64, typeKey *string) (int, error) {
			return 0, nil
		},
		createFunc: func(ctx context.Context, input *models.CreateRelationshipRequest) (*models.Relationship, error) {
			return &models.Relationship{RelationshipTypeKey: input.RelationshipTypeKey}, nil
		},
	}

	mockRelTypeRepo := &mockRelationshipTypeRepository{
		getByTypeKeyFunc: func(ctx context.Context, typeKey string) (*models.RelationshipType, error) {
			return &models.RelationshipType{
				ObjectID:    1,
				TypeKey:     typeKey,
				Cardinality: models.CardinalityManyToMany,
			}, nil
		},
	}

	mockObjRepo := &mockObjectRepository{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Object, error) {
			return &models.Object{ID: 1, PublicID: publicID}, nil
		},
	}

	service := NewRelationshipService(mockRelRepo, mockRelTypeRepo, mockObjRepo)

	input := &models.CreateRelationshipRequest{
		SourceObjectPublicID: testSourcePublicID.String(),
		TargetObjectPublicID: testTargetPublicID.String(),
		RelationshipTypeKey:  "related_to",
	}

	result, err := service.Create(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestRelationshipService_Create_OneToOne_SourceAlreadyHasRelationship(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		existsFunc: func(ctx context.Context, sourceObjectID, targetObjectID int64, typeObjectID int64) (bool, error) {
			return false, nil
		},
		checkCircularFunc: func(ctx context.Context, sourceObjectID, targetObjectID, typeObjectID int64) (bool, error) {
			return false, nil
		},
		countForObjectFunc: func(ctx context.Context, objectID int64, typeKey *string) (int, error) {
			return 1, nil
		},
	}

	mockRelTypeRepo := &mockRelationshipTypeRepository{
		getByTypeKeyFunc: func(ctx context.Context, typeKey string) (*models.RelationshipType, error) {
			return &models.RelationshipType{
				ObjectID:    1,
				TypeKey:     typeKey,
				Cardinality: models.CardinalityOneToOne,
			}, nil
		},
	}

	mockObjRepo := &mockObjectRepository{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Object, error) {
			return &models.Object{ID: 1, PublicID: publicID}, nil
		},
	}

	service := NewRelationshipService(mockRelRepo, mockRelTypeRepo, mockObjRepo)

	input := &models.CreateRelationshipRequest{
		SourceObjectPublicID: testSourcePublicID.String(),
		TargetObjectPublicID: testTargetPublicID.String(),
		RelationshipTypeKey:  "married_to",
	}

	result, err := service.Create(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cardinality")
}

func TestRelationshipService_Create_OneToOne_TargetAlreadyHasRelationship(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		existsFunc: func(ctx context.Context, sourceObjectID, targetObjectID int64, typeObjectID int64) (bool, error) {
			return false, nil
		},
		checkCircularFunc: func(ctx context.Context, sourceObjectID, targetObjectID, typeObjectID int64) (bool, error) {
			return false, nil
		},
		countForObjectFunc: func(ctx context.Context, objectID int64, typeKey *string) (int, error) {
			if objectID == testTargetObjectID {
				return 1, nil
			}
			return 0, nil
		},
	}

	mockRelTypeRepo := &mockRelationshipTypeRepository{
		getByTypeKeyFunc: func(ctx context.Context, typeKey string) (*models.RelationshipType, error) {
			return &models.RelationshipType{
				ObjectID:    1,
				TypeKey:     typeKey,
				Cardinality: models.CardinalityOneToOne,
			}, nil
		},
	}

	mockObjRepo := &mockObjectRepository{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Object, error) {
			if publicID == testSourcePublicID {
				return &models.Object{ID: testSourceObjectID, PublicID: publicID}, nil
			}
			return &models.Object{ID: testTargetObjectID, PublicID: publicID}, nil
		},
	}

	service := NewRelationshipService(mockRelRepo, mockRelTypeRepo, mockObjRepo)

	input := &models.CreateRelationshipRequest{
		SourceObjectPublicID: testSourcePublicID.String(),
		TargetObjectPublicID: testTargetPublicID.String(),
		RelationshipTypeKey:  "married_to",
	}

	result, err := service.Create(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cardinality")
}

func TestRelationshipService_Create_OneToMany_TargetAlreadyHasSource(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		existsFunc: func(ctx context.Context, sourceObjectID, targetObjectID int64, typeObjectID int64) (bool, error) {
			return false, nil
		},
		checkCircularFunc: func(ctx context.Context, sourceObjectID, targetObjectID, typeObjectID int64) (bool, error) {
			return false, nil
		},
		countForObjectFunc: func(ctx context.Context, objectID int64, typeKey *string) (int, error) {
			if objectID == testTargetObjectID {
				return 1, nil
			}
			return 0, nil
		},
	}

	mockRelTypeRepo := &mockRelationshipTypeRepository{
		getByTypeKeyFunc: func(ctx context.Context, typeKey string) (*models.RelationshipType, error) {
			return &models.RelationshipType{
				ObjectID:    1,
				TypeKey:     typeKey,
				Cardinality: models.CardinalityOneToMany,
			}, nil
		},
	}

	mockObjRepo := &mockObjectRepository{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Object, error) {
			if publicID == testSourcePublicID {
				return &models.Object{ID: testSourceObjectID, PublicID: publicID}, nil
			}
			return &models.Object{ID: testTargetObjectID, PublicID: publicID}, nil
		},
	}

	service := NewRelationshipService(mockRelRepo, mockRelTypeRepo, mockObjRepo)

	input := &models.CreateRelationshipRequest{
		SourceObjectPublicID: testSourcePublicID.String(),
		TargetObjectPublicID: testTargetPublicID.String(),
		RelationshipTypeKey:  "contains",
	}

	result, err := service.Create(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "target already has source")
}

func TestRelationshipService_Create_ManyToOne_SourceAlreadyHasTarget(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		existsFunc: func(ctx context.Context, sourceObjectID, targetObjectID int64, typeObjectID int64) (bool, error) {
			return false, nil
		},
		checkCircularFunc: func(ctx context.Context, sourceObjectID, targetObjectID, typeObjectID int64) (bool, error) {
			return false, nil
		},
		countForObjectFunc: func(ctx context.Context, objectID int64, typeKey *string) (int, error) {
			if objectID == testSourceObjectID {
				return 1, nil
			}
			return 0, nil
		},
	}

	mockRelTypeRepo := &mockRelationshipTypeRepository{
		getByTypeKeyFunc: func(ctx context.Context, typeKey string) (*models.RelationshipType, error) {
			return &models.RelationshipType{
				ObjectID:    1,
				TypeKey:     typeKey,
				Cardinality: models.CardinalityManyToOne,
			}, nil
		},
	}

	mockObjRepo := &mockObjectRepository{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Object, error) {
			if publicID == testSourcePublicID {
				return &models.Object{ID: testSourceObjectID, PublicID: publicID}, nil
			}
			return &models.Object{ID: testTargetObjectID, PublicID: publicID}, nil
		},
	}

	service := NewRelationshipService(mockRelRepo, mockRelTypeRepo, mockObjRepo)

	input := &models.CreateRelationshipRequest{
		SourceObjectPublicID: testSourcePublicID.String(),
		TargetObjectPublicID: testTargetPublicID.String(),
		RelationshipTypeKey:  "belongs_to",
	}

	result, err := service.Create(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "source already has target")
}

func TestRelationshipService_Create_MaxCountViolation_OneToMany(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		existsFunc: func(ctx context.Context, sourceObjectID, targetObjectID int64, typeObjectID int64) (bool, error) {
			return false, nil
		},
		checkCircularFunc: func(ctx context.Context, sourceObjectID, targetObjectID, typeObjectID int64) (bool, error) {
			return false, nil
		},
		countForObjectFunc: func(ctx context.Context, objectID int64, typeKey *string) (int, error) {
			if objectID == testSourceObjectID {
				return 5, nil
			}
			return 0, nil
		},
	}

	mockRelTypeRepo := &mockRelationshipTypeRepository{
		getByTypeKeyFunc: func(ctx context.Context, typeKey string) (*models.RelationshipType, error) {
			return &models.RelationshipType{
				ObjectID:    1,
				TypeKey:     typeKey,
				Cardinality: models.CardinalityOneToMany,
				MaxCount:   5,
			}, nil
		},
	}

	mockObjRepo := &mockObjectRepository{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Object, error) {
			if publicID == testSourcePublicID {
				return &models.Object{ID: testSourceObjectID, PublicID: publicID}, nil
			}
			return &models.Object{ID: testTargetObjectID, PublicID: publicID}, nil
		},
	}

	service := NewRelationshipService(mockRelRepo, mockRelTypeRepo, mockObjRepo)

	input := &models.CreateRelationshipRequest{
		SourceObjectPublicID: testSourcePublicID.String(),
		TargetObjectPublicID: testTargetPublicID.String(),
		RelationshipTypeKey:  "contains",
	}

	result, err := service.Create(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "max_count")
}

func TestRelationshipService_Create_MaxCountViolation_ManyToMany(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		existsFunc: func(ctx context.Context, sourceObjectID, targetObjectID int64, typeObjectID int64) (bool, error) {
			return false, nil
		},
		checkCircularFunc: func(ctx context.Context, sourceObjectID, targetObjectID, typeObjectID int64) (bool, error) {
			return false, nil
		},
		countForObjectFunc: func(ctx context.Context, objectID int64, typeKey *string) (int, error) {
			if objectID == testSourceObjectID {
				return 10, nil
			}
			return 0, nil
		},
	}

	mockRelTypeRepo := &mockRelationshipTypeRepository{
		getByTypeKeyFunc: func(ctx context.Context, typeKey string) (*models.RelationshipType, error) {
			return &models.RelationshipType{
				ObjectID:    1,
				TypeKey:     typeKey,
				Cardinality: models.CardinalityManyToMany,
				MaxCount:   10,
			}, nil
		},
	}

	mockObjRepo := &mockObjectRepository{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Object, error) {
			if publicID == testSourcePublicID {
				return &models.Object{ID: testSourceObjectID, PublicID: publicID}, nil
			}
			return &models.Object{ID: testTargetObjectID, PublicID: publicID}, nil
		},
	}

	service := NewRelationshipService(mockRelRepo, mockRelTypeRepo, mockObjRepo)

	input := &models.CreateRelationshipRequest{
		SourceObjectPublicID: testSourcePublicID.String(),
		TargetObjectPublicID: testTargetPublicID.String(),
		RelationshipTypeKey:  "related_to",
	}

	result, err := service.Create(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "max_count")
}

func TestRelationshipService_Create_SourceAndTargetSame(t *testing.T) {
	service := NewRelationshipService(
		&mockRelationshipRepositoryForRelationshipService{},
		&mockRelationshipTypeRepository{},
		&mockObjectRepository{},
	)

	input := &models.CreateRelationshipRequest{
		SourceObjectPublicID: testSourcePublicID.String(),
		TargetObjectPublicID: testSourcePublicID.String(),
		RelationshipTypeKey:  "contains",
	}

	result, err := service.Create(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "source and target cannot be the same")
}

func TestRelationshipService_Create_ExistsCheckFails(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		existsFunc: func(ctx context.Context, sourceObjectID, targetObjectID int64, typeObjectID int64) (bool, error) {
			return false, errors.New("database error")
		},
	}

	mockRelTypeRepo := &mockRelationshipTypeRepository{
		getByTypeKeyFunc: func(ctx context.Context, typeKey string) (*models.RelationshipType, error) {
			return &models.RelationshipType{ObjectID: 1, TypeKey: typeKey}, nil
		},
	}

	mockObjRepo := &mockObjectRepository{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Object, error) {
			return &models.Object{ID: 1, PublicID: publicID}, nil
		},
	}

	service := NewRelationshipService(mockRelRepo, mockRelTypeRepo, mockObjRepo)

	input := &models.CreateRelationshipRequest{
		SourceObjectPublicID: testSourcePublicID.String(),
		TargetObjectPublicID: testTargetPublicID.String(),
		RelationshipTypeKey:  "contains",
	}

	result, err := service.Create(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to check relationship existence")
}

func TestRelationshipService_Create_CheckCircularFails(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		existsFunc: func(ctx context.Context, sourceObjectID, targetObjectID int64, typeObjectID int64) (bool, error) {
			return false, nil
		},
		checkCircularFunc: func(ctx context.Context, sourceObjectID, targetObjectID, typeObjectID int64) (bool, error) {
			return false, errors.New("database error")
		},
	}

	mockRelTypeRepo := &mockRelationshipTypeRepository{
		getByTypeKeyFunc: func(ctx context.Context, typeKey string) (*models.RelationshipType, error) {
			return &models.RelationshipType{ObjectID: 1, TypeKey: typeKey}, nil
		},
	}

	mockObjRepo := &mockObjectRepository{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Object, error) {
			return &models.Object{ID: 1, PublicID: publicID}, nil
		},
	}

	service := NewRelationshipService(mockRelRepo, mockRelTypeRepo, mockObjRepo)

	input := &models.CreateRelationshipRequest{
		SourceObjectPublicID: testSourcePublicID.String(),
		TargetObjectPublicID: testTargetPublicID.String(),
		RelationshipTypeKey:  "contains",
	}

	result, err := service.Create(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to check circular relationship")
}

func TestRelationshipService_Create_MaxCountZero(t *testing.T) {
	mockRelRepo := &mockRelationshipRepositoryForRelationshipService{
		existsFunc: func(ctx context.Context, sourceObjectID, targetObjectID int64, typeObjectID int64) (bool, error) {
			return false, nil
		},
		checkCircularFunc: func(ctx context.Context, sourceObjectID, targetObjectID, typeObjectID int64) (bool, error) {
			return false, nil
		},
		countForObjectFunc: func(ctx context.Context, objectID int64, typeKey *string) (int, error) {
			return 0, nil
		},
		createFunc: func(ctx context.Context, input *models.CreateRelationshipRequest) (*models.Relationship, error) {
			return &models.Relationship{RelationshipTypeKey: input.RelationshipTypeKey}, nil
		},
	}

	mockRelTypeRepo := &mockRelationshipTypeRepository{
		getByTypeKeyFunc: func(ctx context.Context, typeKey string) (*models.RelationshipType, error) {
			return &models.RelationshipType{
				ObjectID:    1,
				TypeKey:     typeKey,
				Cardinality: models.CardinalityOneToMany,
				MaxCount:   0,
			}, nil
		},
	}

	mockObjRepo := &mockObjectRepository{
		getByPublicIDFunc: func(ctx context.Context, publicID uuid.UUID) (*models.Object, error) {
			return &models.Object{ID: 1, PublicID: publicID}, nil
		},
	}

	service := NewRelationshipService(mockRelRepo, mockRelTypeRepo, mockObjRepo)

	input := &models.CreateRelationshipRequest{
		SourceObjectPublicID: testSourcePublicID.String(),
		TargetObjectPublicID: testTargetPublicID.String(),
		RelationshipTypeKey:  "exclusive_contains",
	}

	result, err := service.Create(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}