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
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/permiddleware"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/repository"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type MockAuthClientForRBAC struct {
	mock.Mock
}

func (m *MockAuthClientForRBAC) CheckPermission(ctx context.Context, userID, permission, jwtToken string) (bool, error) {
	args := m.Called(ctx, userID, permission, jwtToken)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthClientForRBAC) GetUserPermissions(ctx context.Context, userID, jwtToken string) ([]string, error) {
	args := m.Called(ctx, userID, jwtToken)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockAuthClientForRBAC) GetUserRoles(ctx context.Context, userID, jwtToken string) ([]string, error) {
	args := m.Called(ctx, userID, jwtToken)
	return args.Get(0).([]string), args.Error(1)
}

func TestPermissionMiddleware_Allowed(t *testing.T) {
	mockAuthClient := new(MockAuthClientForRBAC)
	mockAuthClient.On("CheckPermission", mock.Anything, "user-123", "objects:create", "").Return(true, nil)

	router := gin.New()
	permissionMiddleware := permiddleware.NewPermissionMiddleware(permiddleware.PermissionMiddlewareConfig{
		AuthClient: mockAuthClient,
		Logger:     nil,
	})

	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Next()
	})
	router.POST("/objects", permissionMiddleware("objects:create"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("POST", "/objects", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthClient.AssertExpectations(t)
}

func TestPermissionMiddleware_Denied(t *testing.T) {
	mockAuthClient := new(MockAuthClientForRBAC)
	mockAuthClient.On("CheckPermission", mock.Anything, "user-123", "objects:create", "").Return(false, nil)

	router := gin.New()
	permissionMiddleware := permiddleware.NewPermissionMiddleware(permiddleware.PermissionMiddlewareConfig{
		AuthClient: mockAuthClient,
		Logger:     nil,
	})

	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Next()
	})
	router.POST("/objects", permissionMiddleware("objects:create"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("POST", "/objects", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Insufficient permissions")
	mockAuthClient.AssertExpectations(t)
}

func TestPermissionMiddleware_Unauthorized(t *testing.T) {
	mockAuthClient := new(MockAuthClientForRBAC)

	router := gin.New()
	permissionMiddleware := permiddleware.NewPermissionMiddleware(permiddleware.PermissionMiddlewareConfig{
		AuthClient: mockAuthClient,
		Logger:     nil,
	})

	router.POST("/objects", permissionMiddleware("objects:create"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("POST", "/objects", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authentication required")
}

func TestPermissionMiddleware_AuthServiceDown(t *testing.T) {
	mockAuthClient := new(MockAuthClientForRBAC)
	mockAuthClient.On("CheckPermission", mock.Anything, "user-123", "objects:create", "").Return(false, assert.AnError)

	router := gin.New()
	permissionMiddleware := permiddleware.NewPermissionMiddleware(permiddleware.PermissionMiddlewareConfig{
		AuthClient: mockAuthClient,
		Logger:     nil,
	})

	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Next()
	})
	router.POST("/objects", permissionMiddleware("objects:create"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("POST", "/objects", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Permission check failed")
	mockAuthClient.AssertExpectations(t)
}

func createTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	return logger
}

func TestOwnershipCheck_OwnerCanUpdate(t *testing.T) {
	mockService := new(MockObjectServiceForOwnership)

	existingObj := &models.Object{
		ID:        1,
		Name:      "TestObject",
		CreatedBy: "user-123",
	}

	updatedObj := &models.Object{
		ID:        1,
		Name:      "UpdatedObject",
		CreatedBy: "user-123",
	}

	mockService.On("GetByID", mock.Anything, int64(1)).Return(existingObj, nil)
	mockService.On("Update", mock.Anything, int64(1), mock.AnythingOfType("*models.UpdateObjectRequest")).Return(updatedObj, nil)

	logger := createTestLogger()
	h := handlers.NewObjectHandlerWithInterface(mockService, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("matched_permissions", []string{"objects:update:own"})
		c.Next()
	})
	router.PUT("/objects/:id", h.Update)

	req, _ := http.NewRequest("PUT", "/objects/1", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestOwnershipCheck_NonOwnerDenied(t *testing.T) {
	mockService := new(MockObjectServiceForOwnership)

	existingObj := &models.Object{
		ID:        1,
		Name:      "TestObject",
		CreatedBy: "other-user",
	}

	mockService.On("GetByID", mock.Anything, int64(1)).Return(existingObj, nil)

	logger := createTestLogger()
	h := handlers.NewObjectHandlerWithInterface(mockService, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("matched_permissions", []string{"objects:update:own"})
		c.Next()
	})
	router.PUT("/objects/:id", h.Update)

	req, _ := http.NewRequest("PUT", "/objects/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "You can only update your own objects")
	mockService.AssertExpectations(t)
}

func TestOwnershipCheck_AdminBypass(t *testing.T) {
	mockService := new(MockObjectServiceForOwnership)

	existingObj := &models.Object{
		ID:        1,
		Name:      "TestObject",
		CreatedBy: "other-user",
	}

	updatedObj := &models.Object{
		ID:        1,
		Name:      "UpdatedObject",
		CreatedBy: "other-user",
	}

	mockService.On("GetByID", mock.Anything, int64(1)).Return(existingObj, nil)
	mockService.On("Update", mock.Anything, int64(1), mock.AnythingOfType("*models.UpdateObjectRequest")).Return(updatedObj, nil)

	logger := createTestLogger()
	h := handlers.NewObjectHandlerWithInterface(mockService, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "admin-user")
		c.Set("user_roles", []string{"admin"})
		c.Set("matched_permissions", []string{"objects:update:all"})
		c.Next()
	})
	router.PUT("/objects/:id", h.Update)

	req, _ := http.NewRequest("PUT", "/objects/1", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestOwnershipCheck_DeleteOwner(t *testing.T) {
	mockService := new(MockObjectServiceForOwnership)

	existingObj := &models.Object{
		ID:        1,
		Name:      "TestObject",
		CreatedBy: "user-123",
	}

	mockService.On("GetByID", mock.Anything, int64(1)).Return(existingObj, nil)
	mockService.On("Delete", mock.Anything, int64(1)).Return(nil)

	logger := createTestLogger()
	h := handlers.NewObjectHandlerWithInterface(mockService, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("matched_permissions", []string{"objects:delete:own"})
		c.Next()
	})
	router.DELETE("/objects/:id", h.Delete)

	req, _ := http.NewRequest("DELETE", "/objects/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}

func TestOwnershipCheck_DeleteNonOwnerDenied(t *testing.T) {
	mockService := new(MockObjectServiceForOwnership)

	existingObj := &models.Object{
		ID:        1,
		Name:      "TestObject",
		CreatedBy: "other-user",
	}

	mockService.On("GetByID", mock.Anything, int64(1)).Return(existingObj, nil)

	logger := createTestLogger()
	h := handlers.NewObjectHandlerWithInterface(mockService, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("matched_permissions", []string{"objects:delete:own"})
		c.Next()
	})
	router.DELETE("/objects/:id", h.Delete)

	req, _ := http.NewRequest("DELETE", "/objects/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "You can only delete your own objects")
	mockService.AssertExpectations(t)
}

func TestOwnershipCheck_DeleteAdminBypass(t *testing.T) {
	mockService := new(MockObjectServiceForOwnership)

	existingObj := &models.Object{
		ID:        1,
		Name:      "TestObject",
		CreatedBy: "other-user",
	}

	mockService.On("GetByID", mock.Anything, int64(1)).Return(existingObj, nil)
	mockService.On("Delete", mock.Anything, int64(1)).Return(nil)

	logger := createTestLogger()
	h := handlers.NewObjectHandlerWithInterface(mockService, logger)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "admin-user")
		c.Set("user_roles", []string{"admin"})
		c.Set("matched_permissions", []string{"objects:delete:all"})
		c.Next()
	})
	router.DELETE("/objects/:id", h.Delete)

	req, _ := http.NewRequest("DELETE", "/objects/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}

func TestMatchedPermissions_AllAndOwn(t *testing.T) {
	mockAuthClient := new(MockAuthClientForRBAC)
	mockAuthClient.On("CheckPermission", mock.Anything, "user-123", "objects:read:all", "").Return(false, nil)
	mockAuthClient.On("CheckPermission", mock.Anything, "user-123", "objects:read:own", "").Return(true, nil)

	router := gin.New()
	permissionMiddleware := permiddleware.NewPermissionMiddleware(permiddleware.PermissionMiddlewareConfig{
		AuthClient: mockAuthClient,
		Logger:     nil,
	})

	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Next()
	})
	router.GET("/objects", permissionMiddleware("objects:read:all", "objects:read:own"), func(c *gin.Context) {
		perms, _ := c.Get("matched_permissions")
		c.JSON(http.StatusOK, gin.H{"permissions": perms})
	})

	req, _ := http.NewRequest("GET", "/objects", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	permsIfc := resp["permissions"]
	perms, ok := permsIfc.([]interface{})
	assert.True(t, ok, "permissions should be []interface{}")
	assert.Contains(t, perms, "objects:read:own")
	assert.NotContains(t, perms, "objects:read:all")

	mockAuthClient.AssertExpectations(t)
}

type MockObjectServiceForOwnership struct {
	mock.Mock
}

func (m *MockObjectServiceForOwnership) Create(ctx context.Context, req *models.CreateObjectRequest) (*models.Object, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Object), args.Error(1)
}

func (m *MockObjectServiceForOwnership) GetByID(ctx context.Context, id int64) (*models.Object, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Object), args.Error(1)
}

func (m *MockObjectServiceForOwnership) GetByPublicID(ctx context.Context, publicID uuid.UUID) (*models.Object, error) {
	args := m.Called(ctx, publicID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Object), args.Error(1)
}

func (m *MockObjectServiceForOwnership) GetByName(ctx context.Context, name string) (*models.Object, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Object), args.Error(1)
}

func (m *MockObjectServiceForOwnership) Update(ctx context.Context, id int64, req *models.UpdateObjectRequest) (*models.Object, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Object), args.Error(1)
}

func (m *MockObjectServiceForOwnership) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockObjectServiceForOwnership) List(ctx context.Context, filter *models.ObjectFilter) ([]*models.Object, int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*models.Object), args.Get(1).(int64), args.Error(2)
}

