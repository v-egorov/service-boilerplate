package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/client"
	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/models"
	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/utils"
)

// MockAuthRepository is a mock implementation of AuthRepository for testing
type MockAuthRepository struct {
	createAuthTokenFunc          func(ctx context.Context, token *models.AuthToken) error
	getAuthTokenByHashFunc       func(ctx context.Context, hash string) (*models.AuthToken, error)
	revokeAuthTokenFunc          func(ctx context.Context, tokenID uuid.UUID) error
	createUserSessionFunc        func(ctx context.Context, session *models.UserSession) error
	getUserRolesFunc             func(ctx context.Context, userID uuid.UUID) ([]models.Role, error)
	getRoleByNameFunc            func(ctx context.Context, name string) (*models.Role, error)
	createRoleFunc               func(ctx context.Context, role *models.Role) (*models.Role, error)
	listRolesFunc                func(ctx context.Context) ([]models.Role, error)
	getRoleFunc                  func(ctx context.Context, roleID uuid.UUID) (*models.Role, error)
	updateRoleFunc               func(ctx context.Context, roleID uuid.UUID, name, description string) (*models.Role, error)
	countUsersWithRoleFunc       func(ctx context.Context, roleID uuid.UUID) (int, error)
	deleteRoleFunc               func(ctx context.Context, roleID uuid.UUID) error
	createPermissionFunc         func(ctx context.Context, permission *models.Permission) (*models.Permission, error)
	listPermissionsFunc          func(ctx context.Context) ([]models.Permission, error)
	getPermissionFunc            func(ctx context.Context, permissionID uuid.UUID) (*models.Permission, error)
	updatePermissionFunc         func(ctx context.Context, permissionID uuid.UUID, name, resource, action string) (*models.Permission, error)
	countRolesWithPermissionFunc func(ctx context.Context, permissionID uuid.UUID) (int, error)
	deletePermissionFunc         func(ctx context.Context, permissionID uuid.UUID) error
	assignPermissionToRoleFunc   func(ctx context.Context, roleID, permissionID uuid.UUID) error
	removePermissionFromRoleFunc func(ctx context.Context, roleID, permissionID uuid.UUID) error
	getRolePermissionsFunc       func(ctx context.Context, roleID uuid.UUID) ([]models.Permission, error)
	assignRoleToUserFunc         func(ctx context.Context, userID, roleID uuid.UUID) error
	removeRoleFromUserFunc       func(ctx context.Context, userID, roleID uuid.UUID) error
	updateUserRolesFunc          func(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error
}

func (m *MockAuthRepository) CreateAuthToken(ctx context.Context, token *models.AuthToken) error {
	if m.createAuthTokenFunc != nil {
		return m.createAuthTokenFunc(ctx, token)
	}
	return nil
}

func (m *MockAuthRepository) GetAuthTokenByHash(ctx context.Context, hash string) (*models.AuthToken, error) {
	if m.getAuthTokenByHashFunc != nil {
		return m.getAuthTokenByHashFunc(ctx, hash)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAuthRepository) RevokeAuthToken(ctx context.Context, tokenID uuid.UUID) error {
	if m.revokeAuthTokenFunc != nil {
		return m.revokeAuthTokenFunc(ctx, tokenID)
	}
	return nil
}

func (m *MockAuthRepository) CreateUserSession(ctx context.Context, session *models.UserSession) error {
	if m.createUserSessionFunc != nil {
		return m.createUserSessionFunc(ctx, session)
	}
	return nil
}

func (m *MockAuthRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]models.Role, error) {
	if m.getUserRolesFunc != nil {
		return m.getUserRolesFunc(ctx, userID)
	}
	return []models.Role{}, nil
}

func (m *MockAuthRepository) GetRoleByName(ctx context.Context, name string) (*models.Role, error) {
	if m.getRoleByNameFunc != nil {
		return m.getRoleByNameFunc(ctx, name)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAuthRepository) CountUsersWithRole(ctx context.Context, roleID uuid.UUID) (int, error) {
	if m.countUsersWithRoleFunc != nil {
		return m.countUsersWithRoleFunc(ctx, roleID)
	}
	return 0, nil
}

func (m *MockAuthRepository) CountRolesWithPermission(ctx context.Context, permissionID uuid.UUID) (int, error) {
	if m.countRolesWithPermissionFunc != nil {
		return m.countRolesWithPermissionFunc(ctx, permissionID)
	}
	return 0, nil
}

func (m *MockAuthRepository) CreateRole(ctx context.Context, role *models.Role) (*models.Role, error) {
	if m.createRoleFunc != nil {
		return m.createRoleFunc(ctx, role)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAuthRepository) ListRoles(ctx context.Context) ([]models.Role, error) {
	if m.listRolesFunc != nil {
		return m.listRolesFunc(ctx)
	}
	return []models.Role{}, nil
}

func (m *MockAuthRepository) GetRole(ctx context.Context, roleID uuid.UUID) (*models.Role, error) {
	if m.getRoleFunc != nil {
		return m.getRoleFunc(ctx, roleID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAuthRepository) UpdateRole(ctx context.Context, roleID uuid.UUID, name, description string) (*models.Role, error) {
	if m.updateRoleFunc != nil {
		return m.updateRoleFunc(ctx, roleID, name, description)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAuthRepository) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	if m.deleteRoleFunc != nil {
		return m.deleteRoleFunc(ctx, roleID)
	}
	return nil
}

func (m *MockAuthRepository) CreatePermission(ctx context.Context, permission *models.Permission) (*models.Permission, error) {
	if m.createPermissionFunc != nil {
		return m.createPermissionFunc(ctx, permission)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAuthRepository) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	if m.listPermissionsFunc != nil {
		return m.listPermissionsFunc(ctx)
	}
	return []models.Permission{}, nil
}

func (m *MockAuthRepository) GetPermission(ctx context.Context, permissionID uuid.UUID) (*models.Permission, error) {
	if m.getPermissionFunc != nil {
		return m.getPermissionFunc(ctx, permissionID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAuthRepository) UpdatePermission(ctx context.Context, permissionID uuid.UUID, name, resource, action string) (*models.Permission, error) {
	if m.updatePermissionFunc != nil {
		return m.updatePermissionFunc(ctx, permissionID, name, resource, action)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAuthRepository) DeletePermission(ctx context.Context, permissionID uuid.UUID) error {
	if m.deletePermissionFunc != nil {
		return m.deletePermissionFunc(ctx, permissionID)
	}
	return nil
}

func (m *MockAuthRepository) AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	if m.assignPermissionToRoleFunc != nil {
		return m.assignPermissionToRoleFunc(ctx, roleID, permissionID)
	}
	return nil
}

func (m *MockAuthRepository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	if m.removePermissionFromRoleFunc != nil {
		return m.removePermissionFromRoleFunc(ctx, roleID, permissionID)
	}
	return nil
}

func (m *MockAuthRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]models.Permission, error) {
	if m.getRolePermissionsFunc != nil {
		return m.getRolePermissionsFunc(ctx, roleID)
	}
	return []models.Permission{}, nil
}

func (m *MockAuthRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	if m.assignRoleToUserFunc != nil {
		return m.assignRoleToUserFunc(ctx, userID, roleID)
	}
	return nil
}

func (m *MockAuthRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	if m.removeRoleFromUserFunc != nil {
		return m.removeRoleFromUserFunc(ctx, userID, roleID)
	}
	return nil
}

func (m *MockAuthRepository) UpdateUserRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	if m.updateUserRolesFunc != nil {
		return m.updateUserRolesFunc(ctx, userID, roleIDs)
	}
	return nil
}

// MockUserClient is a mock implementation of UserClient for testing
type MockUserClient struct {
	getUserWithPasswordByEmailFunc func(ctx context.Context, email string) (*client.UserLoginResponse, error)
	getUserByIDFunc                func(ctx context.Context, userID uuid.UUID) (*client.UserData, error)
	getUserByEmailFunc             func(ctx context.Context, email string) (*client.UserData, error)
	createUserFunc                 func(ctx context.Context, req *client.CreateUserRequest) (*client.UserData, error)
}

func (m *MockUserClient) GetUserWithPasswordByEmail(ctx context.Context, email string) (*client.UserLoginResponse, error) {
	if m.getUserWithPasswordByEmailFunc != nil {
		return m.getUserWithPasswordByEmailFunc(ctx, email)
	}
	return nil, errors.New("not implemented")
}

func (m *MockUserClient) GetUserByEmail(ctx context.Context, email string) (*client.UserData, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	return nil, errors.New("not implemented")
}

func (m *MockUserClient) GetUserByID(ctx context.Context, userID uuid.UUID) (*client.UserData, error) {
	if m.getUserByIDFunc != nil {
		return m.getUserByIDFunc(ctx, userID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockUserClient) CreateUser(ctx context.Context, req *client.CreateUserRequest) (*client.UserData, error) {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, req)
	}
	return nil, errors.New("not implemented")
}

// MockJWTUtils is a mock implementation of JWTUtils for testing
type MockJWTUtils struct {
	generateAccessTokenFunc  func(userID uuid.UUID, email string, roles []string, duration time.Duration) (string, error)
	generateRefreshTokenFunc func(userID uuid.UUID, duration time.Duration) (string, error)
	validateTokenFunc        func(tokenString string) (*utils.JWTClaims, error)
	getPublicKeyPEMFunc      func() ([]byte, error)
	rotateKeysFunc           func(ctx context.Context) error
	getKeyIDFunc             func() string
}

func (m *MockJWTUtils) GenerateAccessToken(userID uuid.UUID, email string, roles []string, duration time.Duration) (string, error) {
	if m.generateAccessTokenFunc != nil {
		return m.generateAccessTokenFunc(userID, email, roles, duration)
	}
	return "", errors.New("not implemented")
}

func (m *MockJWTUtils) GenerateRefreshToken(userID uuid.UUID, duration time.Duration) (string, error) {
	if m.generateRefreshTokenFunc != nil {
		return m.generateRefreshTokenFunc(userID, duration)
	}
	return "", errors.New("not implemented")
}

func (m *MockJWTUtils) ValidateToken(tokenString string) (*utils.JWTClaims, error) {
	if m.validateTokenFunc != nil {
		return m.validateTokenFunc(tokenString)
	}
	return nil, errors.New("not implemented")
}

func (m *MockJWTUtils) GetPublicKeyPEM() ([]byte, error) {
	if m.getPublicKeyPEMFunc != nil {
		return m.getPublicKeyPEMFunc()
	}
	return nil, errors.New("not implemented")
}

func (m *MockJWTUtils) RotateKeys(ctx context.Context) error {
	if m.rotateKeysFunc != nil {
		return m.rotateKeysFunc(ctx)
	}
	return nil
}

func (m *MockJWTUtils) GetKeyID() string {
	if m.getKeyIDFunc != nil {
		return m.getKeyIDFunc()
	}
	return ""
}

func TestAuthService_GetPublicKeyPEM(t *testing.T) {
	tests := []struct {
		name        string
		mockKey     []byte
		mockError   error
		expectError bool
		expectedKey []byte
	}{
		{
			name:        "successful key retrieval",
			mockKey:     []byte("-----BEGIN PUBLIC KEY-----\nMIIBIj...\n-----END PUBLIC KEY-----"),
			mockError:   nil,
			expectError: false,
			expectedKey: []byte("-----BEGIN PUBLIC KEY-----\nMIIBIj...\n-----END PUBLIC KEY-----"),
		},
		{
			name:        "key retrieval error",
			mockKey:     nil,
			mockError:   errors.New("key not found"),
			expectError: true,
			expectedKey: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockJWTUtils := &MockJWTUtils{
				getPublicKeyPEMFunc: func() ([]byte, error) {
					return tt.mockKey, tt.mockError
				},
			}

			// Create service
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			service := NewAuthService(nil, nil, mockJWTUtils, logger)

			// Execute
			result, err := service.GetPublicKeyPEM()

			// Assert
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if string(result) != string(tt.expectedKey) {
					t.Errorf("Expected key %s, got %s", string(tt.expectedKey), string(result))
				}
			}
		})
	}
}

func TestAuthService_CreateRole(t *testing.T) {
	tests := []struct {
		name        string
		roleName    string
		description string
		mockRole    *models.Role
		mockError   error
		expectError bool
	}{
		{
			name:        "successful role creation",
			roleName:    "admin",
			description: "Administrator role",
			mockRole: &models.Role{
				ID:          uuid.New(),
				Name:        "admin",
				Description: stringPtr("Administrator role"),
			},
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "role creation error",
			roleName:    "admin",
			description: "Administrator role",
			mockRole:    nil,
			mockError:   errors.New("role creation failed"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &MockAuthRepository{
				createRoleFunc: func(ctx context.Context, role *models.Role) (*models.Role, error) {
					return tt.mockRole, tt.mockError
				},
			}

			// Create service
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			service := NewAuthService(mockRepo, nil, nil, logger)

			// Execute
			result, err := service.CreateRole(context.Background(), tt.roleName, tt.description)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if result == nil {
					t.Errorf("Expected role but got nil")
				} else {
					if result.Name != tt.roleName {
						t.Errorf("Expected role name %s, got %s", tt.roleName, result.Name)
					}
				}
			}
		})
	}
}

func TestAuthService_GetUserRoles(t *testing.T) {
	userID := uuid.New()
	tests := []struct {
		name          string
		userID        uuid.UUID
		mockRoles     []models.Role
		mockError     error
		expectError   bool
		expectedCount int
	}{
		{
			name:   "successful role retrieval",
			userID: userID,
			mockRoles: []models.Role{
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
			mockError:     nil,
			expectError:   false,
			expectedCount: 2,
		},
		{
			name:          "user with no roles",
			userID:        userID,
			mockRoles:     []models.Role{},
			mockError:     nil,
			expectError:   false,
			expectedCount: 0,
		},
		{
			name:          "repository error",
			userID:        userID,
			mockRoles:     nil,
			mockError:     errors.New("database error"),
			expectError:   true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &MockAuthRepository{
				getUserRolesFunc: func(ctx context.Context, uid uuid.UUID) ([]models.Role, error) {
					return tt.mockRoles, tt.mockError
				},
			}

			// Create service
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			service := NewAuthService(mockRepo, nil, nil, logger)

			// Execute
			result, err := service.GetUserRoles(context.Background(), tt.userID)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if len(result) != tt.expectedCount {
					t.Errorf("Expected %d roles, got %d", tt.expectedCount, len(result))
				}
			}
		})
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

