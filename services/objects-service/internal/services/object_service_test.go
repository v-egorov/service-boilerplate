package services

// Stub tests for ObjectService
//
// These are placeholder tests that verify method signatures compile correctly.
// They use minimal mocks and do not test full service logic.
//
// TODO: Replace with proper integration tests in Phase 8
// - Test actual service validation logic
// - Test sealed type restrictions
// - Test parent-child type compatibility
// - Test bulk operations with transactions
// - Test metadata/tags business rules

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/repository"
)

type mockObjectRepository struct {
	createFunc              func(ctx context.Context, input *models.CreateObjectRequest) (*models.Object, error)
	getByIDFunc             func(ctx context.Context, id int64) (*models.Object, error)
	getByPublicIDFunc       func(ctx context.Context, publicID uuid.UUID) (*models.Object, error)
	getByNameFunc           func(ctx context.Context, name string) (*models.Object, error)
	updateFunc              func(ctx context.Context, id int64, input *models.UpdateObjectRequest) (*models.Object, error)
	deleteFunc              func(ctx context.Context, id int64) error
	listFunc                func(ctx context.Context, filter *models.ObjectFilter) ([]*models.Object, int64, error)
	searchFunc              func(ctx context.Context, query string, limit int) ([]*models.Object, error)
	findByMetadataFunc      func(ctx context.Context, key, value string) ([]*models.Object, error)
	findByTagsFunc          func(ctx context.Context, tags []string, matchAll bool) ([]*models.Object, error)
	updateMetadataFunc      func(ctx context.Context, id int64, metadata map[string]interface{}) error
	addTagsFunc             func(ctx context.Context, id int64, tags []string) error
	removeTagsFunc          func(ctx context.Context, id int64, tags []string) error
	getChildrenFunc         func(ctx context.Context, parentID int64) ([]*models.Object, error)
	getDescendantsFunc      func(ctx context.Context, rootID int64, maxDepth *int) ([]*models.Object, error)
	getAncestorsFunc        func(ctx context.Context, id int64) ([]*models.Object, error)
	getPathFunc             func(ctx context.Context, id int64) ([]*models.Object, error)
	bulkCreateFunc          func(ctx context.Context, objects []*models.CreateObjectRequest) ([]*models.Object, error)
	bulkUpdateFunc          func(ctx context.Context, ids []int64, updates *models.UpdateObjectRequest) ([]*models.Object, error)
	bulkDeleteFunc          func(ctx context.Context, ids []int64) error
	validateParentChildFunc func(ctx context.Context, parentID, childID int64) error
	getObjectStatsFunc      func(ctx context.Context, filter *models.ObjectFilter) (*repository.ObjectStats, error)
}

func (m *mockObjectRepository) Create(ctx context.Context, input *models.CreateObjectRequest) (*models.Object, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, input)
	}
	return &models.Object{ID: 1, Name: input.Name}, nil
}

func (m *mockObjectRepository) GetByID(ctx context.Context, id int64) (*models.Object, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return &models.Object{ID: id, Name: "test"}, nil
}

func (m *mockObjectRepository) GetByPublicID(ctx context.Context, publicID uuid.UUID) (*models.Object, error) {
	if m.getByPublicIDFunc != nil {
		return m.getByPublicIDFunc(ctx, publicID)
	}
	return &models.Object{ID: 1}, nil
}

func (m *mockObjectRepository) GetByName(ctx context.Context, name string) (*models.Object, error) {
	if m.getByNameFunc != nil {
		return m.getByNameFunc(ctx, name)
	}
	return &models.Object{ID: 1, Name: name}, nil
}

func (m *mockObjectRepository) Update(ctx context.Context, id int64, input *models.UpdateObjectRequest) (*models.Object, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, id, input)
	}
	return &models.Object{ID: id, Name: "updated"}, nil
}

func (m *mockObjectRepository) Delete(ctx context.Context, id int64) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockObjectRepository) List(ctx context.Context, filter *models.ObjectFilter) ([]*models.Object, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filter)
	}
	return nil, 0, nil
}

