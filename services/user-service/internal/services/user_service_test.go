package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/v-egorov/service-boilerplate/services/user-service/internal/models"
)

// MockUserRepository is a testify mock for UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, id uuid.UUID, user *models.User) (*models.User, error) {
	args := m.Called(ctx, id, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func TestUserService_CreateUser(t *testing.T) {
	tests := []struct {
		name           string
		request        *models.CreateUserRequest
		mockGetByEmail *models.User
		mockGetByEmailErr error
		mockCreate     *models.User
		mockCreateErr error
		expectError    bool
		expectedErrorType string
	}{
		{
			name: "successful user creation",
			request: &models.CreateUserRequest{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			mockGetByEmail: nil,
			mockGetByEmailErr: errors.New("no rows"),
			mockCreate: &models.User{
				ID:        uuid.New(),
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			mockCreateErr: nil,
			expectError:   false,
		},
		{
			name: "user already exists",
			request: &models.CreateUserRequest{
				Email:     "existing@example.com",
				Password:  "password123",
				FirstName: "Jane",
				LastName:  "Smith",
			},
			mockGetByEmail: &models.User{
				ID:    uuid.New(),
				Email: "existing@example.com",
			},
			mockGetByEmailErr: nil,
			expectError:       true,
			expectedErrorType: "conflict",
		},
		{
			name: "invalid email",
			request: &models.CreateUserRequest{
				Email:     "invalid-email",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			expectError:       true,
			expectedErrorType: "validation",
		},
		{
			name: "weak password",
			request: &models.CreateUserRequest{
				Email:     "test@example.com",
				Password:  "123",
				FirstName: "John",
				LastName:  "Doe",
			},
			expectError:       true,
			expectedErrorType: "validation",
		},
		{
			name: "missing first name",
			request: &models.CreateUserRequest{
				Email:    "test@example.com",
				Password: "password123",
				LastName: "Doe",
			},
			expectError:       true,
			expectedErrorType: "validation",
		},
		{
			name: "database error on create",
			request: &models.CreateUserRequest{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			mockGetByEmail: nil,
			mockGetByEmailErr: errors.New("no rows"),
			mockCreate:     nil,
			mockCreateErr:  errors.New("database connection failed"),
			expectError:    true,
			expectedErrorType: "internal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &MockUserRepository{}
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			service := NewUserServiceWithInterface(mockRepo, logger)

			// Setup expectations only if needed
			if tt.mockGetByEmail != nil || tt.mockGetByEmailErr != nil {
				mockRepo.On("GetByEmail", mock.Anything, tt.request.Email).Return(tt.mockGetByEmail, tt.mockGetByEmailErr).Once()
			}

			if tt.mockCreate != nil || tt.mockCreateErr != nil {
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(tt.mockCreate, tt.mockCreateErr).Once()
			}

			// Execute
			result, err := service.CreateUser(context.Background(), tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)

				if tt.expectedErrorType != "" {
					switch tt.expectedErrorType {
					case "validation":
						assert.IsType(t, models.ValidationError{}, err)
					case "conflict":
						assert.IsType(t, models.ConflictError{}, err)
					case "internal":
						assert.IsType(t, models.InternalError{}, err)
					}
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.request.Email, result.Email)
				assert.Equal(t, tt.request.FirstName, result.FirstName)
				assert.Equal(t, tt.request.LastName, result.LastName)
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetUserByEmail(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		mockUser      *models.User
		mockError     error
		expectError   bool
		expectedErrorType string
	}{
		{
			name:  "successful user retrieval",
			email: "test@example.com",
			mockUser: &models.User{
				ID:        uuid.New(),
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			mockError:   nil,
			expectError: false,
		},
		{
			name:            "user not found",
			email:           "notfound@example.com",
			mockUser:        nil,
			mockError:       errors.New("no rows"),
			expectError:     true,
			expectedErrorType: "not_found",
		},
		{
			name:            "empty email",
			email:           "",
			expectError:     true,
			expectedErrorType: "validation",
		},
		{
			name:            "database error",
			email:           "test@example.com",
			mockUser:        nil,
			mockError:       errors.New("database connection failed"),
			expectError:     true,
			expectedErrorType: "internal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &MockUserRepository{}
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			service := NewUserServiceWithInterface(mockRepo, logger)

			// Setup expectations only if needed
			if tt.mockUser != nil || tt.mockError != nil {
				mockRepo.On("GetByEmail", mock.Anything, tt.email).Return(tt.mockUser, tt.mockError).Once()
			}

			// Execute
			result, err := service.GetUserByEmail(context.Background(), tt.email)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)

				if tt.expectedErrorType != "" {
					switch tt.expectedErrorType {
					case "validation":
						assert.IsType(t, models.ValidationError{}, err)
					case "not_found":
						assert.IsType(t, models.NotFoundError{}, err)
					case "internal":
						assert.IsType(t, models.InternalError{}, err)
					}
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockUser.ID, result.ID)
				assert.Equal(t, tt.mockUser.Email, result.Email)
				assert.Equal(t, tt.mockUser.FirstName, result.FirstName)
				assert.Equal(t, tt.mockUser.LastName, result.LastName)
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetUserWithPasswordByEmail(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		mockUser      *models.User
		mockError     error
		expectError   bool
		expectedErrorType string
	}{
		{
			name:  "successful user retrieval with password",
			email: "test@example.com",
			mockUser: &models.User{
				ID:           uuid.New(),
				Email:        "test@example.com",
				PasswordHash: "hashed_password",
				FirstName:    "John",
				LastName:     "Doe",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			mockError:   nil,
			expectError: false,
		},
		{
			name:            "user not found",
			email:           "notfound@example.com",
			mockUser:        nil,
			mockError:       errors.New("no rows"),
			expectError:     true,
			expectedErrorType: "not_found",
		},
		{
			name:            "empty email",
			email:           "",
			expectError:     true,
			expectedErrorType: "validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &MockUserRepository{}
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			service := NewUserServiceWithInterface(mockRepo, logger)

			// Setup expectations only if needed
			if tt.mockUser != nil || tt.mockError != nil {
				mockRepo.On("GetByEmail", mock.Anything, tt.email).Return(tt.mockUser, tt.mockError).Once()
			}

			// Execute
			result, err := service.GetUserWithPasswordByEmail(context.Background(), tt.email)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)

				if tt.expectedErrorType != "" {
					switch tt.expectedErrorType {
					case "validation":
						assert.IsType(t, models.ValidationError{}, err)
					case "not_found":
						assert.IsType(t, models.NotFoundError{}, err)
					}
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockUser.ID, result.User.ID)
				assert.Equal(t, tt.mockUser.Email, result.User.Email)
				assert.Equal(t, tt.mockUser.PasswordHash, result.PasswordHash)
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetUser(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name          string
		userID        uuid.UUID
		mockUser      *models.User
		mockError     error
		expectError   bool
		expectedErrorType string
	}{
		{
			name:   "successful user retrieval",
			userID: userID,
			mockUser: &models.User{
				ID:        userID,
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			mockError:   nil,
			expectError: false,
		},
		{
			name:            "user not found",
			userID:          userID,
			mockUser:        nil,
			mockError:       errors.New("no rows"),
			expectError:     true,
			expectedErrorType: "not_found",
		},
		{
			name:            "nil user ID",
			userID:          uuid.Nil,
			expectError:     true,
			expectedErrorType: "validation",
		},
		{
			name:            "database error",
			userID:          userID,
			mockUser:        nil,
			mockError:       errors.New("database connection failed"),
			expectError:     true,
			expectedErrorType: "internal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &MockUserRepository{}
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			service := NewUserServiceWithInterface(mockRepo, logger)

			// Setup expectations only if needed
			if tt.mockUser != nil || tt.mockError != nil {
				mockRepo.On("GetByID", mock.Anything, tt.userID).Return(tt.mockUser, tt.mockError).Once()
			}

			// Execute
			result, err := service.GetUser(context.Background(), tt.userID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)

				if tt.expectedErrorType != "" {
					switch tt.expectedErrorType {
					case "validation":
						assert.IsType(t, models.ValidationError{}, err)
					case "not_found":
						assert.IsType(t, models.NotFoundError{}, err)
					case "internal":
						assert.IsType(t, models.InternalError{}, err)
					}
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockUser.ID, result.ID)
				assert.Equal(t, tt.mockUser.Email, result.Email)
				assert.Equal(t, tt.mockUser.FirstName, result.FirstName)
				assert.Equal(t, tt.mockUser.LastName, result.LastName)
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_ReplaceUser(t *testing.T) {
	userID := uuid.New()
	existingUser := &models.User{
		ID:        userID,
		Email:     "old@example.com",
		FirstName: "Old",
		LastName:  "User",
		CreatedAt: time.Now().Add(-time.Hour),
	}

	tests := []struct {
		name          string
		userID        uuid.UUID
		request       *models.ReplaceUserRequest
		mockGetByID   *models.User
		mockGetByIDErr error
		mockUpdate    *models.User
		mockUpdateErr error
		expectError   bool
		expectedErrorType string
	}{
		{
			name:   "successful user replacement",
			userID: userID,
			request: &models.ReplaceUserRequest{
				Email:     "new@example.com",
				FirstName: "New",
				LastName:  "User",
			},
			mockGetByID: existingUser,
			mockGetByIDErr: nil,
			mockUpdate: &models.User{
				ID:        userID,
				Email:     "new@example.com",
				FirstName: "New",
				LastName:  "User",
				CreatedAt: existingUser.CreatedAt,
				UpdatedAt: time.Now(),
			},
			mockUpdateErr: nil,
			expectError:   false,
		},
		{
			name:            "user not found",
			userID:          userID,
			request:         &models.ReplaceUserRequest{Email: "test@example.com", FirstName: "Test", LastName: "User"},
			mockGetByID:     nil,
			mockGetByIDErr:  errors.New("user not found"),
			expectError:     true,
			expectedErrorType: "not_found",
		},
		{
			name:            "nil user ID",
			userID:          uuid.Nil,
			request:         &models.ReplaceUserRequest{Email: "test@example.com", FirstName: "Test", LastName: "User"},
			expectError:     true,
			expectedErrorType: "validation",
		},
		{
			name:   "invalid email",
			userID: userID,
			request: &models.ReplaceUserRequest{
				Email:     "invalid-email",
				FirstName: "Test",
				LastName:  "User",
			},
			expectError:       true,
			expectedErrorType: "validation",
		},
		{
			name:   "email conflict",
			userID: userID,
			request: &models.ReplaceUserRequest{
				Email:     "existing@example.com",
				FirstName: "Test",
				LastName:  "User",
			},
			mockGetByID: existingUser,
			mockGetByIDErr: nil,
			mockUpdate:    nil,
			mockUpdateErr: errors.New("duplicate key value violates unique constraint"),
			expectError:   true,
			expectedErrorType: "conflict",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &MockUserRepository{}
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			service := NewUserServiceWithInterface(mockRepo, logger)

			// Setup expectations only if needed
			if tt.mockGetByID != nil || tt.mockGetByIDErr != nil {
				mockRepo.On("GetByID", mock.Anything, tt.userID).Return(tt.mockGetByID, tt.mockGetByIDErr).Once()
			}

			if tt.mockUpdate != nil || tt.mockUpdateErr != nil {
				mockRepo.On("Update", mock.Anything, tt.userID, mock.AnythingOfType("*models.User")).Return(tt.mockUpdate, tt.mockUpdateErr).Once()
			}

			// Execute
			result, err := service.ReplaceUser(context.Background(), tt.userID, tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)

				if tt.expectedErrorType != "" {
					switch tt.expectedErrorType {
					case "validation":
						assert.IsType(t, models.ValidationError{}, err)
					case "not_found":
						assert.IsType(t, models.NotFoundError{}, err)
					case "conflict":
						assert.IsType(t, models.ConflictError{}, err)
					}
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.userID, result.ID)
				assert.Equal(t, tt.request.Email, result.Email)
				assert.Equal(t, tt.request.FirstName, result.FirstName)
				assert.Equal(t, tt.request.LastName, result.LastName)
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	userID := uuid.New()
	existingUser := &models.User{
		ID:        userID,
		Email:     "old@example.com",
		FirstName: "Old",
		LastName:  "Name",
		CreatedAt: time.Now().Add(-time.Hour),
	}

	tests := []struct {
		name          string
		userID        uuid.UUID
		request       *models.UpdateUserRequest
		mockGetByID   *models.User
		mockGetByIDErr error
		mockUpdate    *models.User
		mockUpdateErr error
		expectError   bool
		expectedErrorType string
	}{
		{
			name:   "successful user update",
			userID: userID,
			request: &models.UpdateUserRequest{
				Email:     "new@example.com",
				FirstName: "New",
			},
			mockGetByID: existingUser,
			mockGetByIDErr: nil,
			mockUpdate: &models.User{
				ID:        userID,
				Email:     "new@example.com",
				FirstName: "New",
				LastName:  "Name",
				CreatedAt: existingUser.CreatedAt,
				UpdatedAt: time.Now(),
			},
			mockUpdateErr: nil,
			expectError:   false,
		},
		{
			name:            "user not found",
			userID:          userID,
			request:         &models.UpdateUserRequest{Email: "test@example.com"},
			mockGetByID:     nil,
			mockGetByIDErr:  errors.New("user not found"),
			expectError:     true,
			expectedErrorType: "not_found",
		},
		{
			name:            "nil user ID",
			userID:          uuid.Nil,
			request:         &models.UpdateUserRequest{Email: "test@example.com"},
			expectError:     true,
			expectedErrorType: "validation",
		},
		{
			name:   "empty update request",
			userID: userID,
			request: &models.UpdateUserRequest{},
			expectError:       true,
			expectedErrorType: "validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &MockUserRepository{}
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			service := NewUserServiceWithInterface(mockRepo, logger)

			// Setup expectations only if needed
			if tt.mockGetByID != nil || tt.mockGetByIDErr != nil {
				mockRepo.On("GetByID", mock.Anything, tt.userID).Return(tt.mockGetByID, tt.mockGetByIDErr).Once()
			}

			if tt.mockUpdate != nil || tt.mockUpdateErr != nil {
				mockRepo.On("Update", mock.Anything, tt.userID, mock.AnythingOfType("*models.User")).Return(tt.mockUpdate, tt.mockUpdateErr).Once()
			}

			// Execute
			result, err := service.UpdateUser(context.Background(), tt.userID, tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)

				if tt.expectedErrorType != "" {
					switch tt.expectedErrorType {
					case "validation":
						assert.IsType(t, models.ValidationError{}, err)
					case "not_found":
						assert.IsType(t, models.NotFoundError{}, err)
					}
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.userID, result.ID)
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_DeleteUser(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name          string
		userID        uuid.UUID
		mockDeleteErr error
		expectError   bool
		expectedErrorType string
	}{
		{
			name:          "successful user deletion",
			userID:        userID,
			mockDeleteErr: nil,
			expectError:   false,
		},
		{
			name:            "user not found",
			userID:          userID,
			mockDeleteErr:   errors.New("no rows"),
			expectError:     true,
			expectedErrorType: "not_found",
		},
		{
			name:            "nil user ID",
			userID:          uuid.Nil,
			expectError:     true,
			expectedErrorType: "validation",
		},
		{
			name:            "database error",
			userID:          userID,
			mockDeleteErr:   errors.New("database connection failed"),
			expectError:     true,
			expectedErrorType: "internal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &MockUserRepository{}
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			service := NewUserServiceWithInterface(mockRepo, logger)

			// Setup expectations only if needed
			if tt.mockDeleteErr != nil || tt.userID != uuid.Nil {
				mockRepo.On("Delete", mock.Anything, tt.userID).Return(tt.mockDeleteErr).Once()
			}

			// Execute
			err := service.DeleteUser(context.Background(), tt.userID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)

				if tt.expectedErrorType != "" {
					switch tt.expectedErrorType {
					case "validation":
						assert.IsType(t, models.ValidationError{}, err)
					case "not_found":
						assert.IsType(t, models.NotFoundError{}, err)
					case "internal":
						assert.IsType(t, models.InternalError{}, err)
					}
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_ListUsers(t *testing.T) {
	users := []*models.User{
		{
			ID:        uuid.New(),
			Email:     "user1@example.com",
			FirstName: "User",
			LastName:  "One",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Email:     "user2@example.com",
			FirstName: "User",
			LastName:  "Two",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	tests := []struct {
		name          string
		limit         int
		offset        int
		mockUsers     []*models.User
		mockError     error
		expectError   bool
		expectedCount int
	}{
		{
			name:          "successful user listing",
			limit:         10,
			offset:        0,
			mockUsers:     users,
			mockError:     nil,
			expectError:   false,
			expectedCount: 2,
		},
		{
			name:          "empty user list",
			limit:         10,
			offset:        0,
			mockUsers:     []*models.User{},
			mockError:     nil,
			expectError:   false,
			expectedCount: 0,
		},
		{
			name:          "with pagination",
			limit:         1,
			offset:        1,
			mockUsers:     []*models.User{users[1]},
			mockError:     nil,
			expectError:   false,
			expectedCount: 1,
		},
		{
			name:          "database error",
			limit:         10,
			offset:        0,
			mockUsers:     nil,
			mockError:     errors.New("database connection failed"),
			expectError:   true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &MockUserRepository{}
			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)

			service := NewUserServiceWithInterface(mockRepo, logger)

			// Setup expectations
			mockRepo.On("List", mock.Anything, tt.limit, tt.offset).Return(tt.mockUsers, tt.mockError).Once()

			// Execute
			result, err := service.ListUsers(context.Background(), tt.limit, tt.offset)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)
				if tt.expectedCount > 0 {
					assert.Equal(t, tt.mockUsers[0].ID, result[0].ID)
					assert.Equal(t, tt.mockUsers[0].Email, result[0].Email)
				}
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
		})
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}