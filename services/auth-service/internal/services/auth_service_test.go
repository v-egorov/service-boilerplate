package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedKey, result)
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
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.roleName, result.Name)
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
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	tests := []struct {
		name           string
		request        *models.LoginRequest
		ipAddress      string
		userAgent      string
		mockUserLogin  *client.UserLoginResponse
		mockUserError  error
		mockAccessToken string
		mockRefreshToken string
		mockTokenError error
		mockRepoError  error
		expectedError  string
		expectTokens   bool
	}{
		{
			name: "successful login",
			request: &models.LoginRequest{
				Email:    "user@example.com",
				Password: "password123",
			},
			ipAddress: "192.168.1.1",
			userAgent: "Mozilla/5.0",
			mockUserLogin: &client.UserLoginResponse{
				User: &client.UserData{
					ID:    uuid.New(),
					Email: "user@example.com",
				},
				PasswordHash: "$2a$10$oolyJReLQIIPPeH4XPtEhukeV9D115vs.XbyNQfw/zlTsF4/q8nly",
			},
			mockUserError:   nil,
			mockAccessToken: "access.jwt.token",
			mockRefreshToken: "refresh.jwt.token",
			mockTokenError:  nil,
			mockRepoError:   nil,
			expectedError:  "",
			expectTokens:   true,
		},
		{
			name: "user not found",
			request: &models.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			ipAddress:     "192.168.1.1",
			userAgent:     "Mozilla/5.0",
			mockUserLogin: nil,
			mockUserError: errors.New("user not found"),
			expectedError: "invalid credentials",
			expectTokens:  false,
		},
		{
			name: "invalid password",
			request: &models.LoginRequest{
				Email:    "user@example.com",
				Password: "wrongpassword",
			},
			ipAddress: "192.168.1.1",
			userAgent: "Mozilla/5.0",
			mockUserLogin: &client.UserLoginResponse{
				User: &client.UserData{
					ID:    uuid.New(),
					Email: "user@example.com",
				},
				PasswordHash: "$2a$10$oolyJReLQIIPPeH4XPtEhukeV9D115vs.XbyNQfw/zlTsF4/q8nly",
			},
			mockUserError:  nil,
			expectedError:  "invalid credentials",
			expectTokens:   false,
		},
		{
			name: "token generation failure",
			request: &models.LoginRequest{
				Email:    "user@example.com",
				Password: "password123",
			},
			ipAddress: "192.168.1.1",
			userAgent: "Mozilla/5.0",
			mockUserLogin: &client.UserLoginResponse{
				User: &client.UserData{
					ID:    uuid.New(),
					Email: "user@example.com",
				},
				PasswordHash: "$2a$10$oolyJReLQIIPPeH4XPtEhukeV9D115vs.XbyNQfw/zlTsF4/q8nly",
			},
			mockUserError:   nil,
			mockAccessToken: "",
			mockRefreshToken: "",
			mockTokenError:  errors.New("token generation failed"),
			expectedError:   "failed to generate access token: token generation failed",
			expectTokens:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := &MockAuthRepository{}
			mockUserClient := &MockUserClient{}
			mockJWTUtils := &MockJWTUtils{}

			// Setup expectations
			if tt.mockUserLogin != nil || tt.mockUserError != nil {
				mockUserClient.getUserWithPasswordByEmailFunc = func(ctx context.Context, email string) (*client.UserLoginResponse, error) {
					return tt.mockUserLogin, tt.mockUserError
				}
			}

			if tt.mockUserLogin != nil && tt.mockUserError == nil {
				mockRepo.getUserRolesFunc = func(ctx context.Context, userID uuid.UUID) ([]models.Role, error) {
					return []models.Role{}, nil
				}
				mockJWTUtils.generateAccessTokenFunc = func(userID uuid.UUID, email string, roles []string, duration time.Duration) (string, error) {
					return tt.mockAccessToken, tt.mockTokenError
				}

				if tt.mockTokenError == nil {
					mockJWTUtils.generateRefreshTokenFunc = func(userID uuid.UUID, duration time.Duration) (string, error) {
						return tt.mockRefreshToken, nil
					}
					mockRepo.createAuthTokenFunc = func(ctx context.Context, token *models.AuthToken) error {
						return tt.mockRepoError
					}
					if tt.mockRepoError == nil {
						mockRepo.createUserSessionFunc = func(ctx context.Context, session *models.UserSession) error {
							return nil
						}
					}
				}
			}

			// Create service
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			service := NewAuthService(mockRepo, mockUserClient, mockJWTUtils, logger)

			// Execute
			result, err := service.Login(context.Background(), tt.request, tt.ipAddress, tt.userAgent)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.expectTokens {
					assert.Equal(t, tt.mockAccessToken, result.AccessToken)
					assert.Equal(t, tt.mockRefreshToken, result.RefreshToken)
					assert.Equal(t, "Bearer", result.TokenType)
				}
			}
		})
	}
}

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name         string
		request      *models.RegisterRequest
		mockUserData *client.UserData
		mockUserError error
		mockRole     *models.Role
		mockRoleError error
		mockAssignError error
		expectedError string
		expectUser   bool
	}{
		{
			name: "successful registration",
			request: &models.RegisterRequest{
				Email:     "newuser@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			mockUserData: &client.UserData{
				ID:        uuid.New(),
				Email:     "newuser@example.com",
				FirstName: "John",
				LastName:  "Doe",
			},
			mockUserError: nil,
			mockRole: &models.Role{
				ID:   uuid.New(),
				Name: "user",
			},
			mockRoleError:   nil,
			mockAssignError: nil,
			expectedError:  "",
			expectUser:    true,
		},
		{
			name: "user creation failure",
			request: &models.RegisterRequest{
				Email:     "existing@example.com",
				Password:  "password123",
				FirstName: "Jane",
				LastName:  "Doe",
			},
			mockUserData:   nil,
			mockUserError: errors.New("user already exists"),
			expectedError: "user already exists",
			expectUser:    false,
		},
		{
			name: "default role not found",
			request: &models.RegisterRequest{
				Email:     "newuser@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			mockUserData: &client.UserData{
				ID:        uuid.New(),
				Email:     "newuser@example.com",
				FirstName: "John",
				LastName:  "Doe",
			},
			mockUserError: nil,
			mockRole:      nil,
			mockRoleError: errors.New("role not found"),
			expectedError: "failed to get default role",
			expectUser:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := &MockAuthRepository{}
			mockUserClient := &MockUserClient{}
			mockJWTUtils := &MockJWTUtils{}

			// Setup expectations
			mockUserClient.createUserFunc = func(ctx context.Context, req *client.CreateUserRequest) (*client.UserData, error) {
				return tt.mockUserData, tt.mockUserError
			}

			if tt.mockUserError == nil && tt.mockUserData != nil {
				mockRepo.getRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
					return tt.mockRole, tt.mockRoleError
				}

				if tt.mockRoleError == nil && tt.mockRole != nil {
					mockRepo.assignRoleToUserFunc = func(ctx context.Context, userID, roleID uuid.UUID) error {
						return tt.mockAssignError
					}
				}
			}

			// Create service
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			service := NewAuthService(mockRepo, mockUserClient, mockJWTUtils, logger)

			// Execute
			result, err := service.Register(context.Background(), tt.request)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.expectUser {
					assert.NotNil(t, result)
					assert.Equal(t, tt.mockUserData.Email, result.Email)
					assert.Equal(t, tt.mockUserData.FirstName, result.FirstName)
					assert.Equal(t, tt.mockUserData.LastName, result.LastName)
				}
			}
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	tests := []struct {
		name         string
		tokenString  string
		mockClaims   *utils.JWTClaims
		mockError    error
		expectedError string
		expectClaims bool
	}{
		{
			name:        "valid token",
			tokenString: "valid.jwt.token",
			mockClaims: &utils.JWTClaims{
				UserID:   uuid.New(),
				Email:    "user@example.com",
				Roles:    []string{"user"},
				TokenType: "access",
			},
			mockError:     nil,
			expectedError: "",
			expectClaims: true,
		},
		{
			name:          "invalid token",
			tokenString:  "invalid.jwt.token",
			mockClaims:    nil,
			mockError:     errors.New("invalid token"),
			expectedError: "invalid token",
			expectClaims:  false,
		},
		{
			name:          "expired token",
			tokenString:  "expired.jwt.token",
			mockClaims:    nil,
			mockError:     errors.New("token is expired"),
			expectedError: "token is expired",
			expectClaims:  false,
		},
		{
			name:          "malformed token",
			tokenString:  "malformed.token",
			mockClaims:    nil,
			mockError:     errors.New("token contains an invalid number of segments"),
			expectedError: "token contains an invalid number of segments",
			expectClaims:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := &MockAuthRepository{}
			mockUserClient := &MockUserClient{}
			mockJWTUtils := &MockJWTUtils{}

			// Setup expectations
			mockJWTUtils.validateTokenFunc = func(tokenString string) (*utils.JWTClaims, error) {
				return tt.mockClaims, tt.mockError
			}

			if tt.mockClaims != nil && tt.mockError == nil {
				mockRepo.getAuthTokenByHashFunc = func(ctx context.Context, hash string) (*models.AuthToken, error) {
					return &models.AuthToken{
						ID:        uuid.New(),
						UserID:    tt.mockClaims.UserID,
						TokenHash: hash,
						RevokedAt: nil,
					}, nil
				}
			}

			// Create service
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			service := NewAuthService(mockRepo, mockUserClient, mockJWTUtils, logger)

			// Execute
			result, err := service.ValidateToken(context.Background(), tt.tokenString)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.expectClaims {
					assert.NotNil(t, result)
					assert.Equal(t, tt.mockClaims.UserID, result.UserID)
					assert.Equal(t, tt.mockClaims.Email, result.Email)
					assert.Equal(t, tt.mockClaims.Roles, result.Roles)
				}
			}
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	tests := []struct {
		name           string
		request        *models.RefreshTokenRequest
		mockToken      *models.AuthToken
		mockTokenError error
		mockClaims     *utils.JWTClaims
		mockClaimsError error
		mockAccessToken string
		mockRefreshToken string
		mockTokenGenError error
		mockRepoError  error
		expectedError  string
		expectTokens   bool
	}{
		{
			name: "successful token refresh",
			request: &models.RefreshTokenRequest{
				RefreshToken: "refresh.jwt.token",
			},
			mockToken: &models.AuthToken{
				ID:     uuid.New(),
				UserID: uuid.New(),
				TokenHash: "token.hash",
			},
			mockTokenError: nil,
			mockClaims: &utils.JWTClaims{
				UserID:    uuid.New(),
				Email:     "user@example.com",
				Roles:     []string{"user"},
				TokenType: "refresh",
			},
			mockClaimsError: nil,
			mockAccessToken: "new.access.jwt.token",
			mockRefreshToken: "new.refresh.jwt.token",
			mockTokenGenError: nil,
			mockRepoError:   nil,
			expectedError:  "",
			expectTokens:   true,
		},
		{
			name: "invalid refresh token",
			request: &models.RefreshTokenRequest{
				RefreshToken: "invalid.refresh.token",
			},
			mockToken:      nil,
			mockTokenError: errors.New("token not found"),
			expectedError:  "invalid refresh token",
			expectTokens:   false,
		},
		{
			name: "revoked refresh token",
			request: &models.RefreshTokenRequest{
				RefreshToken: "revoked.refresh.token",
			},
			mockToken: &models.AuthToken{
				ID:        uuid.New(),
				UserID:    uuid.New(),
				TokenHash: "token.hash",
				RevokedAt: &time.Time{},
			},
			mockTokenError: nil,
			mockClaims: &utils.JWTClaims{
				UserID:    uuid.New(),
				Email:     "user@example.com",
				Roles:     []string{"user"},
				TokenType: "refresh",
			},
			mockClaimsError: nil,
			expectedError:   "refresh token has been revoked",
			expectTokens:    false,
		},
		{
			name: "token generation failure",
			request: &models.RefreshTokenRequest{
				RefreshToken: "refresh.jwt.token",
			},
			mockToken: &models.AuthToken{
				ID:     uuid.New(),
				UserID: uuid.New(),
				TokenHash: "token.hash",
			},
			mockTokenError: nil,
			mockClaims: &utils.JWTClaims{
				UserID:    uuid.New(),
				Email:     "user@example.com",
				Roles:     []string{"user"},
				TokenType: "refresh",
			},
			mockClaimsError: nil,
			mockAccessToken: "",
			mockRefreshToken: "",
			mockTokenGenError: errors.New("token generation failed"),
			expectedError:   "failed to generate new access token: token generation failed",
			expectTokens:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := &MockAuthRepository{}
			mockUserClient := &MockUserClient{}
			mockJWTUtils := &MockJWTUtils{}

			// Setup expectations
			mockRepo.getAuthTokenByHashFunc = func(ctx context.Context, hash string) (*models.AuthToken, error) {
				return tt.mockToken, tt.mockTokenError
			}

			if tt.mockToken != nil && tt.mockTokenError == nil {
				mockJWTUtils.validateTokenFunc = func(tokenString string) (*utils.JWTClaims, error) {
					return tt.mockClaims, tt.mockClaimsError
				}

				if tt.mockClaimsError == nil && tt.mockClaims != nil {
					mockRepo.getUserRolesFunc = func(ctx context.Context, userID uuid.UUID) ([]models.Role, error) {
						return []models.Role{}, nil
					}
					mockJWTUtils.generateAccessTokenFunc = func(userID uuid.UUID, email string, roles []string, duration time.Duration) (string, error) {
						return tt.mockAccessToken, tt.mockTokenGenError
					}

					if tt.mockTokenGenError == nil {
						mockJWTUtils.generateRefreshTokenFunc = func(userID uuid.UUID, duration time.Duration) (string, error) {
							return tt.mockRefreshToken, nil
						}
						mockRepo.revokeAuthTokenFunc = func(ctx context.Context, tokenID uuid.UUID) error {
							return nil
						}
						mockRepo.createAuthTokenFunc = func(ctx context.Context, token *models.AuthToken) error {
							return tt.mockRepoError
						}
					}
				}
			}

			// Create service
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			service := NewAuthService(mockRepo, mockUserClient, mockJWTUtils, logger)

			// Execute
			result, err := service.RefreshToken(context.Background(), tt.request)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.expectTokens {
					assert.Equal(t, tt.mockAccessToken, result.AccessToken)
					assert.Equal(t, tt.mockRefreshToken, result.RefreshToken)
					assert.Equal(t, "Bearer", result.TokenType)
				}
			}
		})
	}
}

