# Phase 8: Tests

**Estimated Time**: 4 hours
**Status**: ⬜ Not Started
**Dependencies**: Phase 5 (Handlers), Phase 7 (Test Data)

## Overview

Create comprehensive test suite including unit tests for models, repositories, services, and handlers, plus integration tests for API endpoints. Ensure test coverage for all major functionality.

## Tasks

### 8.1 Create Model Tests

**Files**:
- `internal/models/object_type_test.go`
- `internal/models/object_test.go`

**Steps**:
1. Test model struct initialization
2. Test JSON serialization/deserialization
3. Test validation
4. Test helper methods (IsActive, IsSoftDeleted)

```go
package models

import (
    "testing"
    "time"
    
    "github.com/google/uuid"
)

func TestObjectType_Initialization(t *testing.T) {
    ot := ObjectType{
        Name:  "TestType",
        IsSealed: false,
        Metadata: Metadata{},
    }
    
    if ot.Name != "TestType" {
        t.Errorf("Expected name 'TestType', got '%s'", ot.Name)
    }
    
    if ot.IsSealed {
        t.Error("Expected IsSealed to be false")
    }
}

func TestObject_Initialization(t *testing.T) {
    publicID := uuid.New()
    obj := Object{
        ID:           1,
        PublicID:     publicID,
        ObjectTypeID: 100,
        Name:         "TestObject",
        Status:       StatusActive,
        Tags:         []string{"tag1", "tag2"},
        Metadata:     Metadata{},
    }
    
    if obj.PublicID != publicID {
        t.Error("PublicID mismatch")
    }
    
    if obj.ObjectTypeID != 100 {
        t.Error("ObjectTypeID mismatch")
    }
    
    if !obj.IsActive() {
        t.Error("Object should be active")
    }
    
    if obj.IsSoftDeleted() {
        t.Error("Object should not be soft deleted")
    }
}

func TestObject_SoftDelete(t *testing.T) {
    now := time.Now()
    obj := Object{
        ID:        1,
        DeletedAt: &now,
        Status:    StatusDeleted,
    }
    
    if !obj.IsSoftDeleted() {
        t.Error("Object should be soft deleted")
    }
    
    if !obj.IsActive() {
        t.Error("Soft deleted object should not be active")
    }
}

func TestObject_InactiveStatus(t *testing.T) {
    obj := Object{
        ID:     1,
        Status: StatusInactive,
    }
    
    if obj.IsActive() {
        t.Error("Inactive object should not be active")
    }
    
    if obj.IsSoftDeleted() {
        t.Error("Inactive object should not be soft deleted")
    }
}

func TestMetadata_Scan(t *testing.T) {
    m := Metadata{}
    
    jsonStr := `{"key1": "value1", "key2": 123}`
    
    err := m.Scan([]byte(jsonStr))
    if err != nil {
        t.Fatalf("Failed to scan metadata: %v", err)
    }
    
    if m["key1"] != "value1" {
        t.Error("key1 value mismatch")
    }
    
    if m["key2"] != float64(123) {
        t.Error("key2 value mismatch")
    }
}

func TestMetadata_Value(t *testing.T) {
    m := Metadata{
        "key1": "value1",
        "key2": 123,
    }
    
    value, err := m.Value()
    if err != nil {
        t.Fatalf("Failed to get metadata value: %v", err)
    }
    
    if value == nil {
        t.Error("Expected non-nil value")
    }
}
```

---

### 8.2 Create Repository Tests

**Files**:
- `internal/repository/object_type_repository_test.go`
- `internal/repository/object_repository_test.go`

**Steps**:
1. Setup test database connection
2. Create test fixtures
3. Test CRUD operations
4. Test hierarchical queries
5. Test filtering and search
6. Test transactions

