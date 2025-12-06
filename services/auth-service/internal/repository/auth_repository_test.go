package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/models"
)

// DBInterface defines the database operations needed for testing
type DBInterface interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

// MockDBPool is a mock implementation of DBInterface for testing
type MockDBPool struct {
	QueryFunc    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRowFunc func(ctx context.Context, sql string, args ...any) pgx.Row
	ExecFunc     func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	BeginFunc    func(ctx context.Context) (pgx.Tx, error)
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

func (m *MockDBPool) Begin(ctx context.Context) (pgx.Tx, error) {
	if m.BeginFunc != nil {
		return m.BeginFunc(ctx)
	}
	return nil, nil
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
				case **string:
					if v, ok := val.(*string); ok {
						*d = v
					}
				case *time.Time:
					if v, ok := val.(time.Time); ok {
						*d = v
					}
				case *int:
					if v, ok := val.(int); ok {
						*d = v
					}
				case *bool:
					if v, ok := val.(bool); ok {
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

// MockTx is a mock implementation of pgx.Tx for testing
type MockTx struct {
	ExecFunc     func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	CommitFunc   func(ctx context.Context) error
	RollbackFunc func(ctx context.Context) error
}

func (m *MockTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if m.ExecFunc != nil {
		return m.ExecFunc(ctx, sql, args...)
	}
	return pgconn.CommandTag{}, nil
}

func (m *MockTx) Commit(ctx context.Context) error {
	if m.CommitFunc != nil {
		return m.CommitFunc(ctx)
	}
	return nil
}

func (m *MockTx) Rollback(ctx context.Context) error {
	if m.RollbackFunc != nil {
		return m.RollbackFunc(ctx)
	}
	return nil
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// Helper function to create time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}

func TestAuthRepository_CreateAuthToken(t *testing.T) {
	tests := []struct {
		name        string
		token       *models.AuthToken
		mockExec    func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
		expectError bool
	}{
		{
			name: "successful token creation",
			token: &models.AuthToken{
				ID:        uuid.New(),
				UserID:    uuid.New(),
				TokenHash: "hash123",
				TokenType: "access",
				ExpiresAt: time.Now().Add(time.Hour),
			},
			mockExec: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, nil
			},
			expectError: false,
		},
		{
			name: "database error",
			token: &models.AuthToken{
				ID:        uuid.New(),
				UserID:    uuid.New(),
				TokenHash: "hash123",
				TokenType: "access",
				ExpiresAt: time.Now().Add(time.Hour),
			},
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
			repo := NewAuthRepositoryWithInterface(mockDB)

			// Execute
			err := repo.CreateAuthToken(context.Background(), tt.token)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthRepository_GetAuthTokenByHash(t *testing.T) {
	tokenID := uuid.New()
	userID := uuid.New()
	tokenHash := "hash123"
	expiresAt := time.Now().Add(time.Hour)
	revokedAt := timePtr(time.Now().Add(-time.Hour))
	createdAt := time.Now().Add(-2 * time.Hour)
	updatedAt := time.Now().Add(-time.Hour)

	tests := []struct {
		name          string
		hash          string
		mockRow       *MockRow
		expectedToken *models.AuthToken
		expectError   bool
	}{
		{
			name: "successful token retrieval",
			hash: tokenHash,
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					*dest[0].(*uuid.UUID) = tokenID
					*dest[1].(*uuid.UUID) = userID
					*dest[2].(*string) = tokenHash
					*dest[3].(*string) = "access"
					*dest[4].(*time.Time) = expiresAt
					*dest[5].(**time.Time) = revokedAt
					*dest[6].(*time.Time) = createdAt
					*dest[7].(*time.Time) = updatedAt
					return nil
				},
			},
			expectedToken: &models.AuthToken{
				ID:        tokenID,
				UserID:    userID,
				TokenHash: tokenHash,
				TokenType: "access",
				ExpiresAt: expiresAt,
				RevokedAt: revokedAt,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			},
			expectError: false,
		},
		{
			name: "token not found",
			hash: "nonexistent",
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					return pgx.ErrNoRows
				},
			},
			expectedToken: nil,
			expectError:   true,
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
			repo := NewAuthRepositoryWithInterface(mockDB)

			// Execute
			result, err := repo.GetAuthTokenByHash(context.Background(), tt.hash)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedToken.ID, result.ID)
				assert.Equal(t, tt.expectedToken.UserID, result.UserID)
				assert.Equal(t, tt.expectedToken.TokenHash, result.TokenHash)
				assert.Equal(t, tt.expectedToken.TokenType, result.TokenType)
			}
		})
	}
}

