package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
)

type MockObjectTypeService struct {
	mock.Mock
}

func (m *MockObjectTypeService) Create(ctx context.Context, req *models.CreateObjectTypeRequest) (*models.ObjectType, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ObjectType), args.Error(1)
}

func (m *MockObjectTypeService) GetByID(ctx context.Context, id int64) (*models.ObjectType, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ObjectType), args.Error(1)
}

func (m *MockObjectTypeService) GetByName(ctx context.Context, name string) (*models.ObjectType, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ObjectType), args.Error(1)
}

func (m *MockObjectTypeService) Update(ctx context.Context, id int64, req *models.UpdateObjectTypeRequest) (*models.ObjectType, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ObjectType), args.Error(1)
}

func (m *MockObjectTypeService) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockObjectTypeService) GetTree(ctx context.Context, rootID *int64) ([]*models.ObjectType, error) {
	args := m.Called(ctx, rootID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ObjectType), args.Error(1)
}

func (m *MockObjectTypeService) GetChildren(ctx context.Context, parentID int64) ([]*models.ObjectType, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ObjectType), args.Error(1)
}

func (m *MockObjectTypeService) GetDescendants(ctx context.Context, typeID int64, maxDepth *int) ([]*models.ObjectType, error) {
	args := m.Called(ctx, typeID, maxDepth)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ObjectType), args.Error(1)
}

func (m *MockObjectTypeService) GetAncestors(ctx context.Context, typeID int64) ([]*models.ObjectType, error) {
	args := m.Called(ctx, typeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ObjectType), args.Error(1)
}

func (m *MockObjectTypeService) GetPath(ctx context.Context, typeID int64) ([]*models.ObjectType, error) {
	args := m.Called(ctx, typeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ObjectType), args.Error(1)
}

func (m *MockObjectTypeService) List(ctx context.Context, filter *models.ObjectTypeFilter) ([]*models.ObjectType, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ObjectType), args.Error(1)
}

func (m *MockObjectTypeService) Search(ctx context.Context, query string, limit int) ([]*models.ObjectType, error) {
	args := m.Called(ctx, query, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ObjectType), args.Error(1)
}

func (m *MockObjectTypeService) ValidateMove(ctx context.Context, typeID int64, newParentID *int64) error {
	args := m.Called(ctx, typeID, newParentID)
	return args.Error(0)
}

func (m *MockObjectTypeService) GetSubtreeObjectCount(ctx context.Context, typeID int64) (int64, error) {
	args := m.Called(ctx, typeID)
	return args.Get(0).(int64), args.Error(1)
}

func createTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	return logger
}

func createTestGinContext(method, path string, body interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var jsonBody []byte
	if body != nil {
		jsonBody, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "test-request-id")
	c.Request = req

	return c, w
}

