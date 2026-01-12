package services

import (
	"context"
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
)

// MockEntityRepository is a testify mock for EntityRepository
type MockEntityRepository struct {
	mock.Mock
}

func (m *MockEntityRepository) Create(ctx context.Context, entity *models.Entity) (*models.Entity, error) {
	args := m.Called(ctx, entity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Entity), args.Error(1)
}

func (m *MockEntityRepository) GetByID(ctx context.Context, id int64) (*models.Entity, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Entity), args.Error(1)
}

func (m *MockEntityRepository) Replace(ctx context.Context, id int64, entity *models.Entity) (*models.Entity, error) {
	args := m.Called(ctx, id, entity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Entity), args.Error(1)
}

func (m *MockEntityRepository) Update(ctx context.Context, id int64, updates map[string]interface{}) (*models.Entity, error) {
	args := m.Called(ctx, id, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Entity), args.Error(1)
}

func (m *MockEntityRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEntityRepository) List(ctx context.Context, limit, offset int) ([]*models.Entity, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.Entity), args.Error(1)
}

// Helper function to create a test logger
func createTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Only log fatal errors in tests
	return logger
}

func TestEntityService_Create(t *testing.T) {
	tests := []struct {
		name        string
		request     CreateEntityRequest
		mockSetup   func(*MockEntityRepository)
		expectError bool
	}{
		{
			name: "successful entity creation",
			request: CreateEntityRequest{
				Name:        "Test Entity",
				Description: "Test Description",
			},
			mockSetup: func(m *MockEntityRepository) {
				entity := &models.Entity{
					ID:          1,
					Name:        "Test Entity",
					Description: "Test Description",
				}
				m.On("Create", mock.Anything, mock.AnythingOfType("*models.Entity")).Return(entity, nil)
			},
			expectError: false,
		},
		{
			name: "validation error - empty name",
			request: CreateEntityRequest{
				Name:        "",
				Description: "Test Description",
			},
			mockSetup:   func(m *MockEntityRepository) {},
			expectError: true,
		},
		{
			name: "repository error",
			request: CreateEntityRequest{
				Name:        "Test Entity",
				Description: "Test Description",
			},
			mockSetup: func(m *MockEntityRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*models.Entity")).Return((*models.Entity)(nil), errors.New("database error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockEntityRepository{}
			tt.mockSetup(mockRepo)

			service := NewEntityService(mockRepo, createTestLogger())

			result, err := service.Create(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.request.Name, result.Name)
				assert.Equal(t, tt.request.Description, result.Description)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestEntityService_GetByID(t *testing.T) {
	tests := []struct {
		name        string
		id          int64
		mockSetup   func(*MockEntityRepository)
		expectError bool
	}{
		{
			name: "successful entity retrieval",
			id:   1,
			mockSetup: func(m *MockEntityRepository) {
				entity := &models.Entity{
					ID:          1,
					Name:        "Test Entity",
					Description: "Test Description",
				}
				m.On("GetByID", mock.Anything, int64(1)).Return(entity, nil)
			},
			expectError: false,
		},
		{
			name: "entity not found",
			id:   999,
			mockSetup: func(m *MockEntityRepository) {
				m.On("GetByID", mock.Anything, int64(999)).Return((*models.Entity)(nil), errors.New("entity not found"))
			},
			expectError: true,
		},
		{
			name:        "invalid id",
			id:          0,
			mockSetup:   func(m *MockEntityRepository) {},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockEntityRepository{}
			tt.mockSetup(mockRepo)

			service := NewEntityService(mockRepo, createTestLogger())

			result, err := service.GetByID(context.Background(), tt.id)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.id, result.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestEntityService_Replace(t *testing.T) {
	tests := []struct {
		name        string
		id          int64
		request     ReplaceEntityRequest
		mockSetup   func(*MockEntityRepository)
		expectError bool
	}{
		{
			name: "successful entity replacement",
			id:   1,
			request: ReplaceEntityRequest{
				Name:        "Updated Entity",
				Description: "Updated Description",
			},
			mockSetup: func(m *MockEntityRepository) {
				// Mock GetByID call
				entity := &models.Entity{ID: 1, Name: "Original", Description: "Original"}
				m.On("GetByID", mock.Anything, int64(1)).Return(entity, nil)

				// Mock Replace call
				updatedEntity := &models.Entity{
					ID:          1,
					Name:        "Updated Entity",
					Description: "Updated Description",
				}
				m.On("Replace", mock.Anything, int64(1), mock.AnythingOfType("*models.Entity")).Return(updatedEntity, nil)
			},
			expectError: false,
		},
		{
			name: "validation error - empty name",
			id:   1,
			request: ReplaceEntityRequest{
				Name:        "",
				Description: "Updated Description",
			},
			mockSetup: func(m *MockEntityRepository) {},
			expectError: true,
		},
		{
			name: "entity not found",
			id:   999,
			request: ReplaceEntityRequest{
				Name:        "Updated Entity",
				Description: "Updated Description",
			},
			mockSetup: func(m *MockEntityRepository) {
				m.On("GetByID", mock.Anything, int64(999)).Return((*models.Entity)(nil), errors.New("entity not found"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockEntityRepository{}
			tt.mockSetup(mockRepo)

			service := NewEntityService(mockRepo, createTestLogger())

			result, err := service.Replace(context.Background(), tt.id, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.request.Name, result.Name)
				assert.Equal(t, tt.request.Description, result.Description)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestEntityService_Update(t *testing.T) {
	tests := []struct {
		name        string
		id          int64
		request     UpdateEntityRequest
		mockSetup   func(*MockEntityRepository)
		expectError bool
	}{
		{
			name: "successful entity update",
			id:   1,
			request: UpdateEntityRequest{
				Name:        stringPtr("Updated Entity"),
				Description: stringPtr("Updated Description"),
			},
			mockSetup: func(m *MockEntityRepository) {
				updatedEntity := &models.Entity{
					ID:          1,
					Name:        "Updated Entity",
					Description: "Updated Description",
				}
				m.On("Update", mock.Anything, int64(1), mock.AnythingOfType("map[string]interface {}")).Return(updatedEntity, nil)
			},
			expectError: false,
		},
		{
			name: "validation error - empty name",
			id:   1,
			request: UpdateEntityRequest{
				Name: stringPtr(""),
			},
			mockSetup: func(m *MockEntityRepository) {},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockEntityRepository{}
			tt.mockSetup(mockRepo)

			service := NewEntityService(mockRepo, createTestLogger())

			result, err := service.Update(context.Background(), tt.id, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.request.Name != nil {
					assert.Equal(t, *tt.request.Name, result.Name)
				}
				if tt.request.Description != nil {
					assert.Equal(t, *tt.request.Description, result.Description)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestEntityService_Delete(t *testing.T) {
	tests := []struct {
		name        string
		id          int64
		mockSetup   func(*MockEntityRepository)
		expectError bool
	}{
		{
			name: "successful entity deletion",
			id:   1,
			mockSetup: func(m *MockEntityRepository) {
				m.On("Delete", mock.Anything, int64(1)).Return(nil)
			},
			expectError: false,
		},
		{
			name:        "invalid id",
			id:          0,
			mockSetup:   func(m *MockEntityRepository) {},
			expectError: true,
		},
		{
			name: "repository error",
			id:   1,
			mockSetup: func(m *MockEntityRepository) {
				m.On("Delete", mock.Anything, int64(1)).Return(errors.New("database error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockEntityRepository{}
			tt.mockSetup(mockRepo)

			service := NewEntityService(mockRepo, createTestLogger())

			err := service.Delete(context.Background(), tt.id)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestEntityService_List(t *testing.T) {
	tests := []struct {
		name          string
		limit         int
		offset        int
		mockSetup     func(*MockEntityRepository)
		expectedCount int
		expectError   bool
	}{
		{
			name:   "successful entities listing",
			limit:  10,
			offset: 0,
			mockSetup: func(m *MockEntityRepository) {
				entities := []*models.Entity{
					{ID: 1, Name: "Entity 1", Description: "Description 1"},
					{ID: 2, Name: "Entity 2", Description: "Description 2"},
				}
				m.On("List", mock.Anything, 10, 0).Return(entities, nil)
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:   "empty list",
			limit:  10,
			offset: 0,
			mockSetup: func(m *MockEntityRepository) {
				m.On("List", mock.Anything, 10, 0).Return([]*models.Entity{}, nil)
			},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:   "repository error",
			limit:  10,
			offset: 0,
			mockSetup: func(m *MockEntityRepository) {
				m.On("List", mock.Anything, 10, 0).Return([]*models.Entity(nil), errors.New("database error"))
			},
			expectedCount: 0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockEntityRepository{}
			tt.mockSetup(mockRepo)

			service := NewEntityService(mockRepo, createTestLogger())

			result, err := service.List(context.Background(), tt.limit, tt.offset)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}