func TestAuthRepository_RevokeAuthToken(t *testing.T) {
	tokenID := uuid.New()

	tests := []struct {
		name        string
		tokenID     uuid.UUID
		mockExec    func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
		expectError bool
	}{
		{
			name:    "successful token revocation",
			tokenID: tokenID,
			mockExec: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, nil
			},
			expectError: false,
		},
		{
			name:    "database error",
			tokenID: tokenID,
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
			repo := NewAuthRepositoryWithInterface(mockDB)

			// Execute
			err := repo.RevokeAuthToken(context.Background(), tt.tokenID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthRepository_RevokeUserTokens(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name        string
		userID      uuid.UUID
		mockExec    func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
		expectError bool
	}{
		{
			name:   "successful user tokens revocation",
			userID: userID,
			mockExec: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, nil
			},
			expectError: false,
		},
		{
			name:   "database error",
			userID: userID,
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
			repo := NewAuthRepositoryWithInterface(mockDB)

			// Execute
			err := repo.RevokeUserTokens(context.Background(), tt.userID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthRepository_CreateUserSession(t *testing.T) {
	sessionID := uuid.New()
	userID := uuid.New()
	sessionToken := "session_token_123"
	ipAddress := stringPtr("192.168.1.1")
	userAgent := stringPtr("Mozilla/5.0")
	expiresAt := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name        string
		session     *models.UserSession
		mockExec    func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
		expectError bool
	}{
		{
			name: "successful session creation",
			session: &models.UserSession{
				ID:           sessionID,
				UserID:       userID,
				SessionToken: sessionToken,
				IPAddress:    ipAddress,
				UserAgent:    userAgent,
				ExpiresAt:    expiresAt,
			},
			mockExec: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, nil
			},
			expectError: false,
		},
		{
			name: "database error",
			session: &models.UserSession{
				ID:           sessionID,
				UserID:       userID,
				SessionToken: sessionToken,
				IPAddress:    ipAddress,
				UserAgent:    userAgent,
				ExpiresAt:    expiresAt,
			},
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
			repo := NewAuthRepositoryWithInterface(mockDB)

			// Execute
			err := repo.CreateUserSession(context.Background(), tt.session)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthRepository_GetUserSession(t *testing.T) {
	sessionID := uuid.New()
	userID := uuid.New()
	sessionToken := "session_token_123"
	ipAddress := stringPtr("192.168.1.1")
	userAgent := stringPtr("Mozilla/5.0")
	expiresAt := time.Now().Add(24 * time.Hour)
	createdAt := time.Now().Add(-time.Hour)

	tests := []struct {
		name            string
		sessionToken    string
		mockRow         *MockRow
		expectedSession *models.UserSession
		expectError     bool
	}{
		{
			name:         "successful session retrieval",
			sessionToken: sessionToken,
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					*dest[0].(*uuid.UUID) = sessionID
					*dest[1].(*uuid.UUID) = userID
					*dest[2].(*string) = sessionToken
					*dest[3].(**string) = ipAddress
					*dest[4].(**string) = userAgent
					*dest[5].(*time.Time) = expiresAt
					*dest[6].(*time.Time) = createdAt
					return nil
				},
			},
			expectedSession: &models.UserSession{
				ID:           sessionID,
				UserID:       userID,
				SessionToken: sessionToken,
				IPAddress:    ipAddress,
				UserAgent:    userAgent,
				ExpiresAt:    expiresAt,
				CreatedAt:    createdAt,
			},
			expectError: false,
		},
		{
			name:         "session not found",
			sessionToken: "nonexistent",
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					return pgx.ErrNoRows
				},
			},
			expectedSession: nil,
			expectError:     true,
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
			repo := NewAuthRepositoryWithInterface(mockDB)

			// Execute
			result, err := repo.GetUserSession(context.Background(), tt.sessionToken)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedSession.ID, result.ID)
				assert.Equal(t, tt.expectedSession.UserID, result.UserID)
				assert.Equal(t, tt.expectedSession.SessionToken, result.SessionToken)
			}
		})
	}
}