func TestAuthService_ListRoles(t *testing.T) {
	tests := []struct {
		name         string
		mockRoles    []models.Role
		mockError    error
		expectError  bool
		expectedCount int
	}{
		{
			name: "successful role listing",
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
			name:          "empty role list",
			mockRoles:     []models.Role{},
			mockError:     nil,
			expectError:   false,
			expectedCount: 0,
		},
		{
			name:          "repository error",
			mockRoles:     nil,
			mockError:     errors.New("database connection failed"),
			expectError:   true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &MockAuthRepository{
				listRolesFunc: func(ctx context.Context) ([]models.Role, error) {
					return tt.mockRoles, tt.mockError
				},
			}

			// Create service
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			service := NewAuthService(mockRepo, nil, nil, logger)

			// Execute
			result, err := service.ListRoles(context.Background())

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to list roles")
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)
				if tt.expectedCount > 0 {
					assert.Equal(t, tt.mockRoles[0].Name, result[0].Name)
				}
			}
		})
	}
}

func TestAuthService_GetRole(t *testing.T) {
	roleID := uuid.New()
	tests := []struct {
		name        string
		roleID      uuid.UUID
		mockRole    *models.Role
		mockError   error
		expectError bool
	}{
		{
			name:   "successful role retrieval",
			roleID: roleID,
			mockRole: &models.Role{
				ID:          roleID,
				Name:        "admin",
				Description: stringPtr("Administrator role"),
			},
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "role not found",
			roleID:      roleID,
			mockRole:    nil,
			mockError:   errors.New("role not found"),
			expectError: true,
		},
		{
			name:        "repository error",
			roleID:      roleID,
			mockRole:    nil,
			mockError:   errors.New("database connection failed"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &MockAuthRepository{
				getRoleFunc: func(ctx context.Context, rid uuid.UUID) (*models.Role, error) {
					return tt.mockRole, tt.mockError
				},
			}

			// Create service
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			service := NewAuthService(mockRepo, nil, nil, logger)

			// Execute
			result, err := service.GetRole(context.Background(), tt.roleID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to get role")
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockRole.ID, result.ID)
				assert.Equal(t, tt.mockRole.Name, result.Name)
			}
		})
	}
}

