package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/v-egorov/service-boilerplate/templates/service-template/internal/models"
)

// Helper function to create command tag with specific rows affected
func newCommandTag(rowsAffected int64) pgconn.CommandTag {
	return pgconn.NewCommandTag(fmt.Sprintf("DELETE %d", rowsAffected))
}

// Helper function to create a test logger
func createTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Only log fatal errors in tests
	return logger
}

// MockDBPool is a mock implementation of DBInterface for testing
type MockDBPool struct {
	QueryFunc    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRowFunc func(ctx context.Context, sql string, args ...any) pgx.Row
	ExecFunc     func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func (m *MockDBPool) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(ctx, sql, args...)
	}
	return nil, nil
}

func (m *MockDBPool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if m.QueryRowFunc != nil {
		return m.QueryRowFunc(ctx, sql, args...)
	}
	return nil
}

func (m *MockDBPool) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if m.ExecFunc != nil {
		return m.ExecFunc(ctx, sql, args...)
	}
	return pgconn.CommandTag{}, nil
}

// MockRows is a mock implementation of pgx.Rows for testing
type MockRows struct {
	CloseFunc   func()
	NextFunc    func() bool
	ScanFunc    func(dest ...any) error
	ErrFunc     func() error
	NextCalled  int
	ScanCalled  int
	ScanResults [][]any
	ScanIndex   int
}

func (m *MockRows) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}

func (m *MockRows) Next() bool {
	m.NextCalled++
	if m.NextFunc != nil {
		return m.NextFunc()
	}
	if m.ScanIndex < len(m.ScanResults) {
		m.ScanIndex++
		return true
	}
	return false
}

func (m *MockRows) Scan(dest ...any) error {
	m.ScanCalled++
	if m.ScanFunc != nil {
		return m.ScanFunc(dest...)
	}
	if m.ScanIndex-1 < len(m.ScanResults) {
		values := m.ScanResults[m.ScanIndex-1]
		if len(values) != len(dest) {
			return pgx.ErrNoRows
		}
		for i, val := range values {
			if dest[i] != nil {
				switch d := dest[i].(type) {
				case *int64:
					if v, ok := val.(int64); ok {
						*d = v
					}
				case *string:
					if v, ok := val.(string); ok {
						*d = v
					}
				case *time.Time:
					if v, ok := val.(time.Time); ok {
						*d = v
					}
				}
			}
		}
		return nil
	}
	return pgx.ErrNoRows
}

func (m *MockRows) Err() error {
	if m.ErrFunc != nil {
		return m.ErrFunc()
	}
	return nil
}

func (m *MockRows) CommandTag() pgconn.CommandTag {
	return pgconn.CommandTag{}
}

func (m *MockRows) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}

func (m *MockRows) RawValues() [][]byte {
	return nil
}

func (m *MockRows) Values() ([]any, error) {
	return nil, nil
}

func (m *MockRows) Conn() *pgx.Conn {
	return nil
}

// MockRow is a mock implementation of pgx.Row for testing
type MockRow struct {
	ScanFunc func(dest ...any) error
}

func (m *MockRow) Scan(dest ...any) error {
	if m.ScanFunc != nil {
		return m.ScanFunc(dest...)
	}
	return pgx.ErrNoRows
}

func TestEntityRepository_Create(t *testing.T) {
	entityID := int64(1)
	createdAt := time.Now()
	updatedAt := time.Now()

	tests := []struct {
		name        string
		entity      *models.Entity
		mockRow     *MockRow
		expectedEntity *models.Entity
		expectError bool
	}{
		{
			name: "successful entity creation",
			entity: &models.Entity{
				Name:        "Test Entity",
				Description: "Test Description",
			},
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					*dest[0].(*int64) = entityID
					return nil
				},
			},
			expectedEntity: &models.Entity{
				ID:          entityID,
				Name:        "Test Entity",
				Description: "Test Description",
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
			},
			expectError: false,
		},
		{
			name: "database error",
			entity: &models.Entity{
				Name:        "Test Entity",
				Description: "Test Description",
			},
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					return errors.New("database connection failed")
				},
			},
			expectedEntity: nil,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockDB := &MockDBPool{
				QueryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
					return tt.mockRow
				},
			}

			// Create repository
			repo := NewEntityRepositoryWithInterface(mockDB, createTestLogger())

			// Execute
			result, err := repo.Create(context.Background(), tt.entity)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedEntity.ID, result.ID)
				assert.Equal(t, tt.expectedEntity.Name, result.Name)
				assert.Equal(t, tt.expectedEntity.Description, result.Description)
			}
		})
	}
}