func TestAuthRepository_DeleteUserSession(t *testing.T) {
	sessionID := uuid.New()

	tests := []struct {
		name        string
		sessionID   uuid.UUID
		mockExec    func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
		expectError bool
	}{
		{
			name:      "successful session deletion",
			sessionID: sessionID,
			mockExec: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, nil
			},
			expectError: false,
		},
		{
			name:      "database error",
			sessionID: sessionID,
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
			repo := NewAuthRepositoryWithInterface(mockDB)

			// Execute
			err := repo.DeleteUserSession(context.Background(), tt.sessionID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthRepository_DeleteExpiredSessions(t *testing.T) {
	tests := []struct {
		name        string
		mockExec    func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
		expectError bool
	}{
		{
			name: "successful expired sessions deletion",
			mockExec: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, nil
			},
			expectError: false,
		},
		{
			name: "database error",
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
			repo := NewAuthRepositoryWithInterface(mockDB)

			// Execute
			err := repo.DeleteExpiredSessions(context.Background())

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthRepository_GetUserRoles(t *testing.T) {
	userID := uuid.New()
	roleID1 := uuid.New()
	roleID2 := uuid.New()
	createdAt := time.Now()

	tests := []struct {
		name          string
		userID        uuid.UUID
		mockRows      *MockRows
		expectedRoles []models.Role
		expectError   bool
	}{
		{
			name:   "successful user roles retrieval",
			userID: userID,
			mockRows: &MockRows{
				ScanResults: [][]any{
					{roleID1, "admin", stringPtr("Administrator role"), createdAt},
					{roleID2, "user", stringPtr("Regular user"), createdAt},
				},
			},
			expectedRoles: []models.Role{
				{
					ID:          roleID1,
					Name:        "admin",
					Description: stringPtr("Administrator role"),
					CreatedAt:   createdAt,
				},
				{
					ID:          roleID2,
					Name:        "user",
					Description: stringPtr("Regular user"),
					CreatedAt:   createdAt,
				},
			},
			expectError: false,
		},
		{
			name:   "user with no roles",
			userID: userID,
			mockRows: &MockRows{
				ScanResults: [][]any{},
			},
			expectedRoles: []models.Role{},
			expectError:   false,
		},
		{
			name:   "database error",
			userID: userID,
			mockRows: &MockRows{
				ScanResults: [][]any{},
				ErrFunc: func() error {
					return errors.New("database connection failed")
				},
			},
			expectedRoles: nil,
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
			repo := NewAuthRepositoryWithInterface(mockDB)

			// Execute
			result, err := repo.GetUserRoles(context.Background(), tt.userID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, len(tt.expectedRoles))
				for i, expected := range tt.expectedRoles {
					assert.Equal(t, expected.ID, result[i].ID)
					assert.Equal(t, expected.Name, result[i].Name)
				}
			}
		})
	}
}

func TestAuthRepository_GetUserPermissions(t *testing.T) {
	userID := uuid.New()
	permissionID1 := uuid.New()
	permissionID2 := uuid.New()
	createdAt := time.Now()

	tests := []struct {
		name                string
		userID              uuid.UUID
		mockRows            *MockRows
		expectedPermissions []models.Permission
		expectError         bool
	}{
		{
			name:   "successful user permissions retrieval",
			userID: userID,
			mockRows: &MockRows{
				ScanResults: [][]any{
					{permissionID1, "read_users", "users", "read", createdAt},
					{permissionID2, "write_posts", "posts", "write", createdAt},
				},
			},
			expectedPermissions: []models.Permission{
				{
					ID:        permissionID1,
					Name:      "read_users",
					Resource:  "users",
					Action:    "read",
					CreatedAt: createdAt,
				},
				{
					ID:        permissionID2,
					Name:      "write_posts",
					Resource:  "posts",
					Action:    "write",
					CreatedAt: createdAt,
				},
			},
			expectError: false,
		},
		{
			name:   "user with no permissions",
			userID: userID,
			mockRows: &MockRows{
				ScanResults: [][]any{},
			},
			expectedPermissions: []models.Permission{},
			expectError:         false,
		},
		{
			name:   "database error",
			userID: userID,
			mockRows: &MockRows{
				ScanResults: [][]any{},
				ErrFunc: func() error {
					return errors.New("database connection failed")
				},
			},
			expectedPermissions: nil,
			expectError:         true,
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
			repo := NewAuthRepositoryWithInterface(mockDB)

			// Execute
			result, err := repo.GetUserPermissions(context.Background(), tt.userID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, len(tt.expectedPermissions))
				for i, expected := range tt.expectedPermissions {
					assert.Equal(t, expected.ID, result[i].ID)
					assert.Equal(t, expected.Name, result[i].Name)
					assert.Equal(t, expected.Resource, result[i].Resource)
					assert.Equal(t, expected.Action, result[i].Action)
				}
			}
		})
	}
}

func TestAuthRepository_CheckPermission(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name        string
		userID      uuid.UUID
		resource    string
		action      string
		mockRow     *MockRow
		expected    bool
		expectError bool
	}{
		{
			name:     "user has permission",
			userID:   userID,
			resource: "users",
			action:   "read",
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					*dest[0].(*bool) = true
					return nil
				},
			},
			expected:    true,
			expectError: false,
		},
		{
			name:     "user does not have permission",
			userID:   userID,
			resource: "posts",
			action:   "delete",
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					*dest[0].(*bool) = false
					return nil
				},
			},
			expected:    false,
			expectError: false,
		},
		{
			name:     "database error",
			userID:   userID,
			resource: "users",
			action:   "write",
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					return errors.New("database connection failed")
				},
			},
			expected:    false,
			expectError: true,
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
			repo := NewAuthRepositoryWithInterface(mockDB)

			// Execute
			result, err := repo.CheckPermission(context.Background(), tt.userID, tt.resource, tt.action)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.False(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestAuthRepository_CreateRole(t *testing.T) {
	roleID := uuid.New()
	createdAt := time.Now()

	tests := []struct {
		name         string
		role         *models.Role
		mockRow      *MockRow
		expectedRole *models.Role
		expectError  bool
	}{
		{
			name: "successful role creation",
			role: &models.Role{
				Name:        "admin",
				Description: stringPtr("Administrator role"),
			},
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					*dest[0].(*uuid.UUID) = roleID
					*dest[1].(*string) = "admin"
					*dest[2].(**string) = stringPtr("Administrator role")
					*dest[3].(*time.Time) = createdAt
					return nil
				},
			},
			expectedRole: &models.Role{
				ID:          roleID,
				Name:        "admin",
				Description: stringPtr("Administrator role"),
				CreatedAt:   createdAt,
			},
			expectError: false,
		},
		{
			name: "role creation error",
			role: &models.Role{
				Name:        "user",
				Description: nil,
			},
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					return errors.New("role already exists")
				},
			},
			expectedRole: nil,
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
			repo := NewAuthRepositoryWithInterface(mockDB)

			// Execute
			result, err := repo.CreateRole(context.Background(), tt.role)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedRole.ID, result.ID)
				assert.Equal(t, tt.expectedRole.Name, result.Name)
			}
		})
	}
}

