package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/handlers"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/repository"
)

func init() {
	gin.SetMode(gin.TestMode)
}

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

func createTestRouter() *gin.Engine {
	router := gin.New()
	return router
}

func TestObjectTypeAPI_CreateAndGet(t *testing.T) {
	router := createTestRouter()
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	mockTypeService := &MockObjectTypeService{}
	typeHandler := handlers.NewObjectTypeHandlerWithInterface(mockTypeService, logger)

	mockTypeService.On("Create", mock.Anything, mock.AnythingOfType("*models.CreateObjectTypeRequest")).Run(func(args mock.Arguments) {
		req := args.Get(1).(*models.CreateObjectTypeRequest)
		assert.Equal(t, "TestCategory", req.Name)
	}).Return(&models.ObjectType{
		ID:   1,
		Name: "TestCategory",
	}, nil)

	mockTypeService.On("GetByID", mock.Anything, int64(1)).Return(&models.ObjectType{
		ID:   1,
		Name: "TestCategory",
	}, nil)

	router.POST("/api/v1/object-types", typeHandler.Create)
	router.GET("/api/v1/object-types/:id", typeHandler.GetByID)

	createReq := models.CreateObjectTypeRequest{
		Name: "TestCategory",
	}
	body, _ := json.Marshal(createReq)

	req, _ := http.NewRequest("POST", "/api/v1/object-types", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var createResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResp)
	assert.NotNil(t, createResp["data"])

	req2, _ := http.NewRequest("GET", "/api/v1/object-types/1", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	mockTypeService.AssertExpectations(t)
}

func TestObjectAPI_List(t *testing.T) {
	router := createTestRouter()
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	mockService := &MockObjectService{}
	handler := handlers.NewObjectHandlerWithInterface(mockService, logger)

	mockService.On("List", mock.Anything, mock.Anything).Return([]*models.Object{
		{ID: 1, Name: "Object1"},
		{ID: 2, Name: "Object2"},
	}, int64(2), nil)

	router.GET("/api/v1/objects", handler.List)

	req, _ := http.NewRequest("GET", "/api/v1/objects", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var listResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &listResp)
	assert.NotNil(t, listResp["data"])

	mockService.AssertExpectations(t)
}
