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
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/v-egorov/service-boilerplate/services/user-service/internal/models"
)

// MockUserService is a mock implementation of UserServiceInterface for testing
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.UserResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*models.UserResponse), args.Error(1)
}

func (m *MockUserService) GetUser(ctx context.Context, id uuid.UUID) (*models.UserResponse, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.UserResponse), args.Error(1)
}

func (m *MockUserService) GetUserByEmail(ctx context.Context, email string) (*models.UserResponse, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*models.UserResponse), args.Error(1)
}

func (m *MockUserService) GetUserWithPasswordByEmail(ctx context.Context, email string) (*models.UserLoginResponse, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*models.UserLoginResponse), args.Error(1)
}

func (m *MockUserService) ReplaceUser(ctx context.Context, id uuid.UUID, req *models.ReplaceUserRequest) (*models.UserResponse, error) {
	args := m.Called(ctx, id, req)
	return args.Get(0).(*models.UserResponse), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, id uuid.UUID, req *models.UpdateUserRequest) (*models.UserResponse, error) {
	args := m.Called(ctx, id, req)
	return args.Get(0).(*models.UserResponse), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserService) ListUsers(ctx context.Context, limit, offset int) ([]*models.UserResponse, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.UserResponse), args.Error(1)
}

// Helper functions
func createTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Only log fatal errors in tests
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

func TestUserHandler_handleServiceError(t *testing.T) {
	logger := createTestLogger()

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "validation error",
			err:            models.NewValidationError("email", "invalid email format"),
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "validation error on field 'email': invalid email format",
				"type":  "validation_error",
				"field": "email",
			},
		},
		{
			name:           "conflict error",
			err:            models.NewConflictError("user", "email", "test@example.com"),
			expectedStatus: http.StatusConflict,
			expectedBody: map[string]interface{}{
				"error":    "user with email 'test@example.com' already exists",
				"type":     "conflict_error",
				"resource": "user",
				"field":    "email",
				"value":    "test@example.com",
			},
		},
		{
			name:           "not found error",
			err:            models.NewNotFoundError("user", "id", "123"),
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error":    "user with id '123' not found",
				"type":     "not_found_error",
				"resource": "user",
				"field":    "id",
				"value":    "123",
			},
		},
		{
			name:           "internal error",
			err:            models.NewInternalError("database operation", assert.AnError),
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error":     "Internal server error",
				"type":      "internal_error",
				"operation": "database operation",
			},
		},
		{
			name:           "unknown error",
			err:            assert.AnError,
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "An unexpected error occurred",
				"type":  "unknown_error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &UserHandler{logger: logger}
			c, w := createTestGinContext("GET", "/", nil)

			handler.handleServiceError(c, tt.err, "test operation", "test-request-id")

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			for key, expectedValue := range tt.expectedBody {
				assert.Equal(t, expectedValue, response[key])
			}
		})
	}
}

