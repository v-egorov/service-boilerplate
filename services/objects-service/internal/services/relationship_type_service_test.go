package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/repository"
)

type mockRelationshipTypeRepository struct {
	createFunc              func(ctx context.Context, input *models.CreateRelationshipTypeRequest) (*models.RelationshipType, error)
	getByIDFunc             func(ctx context.Context, id int64) (*models.RelationshipType, error)
	getByTypeKeyFunc        func(ctx context.Context, typeKey string) (*models.RelationshipType, error)
	updateFunc              func(ctx context.Context, id int64, input *models.UpdateRelationshipTypeRequest) (*models.RelationshipType, error)
	deleteFunc              func(ctx context.Context, id int64) error
	listFunc                func(ctx context.Context, filter *models.RelationshipTypeFilter) ([]*models.RelationshipType, error)
	existsFunc              func(ctx context.Context, typeKey string) (bool, error)
	getByReverseTypeKeyFunc func(ctx context.Context, reverseKey string) (*models.RelationshipType, error)
}

func (m *mockRelationshipTypeRepository) DB() repository.DBInterface             { return nil }
func (m *mockRelationshipTypeRepository) Options() *repository.RepositoryOptions { return nil }
func (m *mockRelationshipTypeRepository) Metrics() *repository.RepositoryMetrics { return nil }
func (m *mockRelationshipTypeRepository) ResetMetrics()                          {}
func (m *mockRelationshipTypeRepository) Healthy(ctx context.Context) error      { return nil }

func (m *mockRelationshipTypeRepository) Create(ctx context.Context, input *models.CreateRelationshipTypeRequest) (*models.RelationshipType, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, input)
	}
	return &models.RelationshipType{ObjectID: 1, TypeKey: input.TypeKey}, nil
}

func (m *mockRelationshipTypeRepository) GetByID(ctx context.Context, id int64) (*models.RelationshipType, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return &models.RelationshipType{ObjectID: id, TypeKey: "test"}, nil
}

func (m *mockRelationshipTypeRepository) GetByTypeKey(ctx context.Context, typeKey string) (*models.RelationshipType, error) {
	if m.getByTypeKeyFunc != nil {
		return m.getByTypeKeyFunc(ctx, typeKey)
	}
	return &models.RelationshipType{ObjectID: 1, TypeKey: typeKey}, nil
}

func (m *mockRelationshipTypeRepository) Update(ctx context.Context, id int64, input *models.UpdateRelationshipTypeRequest) (*models.RelationshipType, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, id, input)
	}
	return &models.RelationshipType{ObjectID: id, TypeKey: "updated"}, nil
}

func (m *mockRelationshipTypeRepository) Delete(ctx context.Context, id int64) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockRelationshipTypeRepository) List(ctx context.Context, filter *models.RelationshipTypeFilter) ([]*models.RelationshipType, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filter)
	}
	return []*models.RelationshipType{}, nil
}

func (m *mockRelationshipTypeRepository) Exists(ctx context.Context, typeKey string) (bool, error) {
	if m.existsFunc != nil {
		return m.existsFunc(ctx, typeKey)
	}
	return false, nil
}

func (m *mockRelationshipTypeRepository) GetByReverseTypeKey(ctx context.Context, reverseKey string) (*models.RelationshipType, error) {
	if m.getByReverseTypeKeyFunc != nil {
		return m.getByReverseTypeKeyFunc(ctx, reverseKey)
	}
	return nil, nil
}

func TestRelationshipTypeService_Create_Success(t *testing.T) {
	mockRepo := &mockRelationshipTypeRepository{
		existsFunc: func(ctx context.Context, typeKey string) (bool, error) {
			return false, nil
		},
		createFunc: func(ctx context.Context, input *models.CreateRelationshipTypeRequest) (*models.RelationshipType, error) {
			return &models.RelationshipType{
				ObjectID:    1,
				TypeKey:     input.TypeKey,
				Cardinality: input.Cardinality,
			}, nil
		},
	}

	service := NewRelationshipTypeService(mockRepo)
	req := &models.CreateRelationshipTypeRequest{
		TypeKey:          "contains",
		Cardinality:      models.CardinalityOneToMany,
		RelationshipName: "contains",
	}

	result, err := service.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, "contains", result.TypeKey)
	assert.Equal(t, models.CardinalityOneToMany, result.Cardinality)
}

