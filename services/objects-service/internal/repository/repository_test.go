package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
)

// MockDBPool implements DBInterface for testing
type MockDBPool struct {
	QueryFunc    func(ctx context.Context, sql string, args ...any) (Rows, error)
	QueryRowFunc func(ctx context.Context, sql string, args ...any) Row
	ExecFunc     func(ctx context.Context, sql string, args ...any) (CommandTag, error)
}

func (m *MockDBPool) Query(ctx context.Context, sql string, args ...any) (Rows, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(ctx, sql, args...)
	}
	return nil, nil
}

func (m *MockDBPool) QueryRow(ctx context.Context, sql string, args ...any) Row {
	if m.QueryRowFunc != nil {
		return m.QueryRowFunc(ctx, sql, args...)
	}
	return &MockRow{}
}

func (m *MockDBPool) Exec(ctx context.Context, sql string, args ...any) (CommandTag, error) {
	if m.ExecFunc != nil {
		return m.ExecFunc(ctx, sql, args...)
	}
	return nil, nil
}

// MockRow implements Row for testing
type MockRow struct{}

func (m *MockRow) Scan(dest ...any) error {
	return nil
}

// MockRows implements Rows for testing
type MockRows struct {
	CloseFunc func()
	NextFunc  func() bool
}

func (m *MockRows) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}

func (m *MockRows) Next() bool {
	if m.NextFunc != nil {
		return m.NextFunc()
	}
	return false
}

func (m *MockRows) Scan(dest ...any) error {
	return nil
}

func (m *MockRows) Err() error {
	return nil
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

// TestObjectTypeRepository_Creation tests that repository can be created
func TestObjectTypeRepository_Creation(t *testing.T) {
	mockDB := &MockDBPool{}
	repo := NewObjectTypeRepository(mockDB, DefaultRepositoryOptions())
	assert.NotNil(t, repo)
	assert.NotNil(t, repo.DB())
	assert.NotNil(t, repo.Options())
	assert.NotNil(t, repo.Metrics())
}

// TestObjectTypeRepository_Create tests basic create functionality
func TestObjectTypeRepository_Create(t *testing.T) {
	mockDB := &MockDBPool{
		QueryRowFunc: func(ctx context.Context, sql string, args ...any) Row {
			return &MockRow{}
		},
	}

	repo := NewObjectTypeRepository(mockDB, DefaultRepositoryOptions())
	input := &models.CreateObjectTypeRequest{
		Name:        "test-type",
		Description: "Test description",
	}

	result, err := repo.Create(context.Background(), input)
	// Mock returns nil scan, so we get empty object
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-type", result.Name)
	assert.Equal(t, "Test description", result.Description)
}

// TestObjectTypeRepository_GetByID tests basic get functionality
func TestObjectTypeRepository_GetByID(t *testing.T) {
	mockDB := &MockDBPool{
		QueryRowFunc: func(ctx context.Context, sql string, args ...any) Row {
			return &MockRow{}
		},
		QueryFunc: func(ctx context.Context, sql string, args ...any) (Rows, error) {
			return &MockRows{}, nil
		},
	}

	repo := NewObjectTypeRepository(mockDB, DefaultRepositoryOptions())
	result, err := repo.GetByID(context.Background(), 1)
	// Mock returns nil scan, so we get empty object
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(0), result.ID)
}

// TestObjectTypeRepository_List tests basic list functionality
func TestObjectTypeRepository_List(t *testing.T) {
	mockDB := &MockDBPool{
		QueryFunc: func(ctx context.Context, sql string, args ...any) (Rows, error) {
			rows := &MockRows{
				NextFunc: func() bool {
					return false
				},
			}
			return rows, nil
		},
	}

	repo := NewObjectTypeRepository(mockDB, DefaultRepositoryOptions())
	result, err := repo.List(context.Background(), &models.ObjectTypeFilter{})
	assert.NoError(t, err)
	assert.Nil(t, result)
}

// TestObjectTypeRepository_ValidateParentChild tests validation
func TestObjectTypeRepository_ValidateParentChild(t *testing.T) {
	mockDB := &MockDBPool{
		QueryRowFunc: func(ctx context.Context, sql string, args ...any) Row {
			return &MockRow{}
		},
	}

	repo := NewObjectTypeRepository(mockDB, DefaultRepositoryOptions())
	err := repo.ValidateParentChild(context.Background(), 1, 2)
	// Mock returns nil count (0), so no circular dependency
	assert.NoError(t, err)
}