func TestAuthRepository_ListRoles(t *testing.T) {
	roleID1 := uuid.New()
	roleID2 := uuid.New()
	createdAt := time.Now()

	tests := []struct {
		name          string
		mockRows      *MockRows
		expectedRoles []models.Role
		expectError   bool
	}{
		{
			name: "successful roles listing",
			mockRows: &MockRows{
				ScanResults: [][]any{
					{roleID1, "admin", stringPtr("Administrator role"), createdAt},
					{roleID2, "user", stringPtr("Regular user"), createdAt},
				},
			},
			expectedRoles: []models.Role{
				{
					ID:          roleID1,
					Name:        "admin",
					Description: stringPtr("Administrator role"),
					CreatedAt:   createdAt,
				},
				{
					ID:          roleID2,
					Name:        "user",
					Description: stringPtr("Regular user"),
					CreatedAt:   createdAt,
				},
			},
			expectError: false,
		},
		{
			name: "empty roles list",
			mockRows: &MockRows{
				ScanResults: [][]any{},
			},
			expectedRoles: []models.Role{},
			expectError:   false,
		},
		{
			name: "database error",
			mockRows: &MockRows{
				ScanResults: [][]any{},
				ErrFunc: func() error {
					return errors.New("database connection failed")
				},
			},
			expectedRoles: nil,
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
			repo := NewAuthRepositoryWithInterface(mockDB)

			// Execute
			result, err := repo.ListRoles(context.Background())

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, len(tt.expectedRoles))
				for i, expected := range tt.expectedRoles {
					assert.Equal(t, expected.ID, result[i].ID)
					assert.Equal(t, expected.Name, result[i].Name)
				}
			}
		})
	}
}

