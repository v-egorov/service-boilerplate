package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/models"
	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/utils"
)

// MockAuthService is a mock implementation of AuthService for testing
type MockAuthService struct {
	getPublicKeyPEMFunc func() ([]byte, error)
	loginFunc           func(ctx context.Context, req *models.LoginRequest, ipAddress, userAgent string) (*models.TokenResponse, error)
	registerFunc        func(ctx context.Context, req *models.RegisterRequest) (*models.UserInfo, error)
	getCurrentUserFunc  func(ctx context.Context, userID uuid.UUID, email string) (*models.UserInfo, error)
	logoutFunc          func(ctx context.Context, tokenString string) error
	refreshTokenFunc    func(ctx context.Context, req *models.RefreshTokenRequest) (*models.TokenResponse, error)
	validateTokenFunc   func(ctx context.Context, tokenString string) (*utils.JWTClaims, error)
	rotateKeysFunc      func(ctx context.Context) error
	createRoleFunc      func(ctx context.Context, name, description string) (*models.Role, error)
	listRolesFunc       func(ctx context.Context) ([]models.Role, error)
	getRoleFunc         func(ctx context.Context, roleID uuid.UUID) (*models.Role, error)
	updateRoleFunc      func(ctx context.Context, roleID uuid.UUID, name, description string) (*models.Role, error)
	deleteRoleFunc      func(ctx context.Context, roleID uuid.UUID) error
	createPermissionFunc func(ctx context.Context, name, resource, action string) (*models.Permission, error)
	listPermissionsFunc  func(ctx context.Context) ([]models.Permission, error)
	getPermissionFunc    func(ctx context.Context, permissionID uuid.UUID) (*models.Permission, error)
	updatePermissionFunc func(ctx context.Context, permissionID uuid.UUID, name, resource, action string) (*models.Permission, error)
	deletePermissionFunc func(ctx context.Context, permissionID uuid.UUID) error
	assignPermissionToRoleFunc func(ctx context.Context, roleID, permissionID uuid.UUID) error
	removePermissionFromRoleFunc func(ctx context.Context, roleID, permissionID uuid.UUID) error
	getRolePermissionsFunc func(ctx context.Context, roleID uuid.UUID) ([]models.Permission, error)
	assignRoleToUserFunc func(ctx context.Context, userID, roleID uuid.UUID) error
	removeRoleFromUserFunc func(ctx context.Context, userID, roleID uuid.UUID) error
	getUserRolesFunc     func(ctx context.Context, userID uuid.UUID) ([]models.Role, error)
	updateUserRolesFunc  func(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error
}

func (m *MockAuthService) Login(ctx context.Context, req *models.LoginRequest, ipAddress, userAgent string) (*models.TokenResponse, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, req, ipAddress, userAgent)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.UserInfo, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, req)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAuthService) Logout(ctx context.Context, tokenString string) error {
	if m.logoutFunc != nil {
		return m.logoutFunc(ctx, tokenString)
	}
	return errors.New("not implemented")
}