```go
package repository

import (
    "context"
    "errors"
    "fmt"
    "testing"
    "time"
    
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
    
    "github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
)

// Helper function to create command tag with specific rows affected
func newCommandTag(rowsAffected int64) pgconn.CommandTag {
    return pgconn.NewCommandTag(fmt.Sprintf("DELETE %d", rowsAffected))
}

// Helper function to create a test logger
func createTestLogger() *logrus.Logger {
    logger := logrus.New()
    logger.SetLevel(logrus.FatalLevel)
    return logger
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

// MockTx is a mock implementation of pgx.Tx for testing
type MockTx struct {
    QueryFunc    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
    QueryRowFunc func(ctx context.Context, sql string, args ...any) pgx.Row
    ExecFunc     func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
    CommitFunc   func(ctx context.Context) error
    RollbackFunc func(ctx context.Context) error
}

func (m *MockTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
    if m.QueryFunc != nil {
        return m.QueryFunc(ctx, sql, args...)
    }
    return nil, nil
}

func (m *MockTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
    if m.QueryRowFunc != nil {
        return m.QueryRowFunc(ctx, sql, args...)
    }
    return nil
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
                destVal := dest[i].(interface{})
                switch v := val.(type) {
                case int64:
                    if ptr, ok := destVal.(*int64); ok {
                        *ptr = v
                    }
                case string:
                    if ptr, ok := destVal.(*string); ok {
                        *ptr = v
                    }
                case bool:
                    if ptr, ok := destVal.(*bool); ok {
                        *ptr = v
                    }
                case time.Time:
                    if ptr, ok := destVal.(*time.Time); ok {
                        *ptr = v
                    }
                default:
                    return fmt.Errorf("unsupported type: %T", v)
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

func TestObjectTypeRepository_Create(t *testing.T) {
    mockDB := &MockDBPool{}
    logger := createTestLogger()
    repo := NewObjectTypeRepositoryWithInterface(mockDB, logger)
    
    ot := &models.ObjectType{
        Name:       "TestType",
        Description: stringPtr("Test description"),
        IsSealed:    false,
        Metadata:    make(models.Metadata),
    }
    
    var createdID int64
    var createdTime time.Time
    
    mockDB.QueryRowFunc = func(ctx context.Context, sql string, args ...any) pgx.Row {
        return &mockRow{
            scanFunc: func(dest ...any) error {
                createdID = 1
                createdTime = time.Now()
                dest[0].(*int64).(*createdID)
                dest[2].(*time.Time).(*createdTime)
                dest[3].(*time.Time).(*createdTime)
                return nil
            },
        }
    }
    
    err := repo.Create(context.Background(), ot)
    assert.NoError(t, err)
    assert.Equal(t, int64(1), ot.ID)
}

func TestObjectTypeRepository_GetByID(t *testing.T) {
    mockDB := &MockDBPool{}
    logger := createTestLogger()
    repo := NewObjectTypeRepositoryWithInterface(mockDB, logger)
    
    expectedOT := &models.ObjectType{
        ID:         1,
        Name:       "GetTypeTest",
        IsSealed:   false,
        Metadata:   make(models.Metadata),
        CreatedAt:  time.Now(),
        UpdatedAt:  time.Now(),
    }
    
    mockDB.QueryRowFunc = func(ctx context.Context, sql string, args ...any) pgx.Row {
        return &mockRow{
            scanFunc: func(dest ...any) error {
                dest[0].(*int64).(*expectedOT.ID)
                dest[1].(*string).(*expectedOT.Name)
                dest[6].(*time.Time).(*expectedOT.CreatedAt)
                dest[7].(*time.Time).(*expectedOT.UpdatedAt)
                return nil
            },
        }
    }
    
    fetched, err := repo.GetByID(context.Background(), 1)
    assert.NoError(t, err)
    assert.Equal(t, expectedOT.Name, fetched.Name)
}

func TestObjectTypeRepository_GetByID_NotFound(t *testing.T) {
    mockDB := &MockDBPool{}
    logger := createTestLogger()
    repo := NewObjectTypeRepositoryWithInterface(mockDB, logger)
    
    mockDB.QueryRowFunc = func(ctx context.Context, sql string, args ...any) pgx.Row {
        return &mockRow{
            scanFunc: func(dest ...any) error {
                return pgx.ErrNoRows
            },
        }
    }
    
    _, err := repo.GetByID(context.Background(), 99999)
    assert.Error(t, err)
    assert.Equal(t, ErrObjectTypeNotFound, err)
}

func TestObjectTypeRepository_List(t *testing.T) {
    mockDB := &MockDBPool{}
    logger := createTestLogger()
    repo := NewObjectTypeRepositoryWithInterface(mockDB, logger)
    
    mockRows := &MockRows{
        ScanResults: [][]any{
            {int64(1), "Type1", nil, nil, nil, false, make(models.Metadata), time.Now(), time.Now()},
            {int64(2), "Type2", nil, nil, nil, false, make(models.Metadata), time.Now(), time.Now()},
        },
    }
    
    mockDB.QueryFunc = func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
        return mockRows, nil
    }
    
    filter := &models.ObjectTypeFilter{}
    types, err := repo.List(context.Background(), filter)
    assert.NoError(t, err)
    assert.Len(t, types, 2)
}

func TestObjectTypeRepository_Update(t *testing.T) {
    mockDB := &MockDBPool{}
    logger := createTestLogger()
    repo := NewObjectTypeRepositoryWithInterface(mockDB, logger)
    
    ot := &models.ObjectType{
        ID:   1,
        Name:  "UpdatedType",
        Metadata: make(models.Metadata),
    }
    
    mockDB.QueryRowFunc = func(ctx context.Context, sql string, args ...any) pgx.Row {
        return &mockRow{
            scanFunc: func(dest ...any) error {
                dest[0].(*time.Time).(*time.Now())
                return nil
            },
        }
    }
    
    err := repo.Update(context.Background(), ot)
    assert.NoError(t, err)
}

func TestObjectTypeRepository_Delete(t *testing.T) {
    mockDB := &MockDBPool{}
    logger := createTestLogger()
    repo := NewObjectTypeRepositoryWithInterface(mockDB, logger)
    
    mockDB.ExecFunc = func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
        return newCommandTag(1), nil
    }
    
    err := repo.Delete(context.Background(), 1)
    assert.NoError(t, err)
}

// mockRow is a helper for mocking pgx.Row
type mockRow struct {
    scanFunc func(dest ...any) error
}

func (m *mockRow) Scan(dest ...any) error {
    if m.scanFunc != nil {
        return m.scanFunc(dest...)
    }
    return nil
}

func stringPtr(s string) *string {
    return &s
}
```