func TestAuthRepository_GetRole(t *testing.T) {
	roleID := uuid.New()
	createdAt := time.Now()

	tests := []struct {
		name         string
		roleID       uuid.UUID
		mockRow      *MockRow
		expectedRole *models.Role
		expectError  bool
	}{
		{
			name:   "successful role retrieval",
			roleID: roleID,
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					*dest[0].(*uuid.UUID) = roleID
					*dest[1].(*string) = "admin"
					*dest[2].(**string) = stringPtr("Administrator role")
					*dest[3].(*time.Time) = createdAt
					return nil
				},
			},
			expectedRole: &models.Role{
				ID:          roleID,
				Name:        "admin",
				Description: stringPtr("Administrator role"),
				CreatedAt:   createdAt,
			},
			expectError: false,
		},
		{
			name:   "role not found",
			roleID: roleID,
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					return pgx.ErrNoRows
				},
			},
			expectedRole: nil,
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
			repo := NewAuthRepositoryWithInterface(mockDB)

			// Execute
			result, err := repo.GetRole(context.Background(), tt.roleID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedRole.ID, result.ID)
				assert.Equal(t, tt.expectedRole.Name, result.Name)
			}
		})
	}
}

func TestAuthRepository_UpdateRole(t *testing.T) {
	roleID := uuid.New()
	createdAt := time.Now()

	tests := []struct {
		testName     string
		roleID       uuid.UUID
		roleName     string
		description  string
		mockRow      *MockRow
		expectedRole *models.Role
		expectError  bool
	}{
		{
			testName:    "successful role update",
			roleID:      roleID,
			roleName:    "super-admin",
			description: "Super Administrator",
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					*dest[0].(*uuid.UUID) = roleID
					*dest[1].(*string) = "super-admin"
					*dest[2].(**string) = stringPtr("Super Administrator")
					*dest[3].(*time.Time) = createdAt
					return nil
				},
			},
			expectedRole: &models.Role{
				ID:          roleID,
				Name:        "super-admin",
				Description: stringPtr("Super Administrator"),
				CreatedAt:   createdAt,
			},
			expectError: false,
		},
		{
			testName:    "role update error",
			roleID:      roleID,
			roleName:    "invalid",
			description: "Invalid role",
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					return errors.New("role not found")
				},
			},
			expectedRole: nil,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			// Setup mock
			mockDB := &MockDBPool{
				QueryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
					return tt.mockRow
				},
			}

			// Create repository
			repo := NewAuthRepositoryWithInterface(mockDB)

			// Execute
			result, err := repo.UpdateRole(context.Background(), tt.roleID, tt.roleName, tt.description)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedRole.ID, result.ID)
				assert.Equal(t, tt.expectedRole.Name, result.Name)
			}
		})
	}
}