func (m *mockObjectRepository) Search(ctx context.Context, query string, limit int) ([]*models.Object, error) {
	if m.searchFunc != nil {
		return m.searchFunc(ctx, query, limit)
	}
	return nil, nil
}

func (m *mockObjectRepository) FindByMetadata(ctx context.Context, key, value string) ([]*models.Object, error) {
	if m.findByMetadataFunc != nil {
		return m.findByMetadataFunc(ctx, key, value)
	}
	return nil, nil
}

func (m *mockObjectRepository) FindByTags(ctx context.Context, tags []string, matchAll bool) ([]*models.Object, error) {
	if m.findByTagsFunc != nil {
		return m.findByTagsFunc(ctx, tags, matchAll)
	}
	return nil, nil
}

func (m *mockObjectRepository) UpdateMetadata(ctx context.Context, id int64, metadata map[string]interface{}) error {
	if m.updateMetadataFunc != nil {
		return m.updateMetadataFunc(ctx, id, metadata)
	}
	return nil
}

func (m *mockObjectRepository) AddTags(ctx context.Context, id int64, tags []string) error {
	if m.addTagsFunc != nil {
		return m.addTagsFunc(ctx, id, tags)
	}
	return nil
}

func (m *mockObjectRepository) RemoveTags(ctx context.Context, id int64, tags []string) error {
	if m.removeTagsFunc != nil {
		return m.removeTagsFunc(ctx, id, tags)
	}
	return nil
}

func (m *mockObjectRepository) GetChildren(ctx context.Context, parentID int64) ([]*models.Object, error) {
	if m.getChildrenFunc != nil {
		return m.getChildrenFunc(ctx, parentID)
	}
	return nil, nil
}

func (m *mockObjectRepository) GetDescendants(ctx context.Context, rootID int64, maxDepth *int) ([]*models.Object, error) {
	if m.getDescendantsFunc != nil {
		return m.getDescendantsFunc(ctx, rootID, maxDepth)
	}
	return nil, nil
}