func TestRelationshipTypeService_Create_DuplicateTypeKey(t *testing.T) {
	mockRepo := &mockRelationshipTypeRepository{
		existsFunc: func(ctx context.Context, typeKey string) (bool, error) {
			return true, nil
		},
	}

	service := NewRelationshipTypeService(mockRepo)
	req := &models.CreateRelationshipTypeRequest{
		TypeKey:     "contains",
		Cardinality: models.CardinalityOneToMany,
	}

	result, err := service.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "already exists")
}

func TestRelationshipTypeService_Create_InvalidCardinality(t *testing.T) {
	mockRepo := &mockRelationshipTypeRepository{}

	service := NewRelationshipTypeService(mockRepo)
	req := &models.CreateRelationshipTypeRequest{
		TypeKey:     "test",
		Cardinality: "invalid",
	}

	result, err := service.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid cardinality")
}

func TestRelationshipTypeService_Create_InvalidReverseTypeKey(t *testing.T) {
	reverseKey := "contained_by"
	mockRepo := &mockRelationshipTypeRepository{
		existsFunc: func(ctx context.Context, typeKey string) (bool, error) {
			if typeKey == "contains" {
				return false, nil
			}
			return false, nil
		},
		createFunc: func(ctx context.Context, input *models.CreateRelationshipTypeRequest) (*models.RelationshipType, error) {
			return nil, repository.ErrNotFound
		},
	}

	service := NewRelationshipTypeService(mockRepo)
	req := &models.CreateRelationshipTypeRequest{
		TypeKey:        "contains",
		Cardinality:    models.CardinalityOneToMany,
		ReverseTypeKey: &reverseKey,
	}

	result, err := service.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestRelationshipTypeService_GetByTypeKey_Success(t *testing.T) {
	mockRepo := &mockRelationshipTypeRepository{
		getByTypeKeyFunc: func(ctx context.Context, typeKey string) (*models.RelationshipType, error) {
			return &models.RelationshipType{
				ObjectID:    1,
				TypeKey:     typeKey,
				Cardinality: models.CardinalityOneToMany,
			}, nil
		},
	}

	service := NewRelationshipTypeService(mockRepo)
	result, err := service.GetByTypeKey(context.Background(), "contains")

	assert.NoError(t, err)
	assert.Equal(t, "contains", result.TypeKey)
}

func TestRelationshipTypeService_GetByTypeKey_Empty(t *testing.T) {
	mockRepo := &mockRelationshipTypeRepository{}

	service := NewRelationshipTypeService(mockRepo)
	result, err := service.GetByTypeKey(context.Background(), "")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestRelationshipTypeService_GetByTypeKey_NotFound(t *testing.T) {
	mockRepo := &mockRelationshipTypeRepository{
		getByTypeKeyFunc: func(ctx context.Context, typeKey string) (*models.RelationshipType, error) {
			return nil, repository.ErrNotFound
		},
	}

	service := NewRelationshipTypeService(mockRepo)
	result, err := service.GetByTypeKey(context.Background(), "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
}

func TestRelationshipTypeService_Update_Success(t *testing.T) {
	mockRepo := &mockRelationshipTypeRepository{
		getByTypeKeyFunc: func(ctx context.Context, typeKey string) (*models.RelationshipType, error) {
			return &models.RelationshipType{
				ObjectID:    1,
				TypeKey:     typeKey,
				Cardinality: models.CardinalityOneToMany,
			}, nil
		},
		updateFunc: func(ctx context.Context, id int64, input *models.UpdateRelationshipTypeRequest) (*models.RelationshipType, error) {
			return &models.RelationshipType{
				ObjectID:    id,
				TypeKey:     "contains",
				Cardinality: *input.Cardinality,
			}, nil
		},
	}

	service := NewRelationshipTypeService(mockRepo)
	cardinality := models.CardinalityManyToMany
	req := &models.UpdateRelationshipTypeRequest{
		Cardinality: &cardinality,
	}

	result, err := service.Update(context.Background(), "contains", req)

	assert.NoError(t, err)
	assert.Equal(t, models.CardinalityManyToMany, result.Cardinality)
}

func TestRelationshipTypeService_Update_InvalidReverseTypeKey(t *testing.T) {
	reverseKey := "nonexistent"
	mockRepo := &mockRelationshipTypeRepository{
		getByTypeKeyFunc: func(ctx context.Context, typeKey string) (*models.RelationshipType, error) {
			return &models.RelationshipType{
				ObjectID:    1,
				TypeKey:     typeKey,
				Cardinality: models.CardinalityOneToMany,
			}, nil
		},
		existsFunc: func(ctx context.Context, typeKey string) (bool, error) {
			return false, nil
		},
	}

	service := NewRelationshipTypeService(mockRepo)
	req := &models.UpdateRelationshipTypeRequest{
		ReverseTypeKey: &reverseKey,
	}

	result, err := service.Update(context.Background(), "contains", req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestRelationshipTypeService_Delete_Success(t *testing.T) {
	mockRepo := &mockRelationshipTypeRepository{
		getByTypeKeyFunc: func(ctx context.Context, typeKey string) (*models.RelationshipType, error) {
			return &models.RelationshipType{
				ObjectID:    1,
				TypeKey:     typeKey,
				Cardinality: models.CardinalityOneToMany,
			}, nil
		},
		deleteFunc: func(ctx context.Context, id int64) error {
			return nil
		},
	}

	service := NewRelationshipTypeService(mockRepo)
	err := service.Delete(context.Background(), "contains")

	assert.NoError(t, err)
}

func TestRelationshipTypeService_Delete_NotFound(t *testing.T) {
	mockRepo := &mockRelationshipTypeRepository{
		getByTypeKeyFunc: func(ctx context.Context, typeKey string) (*models.RelationshipType, error) {
			return nil, repository.ErrNotFound
		},
	}

	service := NewRelationshipTypeService(mockRepo)
	err := service.Delete(context.Background(), "nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRelationshipTypeService_List_Success(t *testing.T) {
	mockRepo := &mockRelationshipTypeRepository{
		listFunc: func(ctx context.Context, filter *models.RelationshipTypeFilter) ([]*models.RelationshipType, error) {
			return []*models.RelationshipType{
				{ObjectID: 1, TypeKey: "contains", Cardinality: models.CardinalityOneToMany},
				{ObjectID: 2, TypeKey: "belongs_to", Cardinality: models.CardinalityManyToOne},
			}, nil
		},
	}

	service := NewRelationshipTypeService(mockRepo)
	result, err := service.List(context.Background(), &models.RelationshipTypeFilter{})

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestRelationshipTypeService_List_WithFilter(t *testing.T) {
	mockRepo := &mockRelationshipTypeRepository{
		listFunc: func(ctx context.Context, filter *models.RelationshipTypeFilter) ([]*models.RelationshipType, error) {
			assert.Equal(t, models.CardinalityOneToMany, filter.Cardinality)
			return []*models.RelationshipType{
				{ObjectID: 1, TypeKey: "contains", Cardinality: models.CardinalityOneToMany},
			}, nil
		},
	}

	service := NewRelationshipTypeService(mockRepo)
	filter := &models.RelationshipTypeFilter{
		Cardinality: models.CardinalityOneToMany,
	}
	result, err := service.List(context.Background(), filter)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestRelationshipTypeService_Create_MinCountGreaterThanMaxCount(t *testing.T) {
	mockRepo := &mockRelationshipTypeRepository{}

	service := NewRelationshipTypeService(mockRepo)
	req := &models.CreateRelationshipTypeRequest{
		TypeKey:     "test",
		Cardinality: models.CardinalityOneToMany,
		MinCount:    10,
		MaxCount:    5,
	}

	result, err := service.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "min_count cannot exceed max_count")
}