func TestAuthService_UpdateRole(t *testing.T) {
	roleID := uuid.New()
	tests := []struct {
		name         string
		roleID       uuid.UUID
		nameUpdate   string
		descUpdate   string
		mockRole     *models.Role
		mockError    error
		expectError  bool
	}{
		{
			name:       "successful role update",
			roleID:     roleID,
			nameUpdate: "super-admin",
			descUpdate: "Super Administrator role",
			mockRole: &models.Role{
				ID:          roleID,
				Name:        "super-admin",
				Description: stringPtr("Super Administrator role"),
			},
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "role update error",
			roleID:      roleID,
			nameUpdate:  "invalid-role",
			descUpdate:  "Invalid role",
			mockRole:    nil,
			mockError:   errors.New("role not found"),
			expectError: true,
		},
		{
			name:        "repository error",
			roleID:      roleID,
			nameUpdate:  "admin",
			descUpdate:  "Administrator role",
			mockRole:    nil,
			mockError:   errors.New("database connection failed"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &MockAuthRepository{
				updateRoleFunc: func(ctx context.Context, rid uuid.UUID, name, description string) (*models.Role, error) {
					return tt.mockRole, tt.mockError
				},
			}

			// Create service
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			service := NewAuthService(mockRepo, nil, nil, logger)

			// Execute
			result, err := service.UpdateRole(context.Background(), tt.roleID, tt.nameUpdate, tt.descUpdate)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to update role")
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockRole.ID, result.ID)
				assert.Equal(t, tt.nameUpdate, result.Name)
			}
		})
	}
}

