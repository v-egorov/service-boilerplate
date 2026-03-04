package handlers

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/repository"
)

type MockObjectService struct {
	mock.Mock
}

func (m *MockObjectService) Create(ctx context.Context, req *models.CreateObjectRequest) (*models.Object, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Object), args.Error(1)
}

func (m *MockObjectService) GetByID(ctx context.Context, id int64) (*models.Object, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Object), args.Error(1)
}

func (m *MockObjectService) GetByPublicID(ctx context.Context, publicID uuid.UUID) (*models.Object, error) {
	args := m.Called(ctx, publicID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Object), args.Error(1)
}

func (m *MockObjectService) GetByName(ctx context.Context, name string) (*models.Object, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Object), args.Error(1)
}

func (m *MockObjectService) Update(ctx context.Context, id int64, req *models.UpdateObjectRequest) (*models.Object, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Object), args.Error(1)
}

func (m *MockObjectService) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockObjectService) List(ctx context.Context, filter *models.ObjectFilter) ([]*models.Object, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.Object), args.Get(1).(int64), args.Error(2)
}

func (m *MockObjectService) Search(ctx context.Context, query string, limit int) ([]*models.Object, error) {
	args := m.Called(ctx, query, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Object), args.Error(1)
}

func (m *MockObjectService) FindByMetadata(ctx context.Context, key, value string) ([]*models.Object, error) {
	args := m.Called(ctx, key, value)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Object), args.Error(1)
}

func (m *MockObjectService) FindByTags(ctx context.Context, tags []string, matchAll bool) ([]*models.Object, error) {
	args := m.Called(ctx, tags, matchAll)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Object), args.Error(1)
}

func (m *MockObjectService) UpdateMetadata(ctx context.Context, id int64, metadata map[string]interface{}, updatedBy string) error {
	args := m.Called(ctx, id, metadata, updatedBy)
	return args.Error(0)
}

func (m *MockObjectService) AddTags(ctx context.Context, id int64, tags []string, updatedBy string) error {
	args := m.Called(ctx, id, tags, updatedBy)
	return args.Error(0)
}

func (m *MockObjectService) RemoveTags(ctx context.Context, id int64, tags []string, updatedBy string) error {
	args := m.Called(ctx, id, tags, updatedBy)
	return args.Error(0)
}

func (m *MockObjectService) GetChildren(ctx context.Context, parentID int64) ([]*models.Object, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Object), args.Error(1)
}

func (m *MockObjectService) GetDescendants(ctx context.Context, rootID int64, maxDepth *int) ([]*models.Object, error) {
	args := m.Called(ctx, rootID, maxDepth)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Object), args.Error(1)
}

func (m *MockObjectService) GetAncestors(ctx context.Context, id int64) ([]*models.Object, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Object), args.Error(1)
}

func (m *MockObjectService) GetPath(ctx context.Context, id int64) ([]*models.Object, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Object), args.Error(1)
}

func (m *MockObjectService) BulkCreate(ctx context.Context, objects []*models.CreateObjectRequest) ([]*models.Object, error) {
	args := m.Called(ctx, objects)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Object), args.Error(1)
}

func (m *MockObjectService) BulkUpdate(ctx context.Context, ids []int64, updates *models.UpdateObjectRequest) ([]*models.Object, error) {
	args := m.Called(ctx, ids, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Object), args.Error(1)
}

func (m *MockObjectService) BulkDelete(ctx context.Context, ids []int64) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *MockObjectService) ValidateParentChild(ctx context.Context, parentID, childID int64) error {
	args := m.Called(ctx, parentID, childID)
	return args.Error(0)
}

func (m *MockObjectService) GetObjectStats(ctx context.Context, filter *models.ObjectFilter) (*repository.ObjectStats, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.ObjectStats), args.Error(1)
}