func (m *MockAuthService) RefreshToken(ctx context.Context, req *models.RefreshTokenRequest) (*models.TokenResponse, error) {
	if m.refreshTokenFunc != nil {
		return m.refreshTokenFunc(ctx, req)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAuthService) GetCurrentUser(ctx context.Context, userID uuid.UUID, email string) (*models.UserInfo, error) {
	if m.getCurrentUserFunc != nil {
		return m.getCurrentUserFunc(ctx, userID, email)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAuthService) ValidateToken(ctx context.Context, tokenString string) (*utils.JWTClaims, error) {
	if m.validateTokenFunc != nil {
		return m.validateTokenFunc(ctx, tokenString)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAuthService) GetPublicKeyPEM() ([]byte, error) {
	if m.getPublicKeyPEMFunc != nil {
		return m.getPublicKeyPEMFunc()
	}
	return nil, errors.New("not implemented")
}

func (m *MockAuthService) RotateKeys(ctx context.Context) error {
	if m.rotateKeysFunc != nil {
		return m.rotateKeysFunc(ctx)
	}
	return errors.New("not implemented")
}

func (m *MockAuthService) CreateRole(ctx context.Context, name, description string) (*models.Role, error) {
	if m.createRoleFunc != nil {
		return m.createRoleFunc(ctx, name, description)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAuthService) ListRoles(ctx context.Context) ([]models.Role, error) {
	if m.listRolesFunc != nil {
		return m.listRolesFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAuthService) GetRole(ctx context.Context, roleID uuid.UUID) (*models.Role, error) {
	if m.getRoleFunc != nil {
		return m.getRoleFunc(ctx, roleID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAuthService) UpdateRole(ctx context.Context, roleID uuid.UUID, name, description string) (*models.Role, error) {
	if m.updateRoleFunc != nil {
		return m.updateRoleFunc(ctx, roleID, name, description)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAuthService) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	if m.deleteRoleFunc != nil {
		return m.deleteRoleFunc(ctx, roleID)
	}
	return errors.New("not implemented")
}

func (m *MockAuthService) CreatePermission(ctx context.Context, name, resource, action string) (*models.Permission, error) {
	if m.createPermissionFunc != nil {
		return m.createPermissionFunc(ctx, name, resource, action)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAuthService) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	if m.listPermissionsFunc != nil {
		return m.listPermissionsFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAuthService) GetPermission(ctx context.Context, permissionID uuid.UUID) (*models.Permission, error) {
	if m.getPermissionFunc != nil {
		return m.getPermissionFunc(ctx, permissionID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAuthService) UpdatePermission(ctx context.Context, permissionID uuid.UUID, name, resource, action string) (*models.Permission, error) {
	if m.updatePermissionFunc != nil {
		return m.updatePermissionFunc(ctx, permissionID, name, resource, action)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAuthService) DeletePermission(ctx context.Context, permissionID uuid.UUID) error {
	if m.deletePermissionFunc != nil {
		return m.deletePermissionFunc(ctx, permissionID)
	}
	return errors.New("not implemented")
}

func (m *MockAuthService) AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	if m.assignPermissionToRoleFunc != nil {
		return m.assignPermissionToRoleFunc(ctx, roleID, permissionID)
	}
	return errors.New("not implemented")
}

func (m *MockAuthService) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	if m.removePermissionFromRoleFunc != nil {
		return m.removePermissionFromRoleFunc(ctx, roleID, permissionID)
	}
	return errors.New("not implemented")
}

func (m *MockAuthService) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]models.Permission, error) {
	if m.getRolePermissionsFunc != nil {
		return m.getRolePermissionsFunc(ctx, roleID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAuthService) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	if m.assignRoleToUserFunc != nil {
		return m.assignRoleToUserFunc(ctx, userID, roleID)
	}
	return errors.New("not implemented")
}

func (m *MockAuthService) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	if m.removeRoleFromUserFunc != nil {
		return m.removeRoleFromUserFunc(ctx, userID, roleID)
	}
	return errors.New("not implemented")
}

func (m *MockAuthService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]models.Role, error) {
	if m.getUserRolesFunc != nil {
		return m.getUserRolesFunc(ctx, userID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAuthService) UpdateUserRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	if m.updateUserRolesFunc != nil {
		return m.updateUserRolesFunc(ctx, userID, roleIDs)
	}
	return errors.New("not implemented")
}

// Helper function to create a test Gin context
func createTestContext(method, path string, body interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var jsonBody []byte
	if body != nil {
		jsonBody, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	return c, w
}

func TestAuthHandler_GetPublicKey(t *testing.T) {
	tests := []struct {
		name           string
		mockPublicKey  string
		mockError      error
		expectedStatus int
		expectJSON     bool
		expectedBody   string
	}{
		{
			name:           "successful public key retrieval",
			mockPublicKey:  "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...\n-----END PUBLIC KEY-----",
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectJSON:     false,
			expectedBody:   "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...\n-----END PUBLIC KEY-----",
		},
		{
			name:           "service error",
			mockPublicKey:  "",
			mockError:      errors.New("service error"),
			expectedStatus: http.StatusInternalServerError,
			expectJSON:     true,
			expectedBody:   `{"error":"Failed to get public key"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAuthService{
				getPublicKeyPEMFunc: func() ([]byte, error) {
					return []byte(tt.mockPublicKey), tt.mockError
				},
			}

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

			handler := NewAuthHandler(mockService, logger)

			// Create test context
			c, w := createTestContext("GET", "/public-key", nil)

			// Execute
			handler.GetPublicKey(c)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			body := w.Body.String()
			if tt.expectJSON {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal JSON response: %v", err)
				}
				// For JSON responses, check the structure
				if tt.name == "service error" {
					if errorMsg, exists := response["error"]; !exists || errorMsg != "Failed to get public key" {
						t.Errorf("Expected error message not found or incorrect")
					}
				}
			} else {
				// For plain text responses, check exact match
				if body != tt.expectedBody {
					t.Errorf("Expected body %q, got %q", tt.expectedBody, body)
				}
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    models.LoginRequest
		mockResponse   *models.TokenResponse
		mockError      error
		expectedStatus int
		expectJSON     bool
	}{
		{
			name: "successful login",
			requestBody: models.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockResponse: &models.TokenResponse{
				AccessToken:  "access.jwt.token",
				RefreshToken: "refresh.jwt.token",
				TokenType:    "Bearer",
				ExpiresIn:    900,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectJSON:     true,
		},
		{
			name: "invalid request body",
			requestBody: models.LoginRequest{
				Email:    "",
				Password: "",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectJSON:     true,
		},
		{
			name: "service error",
			requestBody: models.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockResponse:   nil,
			mockError:      errors.New("invalid credentials"),
			expectedStatus: http.StatusUnauthorized,
			expectJSON:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAuthService{
				loginFunc: func(ctx context.Context, req *models.LoginRequest, ipAddress, userAgent string) (*models.TokenResponse, error) {
					return tt.mockResponse, tt.mockError
				},
			}

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			handler := NewAuthHandler(mockService, logger)

			// Create test context
			c, w := createTestContext("POST", "/login", tt.requestBody)

			// Execute
			handler.Login(c)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectJSON {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal JSON response: %v", err)
				}

				if tt.mockResponse != nil {
					// Check successful response structure
					if _, exists := response["access_token"]; !exists {
						t.Errorf("Expected access_token in response")
					}
					if _, exists := response["refresh_token"]; !exists {
						t.Errorf("Expected refresh_token in response")
					}
				}
			}
		})
	}
}

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    models.RegisterRequest
		mockResponse   *models.UserInfo
		mockError      error
		expectedStatus int
		expectJSON     bool
	}{
		{
			name: "successful registration",
			requestBody: models.RegisterRequest{
				Email:     "newuser@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			mockResponse: &models.UserInfo{
				ID:        uuid.New(),
				Email:     "newuser@example.com",
				FirstName: "John",
				LastName:  "Doe",
				Roles:     []string{"user"},
			},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
			expectJSON:     true,
		},
		{
			name: "invalid request body",
			requestBody: models.RegisterRequest{
				Email: "",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectJSON:     true,
		},
		{
			name: "service error",
			requestBody: models.RegisterRequest{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			mockResponse:   nil,
			mockError:      errors.New("user already exists"),
			expectedStatus: http.StatusInternalServerError,
			expectJSON:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAuthService{
				registerFunc: func(ctx context.Context, req *models.RegisterRequest) (*models.UserInfo, error) {
					return tt.mockResponse, tt.mockError
				},
			}

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			handler := NewAuthHandler(mockService, logger)

			// Create test context
			c, w := createTestContext("POST", "/register", tt.requestBody)

			// Execute
			handler.Register(c)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectJSON {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal JSON response: %v", err)
				}

				if tt.mockResponse != nil {
					// Check successful response structure
					if userData, exists := response["user"]; !exists {
						t.Errorf("Expected user in response")
					} else if userMap, ok := userData.(map[string]interface{}); ok {
						if _, exists := userMap["id"]; !exists {
							t.Errorf("Expected id in user response")
						}
						if _, exists := userMap["email"]; !exists {
							t.Errorf("Expected email in user response")
						}
					}
				}
			}
		})
	}
}

func TestAuthHandler_GetCurrentUser(t *testing.T) {
	userID := uuid.New()
	tests := []struct {
		name           string
		userID         string
		mockResponse   *models.UserInfo
		mockError      error
		expectedStatus int
	}{
		{
			name:   "successful user retrieval",
			userID: userID.String(),
			mockResponse: &models.UserInfo{
				ID:        userID,
				Email:     "user@example.com",
				FirstName: "John",
				LastName:  "Doe",
				Roles:     []string{"user", "admin"},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid user ID",
			userID:         "invalid-uuid",
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "user not found",
			userID:         userID.String(),
			mockResponse:   nil,
			mockError:      errors.New("user not found"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAuthService{
				getCurrentUserFunc: func(ctx context.Context, uid uuid.UUID, email string) (*models.UserInfo, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockResponse, nil
				},
			}

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			handler := NewAuthHandler(mockService, logger)

			// Create test context
			c, w := createTestContext("GET", "/me", nil)

			// Mock the user ID and email from JWT middleware (simplified)
			c.Set("user_id", tt.userID)
			c.Set("user_email", "user@example.com")

			// Execute
			handler.GetCurrentUser(c)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal JSON response: %v", err)
				}

				if userData, exists := response["user"]; !exists {
					t.Errorf("Expected user in response")
				} else if userMap, ok := userData.(map[string]interface{}); ok {
					if _, exists := userMap["id"]; !exists {
						t.Errorf("Expected id in user response")
					}
					if _, exists := userMap["email"]; !exists {
						t.Errorf("Expected email in user response")
					}
				}
			}
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		mockError      error
		expectedStatus int
	}{
		{
			name:           "successful logout",
			authHeader:     "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid.token.here",
			mockError:      errors.New("invalid token"),
			expectedStatus: http.StatusOK, // Logout doesn't fail even with invalid tokens
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAuthService{
				logoutFunc: func(ctx context.Context, tokenString string) error {
					return tt.mockError
				},
			}

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			handler := NewAuthHandler(mockService, logger)

			// Create test context
			c, w := createTestContext("POST", "/logout", nil)

			// Set authorization header if provided
			if tt.authHeader != "" {
				c.Request.Header.Set("Authorization", tt.authHeader)
			}

			// Execute
			handler.Logout(c)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    models.RefreshTokenRequest
		mockResponse   *models.TokenResponse
		mockError      error
		expectedStatus int
		expectJSON     bool
	}{
		{
			name: "successful token refresh",
			requestBody: models.RefreshTokenRequest{
				RefreshToken: "refresh.jwt.token",
			},
			mockResponse: &models.TokenResponse{
				AccessToken:  "new.access.jwt.token",
				RefreshToken: "new.refresh.jwt.token",
				TokenType:    "Bearer",
				ExpiresIn:    900,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectJSON:     true,
		},
		{
			name:           "invalid request body",
			requestBody:    models.RefreshTokenRequest{},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectJSON:     true,
		},
		{
			name: "invalid refresh token",
			requestBody: models.RefreshTokenRequest{
				RefreshToken: "invalid.refresh.token",
			},
			mockResponse:   nil,
			mockError:      errors.New("invalid refresh token"),
			expectedStatus: http.StatusUnauthorized,
			expectJSON:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAuthService{
				refreshTokenFunc: func(ctx context.Context, req *models.RefreshTokenRequest) (*models.TokenResponse, error) {
					return tt.mockResponse, tt.mockError
				},
			}

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			handler := NewAuthHandler(mockService, logger)

			// Create test context
			c, w := createTestContext("POST", "/refresh", tt.requestBody)

			// Execute
			handler.RefreshToken(c)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectJSON && tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal JSON response: %v", err)
				}

				if _, exists := response["access_token"]; !exists {
					t.Errorf("Expected access_token in response")
				}
				if _, exists := response["refresh_token"]; !exists {
					t.Errorf("Expected refresh_token in response")
				}
			}
		})
	}
}

func TestAuthHandler_ValidateToken(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		mockClaims     *utils.JWTClaims
		mockError      error
		expectedStatus int
	}{
		{
			name:       "valid token",
			authHeader: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			mockClaims: &utils.JWTClaims{
				UserID:    uuid.New(),
				Email:     "user@example.com",
				Roles:     []string{"user"},
				TokenType: "access",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			mockClaims:     nil,
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token format",
			authHeader:     "InvalidFormat",
			mockClaims:     nil,
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "expired token",
			authHeader:     "Bearer expired.jwt.token",
			mockClaims:     nil,
			mockError:      errors.New("token expired"),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAuthService{
				validateTokenFunc: func(ctx context.Context, tokenString string) (*utils.JWTClaims, error) {
					return tt.mockClaims, tt.mockError
				},
			}

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			handler := NewAuthHandler(mockService, logger)

			// Create test context
			c, w := createTestContext("POST", "/validate-token", nil)

			// Set authorization header
			if tt.authHeader != "" {
				c.Request.Header.Set("Authorization", tt.authHeader)
			}

			// Execute
			handler.ValidateToken(c)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal JSON response: %v", err)
				}

				if valid, exists := response["valid"]; !exists || valid != true {
					t.Errorf("Expected valid=true in response")
				}
			}
		})
	}
}

func TestAuthHandler_CreateRole(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]string
		mockResponse   *models.Role
		mockError      error
		expectedStatus int
	}{
		{
			name: "successful role creation",
			requestBody: map[string]string{
				"name":        "admin",
				"description": "Administrator role",
			},
			mockResponse: &models.Role{
				ID:          uuid.New(),
				Name:        "admin",
				Description: stringPtr("Administrator role"),
			},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing name",
			requestBody: map[string]string{
				"description": "Some description",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			requestBody: map[string]string{
				"name":        "admin",
				"description": "Administrator role",
			},
			mockResponse:   nil,
			mockError:      errors.New("duplicate key value violates unique constraint"),
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAuthService{
				createRoleFunc: func(ctx context.Context, name, description string) (*models.Role, error) {
					return tt.mockResponse, tt.mockError
				},
			}

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			handler := NewAuthHandler(mockService, logger)

			// Create test context
			c, w := createTestContext("POST", "/roles", tt.requestBody)

			// Execute
			handler.CreateRole(c)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusCreated {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal JSON response: %v", err)
				}

				if _, exists := response["id"]; !exists {
					t.Errorf("Expected id in response")
				}
				if _, exists := response["name"]; !exists {
					t.Errorf("Expected name in response")
				}
			}
		})
	}
}

func TestAuthHandler_ListRoles(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   []models.Role
		mockError      error
		expectedStatus int
	}{
		{
			name: "successful role listing",
			mockResponse: []models.Role{
				{
					ID:          uuid.New(),
					Name:        "admin",
					Description: stringPtr("Administrator role"),
				},
				{
					ID:          uuid.New(),
					Name:        "user",
					Description: stringPtr("Regular user role"),
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "service error",
			mockResponse:   nil,
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAuthService{
				listRolesFunc: func(ctx context.Context) ([]models.Role, error) {
					return tt.mockResponse, tt.mockError
				},
			}

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			handler := NewAuthHandler(mockService, logger)

			// Create test context
			c, w := createTestContext("GET", "/roles", nil)

			// Execute
			handler.ListRoles(c)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal JSON response: %v", err)
				}

				if rolesData, exists := response["roles"]; exists {
					if rolesArray, ok := rolesData.([]interface{}); ok {
						if len(rolesArray) != len(tt.mockResponse) {
							t.Errorf("Expected %d roles, got %d", len(tt.mockResponse), len(rolesArray))
						}
					}
				}
			}
		})
	}
}

func TestAuthHandler_GetRole(t *testing.T) {
	roleID := uuid.New()
	tests := []struct {
		name           string
		roleID         string
		mockResponse   *models.Role
		mockError      error
		expectedStatus int
	}{
		{
			name:   "successful role retrieval",
			roleID: roleID.String(),
			mockResponse: &models.Role{
				ID:          roleID,
				Name:        "admin",
				Description: stringPtr("Administrator role"),
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid role ID",
			roleID:         "invalid-uuid",
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "role not found",
			roleID:         roleID.String(),
			mockResponse:   nil,
			mockError:      errors.New("role not found"),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAuthService{
				getRoleFunc: func(ctx context.Context, rid uuid.UUID) (*models.Role, error) {
					return tt.mockResponse, tt.mockError
				},
			}

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			handler := NewAuthHandler(mockService, logger)

			// Create test context
			c, w := createTestContext("GET", "/roles/"+tt.roleID, nil)
			c.Params = []gin.Param{{Key: "role_id", Value: tt.roleID}}

			// Execute
			handler.GetRole(c)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal JSON response: %v", err)
				}

				if _, exists := response["id"]; !exists {
					t.Errorf("Expected id in response")
				}
				if _, exists := response["name"]; !exists {
					t.Errorf("Expected name in response")
				}
			}
		})
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

func TestAuthHandler_AssignPermissionToRole(t *testing.T) {
	roleID := uuid.New()
	permissionID := uuid.New()
	tests := []struct {
		name         string
		roleID       string
		permissionID string
		mockError    error
		expectedStatus int
	}{
		{
			name:           "successful assignment",
			roleID:         roleID.String(),
			permissionID:   permissionID.String(),
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid role ID",
			roleID:         "invalid-uuid",
			permissionID:   permissionID.String(),
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid permission ID",
			roleID:         roleID.String(),
			permissionID:   "invalid-uuid",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "service error",
			roleID:         roleID.String(),
			permissionID:   permissionID.String(),
			mockError:      errors.New("assignment failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAuthService{
				assignPermissionToRoleFunc: func(ctx context.Context, rid, pid uuid.UUID) error {
					return tt.mockError
				},
			}

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			handler := NewAuthHandler(mockService, logger)

			// Create test context
			c, w := createTestContext("POST", "/roles/"+tt.roleID+"/permissions", map[string]string{"permission_id": tt.permissionID})
			c.Params = []gin.Param{{Key: "role_id", Value: tt.roleID}}

			// Mock admin role
			c.Set("user_roles", []string{"admin"})

			// Execute
			handler.AssignPermissionToRole(c)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestAuthHandler_RemovePermissionFromRole(t *testing.T) {
	roleID := uuid.New()
	permissionID := uuid.New()
	tests := []struct {
		name         string
		roleID       string
		permissionID string
		mockError    error
		expectedStatus int
	}{
		{
			name:           "successful removal",
			roleID:         roleID.String(),
			permissionID:   permissionID.String(),
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid role ID",
			roleID:         "invalid-uuid",
			permissionID:   permissionID.String(),
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid permission ID",
			roleID:         roleID.String(),
			permissionID:   "invalid-uuid",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "service error",
			roleID:         roleID.String(),
			permissionID:   permissionID.String(),
			mockError:      errors.New("removal failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAuthService{
				removePermissionFromRoleFunc: func(ctx context.Context, rid, pid uuid.UUID) error {
					return tt.mockError
				},
			}

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			handler := NewAuthHandler(mockService, logger)

			// Create test context
			c, w := createTestContext("DELETE", "/roles/"+tt.roleID+"/permissions/"+tt.permissionID, nil)
			c.Params = []gin.Param{{Key: "role_id", Value: tt.roleID}, {Key: "perm_id", Value: tt.permissionID}}

			// Mock admin role
			c.Set("user_roles", []string{"admin"})

			// Execute
			handler.RemovePermissionFromRole(c)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestAuthHandler_GetRolePermissions(t *testing.T) {
	roleID := uuid.New()
	tests := []struct {
		name           string
		roleID         string
		mockResponse   []models.Permission
		mockError      error
		expectedStatus int
	}{
		{
			name:   "successful retrieval",
			roleID: roleID.String(),
			mockResponse: []models.Permission{
				{
					ID:       uuid.New(),
					Name:     "read:users",
					Resource: "users",
					Action:   "read",
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid role ID",
			roleID:         "invalid-uuid",
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "service error",
			roleID:         roleID.String(),
			mockResponse:   nil,
			mockError:      errors.New("retrieval failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAuthService{
				getRolePermissionsFunc: func(ctx context.Context, rid uuid.UUID) ([]models.Permission, error) {
					return tt.mockResponse, tt.mockError
				},
			}

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			handler := NewAuthHandler(mockService, logger)

			// Create test context
			c, w := createTestContext("GET", "/roles/"+tt.roleID+"/permissions", nil)
			c.Params = []gin.Param{{Key: "role_id", Value: tt.roleID}}

			// Mock admin role
			c.Set("user_roles", []string{"admin"})

			// Execute
			handler.GetRolePermissions(c)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestAuthHandler_AssignRoleToUser(t *testing.T) {
	userID := uuid.New()
	roleID := uuid.New()
	tests := []struct {
		name         string
		userID       string
		roleID       string
		mockError    error
		expectedStatus int
	}{
		{
			name:           "successful assignment",
			userID:         userID.String(),
			roleID:         roleID.String(),
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid user ID",
			userID:         "invalid-uuid",
			roleID:         roleID.String(),
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid role ID",
			userID:         userID.String(),
			roleID:         "invalid-uuid",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "service error",
			userID:         userID.String(),
			roleID:         roleID.String(),
			mockError:      errors.New("assignment failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAuthService{
				assignRoleToUserFunc: func(ctx context.Context, uid, rid uuid.UUID) error {
					return tt.mockError
				},
			}

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			handler := NewAuthHandler(mockService, logger)

			// Create test context
			c, w := createTestContext("POST", "/users/"+tt.userID+"/roles", map[string]string{"role_id": tt.roleID})
			c.Params = []gin.Param{{Key: "user_id", Value: tt.userID}}

			// Mock admin role
			c.Set("user_roles", []string{"admin"})

			// Execute
			handler.AssignRoleToUser(c)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestAuthHandler_RemoveRoleFromUser(t *testing.T) {
	userID := uuid.New()
	roleID := uuid.New()
	tests := []struct {
		name         string
		userID       string
		roleID       string
		mockError    error
		expectedStatus int
	}{
		{
			name:           "successful removal",
			userID:         userID.String(),
			roleID:         roleID.String(),
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid user ID",
			userID:         "invalid-uuid",
			roleID:         roleID.String(),
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid role ID",
			userID:         userID.String(),
			roleID:         "invalid-uuid",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "service error",
			userID:         userID.String(),
			roleID:         roleID.String(),
			mockError:      errors.New("removal failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAuthService{
				removeRoleFromUserFunc: func(ctx context.Context, uid, rid uuid.UUID) error {
					return tt.mockError
				},
			}

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			handler := NewAuthHandler(mockService, logger)

			// Create test context
			c, w := createTestContext("DELETE", "/users/"+tt.userID+"/roles/"+tt.roleID, nil)
			c.Params = []gin.Param{{Key: "user_id", Value: tt.userID}, {Key: "role_id", Value: tt.roleID}}

			// Mock admin role
			c.Set("user_roles", []string{"admin"})

			// Execute
			handler.RemoveRoleFromUser(c)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestAuthHandler_GetUserRoles(t *testing.T) {
	userID := uuid.New()
	tests := []struct {
		name           string
		userID         string
		mockResponse   []models.Role
		mockError      error
		expectedStatus int
	}{
		{
			name:   "successful retrieval",
			userID: userID.String(),
			mockResponse: []models.Role{
				{
					ID:          uuid.New(),
					Name:        "admin",
					Description: stringPtr("Administrator role"),
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid user ID",
			userID:         "invalid-uuid",
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "service error",
			userID:         userID.String(),
			mockResponse:   nil,
			mockError:      errors.New("retrieval failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAuthService{
				getUserRolesFunc: func(ctx context.Context, uid uuid.UUID) ([]models.Role, error) {
					return tt.mockResponse, tt.mockError
				},
			}

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			handler := NewAuthHandler(mockService, logger)

			// Create test context
			c, w := createTestContext("GET", "/users/"+tt.userID+"/roles", nil)
			c.Params = []gin.Param{{Key: "user_id", Value: tt.userID}}

			// Mock admin role
			c.Set("user_roles", []string{"admin"})

			// Execute
			handler.GetUserRoles(c)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestAuthHandler_UpdateUserRoles(t *testing.T) {
	userID := uuid.New()
	roleID1 := uuid.New()
	roleID2 := uuid.New()
	tests := []struct {
		name         string
		userID       string
		requestBody  map[string][]string
		mockError    error
		expectedStatus int
	}{
		{
			name:   "successful update",
			userID: userID.String(),
			requestBody: map[string][]string{
				"role_ids": {roleID1.String(), roleID2.String()},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:   "invalid user ID",
			userID: "invalid-uuid",
			requestBody: map[string][]string{
				"role_ids": {roleID1.String()},
			},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "service error",
			userID: userID.String(),
			requestBody: map[string][]string{
				"role_ids": {roleID1.String()},
			},
			mockError:      errors.New("update failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAuthService{
				updateUserRolesFunc: func(ctx context.Context, uid uuid.UUID, roleIDs []uuid.UUID) error {
					return tt.mockError
				},
			}

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			handler := NewAuthHandler(mockService, logger)

			// Create test context
			c, w := createTestContext("PUT", "/users/"+tt.userID+"/roles", tt.requestBody)
			c.Params = []gin.Param{{Key: "user_id", Value: tt.userID}}

			// Mock admin role
			c.Set("user_roles", []string{"admin"})

			// Execute
			handler.UpdateUserRoles(c)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestAuthHandler_RotateKeys(t *testing.T) {
	tests := []struct {
		name        string
		mockError   error
		expectedStatus int
	}{
		{
			name:          "successful key rotation",
			mockError:     nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:          "service error",
			mockError:     errors.New("key rotation failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAuthService{
				rotateKeysFunc: func(ctx context.Context) error {
					return tt.mockError
				},
			}

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			handler := NewAuthHandler(mockService, logger)

			// Create test context
			c, w := createTestContext("POST", "/rotate-keys", nil)

			// Mock admin role
			c.Set("user_roles", []string{"admin"})

			// Execute
			handler.RotateKeys(c)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestAuthHandler_UpdateRole(t *testing.T) {
	roleID := uuid.New()
	tests := []struct {
		name           string
		roleID         string
		requestBody    map[string]string
		mockResponse   *models.Role
		mockError      error
		expectedStatus int
	}{
		{
			name:   "successful role update",
			roleID: roleID.String(),
			requestBody: map[string]string{
				"name":        "updated-admin",
				"description": "Updated administrator role",
			},
			mockResponse: &models.Role{
				ID:          roleID,
				Name:        "updated-admin",
				Description: stringPtr("Updated administrator role"),
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:   "invalid role ID",
			roleID: "invalid-uuid",
			requestBody: map[string]string{
				"name": "updated-admin",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:         "role not found",
			roleID:       roleID.String(),
			requestBody: map[string]string{
				"name": "updated-admin",
			},
			mockResponse:   nil,
			mockError:      errors.New("role not found"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAuthService{
				updateRoleFunc: func(ctx context.Context, rid uuid.UUID, name, description string) (*models.Role, error) {
					return tt.mockResponse, tt.mockError
				},
			}

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			handler := NewAuthHandler(mockService, logger)

			// Create test context
			c, w := createTestContext("PUT", "/roles/"+tt.roleID, tt.requestBody)
			c.Params = []gin.Param{{Key: "role_id", Value: tt.roleID}}

			// Mock admin role
			c.Set("user_roles", []string{"admin"})

			// Execute
			handler.UpdateRole(c)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal JSON response: %v", err)
				}

				if _, exists := response["id"]; !exists {
					t.Errorf("Expected id in response")
				}
				if _, exists := response["name"]; !exists {
					t.Errorf("Expected name in response")
				}
			}
		})
	}
}

func TestAuthHandler_DeleteRole(t *testing.T) {
	roleID := uuid.New()
	tests := []struct {
		name           string
		roleID         string
		mockError      error
		expectedStatus int
	}{
		{
			name:           "successful role deletion",
			roleID:         roleID.String(),
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid role ID",
			roleID:         "invalid-uuid",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "role not found",
			roleID:         roleID.String(),
			mockError:      errors.New("role not found"),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAuthService{
				deleteRoleFunc: func(ctx context.Context, rid uuid.UUID) error {
					return tt.mockError
				},
			}

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			handler := NewAuthHandler(mockService, logger)

			// Create test context
			c, w := createTestContext("DELETE", "/roles/"+tt.roleID, nil)
			c.Params = []gin.Param{{Key: "role_id", Value: tt.roleID}}

			// Mock admin role
			c.Set("user_roles", []string{"admin"})

			// Execute
			handler.DeleteRole(c)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestAuthHandler_CreatePermission(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]string
		mockResponse   *models.Permission
		mockError      error
		expectedStatus int
	}{
		{
			name: "successful permission creation",
			requestBody: map[string]string{
				"name":    "read:users",
				"resource": "users",
				"action":   "read",
			},
			mockResponse: &models.Permission{
				ID:       uuid.New(),
				Name:     "read:users",
				Resource: "users",
				Action:   "read",
			},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing required fields",
			requestBody: map[string]string{
				"name": "read:users",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			requestBody: map[string]string{
				"name":    "read:users",
				"resource": "users",
				"action":   "read",
			},
			mockResponse:   nil,
			mockError:      errors.New("duplicate key value violates unique constraint"),
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAuthService{
				createPermissionFunc: func(ctx context.Context, name, resource, action string) (*models.Permission, error) {
					return tt.mockResponse, tt.mockError
				},
			}

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			handler := NewAuthHandler(mockService, logger)

			// Create test context
			c, w := createTestContext("POST", "/permissions", tt.requestBody)

			// Mock admin role
			c.Set("user_roles", []string{"admin"})

			// Execute
			handler.CreatePermission(c)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusCreated {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal JSON response: %v", err)
				}

				if _, exists := response["id"]; !exists {
					t.Errorf("Expected id in response")
				}
				if _, exists := response["name"]; !exists {
					t.Errorf("Expected name in response")
				}
			}
		})
	}
}

func TestAuthHandler_ListPermissions(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   []models.Permission
		mockError      error
		expectedStatus int
	}{
		{
			name: "successful permission listing",
			mockResponse: []models.Permission{
				{
					ID:       uuid.New(),
					Name:     "read:users",
					Resource: "users",
					Action:   "read",
				},
				{
					ID:       uuid.New(),
					Name:     "write:users",
					Resource: "users",
					Action:   "write",
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "service error",
			mockResponse:   nil,
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAuthService{
				listPermissionsFunc: func(ctx context.Context) ([]models.Permission, error) {
					return tt.mockResponse, tt.mockError
				},
			}

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			handler := NewAuthHandler(mockService, logger)

			// Create test context
			c, w := createTestContext("GET", "/permissions", nil)

			// Mock admin role
			c.Set("user_roles", []string{"admin"})

			// Execute
			handler.ListPermissions(c)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal JSON response: %v", err)
				}

				if permissionsData, exists := response["permissions"]; exists {
					if permissionsArray, ok := permissionsData.([]interface{}); ok {
						if len(permissionsArray) != len(tt.mockResponse) {
							t.Errorf("Expected %d permissions, got %d", len(tt.mockResponse), len(permissionsArray))
						}
					}
				}
			}
		})
	}
}

func TestAuthHandler_GetPermission(t *testing.T) {
	permissionID := uuid.New()
	tests := []struct {
		name             string
		permissionID     string
		mockResponse     *models.Permission
		mockError        error
		expectedStatus   int
	}{
		{
			name:         "successful permission retrieval",
			permissionID: permissionID.String(),
			mockResponse: &models.Permission{
				ID:       permissionID,
				Name:     "read:users",
				Resource: "users",
				Action:   "read",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid permission ID",
			permissionID:   "invalid-uuid",
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "permission not found",
			permissionID:   permissionID.String(),
			mockResponse:   nil,
			mockError:      errors.New("permission not found"),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAuthService{
				getPermissionFunc: func(ctx context.Context, pid uuid.UUID) (*models.Permission, error) {
					return tt.mockResponse, tt.mockError
				},
			}

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			handler := NewAuthHandler(mockService, logger)

			// Create test context
			c, w := createTestContext("GET", "/permissions/"+tt.permissionID, nil)
			c.Params = []gin.Param{{Key: "permission_id", Value: tt.permissionID}}

			// Mock admin role
			c.Set("user_roles", []string{"admin"})

			// Execute
			handler.GetPermission(c)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal JSON response: %v", err)
				}

				if _, exists := response["id"]; !exists {
					t.Errorf("Expected id in response")
				}
				if _, exists := response["name"]; !exists {
					t.Errorf("Expected name in response")
				}
			}
		})
	}
}

func TestAuthHandler_UpdatePermission(t *testing.T) {
	permissionID := uuid.New()
	tests := []struct {
		name             string
		permissionID     string
		requestBody      map[string]string
		mockResponse     *models.Permission
		mockError        error
		expectedStatus   int
	}{
		{
			name:         "successful permission update",
			permissionID: permissionID.String(),
			requestBody: map[string]string{
				"name":    "write:users",
				"resource": "users",
				"action":   "write",
			},
			mockResponse: &models.Permission{
				ID:       permissionID,
				Name:     "write:users",
				Resource: "users",
				Action:   "write",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:         "invalid permission ID",
			permissionID: "invalid-uuid",
			requestBody: map[string]string{
				"name": "write:users",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:         "permission not found",
			permissionID: permissionID.String(),
			requestBody: map[string]string{
				"name":     "write:users",
				"resource": "users",
				"action":   "write",
			},
			mockResponse:   nil,
			mockError:      errors.New("permission not found"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAuthService{
				updatePermissionFunc: func(ctx context.Context, pid uuid.UUID, name, resource, action string) (*models.Permission, error) {
					return tt.mockResponse, tt.mockError
				},
			}

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			handler := NewAuthHandler(mockService, logger)

			// Create test context
			c, w := createTestContext("PUT", "/permissions/"+tt.permissionID, tt.requestBody)
			c.Params = []gin.Param{{Key: "permission_id", Value: tt.permissionID}}

			// Mock admin role
			c.Set("user_roles", []string{"admin"})

			// Execute
			handler.UpdatePermission(c)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal JSON response: %v", err)
				}

				if _, exists := response["id"]; !exists {
					t.Errorf("Expected id in response")
				}
				if _, exists := response["name"]; !exists {
					t.Errorf("Expected name in response")
				}
			}
		})
	}
}

func TestAuthHandler_DeletePermission(t *testing.T) {
	permissionID := uuid.New()
	tests := []struct {
		name             string
		permissionID     string
		mockError        error
		expectedStatus   int
	}{
		{
			name:           "successful permission deletion",
			permissionID:   permissionID.String(),
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid permission ID",
			permissionID:   "invalid-uuid",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "permission not found",
			permissionID:   permissionID.String(),
			mockError:      errors.New("permission not found"),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockAuthService{
				deletePermissionFunc: func(ctx context.Context, pid uuid.UUID) error {
					return tt.mockError
				},
			}

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			handler := NewAuthHandler(mockService, logger)

			// Create test context
			c, w := createTestContext("DELETE", "/permissions/"+tt.permissionID, nil)
			c.Params = []gin.Param{{Key: "permission_id", Value: tt.permissionID}}

			// Mock admin role
			c.Set("user_roles", []string{"admin"})

			// Execute
			handler.DeletePermission(c)

			// Assert
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}