func TestAuthService_DeleteRole(t *testing.T) {
	roleID := uuid.New()
	tests := []struct {
		name           string
		roleID         uuid.UUID
		mockUserCount  int
		mockCountError error
		mockDeleteError error
		expectError    bool
		expectedErrorMsg string
	}{
		{
			name:            "successful role deletion",
			roleID:          roleID,
			mockUserCount:   0,
			mockCountError:  nil,
			mockDeleteError: nil,
			expectError:     false,
		},
		{
			name:             "role in use",
			roleID:           roleID,
			mockUserCount:    3,
			mockCountError:   nil,
			mockDeleteError:  nil,
			expectError:      true,
			expectedErrorMsg: "cannot delete role: 3 users are assigned to this role",
		},
		{
			name:            "count users error",
			roleID:          roleID,
			mockUserCount:   0,
			mockCountError:  errors.New("database error"),
			mockDeleteError: nil,
			expectError:     true,
			expectedErrorMsg: "failed to check role usage",
		},
		{
			name:            "delete role error",
			roleID:          roleID,
			mockUserCount:   0,
			mockCountError:  nil,
			mockDeleteError: errors.New("delete failed"),
			expectError:     true,
			expectedErrorMsg: "failed to delete role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &MockAuthRepository{
				countUsersWithRoleFunc: func(ctx context.Context, rid uuid.UUID) (int, error) {
					return tt.mockUserCount, tt.mockCountError
				},
				deleteRoleFunc: func(ctx context.Context, rid uuid.UUID) error {
					return tt.mockDeleteError
				},
			}

			// Create service
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			service := NewAuthService(mockRepo, nil, nil, logger)

			// Execute
			err := service.DeleteRole(context.Background(), tt.roleID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthService_CreatePermission(t *testing.T) {
	tests := []struct {
		name         string
		permName     string
		resource     string
		action       string
		mockPermission *models.Permission
		mockError    error
		expectError  bool
	}{
		{
			name:     "successful permission creation",
			permName: "read_users",
			resource: "users",
			action:   "read",
			mockPermission: &models.Permission{
				ID:       uuid.New(),
				Name:     "read_users",
				Resource: "users",
				Action:   "read",
			},
			mockError:   nil,
			expectError: false,
		},
		{
			name:         "permission creation error",
			permName:     "invalid_perm",
			resource:     "invalid",
			action:       "invalid",
			mockPermission: nil,
			mockError:    errors.New("permission already exists"),
			expectError:  true,
		},
		{
			name:         "repository error",
			permName:     "write_posts",
			resource:     "posts",
			action:       "write",
			mockPermission: nil,
			mockError:    errors.New("database connection failed"),
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &MockAuthRepository{
				createPermissionFunc: func(ctx context.Context, permission *models.Permission) (*models.Permission, error) {
					return tt.mockPermission, tt.mockError
				},
			}

			// Create service
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			service := NewAuthService(mockRepo, nil, nil, logger)

			// Execute
			result, err := service.CreatePermission(context.Background(), tt.permName, tt.resource, tt.action)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to create permission")
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.permName, result.Name)
				assert.Equal(t, tt.resource, result.Resource)
				assert.Equal(t, tt.action, result.Action)
			}
		})
	}
}

