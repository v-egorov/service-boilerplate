package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/v-egorov/service-boilerplate/services/user-service/internal/models"
)

// Helper function to create command tag with specific rows affected
func newCommandTag(rowsAffected int64) pgconn.CommandTag {
	// Use NewCommandTag with a string that represents the affected rows
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
				case *uuid.UUID:
					if v, ok := val.(uuid.UUID); ok {
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

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// Helper function to create time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}

func TestUserRepository_Create(t *testing.T) {
	userID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	tests := []struct {
		name        string
		user        *models.User
		mockRow     *MockRow
		expectedUser *models.User
		expectError bool
	}{
		{
			name: "successful user creation",
			user: &models.User{
				Email:     "test@example.com",
				PasswordHash: "hashedpassword",
				FirstName: "John",
				LastName:  "Doe",
			},
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					*dest[0].(*uuid.UUID) = userID
					*dest[1].(*time.Time) = createdAt
					*dest[2].(*time.Time) = updatedAt
					return nil
				},
			},
			expectedUser: &models.User{
				ID:          userID,
				Email:       "test@example.com",
				PasswordHash: "hashedpassword",
				FirstName:   "John",
				LastName:    "Doe",
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
			},
			expectError: false,
		},
		{
			name: "database error",
			user: &models.User{
				Email:     "test@example.com",
				PasswordHash: "hashedpassword",
				FirstName: "John",
				LastName:  "Doe",
			},
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					return errors.New("database connection failed")
				},
			},
			expectedUser: nil,
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
			repo := NewUserRepositoryWithInterface(mockDB, createTestLogger())

			// Execute
			result, err := repo.Create(context.Background(), tt.user)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedUser.ID, result.ID)
				assert.Equal(t, tt.expectedUser.Email, result.Email)
				assert.Equal(t, tt.expectedUser.FirstName, result.FirstName)
				assert.Equal(t, tt.expectedUser.LastName, result.LastName)
			}
		})
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	userID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	tests := []struct {
		name         string
		id           uuid.UUID
		mockRow      *MockRow
		expectedUser *models.User
		expectError  bool
	}{
		{
			name: "successful user retrieval",
			id:   userID,
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					*dest[0].(*uuid.UUID) = userID
					*dest[1].(*string) = "test@example.com"
					*dest[2].(*string) = "hashedpassword"
					*dest[3].(*string) = "John"
					*dest[4].(*string) = "Doe"
					*dest[5].(*time.Time) = createdAt
					*dest[6].(*time.Time) = updatedAt
					return nil
				},
			},
			expectedUser: &models.User{
				ID:          userID,
				Email:       "test@example.com",
				PasswordHash: "hashedpassword",
				FirstName:   "John",
				LastName:    "Doe",
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
			},
			expectError: false,
		},
		{
			name: "user not found",
			id:   userID,
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					return pgx.ErrNoRows
				},
			},
			expectedUser: nil,
			expectError:  true,
		},
		{
			name: "database error",
			id:   userID,
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					return errors.New("database connection failed")
				},
			},
			expectedUser: nil,
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
			repo := NewUserRepositoryWithInterface(mockDB, createTestLogger())

			// Execute
			result, err := repo.GetByID(context.Background(), tt.id)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedUser.ID, result.ID)
				assert.Equal(t, tt.expectedUser.Email, result.Email)
				assert.Equal(t, tt.expectedUser.FirstName, result.FirstName)
				assert.Equal(t, tt.expectedUser.LastName, result.LastName)
			}
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	userID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	tests := []struct {
		name         string
		id           uuid.UUID
		user         *models.User
		mockRow      *MockRow
		expectedUser *models.User
		expectError  bool
	}{
		{
			name: "successful user update",
			id:   userID,
			user: &models.User{
				Email:     "updated@example.com",
				PasswordHash: "newhashedpassword",
				FirstName: "Jane",
				LastName:  "Smith",
			},
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					*dest[0].(*uuid.UUID) = userID
					*dest[1].(*string) = "updated@example.com"
					*dest[2].(*string) = "newhashedpassword"
					*dest[3].(*string) = "Jane"
					*dest[4].(*string) = "Smith"
					*dest[5].(*time.Time) = createdAt
					*dest[6].(*time.Time) = updatedAt
					return nil
				},
			},
			expectedUser: &models.User{
				ID:          userID,
				Email:       "updated@example.com",
				PasswordHash: "newhashedpassword",
				FirstName:   "Jane",
				LastName:    "Smith",
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
			},
			expectError: false,
		},
		{
			name: "user not found",
			id:   userID,
			user: &models.User{
				Email:     "updated@example.com",
				PasswordHash: "newhashedpassword",
				FirstName: "Jane",
				LastName:  "Smith",
			},
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					return pgx.ErrNoRows
				},
			},
			expectedUser: nil,
			expectError:  true,
		},
		{
			name: "database error",
			id:   userID,
			user: &models.User{
				Email:     "updated@example.com",
				PasswordHash: "newhashedpassword",
				FirstName: "Jane",
				LastName:  "Smith",
			},
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					return errors.New("database connection failed")
				},
			},
			expectedUser: nil,
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
			repo := NewUserRepositoryWithInterface(mockDB, createTestLogger())

			// Execute
			result, err := repo.Update(context.Background(), tt.id, tt.user)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedUser.ID, result.ID)
				assert.Equal(t, tt.expectedUser.Email, result.Email)
				assert.Equal(t, tt.expectedUser.FirstName, result.FirstName)
				assert.Equal(t, tt.expectedUser.LastName, result.LastName)
			}
		})
	}
}