---

### 8.3 Create Handler Tests

**Files**:
- `internal/handlers/object_type_handler_test.go`
- `internal/handlers/object_handler_test.go`

**Steps**:
1. Setup test HTTP server
2. Test all endpoints
3. Test request validation
4. Test error handling
5. Test authentication/authorization

```go
package handlers

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/gin-gonic/gin"
    
    "your-project/services/objects-service/internal/models"
)

func init() {
    gin.SetMode(gin.TestMode)
}

func TestObjectTypeHandler_Create(t *testing.T) {
    router := setupTestRouter()
    
    reqBody := models.CreateObjectTypeRequest{
        Name:        stringPtr("TestType"),
        Description: stringPtr("Test description"),
    }
    
    body, _ := json.Marshal(reqBody)
    
    req, _ := http.NewRequest("POST", "/api/v1/object-types", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    if w.Code != http.StatusCreated {
        t.Errorf("Expected status 201, got %d", w.Code)
    }
    
    var response models.ObjectType
    if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
        t.Fatalf("Failed to unmarshal response: %v", err)
    }
    
    if response.Name != *reqBody.Name {
        t.Errorf("Name mismatch: expected %s, got %s", *reqBody.Name, response.Name)
    }
}

func TestObjectTypeHandler_GetByID(t *testing.T) {
    router := setupTestRouter()
    
    req, _ := http.NewRequest("GET", "/api/v1/object-types/1", nil)
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
}

func TestObjectTypeHandler_GetByID_InvalidID(t *testing.T) {
    router := setupTestRouter()
    
    req, _ := http.NewRequest("GET", "/api/v1/object-types/invalid", nil)
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    if w.Code != http.StatusBadRequest {
        t.Errorf("Expected status 400, got %d", w.Code)
    }
}

func TestObjectHandler_Create(t *testing.T) {
    router := setupTestRouter()
    
    reqBody := models.CreateObjectRequest{
        ObjectTypeID: 1,
        Name:         "TestObject",
        Description:  stringPtr("Test description"),
    }
    
    body, _ := json.Marshal(reqBody)
    
    req, _ := http.NewRequest("POST", "/api/v1/objects", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    if w.Code != http.StatusCreated {
        t.Errorf("Expected status 201, got %d", w.Code)
    }
    
    var response models.Object
    if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
        t.Fatalf("Failed to unmarshal response: %v", err)
    }
    
    if response.Name != reqBody.Name {
        t.Errorf("Name mismatch: expected %s, got %s", reqBody.Name, response.Name)
    }
}

func TestObjectHandler_List(t *testing.T) {
    router := setupTestRouter()
    
    req, _ := http.NewRequest("GET", "/api/v1/objects?page=1&page_size=10", nil)
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
    
    var response models.ObjectListResponse
    if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
        t.Fatalf("Failed to unmarshal response: %v", err)
    }
    
    if response.Page != 1 {
        t.Errorf("Page mismatch: expected 1, got %d", response.Page)
    }
}

func TestObjectHandler_GetByPublicID(t *testing.T) {
    router := setupTestRouter()
    
    req, _ := http.NewRequest("GET", "/api/v1/objects/some-uuid-here", nil)
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    if w.Code != http.StatusOK || w.Code == http.StatusNotFound {
        t.Logf("Status: %d", w.Code)
    }
}

func stringPtr(s string) *string {
    return &s
}

func setupTestRouter() *gin.Engine {
    router := gin.New()
    return router
}
```

---

### 8.4 Create Integration Tests

**File**: `tests/integration/api_integration_test.go`