func TestAuthRepository_DeleteRole(t *testing.T) {
	roleID := uuid.New()

	tests := []struct {
		name        string
		roleID      uuid.UUID
		mockExec    func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
		expectError bool
	}{
		{
			name:   "successful role deletion",
			roleID: roleID,
			mockExec: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, nil
			},
			expectError: false,
		},
		{
			name:   "database error",
			roleID: roleID,
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
			repo := NewAuthRepositoryWithInterface(mockDB)

			// Execute
			err := repo.DeleteRole(context.Background(), tt.roleID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthRepository_GetRoleByName(t *testing.T) {
	roleID := uuid.New()
	createdAt := time.Now()

	tests := []struct {
		name         string
		roleName     string
		mockRow      *MockRow
		expectedRole *models.Role
		expectError  bool
	}{
		{
			name:     "successful role retrieval by name",
			roleName: "admin",
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					*dest[0].(*uuid.UUID) = roleID
					*dest[1].(*string) = "admin"
					*dest[2].(**string) = stringPtr("Administrator role")
					*dest[3].(*time.Time) = createdAt
					return nil
				},
			},
			expectedRole: &models.Role{
				ID:          roleID,
				Name:        "admin",
				Description: stringPtr("Administrator role"),
				CreatedAt:   createdAt,
			},
			expectError: false,
		},
		{
			name:     "role not found",
			roleName: "nonexistent",
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					return pgx.ErrNoRows
				},
			},
			expectedRole: nil,
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
			repo := NewAuthRepositoryWithInterface(mockDB)

			// Execute
			result, err := repo.GetRoleByName(context.Background(), tt.roleName)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedRole.ID, result.ID)
				assert.Equal(t, tt.expectedRole.Name, result.Name)
			}
		})
	}
}

func TestAuthRepository_CreatePermission(t *testing.T) {
	permissionID := uuid.New()
	createdAt := time.Now()

	tests := []struct {
		name         string
		permission   *models.Permission
		mockRow      *MockRow
		expectedPerm *models.Permission
		expectError  bool
	}{
		{
			name: "successful permission creation",
			permission: &models.Permission{
				Name:     "read_users",
				Resource: "users",
				Action:   "read",
			},
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					*dest[0].(*uuid.UUID) = permissionID
					*dest[1].(*string) = "read_users"
					*dest[2].(*string) = "users"
					*dest[3].(*string) = "read"
					*dest[4].(*time.Time) = createdAt
					return nil
				},
			},
			expectedPerm: &models.Permission{
				ID:        permissionID,
				Name:      "read_users",
				Resource:  "users",
				Action:    "read",
				CreatedAt: createdAt,
			},
			expectError: false,
		},
		{
			name: "permission creation error",
			permission: &models.Permission{
				Name:     "invalid_perm",
				Resource: "invalid",
				Action:   "invalid",
			},
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					return errors.New("permission already exists")
				},
			},
			expectedPerm: nil,
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
			repo := NewAuthRepositoryWithInterface(mockDB)

			// Execute
			result, err := repo.CreatePermission(context.Background(), tt.permission)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedPerm.ID, result.ID)
				assert.Equal(t, tt.expectedPerm.Name, result.Name)
				assert.Equal(t, tt.expectedPerm.Resource, result.Resource)
				assert.Equal(t, tt.expectedPerm.Action, result.Action)
			}
		})
	}
}

func TestAuthRepository_ListPermissions(t *testing.T) {
	permissionID1 := uuid.New()
	permissionID2 := uuid.New()
	createdAt := time.Now()

	tests := []struct {
		name          string
		mockRows      *MockRows
		expectedPerms []models.Permission
		expectError   bool
	}{
		{
			name: "successful permissions listing",
			mockRows: &MockRows{
				ScanResults: [][]any{
					{permissionID1, "read_users", "users", "read", createdAt},
					{permissionID2, "write_posts", "posts", "write", createdAt},
				},
			},
			expectedPerms: []models.Permission{
				{
					ID:        permissionID1,
					Name:      "read_users",
					Resource:  "users",
					Action:    "read",
					CreatedAt: createdAt,
				},
				{
					ID:        permissionID2,
					Name:      "write_posts",
					Resource:  "posts",
					Action:    "write",
					CreatedAt: createdAt,
				},
			},
			expectError: false,
		},
		{
			name: "empty permissions list",
			mockRows: &MockRows{
				ScanResults: [][]any{},
			},
			expectedPerms: []models.Permission{},
			expectError:   false,
		},
		{
			name: "database error",
			mockRows: &MockRows{
				ScanResults: [][]any{},
				ErrFunc: func() error {
					return errors.New("database connection failed")
				},
			},
			expectedPerms: nil,
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
			repo := NewAuthRepositoryWithInterface(mockDB)

			// Execute
			result, err := repo.ListPermissions(context.Background())

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, len(tt.expectedPerms))
				for i, expected := range tt.expectedPerms {
					assert.Equal(t, expected.ID, result[i].ID)
					assert.Equal(t, expected.Name, result[i].Name)
					assert.Equal(t, expected.Resource, result[i].Resource)
					assert.Equal(t, expected.Action, result[i].Action)
				}
			}
		})
	}
}

