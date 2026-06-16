package permiddleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthClient struct {
	mock.Mock
}

func (m *MockAuthClient) CheckPermission(ctx context.Context, userID, permission, jwtToken string) (bool, error) {
	args := m.Called(ctx, userID, permission, jwtToken)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthClient) GetUserPermissions(ctx context.Context, userID, jwtToken string) ([]string, error) {
	args := m.Called(ctx, userID, jwtToken)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockAuthClient) GetUserRoles(ctx context.Context, userID, jwtToken string) ([]string, error) {
	args := m.Called(ctx, userID, jwtToken)
	return args.Get(0).([]string), args.Error(1)
}

func TestRequirePermission_Allowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockClient := new(MockAuthClient)
	mockClient.On("CheckPermission", mock.Anything, "user-123", "objects:create", "").Return(true, nil)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Next()
	})
	router.Use(NewPermissionMiddleware(RouteConfig{TypeKey: "objects", HTTPMethod: "POST"}, mockClient, nil))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockClient.AssertExpectations(t)
}

func TestRequirePermission_Denied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockClient := new(MockAuthClient)
	mockClient.On("CheckPermission", mock.Anything, "user-123", "objects:delete:all", "").Return(false, nil)
	mockClient.On("CheckPermission", mock.Anything, "user-123", "objects:delete:own", "").Return(false, nil)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Next()
	})
	router.Use(NewPermissionMiddleware(RouteConfig{TypeKey: "objects", HTTPMethod: "DELETE"}, mockClient, nil))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Insufficient permissions")
	mockClient.AssertExpectations(t)
}

func TestRequirePermission_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockClient := new(MockAuthClient)

	router := gin.New()
	router.Use(NewPermissionMiddleware(RouteConfig{TypeKey: "objects", HTTPMethod: "POST"}, mockClient, nil))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authentication required")
}

func TestRequirePermission_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockClient := new(MockAuthClient)
	mockClient.On("CheckPermission", mock.Anything, "user-123", "objects:create", "").Return(false, assert.AnError)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Next()
	})
	router.Use(NewPermissionMiddleware(RouteConfig{TypeKey: "objects", HTTPMethod: "POST"}, mockClient, nil))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Permission check failed")
	mockClient.AssertExpectations(t)
}

func TestRequirePermission_AnyOfMultiplePermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockClient := new(MockAuthClient)
	mockClient.On("CheckPermission", mock.Anything, "user-123", "objects:read:all", "").Return(false, nil)
	mockClient.On("CheckPermission", mock.Anything, "user-123", "objects:read:own", "").Return(true, nil)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Next()
	})
	router.Use(NewPermissionMiddleware(RouteConfig{TypeKey: "objects", HTTPMethod: "GET"}, mockClient, nil))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockClient.AssertExpectations(t)
}

func TestBuildPermissionStrings_POST(t *testing.T) {
	cfg := RouteConfig{TypeKey: "relationships", HTTPMethod: "POST"}
	perms := buildPermissionStrings(cfg)
	assert.Equal(t, []string{"relationships:create"}, perms)
}

func TestBuildPermissionStrings_GET(t *testing.T) {
	cfg := RouteConfig{TypeKey: "objects", HTTPMethod: "GET"}
	perms := buildPermissionStrings(cfg)
	assert.Equal(t, []string{"objects:read:all", "objects:read:own"}, perms)
}

func TestBuildPermissionStrings_PUT(t *testing.T) {
	cfg := RouteConfig{TypeKey: "documents", HTTPMethod: "PUT"}
	perms := buildPermissionStrings(cfg)
	assert.Equal(t, []string{"documents:update:all", "documents:update:own"}, perms)
}

func TestBuildPermissionStrings_DELETE(t *testing.T) {
	cfg := RouteConfig{TypeKey: "files", HTTPMethod: "DELETE"}
	perms := buildPermissionStrings(cfg)
	assert.Equal(t, []string{"files:delete:all", "files:delete:own"}, perms)
}