**Steps**:
1. Start test server
2. Test full request/response cycles
3. Test authentication flow
4. Test error scenarios

```go
package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/gin-gonic/gin"
    
    "your-project/services/objects-service/internal/models"
)

func TestAPI_Integration(t *testing.T) {
    router := setupIntegrationRouter()
    
    t.Run("Create and retrieve object type", func(t *testing.T) {
        createReq := models.CreateObjectTypeRequest{
            Name:        stringPtr("IntegrationTestType"),
            Description: stringPtr("Integration test type"),
        }
        
        body, _ := json.Marshal(createReq)
        req, _ := http.NewRequest("POST", "/api/v1/object-types", bytes.NewBuffer(body))
        req.Header.Set("Content-Type", "application/json")
        
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        if w.Code != http.StatusCreated {
            t.Fatalf("Expected status 201, got %d", w.Code)
        }
        
        var createdType models.ObjectType
        if err := json.Unmarshal(w.Body.Bytes(), &createdType); err != nil {
            t.Fatalf("Failed to unmarshal response: %v", err)
        }
        
        getReq, _ := http.NewRequest("GET", "/api/v1/object-types/"+string(rune(createdType.ID)), nil)
        w2 := httptest.NewRecorder()
        router.ServeHTTP(w2, getReq)
        
        if w2.Code != http.StatusOK {
            t.Fatalf("Expected status 200, got %d", w2.Code)
        }
    })
    
    t.Run("Create and retrieve object", func(t *testing.T) {
        createReq := models.CreateObjectRequest{
            ObjectTypeID: 1,
            Name:         "IntegrationTestObject",
            Description:  stringPtr("Integration test object"),
        }
        
        body, _ := json.Marshal(createReq)
        req, _ := http.NewRequest("POST", "/api/v1/objects", bytes.NewBuffer(body))
        req.Header.Set("Content-Type", "application/json")
        
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        if w.Code != http.StatusCreated {
            t.Fatalf("Expected status 201, got %d", w.Code)
        }
        
        var createdObj models.Object
        if err := json.Unmarshal(w.Body.Bytes(), &createdObj); err != nil {
            t.Fatalf("Failed to unmarshal response: %v", err)
        }
        
        getReq, _ := http.NewRequest("GET", "/api/v1/objects/"+createdObj.PublicID.String(), nil)
        w2 := httptest.NewRecorder()
        router.ServeHTTP(w2, getReq)
        
        if w2.Code != http.StatusOK {
            t.Fatalf("Expected status 200, got %d", w2.Code)
        }
    })
}

func setupIntegrationRouter() *gin.Engine {
    router := gin.New()
    return router
}
```

---

## Checklist

- [ ] Create `internal/models/object_type_test.go`
- [ ] Create `internal/models/object_test.go`
- [ ] Create `internal/repository/object_type_repository_test.go`
- [ ] Create `internal/repository/object_repository_test.go`
- [ ] Create `internal/handlers/object_type_handler_test.go`
- [ ] Create `internal/handlers/object_handler_test.go`
- [ ] Create `tests/integration/api_integration_test.go`
- [ ] Run all tests: `go test ./...`
- [ ] Verify test coverage: `go test -cover ./...`
- [ ] Fix failing tests
- [ ] Update progress.md

## Testing

```bash
# Run all tests
cd services/objects-service
go test ./... -v

# Run tests with coverage
go test ./... -cover

# Run tests with coverage profile
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific package tests
go test ./internal/models/... -v
go test ./internal/repository/... -v
go test ./internal/services/... -v
go test ./internal/handlers/... -v

# Run tests with testify/assert
go test ./internal/repository/... -v -run TestObjectTypeRepository

# Run integration tests
go test ./tests/integration/... -v
```

## Common Issues

**Issue**: Tests fail due to missing test database
**Solution**: Set up test database connection string in environment or test setup

**Issue**: Concurrent test failures
**Solution**: Use test database transactions and rollback after each test

**Issue**: pgx.ErrNoRows not recognized
**Solution**: Import `github.com/jackc/pgx/v5` and use `errors.Is(err, pgx.ErrNoRows)`

**Issue**: Mock rows not scanning correctly
**Solution**: Ensure mockRow implements pgx.Row interface correctly with Scan method

**Issue**: Command tag not returning rows affected
**Solution**: Use `pgconn.NewCommandTag(fmt.Sprintf("DELETE %d", rowsAffected))` to create proper command tags

**Issue**: UUID parsing errors in tests
**Solution**: Use `github.com/google/uuid` package with `uuid.New()` for valid UUIDs in tests

## Next Phase

Proceed to [Phase 9: Documentation](phase-09-documentation.md) once all tasks in this phase are complete.
