package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/v-egorov/service-boilerplate/common/logging"
	"github.com/v-egorov/service-boilerplate/templates/service-template/internal/models"
	"github.com/v-egorov/service-boilerplate/templates/service-template/internal/services"
)

// MockEntityService is a mock implementation of Service interface for testing
type MockEntityService struct {
	mock.Mock
}

func (m *MockEntityService) Create(ctx context.Context, req services.CreateEntityRequest) (*services.EntityResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*services.EntityResponse), args.Error(1)
}

func (m *MockEntityService) GetByID(ctx context.Context, id int64) (*services.EntityResponse, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*services.EntityResponse), args.Error(1)
}

func (m *MockEntityService) Replace(ctx context.Context, id int64, req services.ReplaceEntityRequest) (*services.EntityResponse, error) {
	args := m.Called(ctx, id, req)
	return args.Get(0).(*services.EntityResponse), args.Error(1)
}

func (m *MockEntityService) Update(ctx context.Context, id int64, req services.UpdateEntityRequest) (*services.EntityResponse, error) {
	args := m.Called(ctx, id, req)
	return args.Get(0).(*services.EntityResponse), args.Error(1)
}

func (m *MockEntityService) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEntityService) List(ctx context.Context, limit, offset int) ([]*services.EntityResponse, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*services.EntityResponse), args.Error(1)
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

// createTestLoggers creates loggers that don't output anything for testing
func createTestLoggers() (*logging.AuditLogger, *logging.StandardLogger) {
	testLogger := createTestLogger()
	auditLogger := logging.NewAuditLogger(testLogger, "test-service")
	standardLogger := logging.NewStandardLogger(testLogger, "test-service")
	return auditLogger, standardLogger
}