func TestObjectTypeHandler_Create(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectTypeService{}

	handler := NewObjectTypeHandlerWithInterface(mockService, logger)

	req := models.CreateObjectTypeRequest{
		Name:        "TestType",
		Description: "Test description",
	}

	createdType := &models.ObjectType{
		ID:          1,
		Name:        "TestType",
		Description: "Test description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockService.On("Create", mock.Anything, mock.AnythingOfType("*models.CreateObjectTypeRequest")).Return(createdType, nil)

	c, w := createTestGinContext("POST", "/api/v1/object-types", req)

	handler.Create(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "TestType", response["data"].(map[string]interface{})["name"])

	mockService.AssertExpectations(t)
}

func TestObjectTypeHandler_Create_InvalidBody(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectTypeService{}

	handler := NewObjectTypeHandlerWithInterface(mockService, logger)

	c, w := createTestGinContext("POST", "/api/v1/object-types", "invalid json")

	handler.Create(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	mockService.AssertExpectations(t)
}

func TestObjectTypeHandler_GetByID(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectTypeService{}

	handler := NewObjectTypeHandlerWithInterface(mockService, logger)

	objectType := &models.ObjectType{
		ID:          1,
		Name:        "TestType",
		Description: "Test description",
	}

	mockService.On("GetByID", mock.Anything, int64(1)).Return(objectType, nil)

	c, w := createTestGinContext("GET", "/api/v1/object-types/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.GetByID(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), response["data"].(map[string]interface{})["id"])

	mockService.AssertExpectations(t)
}

func TestObjectTypeHandler_GetByID_InvalidID(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectTypeService{}

	handler := NewObjectTypeHandlerWithInterface(mockService, logger)

	c, w := createTestGinContext("GET", "/api/v1/object-types/invalid", nil)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}

	handler.GetByID(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	mockService.AssertExpectations(t)
}

func TestObjectTypeHandler_GetByID_NotFound(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectTypeService{}

	handler := NewObjectTypeHandlerWithInterface(mockService, logger)

	mockService.On("GetByID", mock.Anything, int64(999)).Return(nil, assert.AnError)

	c, w := createTestGinContext("GET", "/api/v1/object-types/999", nil)
	c.Params = gin.Params{{Key: "id", Value: "999"}}

	handler.GetByID(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}

func TestObjectTypeHandler_GetByName(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectTypeService{}

	handler := NewObjectTypeHandlerWithInterface(mockService, logger)

	objectType := &models.ObjectType{
		ID:   1,
		Name: "TestType",
	}

	mockService.On("GetByName", mock.Anything, "TestType").Return(objectType, nil)

	c, w := createTestGinContext("GET", "/api/v1/object-types/name/TestType", nil)
	c.Params = gin.Params{{Key: "name", Value: "TestType"}}

	handler.GetByName(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestObjectTypeHandler_Update(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectTypeService{}

	handler := NewObjectTypeHandlerWithInterface(mockService, logger)

	req := models.UpdateObjectTypeRequest{
		Description: stringPtr("Updated description"),
	}

	updatedType := &models.ObjectType{
		ID:          1,
		Name:        "TestType",
		Description: "Updated description",
	}

	mockService.On("Update", mock.Anything, int64(1), mock.AnythingOfType("*models.UpdateObjectTypeRequest")).Return(updatedType, nil)

	c, w := createTestGinContext("PUT", "/api/v1/object-types/1", req)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.Update(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestObjectTypeHandler_Delete(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectTypeService{}

	handler := NewObjectTypeHandlerWithInterface(mockService, logger)

	mockService.On("Delete", mock.Anything, int64(1)).Return(nil)

	c, w := createTestGinContext("DELETE", "/api/v1/object-types/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.Delete(c)

	assert.Equal(t, http.StatusNoContent, w.Code)

	mockService.AssertExpectations(t)
}

func TestObjectTypeHandler_List(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectTypeService{}

	handler := NewObjectTypeHandlerWithInterface(mockService, logger)

	types := []*models.ObjectType{
		{ID: 1, Name: "Type1"},
		{ID: 2, Name: "Type2"},
	}

	mockService.On("List", mock.Anything, mock.Anything).Return(types, nil)

	c, w := createTestGinContext("GET", "/api/v1/object-types", nil)

	handler.List(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestObjectTypeHandler_Search(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectTypeService{}

	handler := NewObjectTypeHandlerWithInterface(mockService, logger)

	types := []*models.ObjectType{
		{ID: 1, Name: "SearchResult"},
	}

	mockService.On("Search", mock.Anything, "query", 50).Return(types, nil)

	c, w := createTestGinContext("GET", "/api/v1/object-types/search?q=query", nil)

	handler.Search(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestObjectTypeHandler_GetTree(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectTypeService{}

	handler := NewObjectTypeHandlerWithInterface(mockService, logger)

	types := []*models.ObjectType{
		{ID: 1, Name: "Root"},
		{ID: 2, Name: "Child"},
	}

	mockService.On("GetTree", mock.Anything, (*int64)(nil)).Return(types, nil)

	c, w := createTestGinContext("GET", "/api/v1/object-types/1/tree", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.GetTree(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestObjectTypeHandler_GetChildren(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectTypeService{}

	handler := NewObjectTypeHandlerWithInterface(mockService, logger)

	types := []*models.ObjectType{
		{ID: 2, Name: "Child"},
	}

	mockService.On("GetChildren", mock.Anything, int64(1)).Return(types, nil)

	c, w := createTestGinContext("GET", "/api/v1/object-types/1/children", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.GetChildren(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestObjectTypeHandler_ValidateMove(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectTypeService{}

	handler := NewObjectTypeHandlerWithInterface(mockService, logger)

	newParentID := int64(2)
	mockService.On("ValidateMove", mock.Anything, int64(1), &newParentID).Return(nil)

	c, w := createTestGinContext("GET", "/api/v1/object-types/1/validate-move?new_parent_id=2", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.ValidateMove(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestObjectTypeHandler_GetSubtreeObjectCount(t *testing.T) {
	logger := createTestLogger()
	mockService := &MockObjectTypeService{}

	handler := NewObjectTypeHandlerWithInterface(mockService, logger)

	mockService.On("GetSubtreeObjectCount", mock.Anything, int64(1)).Return(int64(100), nil)

	c, w := createTestGinContext("GET", "/api/v1/object-types/1/subtree-count", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.GetSubtreeObjectCount(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func stringPtr(s string) *string {
	return &s
}