func (m *mockObjectRepository) GetAncestors(ctx context.Context, id int64) ([]*models.Object, error) {
	if m.getAncestorsFunc != nil {
		return m.getAncestorsFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockObjectRepository) GetPath(ctx context.Context, id int64) ([]*models.Object, error) {
	if m.getPathFunc != nil {
		return m.getPathFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockObjectRepository) BulkCreate(ctx context.Context, objects []*models.CreateObjectRequest) ([]*models.Object, error) {
	if m.bulkCreateFunc != nil {
		return m.bulkCreateFunc(ctx, objects)
	}
	return nil, nil
}

func (m *mockObjectRepository) BulkUpdate(ctx context.Context, ids []int64, updates *models.UpdateObjectRequest) ([]*models.Object, error) {
	if m.bulkUpdateFunc != nil {
		return m.bulkUpdateFunc(ctx, ids, updates)
	}
	return nil, nil
}

func (m *mockObjectRepository) BulkDelete(ctx context.Context, ids []int64) error {
	if m.bulkDeleteFunc != nil {
		return m.bulkDeleteFunc(ctx, ids)
	}
	return nil
}

func (m *mockObjectRepository) ValidateObjectType(ctx context.Context, objectTypeID int64) error {
	return nil
}

func (m *mockObjectRepository) ValidateParentChild(ctx context.Context, parentID, childID int64) error {
	if m.validateParentChildFunc != nil {
		return m.validateParentChildFunc(ctx, parentID, childID)
	}
	return nil
}

func (m *mockObjectRepository) CanDelete(ctx context.Context, id int64) (bool, error) {
	return true, nil
}

func (m *mockObjectRepository) GetObjectStats(ctx context.Context, filter *models.ObjectFilter) (*repository.ObjectStats, error) {
	if m.getObjectStatsFunc != nil {
		return m.getObjectStatsFunc(ctx, filter)
	}
	return &repository.ObjectStats{}, nil
}

func (m *mockObjectRepository) DB() repository.DBInterface             { return nil }
func (m *mockObjectRepository) Options() *repository.RepositoryOptions { return nil }
func (m *mockObjectRepository) Metrics() *repository.RepositoryMetrics { return nil }
func (m *mockObjectRepository) ResetMetrics()                          {}
func (m *mockObjectRepository) Healthy(ctx context.Context) error      { return nil }

type mockObjectTypeRepositoryForObjectService struct {
	getByIDFunc func(ctx context.Context, id int64) (*models.ObjectType, error)
}

func (m *mockObjectTypeRepositoryForObjectService) Create(ctx context.Context, input *models.CreateObjectTypeRequest) (*models.ObjectType, error) {
	return &models.ObjectType{}, nil
}

func (m *mockObjectTypeRepositoryForObjectService) GetByID(ctx context.Context, id int64) (*models.ObjectType, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return &models.ObjectType{ID: id, Name: "test"}, nil
}

func (m *mockObjectTypeRepositoryForObjectService) GetByName(ctx context.Context, name string) (*models.ObjectType, error) {
	return &models.ObjectType{ID: 1, Name: name}, nil
}

func (m *mockObjectTypeRepositoryForObjectService) Update(ctx context.Context, id int64, input *models.UpdateObjectTypeRequest) (*models.ObjectType, error) {
	return &models.ObjectType{ID: id}, nil
}

func (m *mockObjectTypeRepositoryForObjectService) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockObjectTypeRepositoryForObjectService) GetTree(ctx context.Context, rootID *int64) ([]*models.ObjectType, error) {
	return nil, nil
}

func (m *mockObjectTypeRepositoryForObjectService) GetChildren(ctx context.Context, parentID int64) ([]*models.ObjectType, error) {
	return nil, nil
}

func (m *mockObjectTypeRepositoryForObjectService) GetDescendants(ctx context.Context, rootID int64, maxDepth *int) ([]*models.ObjectType, error) {
	return nil, nil
}

func (m *mockObjectTypeRepositoryForObjectService) GetAncestors(ctx context.Context, id int64) ([]*models.ObjectType, error) {
	return nil, nil
}

func (m *mockObjectTypeRepositoryForObjectService) GetPath(ctx context.Context, id int64) ([]*models.ObjectType, error) {
	return nil, nil
}

func (m *mockObjectTypeRepositoryForObjectService) List(ctx context.Context, filter *models.ObjectTypeFilter) ([]*models.ObjectType, error) {
	return nil, nil
}

func (m *mockObjectTypeRepositoryForObjectService) Search(ctx context.Context, query string, limit int) ([]*models.ObjectType, error) {
	return nil, nil
}

func (m *mockObjectTypeRepositoryForObjectService) ValidateMove(ctx context.Context, id int64, newParentID *int64) error {
	return nil
}

func (m *mockObjectTypeRepositoryForObjectService) GetSubtreeObjectCount(ctx context.Context, id int64) (int64, error) {
	return 0, nil
}

func (m *mockObjectTypeRepositoryForObjectService) CanDelete(ctx context.Context, id int64) (bool, error) {
	return true, nil
}

func (m *mockObjectTypeRepositoryForObjectService) ValidateParentChild(ctx context.Context, parentID, childID int64) error {
	return nil
}

func (m *mockObjectTypeRepositoryForObjectService) DB() repository.DBInterface { return nil }
func (m *mockObjectTypeRepositoryForObjectService) Options() *repository.RepositoryOptions {
	return nil
}
func (m *mockObjectTypeRepositoryForObjectService) Metrics() *repository.RepositoryMetrics {
	return nil
}
func (m *mockObjectTypeRepositoryForObjectService) ResetMetrics()                     {}
func (m *mockObjectTypeRepositoryForObjectService) Healthy(ctx context.Context) error { return nil }

// TODO: Replace with proper test - minimal stub
func TestObjectService_Create_EmptyName(t *testing.T) {
	mockRepo := &mockObjectRepository{}
	mockTypeRepo := &mockObjectTypeRepositoryForObjectService{}
	service := NewObjectService(mockRepo, mockTypeRepo)

	_, err := service.Create(context.Background(), &models.CreateObjectRequest{Name: ""})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

// TODO: Replace with proper test - minimal stub
func TestObjectService_Create_InvalidObjectTypeID(t *testing.T) {
	mockRepo := &mockObjectRepository{}
	mockTypeRepo := &mockObjectTypeRepositoryForObjectService{}
	service := NewObjectService(mockRepo, mockTypeRepo)

	_, err := service.Create(context.Background(), &models.CreateObjectRequest{Name: "test", ObjectTypeID: 0})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "object_type_id is required")
}

// TODO: Replace with proper test - minimal stub
func TestObjectService_Create_SealedType(t *testing.T) {
	mockRepo := &mockObjectRepository{}
	mockTypeRepo := &mockObjectTypeRepositoryForObjectService{
		getByIDFunc: func(ctx context.Context, id int64) (*models.ObjectType, error) {
			return &models.ObjectType{ID: id, IsSealed: true}, nil
		},
	}
	service := NewObjectService(mockRepo, mockTypeRepo)

	_, err := service.Create(context.Background(), &models.CreateObjectRequest{Name: "test", ObjectTypeID: 1})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sealed type")
}

// TODO: Replace with proper test - minimal stub
func TestObjectService_Create_Success(t *testing.T) {
	mockRepo := &mockObjectRepository{
		createFunc: func(ctx context.Context, input *models.CreateObjectRequest) (*models.Object, error) {
			return &models.Object{ID: 1, Name: input.Name}, nil
		},
	}
	mockTypeRepo := &mockObjectTypeRepositoryForObjectService{
		getByIDFunc: func(ctx context.Context, id int64) (*models.ObjectType, error) {
			return &models.ObjectType{ID: id, IsSealed: false}, nil
		},
	}
	service := NewObjectService(mockRepo, mockTypeRepo)

	result, err := service.Create(context.Background(), &models.CreateObjectRequest{Name: "test", ObjectTypeID: 1})
	assert.NoError(t, err)
	assert.Equal(t, "test", result.Name)
}

// TODO: Replace with proper test - minimal stub
func TestObjectService_GetByID_InvalidID(t *testing.T) {
	mockRepo := &mockObjectRepository{}
	mockTypeRepo := &mockObjectTypeRepositoryForObjectService{}
	service := NewObjectService(mockRepo, mockTypeRepo)

	_, err := service.GetByID(context.Background(), 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid id")
}

// TODO: Replace with proper test - minimal stub
func TestObjectService_GetByID_NotFound(t *testing.T) {
	mockRepo := &mockObjectRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Object, error) {
			return nil, repository.ErrNotFound
		},
	}
	mockTypeRepo := &mockObjectTypeRepositoryForObjectService{}
	service := NewObjectService(mockRepo, mockTypeRepo)

	_, err := service.GetByID(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TODO: Replace with proper test - minimal stub
func TestObjectService_GetByPublicID_InvalidID(t *testing.T) {
	mockRepo := &mockObjectRepository{}
	mockTypeRepo := &mockObjectTypeRepositoryForObjectService{}
	service := NewObjectService(mockRepo, mockTypeRepo)

	_, err := service.GetByPublicID(context.Background(), uuid.Nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid public id")
}

// TODO: Replace with proper test - minimal stub
func TestObjectService_Update_SelfParent(t *testing.T) {
	mockRepo := &mockObjectRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Object, error) {
			return &models.Object{ID: id, Name: "test"}, nil
		},
	}
	mockTypeRepo := &mockObjectTypeRepositoryForObjectService{
		getByIDFunc: func(ctx context.Context, id int64) (*models.ObjectType, error) {
			return &models.ObjectType{ID: id, IsSealed: false}, nil
		},
	}
	service := NewObjectService(mockRepo, mockTypeRepo)

	parentID := int64(1)
	_, err := service.Update(context.Background(), 1, &models.UpdateObjectRequest{ParentObjectID: &parentID})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "own parent")
}