func TestEntityHandler_handleServiceError(t *testing.T) {
	logger := createTestLogger()

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "validation error",
			err:            models.NewValidationError("name", "name is required"),
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "validation error on field 'name': name is required",
				"type":  "validation_error",
				"field": "name",
			},
		},
		{
			name:           "not found error",
			err:            models.NewNotFoundError("entity", "id", "123"),
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error":    "entity with id '123' not found",
				"type":     "not_found_error",
				"resource": "entity",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &EntityHandler{logger: logger}
			c, w := createTestGinContext("GET", "/", nil)

			handler.handleServiceError(c, tt.err, "test operation")

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

func TestEntityHandler_CreateEntity(t *testing.T) {
	logger := createTestLogger()

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockEntityService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful entity creation",
			requestBody: services.CreateEntityRequest{
				Name:        "Test Entity",
				Description: "Test Description",
			},
			mockSetup: func(m *MockEntityService) {
				entityResp := &services.EntityResponse{
					ID:          1,
					Name:        "Test Entity",
					Description: "Test Description",
				}
				m.On("Create", mock.Anything, mock.AnythingOfType("models.CreateEntityRequest")).Return(entityResp, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"message": "Entity created successfully",
			},
		},
		{
			name:        "invalid request body",
			requestBody: "invalid json",
			mockSetup:   func(m *MockEntityService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Invalid request format",
				"type":  "validation_error",
			},
		},
		{
			name: "service validation error",
			requestBody: services.CreateEntityRequest{
				Name:        "",
				Description: "Test Description",
			},
			mockSetup: func(m *MockEntityService) {
				m.On("Create", mock.Anything, mock.AnythingOfType("models.CreateEntityRequest")).Return((*services.EntityResponse)(nil), models.NewValidationError("name", "name is required"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "validation error on field 'name': name is required",
				"type":  "validation_error",
				"field": "name",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockEntityService{}
			tt.mockSetup(mockService)

			testAuditLogger, testStandardLogger := createTestLoggers()
			handler := &EntityHandler{
				service:        mockService,
				logger:         logger,
				auditLogger:    testAuditLogger,
				standardLogger: testStandardLogger,
			}
			c, w := createTestGinContext("POST", "/entities", tt.requestBody)

			handler.CreateEntity(c)

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

func TestEntityHandler_GetEntity(t *testing.T) {
	logger := createTestLogger()

	tests := []struct {
		name           string
		entityID       string
		mockSetup      func(*MockEntityService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:     "successful entity retrieval",
			entityID: "1",
			mockSetup: func(m *MockEntityService) {
				entityResp := &services.EntityResponse{
					ID:          1,
					Name:        "Test Entity",
					Description: "Test Description",
				}
				m.On("GetByID", mock.Anything, int64(1)).Return(entityResp, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid entity ID",
			entityID:       "invalid",
			mockSetup:      func(m *MockEntityService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error":   "Invalid entity ID format",
				"details": "Entity ID must be a valid integer",
				"type":    "validation_error",
				"field":   "id",
			},
		},
		{
			name:     "entity not found",
			entityID: "999",
			mockSetup: func(m *MockEntityService) {
				m.On("GetByID", mock.Anything, int64(999)).Return((*services.EntityResponse)(nil), models.NewNotFoundError("entity", "id", "999"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error":    "entity with id '999' not found",
				"type":     "not_found_error",
				"resource": "entity",
				"field":    "id",
				"value":    "999",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockEntityService{}
			tt.mockSetup(mockService)

			testAuditLogger, testStandardLogger := createTestLoggers()
			handler := &EntityHandler{
				service:        mockService,
				logger:         logger,
				auditLogger:    testAuditLogger,
				standardLogger: testStandardLogger,
			}
			c, w := createTestGinContext("GET", "/entities/"+tt.entityID, nil)
			c.Params = gin.Params{{Key: "id", Value: tt.entityID}}

			handler.GetEntity(c)

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

func TestEntityHandler_ReplaceEntity(t *testing.T) {
	logger := createTestLogger()

	tests := []struct {
		name           string
		entityID       string
		requestBody    interface{}
		mockSetup      func(*MockEntityService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:     "successful entity replacement",
			entityID: "1",
			requestBody: services.ReplaceEntityRequest{
				Name:        "Updated Entity",
				Description: "Updated Description",
			},
			mockSetup: func(m *MockEntityService) {
				entityResp := &services.EntityResponse{
					ID:          1,
					Name:        "Updated Entity",
					Description: "Updated Description",
				}
				m.On("Replace", mock.Anything, int64(1), mock.AnythingOfType("models.ReplaceEntityRequest")).Return(entityResp, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "Entity replaced successfully",
			},
		},
		{
			name:           "invalid entity ID",
			entityID:       "invalid",
			requestBody:    services.ReplaceEntityRequest{},
			mockSetup:      func(m *MockEntityService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error":   "Invalid entity ID format",
				"details": "Entity ID must be a valid integer",
				"type":    "validation_error",
				"field":   "id",
			},
		},
		{
			name:        "invalid request body",
			entityID:    "1",
			requestBody: "invalid json",
			mockSetup:   func(m *MockEntityService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Invalid request format",
				"type":  "validation_error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockEntityService{}
			tt.mockSetup(mockService)

			testAuditLogger, testStandardLogger := createTestLoggers()
			handler := &EntityHandler{
				service:        mockService,
				logger:         logger,
				auditLogger:    testAuditLogger,
				standardLogger: testStandardLogger,
			}
			c, w := createTestGinContext("PUT", "/entities/"+tt.entityID, tt.requestBody)
			c.Params = gin.Params{{Key: "id", Value: tt.entityID}}

			handler.ReplaceEntity(c)

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

func TestEntityHandler_UpdateEntity(t *testing.T) {
	logger := createTestLogger()

	tests := []struct {
		name           string
		entityID       string
		requestBody    interface{}
		mockSetup      func(*MockEntityService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:     "successful entity update",
			entityID: "1",
			requestBody: services.UpdateEntityRequest{
				Name:        stringPtr("Updated Entity"),
				Description: stringPtr("Updated Description"),
			},
			mockSetup: func(m *MockEntityService) {
				entityResp := &services.EntityResponse{
					ID:          1,
					Name:        "Updated Entity",
					Description: "Updated Description",
				}
				m.On("Update", mock.Anything, int64(1), mock.AnythingOfType("models.UpdateEntityRequest")).Return(entityResp, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "Entity updated successfully",
			},
		},
		{
			name:           "invalid entity ID",
			entityID:       "invalid",
			requestBody:    services.UpdateEntityRequest{},
			mockSetup:      func(m *MockEntityService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error":   "Invalid entity ID format",
				"details": "Entity ID must be a valid integer",
				"type":    "validation_error",
				"field":   "id",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockEntityService{}
			tt.mockSetup(mockService)

			testAuditLogger, testStandardLogger := createTestLoggers()
			handler := &EntityHandler{
				service:        mockService,
				logger:         logger,
				auditLogger:    testAuditLogger,
				standardLogger: testStandardLogger,
			}
			c, w := createTestGinContext("PATCH", "/entities/"+tt.entityID, tt.requestBody)
			c.Params = gin.Params{{Key: "id", Value: tt.entityID}}

			handler.UpdateEntity(c)

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

func TestEntityHandler_DeleteEntity(t *testing.T) {
	logger := createTestLogger()

	tests := []struct {
		name           string
		entityID       string
		mockSetup      func(*MockEntityService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:     "successful entity deletion",
			entityID: "1",
			mockSetup: func(m *MockEntityService) {
				m.On("Delete", mock.Anything, int64(1)).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "invalid entity ID",
			entityID:       "invalid",
			mockSetup:      func(m *MockEntityService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error":   "Invalid entity ID format",
				"details": "Entity ID must be a valid integer",
				"type":    "validation_error",
				"field":   "id",
			},
		},
		{
			name:     "entity not found",
			entityID: "999",
			mockSetup: func(m *MockEntityService) {
				m.On("Delete", mock.Anything, int64(999)).Return(models.NewNotFoundError("entity", "id", "999"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error":    "entity with id '999' not found",
				"type":     "not_found_error",
				"resource": "entity",
				"field":    "id",
				"value":    "999",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockEntityService{}
			tt.mockSetup(mockService)

			testAuditLogger, testStandardLogger := createTestLoggers()
			handler := &EntityHandler{
				service:        mockService,
				logger:         logger,
				auditLogger:    testAuditLogger,
				standardLogger: testStandardLogger,
			}
			c, w := createTestGinContext("DELETE", "/entities/"+tt.entityID, nil)
			c.Params = gin.Params{{Key: "id", Value: tt.entityID}}

			handler.DeleteEntity(c)

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

func TestEntityHandler_ListEntities(t *testing.T) {
	logger := createTestLogger()

	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(*MockEntityService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:        "successful entities listing",
			queryParams: "?limit=10&offset=0",
			mockSetup: func(m *MockEntityService) {
				entities := []*services.EntityResponse{
					{
						ID:          1,
						Name:        "Entity 1",
						Description: "Description 1",
					},
					{
						ID:          2,
						Name:        "Entity 2",
						Description: "Description 2",
					},
				}
				m.On("List", mock.Anything, 10, 0).Return(entities, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "empty entities list",
			queryParams: "?limit=10&offset=0",
			mockSetup: func(m *MockEntityService) {
				m.On("List", mock.Anything, 10, 0).Return([]*services.EntityResponse{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "limit too high",
			queryParams: "?limit=200&offset=0",
			mockSetup:   func(m *MockEntityService) {},
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
			mockSetup: func(m *MockEntityService) {
				m.On("List", mock.Anything, 10, 0).Return([]*services.EntityResponse{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockEntityService{}
			tt.mockSetup(mockService)

			testAuditLogger, testStandardLogger := createTestLoggers()
			handler := &EntityHandler{
				service:        mockService,
				logger:         logger,
				auditLogger:    testAuditLogger,
				standardLogger: testStandardLogger,
			}
			c, w := createTestGinContext("GET", "/entities"+tt.queryParams, nil)

			handler.ListEntities(c)

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

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}