func (m *MockObjectServiceForOwnership) Search(ctx context.Context, query string, limit int) ([]*models.Object, error) {
	args := m.Called(ctx, query, limit)
	return args.Get(0).([]*models.Object), args.Error(1)
}

func (m *MockObjectServiceForOwnership) FindByMetadata(ctx context.Context, key, value string) ([]*models.Object, error) {
	args := m.Called(ctx, key, value)
	return args.Get(0).([]*models.Object), args.Error(1)
}

func (m *MockObjectServiceForOwnership) FindByTags(ctx context.Context, tags []string, matchAll bool) ([]*models.Object, error) {
	args := m.Called(ctx, tags, matchAll)
	return args.Get(0).([]*models.Object), args.Error(1)
}

func (m *MockObjectServiceForOwnership) UpdateMetadata(ctx context.Context, id int64, metadata map[string]interface{}, updatedBy string) error {
	args := m.Called(ctx, id, metadata, updatedBy)
	return args.Error(0)
}

func (m *MockObjectServiceForOwnership) AddTags(ctx context.Context, id int64, tags []string, updatedBy string) error {
	args := m.Called(ctx, id, tags, updatedBy)
	return args.Error(0)
}

func (m *MockObjectServiceForOwnership) RemoveTags(ctx context.Context, id int64, tags []string, updatedBy string) error {
	args := m.Called(ctx, id, tags, updatedBy)
	return args.Error(0)
}