func TestUserRepository_Delete(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name        string
		id          uuid.UUID
		mockExec    func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
		expectError bool
	}{
		{
			name: "successful user deletion",
			id:   userID,
			mockExec: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
				return newCommandTag(1), nil
			},
			expectError: false,
		},
		{
			name: "user not found",
			id:   userID,
			mockExec: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
				return newCommandTag(0), nil
			},
			expectError: true,
		},
		{
			name: "database error",
			id:   userID,
			mockExec: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, errors.New("database connection failed")
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
			repo := NewUserRepositoryWithInterface(mockDB, createTestLogger())

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

func TestUserRepository_List(t *testing.T) {
	userID1 := uuid.New()
	userID2 := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	tests := []struct {
		name          string
		limit         int
		offset        int
		mockRows      *MockRows
		expectedUsers []*models.User
		expectError   bool
	}{
		{
			name:   "successful users listing",
			limit:  10,
			offset: 0,
			mockRows: &MockRows{
				ScanResults: [][]any{
					{userID1, "user1@example.com", "hash1", "John", "Doe", createdAt, updatedAt},
					{userID2, "user2@example.com", "hash2", "Jane", "Smith", createdAt, updatedAt},
				},
			},
			expectedUsers: []*models.User{
				{
					ID:          userID1,
					Email:       "user1@example.com",
					PasswordHash: "hash1",
					FirstName:   "John",
					LastName:    "Doe",
					CreatedAt:   createdAt,
					UpdatedAt:   updatedAt,
				},
				{
					ID:          userID2,
					Email:       "user2@example.com",
					PasswordHash: "hash2",
					FirstName:   "Jane",
					LastName:    "Smith",
					CreatedAt:   createdAt,
					UpdatedAt:   updatedAt,
				},
			},
			expectError: false,
		},
		{
			name:   "empty users list",
			limit:  10,
			offset: 0,
			mockRows: &MockRows{
				ScanResults: [][]any{},
			},
			expectedUsers: []*models.User{},
			expectError:   false,
		},
		{
			name:   "database error",
			limit:  10,
			offset: 0,
			mockRows: &MockRows{
				ScanResults: [][]any{},
				ErrFunc: func() error {
					return errors.New("database connection failed")
				},
			},
			expectedUsers: nil,
			expectError:   true,
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
			repo := NewUserRepositoryWithInterface(mockDB, createTestLogger())

			// Execute
			result, err := repo.List(context.Background(), tt.limit, tt.offset)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, len(tt.expectedUsers))
				for i, expected := range tt.expectedUsers {
					assert.Equal(t, expected.ID, result[i].ID)
					assert.Equal(t, expected.Email, result[i].Email)
					assert.Equal(t, expected.FirstName, result[i].FirstName)
					assert.Equal(t, expected.LastName, result[i].LastName)
				}
			}
		})
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	userID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	tests := []struct {
		name         string
		email        string
		mockRow      *MockRow
		expectedUser *models.User
		expectError  bool
	}{
		{
			name:  "successful user retrieval by email",
			email: "test@example.com",
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					*dest[0].(*uuid.UUID) = userID
					*dest[1].(*string) = "test@example.com"
					*dest[2].(*string) = "hashedpassword"
					*dest[3].(*string) = "John"
					*dest[4].(*string) = "Doe"
					*dest[5].(*time.Time) = createdAt
					*dest[6].(*time.Time) = updatedAt
					return nil
				},
			},
			expectedUser: &models.User{
				ID:          userID,
				Email:       "test@example.com",
				PasswordHash: "hashedpassword",
				FirstName:   "John",
				LastName:    "Doe",
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
			},
			expectError: false,
		},
		{
			name:  "user not found",
			email: "nonexistent@example.com",
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					return pgx.ErrNoRows
				},
			},
			expectedUser: nil,
			expectError:  true,
		},
		{
			name:  "database error",
			email: "test@example.com",
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					return errors.New("database connection failed")
				},
			},
			expectedUser: nil,
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
			repo := NewUserRepositoryWithInterface(mockDB, createTestLogger())

			// Execute
			result, err := repo.GetByEmail(context.Background(), tt.email)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedUser.ID, result.ID)
				assert.Equal(t, tt.expectedUser.Email, result.Email)
				assert.Equal(t, tt.expectedUser.FirstName, result.FirstName)
				assert.Equal(t, tt.expectedUser.LastName, result.LastName)
			}
		})
	}
}