func TestEntityRepository_GetByID(t *testing.T) {
	entityID := int64(1)
	createdAt := time.Now()
	updatedAt := time.Now()

	tests := []struct {
		name         string
		id           int64
		mockRow      *MockRow
		expectedEntity *models.Entity
		expectError  bool
	}{
		{
			name: "successful entity retrieval",
			id:   entityID,
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					*dest[0].(*int64) = entityID
					*dest[1].(*string) = "Test Entity"
					*dest[2].(*string) = "Test Description"
					*dest[3].(*time.Time) = createdAt
					*dest[4].(*time.Time) = updatedAt
					return nil
				},
			},
			expectedEntity: &models.Entity{
				ID:          entityID,
				Name:        "Test Entity",
				Description: "Test Description",
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
			},
			expectError: false,
		},
		{
			name: "entity not found",
			id:   entityID,
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					return pgx.ErrNoRows
				},
			},
			expectedEntity: nil,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockDB := &MockDBPool{
				QueryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
					return tt.mockRow
				},
			}

			// Create repository
			repo := NewEntityRepositoryWithInterface(mockDB, createTestLogger())

			// Execute
			result, err := repo.GetByID(context.Background(), tt.id)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedEntity.ID, result.ID)
				assert.Equal(t, tt.expectedEntity.Name, result.Name)
				assert.Equal(t, tt.expectedEntity.Description, result.Description)
			}
		})
	}
}

func TestEntityRepository_Update(t *testing.T) {
	entityID := int64(1)
	createdAt := time.Now()
	updatedAt := time.Now()

	tests := []struct {
		name         string
		id           int64
		updates      map[string]interface{}
		mockRow      *MockRow
		expectedEntity *models.Entity
		expectError  bool
	}{
		{
			name: "successful entity update",
			id:   entityID,
			updates: map[string]interface{}{
				"name":        "Updated Entity",
				"description": "Updated Description",
			},
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					*dest[0].(*int64) = entityID
					*dest[1].(*string) = "Updated Entity"
					*dest[2].(*string) = "Updated Description"
					*dest[3].(*time.Time) = createdAt
					*dest[4].(*time.Time) = updatedAt
					return nil
				},
			},
			expectedEntity: &models.Entity{
				ID:          entityID,
				Name:        "Updated Entity",
				Description: "Updated Description",
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockDB := &MockDBPool{
				QueryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
					return tt.mockRow
				},
			}

			// Create repository
			repo := NewEntityRepositoryWithInterface(mockDB, createTestLogger())

			// Execute
			result, err := repo.Update(context.Background(), tt.id, tt.updates)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedEntity.ID, result.ID)
				assert.Equal(t, tt.expectedEntity.Name, result.Name)
				assert.Equal(t, tt.expectedEntity.Description, result.Description)
			}
		})
	}
}

func TestEntityRepository_Delete(t *testing.T) {
	entityID := int64(1)

	tests := []struct {
		name        string
		id          int64
		mockExec    func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
		expectError bool
	}{
		{
			name: "successful entity deletion",
			id:   entityID,
			mockExec: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
				return newCommandTag(1), nil
			},
			expectError: false,
		},
		{
			name: "entity not found",
			id:   entityID,
			mockExec: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
				return newCommandTag(0), nil
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockDB := &MockDBPool{
				ExecFunc: tt.mockExec,
			}

			// Create repository
			repo := NewEntityRepositoryWithInterface(mockDB, createTestLogger())

			// Execute
			err := repo.Delete(context.Background(), tt.id)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEntityRepository_List(t *testing.T) {
	entityID1 := int64(1)
	entityID2 := int64(2)
	createdAt := time.Now()
	updatedAt := time.Now()

	tests := []struct {
		name          string
		limit         int
		offset        int
		mockRows      *MockRows
		expectedEntities []*models.Entity
		expectError   bool
	}{
		{
			name:   "successful entities listing",
			limit:  10,
			offset: 0,
			mockRows: &MockRows{
				ScanResults: [][]any{
					{entityID1, "Entity 1", "Description 1", createdAt, updatedAt},
					{entityID2, "Entity 2", "Description 2", createdAt, updatedAt},
				},
			},
			expectedEntities: []*models.Entity{
				{
					ID:          entityID1,
					Name:        "Entity 1",
					Description: "Description 1",
					CreatedAt:   createdAt,
					UpdatedAt:   updatedAt,
				},
				{
					ID:          entityID2,
					Name:        "Entity 2",
					Description: "Description 2",
					CreatedAt:   createdAt,
					UpdatedAt:   updatedAt,
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockDB := &MockDBPool{
				QueryFunc: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
					return tt.mockRows, nil
				},
			}

			// Create repository
			repo := NewEntityRepositoryWithInterface(mockDB, createTestLogger())

			// Execute
			result, err := repo.List(context.Background(), tt.limit, tt.offset)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, len(tt.expectedEntities))
				for i, expected := range tt.expectedEntities {
					assert.Equal(t, expected.ID, result[i].ID)
					assert.Equal(t, expected.Name, result[i].Name)
					assert.Equal(t, expected.Description, result[i].Description)
				}
			}
		})
	}
}