func (m *MockObjectServiceForOwnership) GetChildren(ctx context.Context, parentID int64) ([]*models.Object, error) {
	args := m.Called(ctx, parentID)
	return args.Get(0).([]*models.Object), args.Error(1)
}

func (m *MockObjectServiceForOwnership) GetDescendants(ctx context.Context, rootID int64, maxDepth *int) ([]*models.Object, error) {
	args := m.Called(ctx, rootID, maxDepth)
	return args.Get(0).([]*models.Object), args.Error(1)
}

func (m *MockObjectServiceForOwnership) GetAncestors(ctx context.Context, id int64) ([]*models.Object, error) {
	args := m.Called(ctx, id)
	return args.Get(0).([]*models.Object), args.Error(1)
}

func (m *MockObjectServiceForOwnership) GetPath(ctx context.Context, id int64) ([]*models.Object, error) {
	args := m.Called(ctx, id)
	return args.Get(0).([]*models.Object), args.Error(1)
}

func (m *MockObjectServiceForOwnership) BulkCreate(ctx context.Context, objects []*models.CreateObjectRequest) ([]*models.Object, error) {
	args := m.Called(ctx, objects)
	return args.Get(0).([]*models.Object), args.Error(1)
}

func (m *MockObjectServiceForOwnership) BulkUpdate(ctx context.Context, ids []int64, updates *models.UpdateObjectRequest) ([]*models.Object, error) {
	args := m.Called(ctx, ids, updates)
	return args.Get(0).([]*models.Object), args.Error(1)
}

func (m *MockObjectServiceForOwnership) BulkDelete(ctx context.Context, ids []int64) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *MockObjectServiceForOwnership) ValidateParentChild(ctx context.Context, parentID, childID int64) error {
	args := m.Called(ctx, parentID, childID)
	return args.Error(0)
}

func (m *MockObjectServiceForOwnership) GetObjectStats(ctx context.Context, filter *models.ObjectFilter) (*repository.ObjectStats, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*repository.ObjectStats), args.Error(1)
}