// TODO: Replace with proper test - minimal stub
func TestObjectService_Delete_Success(t *testing.T) {
	mockRepo := &mockObjectRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.Object, error) {
			return &models.Object{ID: id, ObjectTypeID: 1}, nil
		},
		deleteFunc: func(ctx context.Context, id int64) error {
			return nil
		},
	}
	mockTypeRepo := &mockObjectTypeRepositoryForObjectService{
		getByIDFunc: func(ctx context.Context, id int64) (*models.ObjectType, error) {
			return &models.ObjectType{ID: id, IsSealed: false}, nil
		},
	}
	service := NewObjectService(mockRepo, mockTypeRepo)

	err := service.Delete(context.Background(), 1)
	assert.NoError(t, err)
}

// TODO: Replace with proper test - minimal stub
func TestObjectService_Search_EmptyQuery(t *testing.T) {
	mockRepo := &mockObjectRepository{}
	mockTypeRepo := &mockObjectTypeRepositoryForObjectService{}
	service := NewObjectService(mockRepo, mockTypeRepo)

	_, err := service.Search(context.Background(), "", 10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query is required")
}

// TODO: Replace with proper test - minimal stub
func TestObjectService_FindByTags_EmptyTags(t *testing.T) {
	mockRepo := &mockObjectRepository{}
	mockTypeRepo := &mockObjectTypeRepositoryForObjectService{}
	service := NewObjectService(mockRepo, mockTypeRepo)

	_, err := service.FindByTags(context.Background(), []string{}, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tag is required")
}

// TODO: Replace with proper test - minimal stub
func TestObjectService_UpdateMetadata_InvalidID(t *testing.T) {
	mockRepo := &mockObjectRepository{}
	mockTypeRepo := &mockObjectTypeRepositoryForObjectService{}
	service := NewObjectService(mockRepo, mockTypeRepo)

	err := service.UpdateMetadata(context.Background(), 0, map[string]interface{}{"key": "value"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid id")
}

// TODO: Replace with proper test - minimal stub
func TestObjectService_UpdateMetadata_NilMetadata(t *testing.T) {
	mockRepo := &mockObjectRepository{}
	mockTypeRepo := &mockObjectTypeRepositoryForObjectService{}
	service := NewObjectService(mockRepo, mockTypeRepo)

	err := service.UpdateMetadata(context.Background(), 1, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "metadata cannot be nil")
}

// TODO: Replace with proper test - minimal stub
func TestObjectService_AddTags_InvalidID(t *testing.T) {
	mockRepo := &mockObjectRepository{}
	mockTypeRepo := &mockObjectTypeRepositoryForObjectService{}
	service := NewObjectService(mockRepo, mockTypeRepo)

	err := service.AddTags(context.Background(), 0, []string{"tag1"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid id")
}

// TODO: Replace with proper test - minimal stub
func TestObjectService_AddTags_EmptyTags(t *testing.T) {
	mockRepo := &mockObjectRepository{}
	mockTypeRepo := &mockObjectTypeRepositoryForObjectService{}
	service := NewObjectService(mockRepo, mockTypeRepo)

	err := service.AddTags(context.Background(), 1, []string{})
	assert.NoError(t, err)
}

// TODO: Replace with proper test - minimal stub
func TestObjectService_RemoveTags_InvalidID(t *testing.T) {
	mockRepo := &mockObjectRepository{}
	mockTypeRepo := &mockObjectTypeRepositoryForObjectService{}
	service := NewObjectService(mockRepo, mockTypeRepo)

	err := service.RemoveTags(context.Background(), 0, []string{"tag1"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid id")
}

// TODO: Replace with proper test - minimal stub
func TestObjectService_GetChildren_InvalidParentID(t *testing.T) {
	mockRepo := &mockObjectRepository{}
	mockTypeRepo := &mockObjectTypeRepositoryForObjectService{}
	service := NewObjectService(mockRepo, mockTypeRepo)

	_, err := service.GetChildren(context.Background(), 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid parent id")
}

// TODO: Replace with proper test - minimal stub
func TestObjectService_GetDescendants_InvalidRootID(t *testing.T) {
	mockRepo := &mockObjectRepository{}
	mockTypeRepo := &mockObjectTypeRepositoryForObjectService{}
	service := NewObjectService(mockRepo, mockTypeRepo)

	_, err := service.GetDescendants(context.Background(), 0, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid root id")
}

// TODO: Replace with proper test - minimal stub
func TestObjectService_ValidateParentChild_InvalidParentID(t *testing.T) {
	mockRepo := &mockObjectRepository{}
	mockTypeRepo := &mockObjectTypeRepositoryForObjectService{}
	service := NewObjectService(mockRepo, mockTypeRepo)

	err := service.ValidateParentChild(context.Background(), 0, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid parent id")
}

// TODO: Replace with proper test - minimal stub
func TestObjectService_ValidateParentChild_InvalidChildID(t *testing.T) {
	mockRepo := &mockObjectRepository{}
	mockTypeRepo := &mockObjectTypeRepositoryForObjectService{}
	service := NewObjectService(mockRepo, mockTypeRepo)

	err := service.ValidateParentChild(context.Background(), 1, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid child id")
}

// TODO: Replace with proper test - minimal stub
func TestObjectService_BulkCreate_EmptyObjects(t *testing.T) {
	mockRepo := &mockObjectRepository{}
	mockTypeRepo := &mockObjectTypeRepositoryForObjectService{}
	service := NewObjectService(mockRepo, mockTypeRepo)

	result, err := service.BulkCreate(context.Background(), []*models.CreateObjectRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

// TODO: Replace with proper test - minimal stub
func TestObjectService_BulkCreate_InvalidObjectType(t *testing.T) {
	mockRepo := &mockObjectRepository{}
	mockTypeRepo := &mockObjectTypeRepositoryForObjectService{
		getByIDFunc: func(ctx context.Context, id int64) (*models.ObjectType, error) {
			return nil, repository.ErrNotFound
		},
	}
	service := NewObjectService(mockRepo, mockTypeRepo)

	_, err := service.BulkCreate(context.Background(), []*models.CreateObjectRequest{{ObjectTypeID: 1}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid object type")
}

// TODO: Replace with proper test - minimal stub
func TestObjectService_BulkCreate_Success(t *testing.T) {
	mockRepo := &mockObjectRepository{
		bulkCreateFunc: func(ctx context.Context, objects []*models.CreateObjectRequest) ([]*models.Object, error) {
			results := make([]*models.Object, len(objects))
			for i, obj := range objects {
				results[i] = &models.Object{ID: int64(i + 1), Name: obj.Name}
			}
			return results, nil
		},
	}
	mockTypeRepo := &mockObjectTypeRepositoryForObjectService{
		getByIDFunc: func(ctx context.Context, id int64) (*models.ObjectType, error) {
			return &models.ObjectType{ID: id, IsSealed: false}, nil
		},
	}
	service := NewObjectService(mockRepo, mockTypeRepo)

	result, err := service.BulkCreate(context.Background(), []*models.CreateObjectRequest{
		{Name: "obj1", ObjectTypeID: 1},
		{Name: "obj2", ObjectTypeID: 1},
	})
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// TODO: Replace with proper test - minimal stub
func TestObjectService_BulkUpdate_EmptyIDs(t *testing.T) {
	mockRepo := &mockObjectRepository{}
	mockTypeRepo := &mockObjectTypeRepositoryForObjectService{}
	service := NewObjectService(mockRepo, mockTypeRepo)

	result, err := service.BulkUpdate(context.Background(), []int64{}, &models.UpdateObjectRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

// TODO: Replace with proper test - minimal stub
func TestObjectService_BulkUpdate_NilUpdates(t *testing.T) {
	mockRepo := &mockObjectRepository{}
	mockTypeRepo := &mockObjectTypeRepositoryForObjectService{}
	service := NewObjectService(mockRepo, mockTypeRepo)

	_, err := service.BulkUpdate(context.Background(), []int64{1, 2}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "updates cannot be nil")
}

// TODO: Replace with proper test - minimal stub
func TestObjectService_BulkDelete_EmptyIDs(t *testing.T) {
	mockRepo := &mockObjectRepository{}
	mockTypeRepo := &mockObjectTypeRepositoryForObjectService{}
	service := NewObjectService(mockRepo, mockTypeRepo)

	err := service.BulkDelete(context.Background(), []int64{})
	assert.NoError(t, err)
}