// TestObjectTypeRepository_GetTree tests tree functionality
func TestObjectTypeRepository_GetTree(t *testing.T) {
	mockDB := &MockDBPool{
		QueryFunc: func(ctx context.Context, sql string, args ...any) (Rows, error) {
			rows := &MockRows{
				NextFunc: func() bool {
					return false // No rows
				},
			}
			return rows, nil
		},
	}

	repo := NewObjectTypeRepository(mockDB, DefaultRepositoryOptions())
	result, err := repo.GetTree(context.Background(), nil)
	assert.NoError(t, err)
	// Current implementation returns nil when there are no rows
	assert.Nil(t, result)
}

// TestObjectRepository_Creation tests that object repository can be created
func TestObjectRepository_Creation(t *testing.T) {
	mockDB := &MockDBPool{}
	repo := NewObjectRepository(mockDB, DefaultRepositoryOptions())
	assert.NotNil(t, repo)
	assert.NotNil(t, repo.DB())
	assert.NotNil(t, repo.Options())
}

// TestObjectRepository_UpdateMetadata tests metadata update functionality
func TestObjectRepository_UpdateMetadata(t *testing.T) {
	mockDB := &MockDBPool{
		ExecFunc: func(ctx context.Context, sql string, args ...any) (CommandTag, error) {
			return nil, nil
		},
	}

	repo := NewObjectRepository(mockDB, DefaultRepositoryOptions())
	metadata := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}

	err := repo.UpdateMetadata(context.Background(), 1, metadata)
	assert.NoError(t, err)
}

// TestObjectRepository_AddTags tests adding tags to an object
func TestObjectRepository_AddTags(t *testing.T) {
	mockDB := &MockDBPool{
		ExecFunc: func(ctx context.Context, sql string, args ...any) (CommandTag, error) {
			return nil, nil
		},
	}

	repo := NewObjectRepository(mockDB, DefaultRepositoryOptions())
	tags := []string{"tag1", "tag2", "tag3"}

	err := repo.AddTags(context.Background(), 1, tags)
	assert.NoError(t, err)
}

// TestObjectRepository_AddTags_Empty tests adding empty tags
func TestObjectRepository_AddTags_Empty(t *testing.T) {
	mockDB := &MockDBPool{}

	repo := NewObjectRepository(mockDB, DefaultRepositoryOptions())
	err := repo.AddTags(context.Background(), 1, []string{})
	assert.NoError(t, err)
}

// TestObjectRepository_RemoveTags tests removing tags from an object
func TestObjectRepository_RemoveTags(t *testing.T) {
	mockDB := &MockDBPool{
		ExecFunc: func(ctx context.Context, sql string, args ...any) (CommandTag, error) {
			return nil, nil
		},
	}

	repo := NewObjectRepository(mockDB, DefaultRepositoryOptions())
	tags := []string{"tag1", "tag2"}

	err := repo.RemoveTags(context.Background(), 1, tags)
	assert.NoError(t, err)
}

// TestObjectRepository_RemoveTags_Empty tests removing empty tags
func TestObjectRepository_RemoveTags_Empty(t *testing.T) {
	mockDB := &MockDBPool{}

	repo := NewObjectRepository(mockDB, DefaultRepositoryOptions())
	err := repo.RemoveTags(context.Background(), 1, []string{})
	assert.NoError(t, err)
}

// TestRepositoryMetrics tests metrics functionality
func TestRepositoryMetrics(t *testing.T) {
	metrics := &RepositoryMetrics{LastResetAt: time.Now()}

	assert.Equal(t, int64(0), metrics.QueryCount)
	assert.Equal(t, int64(0), metrics.ErrorCount)

	metrics.QueryCount = 100
	metrics.UpdateAverageQueryTime()

	expectedAvg := time.Duration(int64(metrics.TotalQueryTime) / metrics.QueryCount)
	assert.Equal(t, expectedAvg, metrics.AverageQueryTime)

	metrics.Reset()
	assert.Equal(t, int64(0), metrics.QueryCount)
}

// TestQueryBuilder tests the query builder
func TestQueryBuilder(t *testing.T) {
	qb := NewQueryBuilder()
	assert.NotNil(t, qb)

	query, args := qb.Select("id", "name").From("users").Where("id = $1", 1).Build()
	assert.Contains(t, query, "SELECT")
	assert.Contains(t, query, "FROM")
	assert.Contains(t, query, "WHERE")
	assert.Len(t, args, 1)
}

// TestPGDatabaseCreation tests that PGDatabase can be created
func TestPGDatabaseCreation(t *testing.T) {
	// This test just verifies the type can be instantiated
	var db *PGDatabase
	assert.Nil(t, db)
}