func TestAuthService_ListPermissions(t *testing.T) {
	tests := []struct {
		name            string
		mockPermissions []models.Permission
		mockError       error
		expectError     bool
		expectedCount   int
	}{
		{
			name: "successful permission listing",
			mockPermissions: []models.Permission{
				{
					ID:       uuid.New(),
					Name:     "read_users",
					Resource: "users",
					Action:   "read",
				},
				{
					ID:       uuid.New(),
					Name:     "write_posts",
					Resource: "posts",
					Action:   "write",
				},
			},
			mockError:     nil,
			expectError:   false,
			expectedCount: 2,
		},
		{
			name:            "empty permission list",
			mockPermissions: []models.Permission{},
			mockError:       nil,
			expectError:     false,
			expectedCount:   0,
		},
		{
			name:            "repository error",
			mockPermissions: nil,
			mockError:       errors.New("database connection failed"),
			expectError:     true,
			expectedCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &MockAuthRepository{
				listPermissionsFunc: func(ctx context.Context) ([]models.Permission, error) {
					return tt.mockPermissions, tt.mockError
				},
			}

			// Create service
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			service := NewAuthService(mockRepo, nil, nil, logger)

			// Execute
			result, err := service.ListPermissions(context.Background())

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to list permissions")
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)
				if tt.expectedCount > 0 {
					assert.Equal(t, tt.mockPermissions[0].Name, result[0].Name)
				}
			}
		})
	}
}