func TestObjectHandler_Create(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectService{}

	handler := NewObjectHandlerWithInterface(mockService, logger)

	req := models.CreateObjectRequest{
		ObjectTypeID: 1,
		Name:         "TestObject",
		Description:  "Test description",
	}

	createdObj := &models.Object{
		ID:           1,
		ObjectTypeID: 1,
		Name:         "TestObject",
		Description:  "Test description",
		Status:       models.StatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	mockService.On("Create", mock.Anything, mock.AnythingOfType("*models.CreateObjectRequest")).Return(createdObj, nil)

	c, w := createTestGinContext("POST", "/api/v1/objects", req)

	handler.Create(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	mockService.AssertExpectations(t)
}

func TestObjectHandler_Create_InvalidBody(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectService{}

	handler := NewObjectHandlerWithInterface(mockService, logger)

	c, w := createTestGinContext("POST", "/api/v1/objects", "invalid json")

	handler.Create(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	mockService.AssertExpectations(t)
}

func TestObjectHandler_GetByID(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectService{}

	handler := NewObjectHandlerWithInterface(mockService, logger)

	object := &models.Object{
		ID:           1,
		ObjectTypeID: 1,
		Name:         "TestObject",
		Status:       models.StatusActive,
	}

	mockService.On("GetByID", mock.Anything, int64(1)).Return(object, nil)

	c, w := createTestGinContext("GET", "/api/v1/objects/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.GetByID(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestObjectHandler_GetByID_InvalidID(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectService{}

	handler := NewObjectHandlerWithInterface(mockService, logger)

	c, w := createTestGinContext("GET", "/api/v1/objects/invalid", nil)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}

	handler.GetByID(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	mockService.AssertExpectations(t)
}

func TestObjectHandler_GetByPublicID(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectService{}

	handler := NewObjectHandlerWithInterface(mockService, logger)

	publicID := uuid.New()
	object := &models.Object{
		ID:           1,
		PublicID:     publicID,
		ObjectTypeID: 1,
		Name:         "TestObject",
	}

	mockService.On("GetByPublicID", mock.Anything, publicID).Return(object, nil)

	c, w := createTestGinContext("GET", "/api/v1/objects/public-id/"+publicID.String(), nil)
	c.Params = gin.Params{{Key: "public_id", Value: publicID.String()}}

	handler.GetByPublicID(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestObjectHandler_GetByName(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectService{}

	handler := NewObjectHandlerWithInterface(mockService, logger)

	object := &models.Object{
		ID:           1,
		ObjectTypeID: 1,
		Name:         "TestObject",
	}

	mockService.On("GetByName", mock.Anything, "TestObject").Return(object, nil)

	c, w := createTestGinContext("GET", "/api/v1/objects/name/TestObject", nil)
	c.Params = gin.Params{{Key: "name", Value: "TestObject"}}

	handler.GetByName(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestObjectHandler_Update(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectService{}

	handler := NewObjectHandlerWithInterface(mockService, logger)

	req := models.UpdateObjectRequest{
		Name:        stringPtr("UpdatedObject"),
		Description: stringPtr("Updated description"),
	}

	existingObj := &models.Object{
		ID:           1,
		ObjectTypeID: 1,
		Name:         "ExistingObject",
		CreatedBy:    "user-123",
	}

	updatedObj := &models.Object{
		ID:           1,
		ObjectTypeID: 1,
		Name:         "UpdatedObject",
		Description:  "Updated description",
		CreatedBy:    "user-123",
	}

	mockService.On("GetByID", mock.Anything, int64(1)).Return(existingObj, nil)
	mockService.On("Update", mock.Anything, int64(1), mock.AnythingOfType("*models.UpdateObjectRequest")).Return(updatedObj, nil)

	c, w := createTestGinContext("PUT", "/api/v1/objects/1", req)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_id", "user-123")

	handler.Update(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestObjectHandler_Update_OwnershipDenied(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectService{}

	handler := NewObjectHandlerWithInterface(mockService, logger)

	req := models.UpdateObjectRequest{
		Name: stringPtr("UpdatedObject"),
	}

	existingObj := &models.Object{
		ID:        1,
		Name:      "ExistingObject",
		CreatedBy: "other-user",
	}

	mockService.On("GetByID", mock.Anything, int64(1)).Return(existingObj, nil)

	c, w := createTestGinContext("PUT", "/api/v1/objects/1", req)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_id", "user-123")

	handler.Update(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "You can only update your own objects")

	mockService.AssertExpectations(t)
}

func TestObjectHandler_Update_AdminCanUpdateAny(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectService{}

	handler := NewObjectHandlerWithInterface(mockService, logger)

	req := models.UpdateObjectRequest{
		Name: stringPtr("UpdatedObject"),
	}

	existingObj := &models.Object{
		ID:        1,
		Name:      "ExistingObject",
		CreatedBy: "other-user",
	}

	updatedObj := &models.Object{
		ID:        1,
		Name:      "UpdatedObject",
		CreatedBy: "other-user",
	}

	mockService.On("GetByID", mock.Anything, int64(1)).Return(existingObj, nil)
	mockService.On("Update", mock.Anything, int64(1), mock.AnythingOfType("*models.UpdateObjectRequest")).Return(updatedObj, nil)

	c, w := createTestGinContext("PUT", "/api/v1/objects/1", req)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_id", "admin-user")
	c.Set("user_roles", []string{"admin"})

	handler.Update(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestObjectHandler_Delete(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectService{}

	handler := NewObjectHandlerWithInterface(mockService, logger)

	existingObj := &models.Object{
		ID:        1,
		Name:      "Object",
		CreatedBy: "user-123",
	}

	mockService.On("GetByID", mock.Anything, int64(1)).Return(existingObj, nil)
	mockService.On("Delete", mock.Anything, int64(1)).Return(nil)

	c, w := createTestGinContext("DELETE", "/api/v1/objects/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_id", "user-123")

	handler.Delete(c)

	assert.Equal(t, http.StatusNoContent, w.Code)

	mockService.AssertExpectations(t)
}

func TestObjectHandler_Delete_OwnershipDenied(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectService{}

	handler := NewObjectHandlerWithInterface(mockService, logger)

	existingObj := &models.Object{
		ID:        1,
		Name:      "Object",
		CreatedBy: "other-user",
	}

	mockService.On("GetByID", mock.Anything, int64(1)).Return(existingObj, nil)

	c, w := createTestGinContext("DELETE", "/api/v1/objects/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_id", "user-123")

	handler.Delete(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "You can only delete your own objects")

	mockService.AssertExpectations(t)
}

func TestObjectHandler_Delete_AdminCanDeleteAny(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectService{}

	handler := NewObjectHandlerWithInterface(mockService, logger)

	existingObj := &models.Object{
		ID:        1,
		Name:      "Object",
		CreatedBy: "other-user",
	}

	mockService.On("GetByID", mock.Anything, int64(1)).Return(existingObj, nil)
	mockService.On("Delete", mock.Anything, int64(1)).Return(nil)

	c, w := createTestGinContext("DELETE", "/api/v1/objects/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_id", "admin-user")
	c.Set("user_roles", []string{"admin"})

	handler.Delete(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}

func TestObjectHandler_List(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectService{}

	handler := NewObjectHandlerWithInterface(mockService, logger)

	objects := []*models.Object{
		{ID: 1, Name: "Object1"},
		{ID: 2, Name: "Object2"},
	}

	mockService.On("List", mock.Anything, mock.Anything).Return(objects, int64(2), nil)

	c, w := createTestGinContext("GET", "/api/v1/objects", nil)

	handler.List(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestObjectHandler_Search(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectService{}

	handler := NewObjectHandlerWithInterface(mockService, logger)

	objects := []*models.Object{
		{ID: 1, Name: "SearchResult"},
	}

	mockService.On("Search", mock.Anything, "query", 50).Return(objects, nil)

	c, w := createTestGinContext("GET", "/api/v1/objects/search?q=query", nil)

	handler.Search(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestObjectHandler_AddTags(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectService{}

	handler := NewObjectHandlerWithInterface(mockService, logger)

	mockService.On("AddTags", mock.Anything, int64(1), []string{"tag1", "tag2"}, mock.Anything).Return(nil)

	c, w := createTestGinContext("POST", "/api/v1/objects/1/tags", map[string]interface{}{"tags": []string{"tag1", "tag2"}})
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.AddTags(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestObjectHandler_RemoveTags(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectService{}

	handler := NewObjectHandlerWithInterface(mockService, logger)

	mockService.On("RemoveTags", mock.Anything, int64(1), []string{"tag1"}, mock.Anything).Return(nil)

	c, w := createTestGinContext("DELETE", "/api/v1/objects/1/tags", map[string]interface{}{"tags": []string{"tag1"}})
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.RemoveTags(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestObjectHandler_GetChildren(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectService{}

	handler := NewObjectHandlerWithInterface(mockService, logger)

	objects := []*models.Object{
		{ID: 2, Name: "Child1"},
	}

	mockService.On("GetChildren", mock.Anything, int64(1)).Return(objects, nil)

	c, w := createTestGinContext("GET", "/api/v1/objects/1/children", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.GetChildren(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestObjectHandler_GetStats(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectService{}

	handler := NewObjectHandlerWithInterface(mockService, logger)

	stats := &repository.ObjectStats{
		Total:    100,
		ByStatus: map[string]int64{"active": 80, "archived": 20},
		ByType:   map[int64]int64{1: 100},
		ByTags:   map[string]int64{},
		Recent:   50,
	}

	mockService.On("GetObjectStats", mock.Anything, mock.Anything).Return(stats, nil)

	c, w := createTestGinContext("GET", "/api/v1/objects/stats", nil)

	handler.GetStats(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestObjectHandler_BulkCreate(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectService{}

	handler := NewObjectHandlerWithInterface(mockService, logger)

	req := []models.CreateObjectRequest{
		{ObjectTypeID: 1, Name: "Object1"},
		{ObjectTypeID: 1, Name: "Object2"},
	}

	objects := []*models.Object{
		{ID: 1, Name: "Object1"},
		{ID: 2, Name: "Object2"},
	}

	mockService.On("BulkCreate", mock.Anything, mock.Anything).Return(objects, nil)

	c, w := createTestGinContext("POST", "/api/v1/objects/bulk", req)

	handler.BulkCreate(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	mockService.AssertExpectations(t)
}

func TestObjectHandler_BulkDelete(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectService{}

	handler := NewObjectHandlerWithInterface(mockService, logger)

	mockService.On("BulkDelete", mock.Anything, []int64{1, 2}).Return(nil)

	c, w := createTestGinContext("DELETE", "/api/v1/objects/bulk", map[string]interface{}{"ids": []int64{1, 2}})

	handler.BulkDelete(c)

	assert.Equal(t, http.StatusNoContent, w.Code)

	mockService.AssertExpectations(t)
}