func TestUserHandler_CreateUser(t *testing.T) {
	logger := createTestLogger()
	userID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockUserService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful user creation",
			requestBody: models.CreateUserRequest{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			mockSetup: func(m *MockUserService) {
				userResp := &models.UserResponse{
					ID:        userID,
					Email:     "test@example.com",
					FirstName: "John",
					LastName:  "Doe",
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
				}
				m.On("CreateUser", mock.Anything, mock.AnythingOfType("*models.CreateUserRequest")).Return(userResp, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"message": "User created successfully",
			},
		},
		{
			name:        "invalid request body",
			requestBody: "invalid json",
			mockSetup: func(m *MockUserService) {
				// No mock setup needed for validation failure
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Invalid request format",
				"type":  "validation_error",
			},
		},

		{
			name: "service internal error",
			requestBody: models.CreateUserRequest{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			mockSetup: func(m *MockUserService) {
				m.On("CreateUser", mock.Anything, mock.AnythingOfType("*models.CreateUserRequest")).Return((*models.UserResponse)(nil), models.NewInternalError("create user", assert.AnError))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error":     "Internal server error",
				"type":      "internal_error",
				"operation": "create user",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockUserService{}
			tt.mockSetup(mockService)

			handler := NewUserHandlerWithInterface(mockService, logger)
			c, w := createTestGinContext("POST", "/users", tt.requestBody)

			// Set authenticated user ID for audit logging
			c.Set("user_id", userID.String())

			handler.CreateUser(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			for key, expectedValue := range tt.expectedBody {
				assert.Equal(t, expectedValue, response[key])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetUser(t *testing.T) {
	logger := createTestLogger()
	userID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*MockUserService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:   "successful user retrieval",
			userID: userID.String(),
			mockSetup: func(m *MockUserService) {
				userResp := &models.UserResponse{
					ID:        userID,
					Email:     "test@example.com",
					FirstName: "John",
					LastName:  "Doe",
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
				}
				m.On("GetUser", mock.Anything, userID).Return(userResp, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid user ID",
			userID:         "invalid-uuid",
			mockSetup:      func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error":   "Invalid user ID format",
				"details": "User ID must be a valid UUID",
				"type":    "validation_error",
				"field":   "id",
			},
		},
		{
			name:   "user not found",
			userID: userID.String(),
			mockSetup: func(m *MockUserService) {
				m.On("GetUser", mock.Anything, userID).Return((*models.UserResponse)(nil), models.NewNotFoundError("user", "id", userID.String()))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error":    "user with id '" + userID.String() + "' not found",
				"type":     "not_found_error",
				"resource": "user",
				"field":    "id",
				"value":    userID.String(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockUserService{}
			tt.mockSetup(mockService)

			handler := NewUserHandlerWithInterface(mockService, logger)
			c, w := createTestGinContext("GET", "/users/"+tt.userID, nil)
			c.Params = gin.Params{{Key: "id", Value: tt.userID}}

			handler.GetUser(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				for key, expectedValue := range tt.expectedBody {
					assert.Equal(t, expectedValue, response[key])
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_ReplaceUser(t *testing.T) {
	logger := createTestLogger()
	userID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	tests := []struct {
		name           string
		userID         string
		requestBody    interface{}
		mockSetup      func(*MockUserService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:   "successful user replacement",
			userID: userID.String(),
			requestBody: models.ReplaceUserRequest{
				Email:     "updated@example.com",
				FirstName: "Jane",
				LastName:  "Smith",
			},
			mockSetup: func(m *MockUserService) {
				userResp := &models.UserResponse{
					ID:        userID,
					Email:     "updated@example.com",
					FirstName: "Jane",
					LastName:  "Smith",
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
				}
				m.On("ReplaceUser", mock.Anything, userID, mock.AnythingOfType("*models.ReplaceUserRequest")).Return(userResp, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "User replaced successfully",
			},
		},
		{
			name:           "invalid user ID",
			userID:         "invalid-uuid",
			requestBody:    models.ReplaceUserRequest{},
			mockSetup:      func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error":   "Invalid user ID format",
				"details": "User ID must be a valid UUID",
				"type":    "validation_error",
				"field":   "id",
			},
		},
		{
			name:   "invalid request body",
			userID: userID.String(),
			requestBody: "invalid json",
			mockSetup: func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Invalid request format",
				"type":  "validation_error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockUserService{}
			tt.mockSetup(mockService)

			handler := NewUserHandlerWithInterface(mockService, logger)
			c, w := createTestGinContext("PUT", "/users/"+tt.userID, tt.requestBody)
			c.Params = gin.Params{{Key: "id", Value: tt.userID}}

			handler.ReplaceUser(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				for key, expectedValue := range tt.expectedBody {
					assert.Equal(t, expectedValue, response[key])
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_UpdateUser(t *testing.T) {
	logger := createTestLogger()
	userID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	tests := []struct {
		name           string
		userID         string
		requestBody    interface{}
		mockSetup      func(*MockUserService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:   "successful user update",
			userID: userID.String(),
			requestBody: models.UpdateUserRequest{
				Email:     "updated@example.com",
				FirstName: "Jane",
			},
			mockSetup: func(m *MockUserService) {
				userResp := &models.UserResponse{
					ID:        userID,
					Email:     "updated@example.com",
					FirstName: "Jane",
					LastName:  "Doe",
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
				}
				m.On("UpdateUser", mock.Anything, userID, mock.AnythingOfType("*models.UpdateUserRequest")).Return(userResp, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "User updated successfully",
			},
		},
		{
			name:           "invalid user ID",
			userID:         "invalid-uuid",
			requestBody:    models.UpdateUserRequest{},
			mockSetup:      func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error":   "Invalid user ID format",
				"details": "User ID must be a valid UUID",
				"type":    "validation_error",
				"field":   "id",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockUserService{}
			tt.mockSetup(mockService)

			handler := NewUserHandlerWithInterface(mockService, logger)
			c, w := createTestGinContext("PATCH", "/users/"+tt.userID, tt.requestBody)
			c.Params = gin.Params{{Key: "id", Value: tt.userID}}

			handler.UpdateUser(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				for key, expectedValue := range tt.expectedBody {
					assert.Equal(t, expectedValue, response[key])
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_DeleteUser(t *testing.T) {
	logger := createTestLogger()
	userID := uuid.New()

	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*MockUserService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:   "successful user deletion",
			userID: userID.String(),
			mockSetup: func(m *MockUserService) {
				m.On("DeleteUser", mock.Anything, userID).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "invalid user ID",
			userID:         "invalid-uuid",
			mockSetup:      func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error":   "Invalid user ID format",
				"details": "User ID must be a valid UUID",
				"type":    "validation_error",
				"field":   "id",
			},
		},
		{
			name:   "user not found",
			userID: userID.String(),
			mockSetup: func(m *MockUserService) {
				m.On("DeleteUser", mock.Anything, userID).Return(models.NewNotFoundError("user", "id", userID.String()))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error":    "user with id '" + userID.String() + "' not found",
				"type":     "not_found_error",
				"resource": "user",
				"field":    "id",
				"value":    userID.String(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockUserService{}
			tt.mockSetup(mockService)

			handler := NewUserHandlerWithInterface(mockService, logger)
			c, w := createTestGinContext("DELETE", "/users/"+tt.userID, nil)
			c.Params = gin.Params{{Key: "id", Value: tt.userID}}

			handler.DeleteUser(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				for key, expectedValue := range tt.expectedBody {
					assert.Equal(t, expectedValue, response[key])
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_ListUsers(t *testing.T) {
	logger := createTestLogger()
	userID1 := uuid.New()
	userID2 := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(*MockUserService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:        "successful users listing",
			queryParams: "?limit=10&offset=0",
			mockSetup: func(m *MockUserService) {
				users := []*models.UserResponse{
					{
						ID:        userID1,
						Email:     "user1@example.com",
						FirstName: "John",
						LastName:  "Doe",
						CreatedAt: createdAt,
						UpdatedAt: updatedAt,
					},
					{
						ID:        userID2,
						Email:     "user2@example.com",
						FirstName: "Jane",
						LastName:  "Smith",
						CreatedAt: createdAt,
						UpdatedAt: updatedAt,
					},
				}
				m.On("ListUsers", mock.Anything, 10, 0).Return(users, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "empty users list",
			queryParams: "?limit=10&offset=0",
			mockSetup: func(m *MockUserService) {
				m.On("ListUsers", mock.Anything, 10, 0).Return([]*models.UserResponse{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "limit too high",
			queryParams: "?limit=200&offset=0",
			mockSetup: func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error":   "Limit too high",
				"details": "Maximum limit is 100",
				"type":    "validation_error",
				"field":   "limit",
			},
		},
		{
			name:        "default parameters",
			queryParams: "",
			mockSetup: func(m *MockUserService) {
				m.On("ListUsers", mock.Anything, 10, 0).Return([]*models.UserResponse{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockUserService{}
			tt.mockSetup(mockService)

			handler := NewUserHandlerWithInterface(mockService, logger)
			c, w := createTestGinContext("GET", "/users"+tt.queryParams, nil)

			handler.ListUsers(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				for key, expectedValue := range tt.expectedBody {
					assert.Equal(t, expectedValue, response[key])
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetUserByEmail(t *testing.T) {
	logger := createTestLogger()
	userID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	tests := []struct {
		name           string
		email          string
		mockSetup      func(*MockUserService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:  "successful user retrieval by email",
			email: "test@example.com",
			mockSetup: func(m *MockUserService) {
				userResp := &models.UserResponse{
					ID:        userID,
					Email:     "test@example.com",
					FirstName: "John",
					LastName:  "Doe",
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
				}
				m.On("GetUserByEmail", mock.Anything, "test@example.com").Return(userResp, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "empty email parameter",
			email:          "",
			mockSetup:      func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Email parameter is required",
				"type":  "validation_error",
				"field": "email",
			},
		},
		{
			name:  "user not found",
			email: "notfound@example.com",
			mockSetup: func(m *MockUserService) {
				m.On("GetUserByEmail", mock.Anything, "notfound@example.com").Return((*models.UserResponse)(nil), models.NewNotFoundError("user", "email", "notfound@example.com"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error":    "user with email 'notfound@example.com' not found",
				"type":     "not_found_error",
				"resource": "user",
				"field":    "email",
				"value":    "notfound@example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockUserService{}
			tt.mockSetup(mockService)

			handler := NewUserHandlerWithInterface(mockService, logger)
			c, w := createTestGinContext("GET", "/users/email/"+tt.email, nil)
			c.Params = gin.Params{{Key: "email", Value: tt.email}}

			handler.GetUserByEmail(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				for key, expectedValue := range tt.expectedBody {
					assert.Equal(t, expectedValue, response[key])
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetUserWithPasswordByEmail(t *testing.T) {
	logger := createTestLogger()
	userID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	tests := []struct {
		name           string
		email          string
		mockSetup      func(*MockUserService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:  "successful user with password retrieval",
			email: "test@example.com",
			mockSetup: func(m *MockUserService) {
				userResp := &models.UserResponse{
					ID:        userID,
					Email:     "test@example.com",
					FirstName: "John",
					LastName:  "Doe",
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
				}
				loginResp := &models.UserLoginResponse{
					User:         userResp,
					PasswordHash: "hashedpassword",
				}
				m.On("GetUserWithPasswordByEmail", mock.Anything, "test@example.com").Return(loginResp, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "empty email parameter",
			email:          "",
			mockSetup:      func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Email parameter is required",
				"type":  "validation_error",
				"field": "email",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockUserService{}
			tt.mockSetup(mockService)

			handler := NewUserHandlerWithInterface(mockService, logger)
			c, w := createTestGinContext("GET", "/users/login/"+tt.email, nil)
			c.Params = gin.Params{{Key: "email", Value: tt.email}}

			handler.GetUserWithPasswordByEmail(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				for key, expectedValue := range tt.expectedBody {
					assert.Equal(t, expectedValue, response[key])
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}