func TestAuthService_GetPermission(t *testing.T) {
	permissionID := uuid.New()
	tests := []struct {
		name           string
		permissionID   uuid.UUID
		mockPermission *models.Permission
		mockError      error
		expectError    bool
	}{
		{
			name:         "successful permission retrieval",
			permissionID: permissionID,
			mockPermission: &models.Permission{
				ID:       permissionID,
				Name:     "read_users",
				Resource: "users",
				Action:   "read",
			},
			mockError:   nil,
			expectError: false,
		},
		{
			name:           "permission not found",
			permissionID:   permissionID,
			mockPermission: nil,
			mockError:      errors.New("permission not found"),
			expectError:    true,
		},
		{
			name:           "repository error",
			permissionID:   permissionID,
			mockPermission: nil,
			mockError:      errors.New("database connection failed"),
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &MockAuthRepository{
				getPermissionFunc: func(ctx context.Context, pid uuid.UUID) (*models.Permission, error) {
					return tt.mockPermission, tt.mockError
				},
			}

			// Create service
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			service := NewAuthService(mockRepo, nil, nil, logger)

			// Execute
			result, err := service.GetPermission(context.Background(), tt.permissionID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to get permission")
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockPermission.ID, result.ID)
				assert.Equal(t, tt.mockPermission.Name, result.Name)
			}
		})
	}
}