func TestAuthRepository_GetPermission(t *testing.T) {
	permissionID := uuid.New()
	createdAt := time.Now()

	tests := []struct {
		name         string
		permissionID uuid.UUID
		mockRow      *MockRow
		expectedPerm *models.Permission
		expectError  bool
	}{
		{
			name:         "successful permission retrieval",
			permissionID: permissionID,
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					*dest[0].(*uuid.UUID) = permissionID
					*dest[1].(*string) = "read_users"
					*dest[2].(*string) = "users"
					*dest[3].(*string) = "read"
					*dest[4].(*time.Time) = createdAt
					return nil
				},
			},
			expectedPerm: &models.Permission{
				ID:        permissionID,
				Name:      "read_users",
				Resource:  "users",
				Action:    "read",
				CreatedAt: createdAt,
			},
			expectError: false,
		},
		{
			name:         "permission not found",
			permissionID: permissionID,
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					return pgx.ErrNoRows
				},
			},
			expectedPerm: nil,
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
			repo := NewAuthRepositoryWithInterface(mockDB)

			// Execute
			result, err := repo.GetPermission(context.Background(), tt.permissionID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedPerm.ID, result.ID)
				assert.Equal(t, tt.expectedPerm.Name, result.Name)
				assert.Equal(t, tt.expectedPerm.Resource, result.Resource)
				assert.Equal(t, tt.expectedPerm.Action, result.Action)
			}
		})
	}
}

func TestAuthRepository_UpdatePermission(t *testing.T) {
	permissionID := uuid.New()
	createdAt := time.Now()

	tests := []struct {
		name           string
		permissionID   uuid.UUID
		nameUpdate     string
		resourceUpdate string
		actionUpdate   string
		mockRow        *MockRow
		expectedPerm   *models.Permission
		expectError    bool
	}{
		{
			name:           "successful permission update",
			permissionID:   permissionID,
			nameUpdate:     "write_users",
			resourceUpdate: "users",
			actionUpdate:   "write",
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					*dest[0].(*uuid.UUID) = permissionID
					*dest[1].(*string) = "write_users"
					*dest[2].(*string) = "users"
					*dest[3].(*string) = "write"
					*dest[4].(*time.Time) = createdAt
					return nil
				},
			},
			expectedPerm: &models.Permission{
				ID:        permissionID,
				Name:      "write_users",
				Resource:  "users",
				Action:    "write",
				CreatedAt: createdAt,
			},
			expectError: false,
		},
		{
			name:           "permission update error",
			permissionID:   permissionID,
			nameUpdate:     "invalid_perm",
			resourceUpdate: "invalid",
			actionUpdate:   "invalid",
			mockRow: &MockRow{
				ScanFunc: func(dest ...any) error {
					return errors.New("permission not found")
				},
			},
			expectedPerm: nil,
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
			repo := NewAuthRepositoryWithInterface(mockDB)

			// Execute
			result, err := repo.UpdatePermission(context.Background(), tt.permissionID, tt.nameUpdate, tt.resourceUpdate, tt.actionUpdate)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedPerm.ID, result.ID)
				assert.Equal(t, tt.expectedPerm.Name, result.Name)
				assert.Equal(t, tt.expectedPerm.Resource, result.Resource)
				assert.Equal(t, tt.expectedPerm.Action, result.Action)
			}
		})
	}
}

func TestAuthRepository_DeletePermission(t *testing.T) {
	permissionID := uuid.New()

	tests := []struct {
		name         string
		permissionID uuid.UUID
		mockExec     func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
		expectError  bool
	}{
		{
			name:         "successful permission deletion",
			permissionID: permissionID,
			mockExec: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, nil
			},
			expectError: false,
		},
		{
			name:         "database error",
			permissionID: permissionID,
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
			repo := NewAuthRepositoryWithInterface(mockDB)

			// Execute
			err := repo.DeletePermission(context.Background(), tt.permissionID)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