func TestAuthService_UpdatePermission(t *testing.T) {
	permissionID := uuid.New()
	tests := []struct {
		name           string
		permissionID   uuid.UUID
		nameUpdate     string
		resourceUpdate string
		actionUpdate   string
		mockPermission *models.Permission
		mockError      error
		expectError    bool
	}{
		{
			name:           "successful permission update",
			permissionID:   permissionID,
			nameUpdate:     "write_users",
			resourceUpdate: "users",
			actionUpdate:   "write",
			mockPermission: &models.Permission{
				ID:       permissionID,
				Name:     "write_users",
				Resource: "users",
				Action:   "write",
			},
			mockError:   nil,
			expectError: false,
		},
		{
			name:           "permission update error",
			permissionID:   permissionID,
			nameUpdate:     "invalid_perm",
			resourceUpdate: "invalid",
			actionUpdate:   "invalid",
			mockPermission: nil,
			mockError:      errors.New("permission not found"),
			expectError:    true,
		},
		{
			name:           "repository error",
			permissionID:   permissionID,
			nameUpdate:     "delete_posts",
			resourceUpdate: "posts",
			actionUpdate:   "delete",
			mockPermission: nil,
			mockError:      errors.New("database connection failed"),
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &MockAuthRepository{
				updatePermissionFunc: func(ctx context.Context, pid uuid.UUID, name, resource, action string) (*models.Permission, error) {
					return tt.mockPermission, tt.mockError
				},
			}

			// Create service
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			service := NewAuthService(mockRepo, nil, nil, logger)

			// Execute
			result, err := service.UpdatePermission(context.Background(), tt.permissionID, tt.nameUpdate, tt.resourceUpdate, tt.actionUpdate)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to update permission")
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockPermission.ID, result.ID)
				assert.Equal(t, tt.nameUpdate, result.Name)
				assert.Equal(t, tt.resourceUpdate, result.Resource)
				assert.Equal(t, tt.actionUpdate, result.Action)
			}
		})
	}
}

func TestAuthService_DeletePermission(t *testing.T) {
	permissionID := uuid.New()
	tests := []struct {
		name             string
		permissionID     uuid.UUID
		mockRoleCount    int
		mockCountError   error
		mockDeleteError  error
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:            "successful permission deletion",
			permissionID:    permissionID,
			mockRoleCount:   0,
			mockCountError:  nil,
			mockDeleteError: nil,
			expectError:     false,
		},
		{
			name:             "permission in use",
			permissionID:     permissionID,
			mockRoleCount:    2,
			mockCountError:   nil,
			mockDeleteError:  nil,
			expectError:      true,
			expectedErrorMsg: "cannot delete permission: 2 roles are assigned this permission",
		},
		{
			name:            "count roles error",
			permissionID:    permissionID,
			mockRoleCount:   0,
			mockCountError:  errors.New("database error"),
			mockDeleteError: nil,
			expectError:     true,
			expectedErrorMsg: "failed to check permission usage",
		},
		{
			name:            "delete permission error",
			permissionID:    permissionID,
			mockRoleCount:   0,
			mockCountError:  nil,
			mockDeleteError: errors.New("delete failed"),
			expectError:     true,
			expectedErrorMsg: "failed to delete permission",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &MockAuthRepository{
				countRolesWithPermissionFunc: func(ctx context.Context, pid uuid.UUID) (int, error) {
					return tt.mockRoleCount, tt.mockCountError
				},
				deletePermissionFunc: func(ctx context.Context, pid uuid.UUID) error {
					return tt.mockDeleteError
				},
			}

			// Create service
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			service := NewAuthService(mockRepo, nil, nil, logger)

			// Execute
			err := service.DeletePermission(context.Background(), tt.permissionID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
