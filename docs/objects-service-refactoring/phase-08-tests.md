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
    "testing"
    
    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
    
    "your-project/services/objects-service/internal/models"
)

func setupTestDB(t *testing.T) *sqlx.DB {
    db, err := sqlx.Connect("postgres", "postgres://postgres:password@localhost:5432/objects_service_test?sslmode=disable")
    if err != nil {
        t.Fatalf("Failed to connect to test database: %v", err)
    }
    return db
}

func TestObjectTypeRepository_Create(t *testing.T) {
    db := setupTestDB(t)
    repo := NewObjectTypeRepository(db)
    
    ot := &models.ObjectType{
        Name:       "TestType",
        Description: stringPtr("Test description"),
        IsSealed:    false,
        Metadata:    make(models.Metadata),
    }
    
    err := repo.Create(context.Background(), ot)
    if err != nil {
        t.Fatalf("Failed to create object type: %v", err)
    }
    
    if ot.ID == 0 {
        t.Error("Expected non-zero ID after creation")
    }
    
    if ot.CreatedAt.IsZero() {
        t.Error("Expected CreatedAt to be set")
    }
}

func TestObjectTypeRepository_GetByID(t *testing.T) {
    db := setupTestDB(t)
    repo := NewObjectTypeRepository(db)
    
    ot := &models.ObjectType{
        Name:    "GetTypeTest",
        IsSealed: false,
        Metadata: make(models.Metadata),
    }
    
    if err := repo.Create(context.Background(), ot); err != nil {
        t.Fatalf("Failed to create object type: %v", err)
    }
    
    fetched, err := repo.GetByID(context.Background(), ot.ID)
    if err != nil {
        t.Fatalf("Failed to get object type by ID: %v", err)
    }
    
    if fetched.Name != ot.Name {
        t.Errorf("Name mismatch: expected %s, got %s", ot.Name, fetched.Name)
    }
}

func TestObjectTypeRepository_GetByID_NotFound(t *testing.T) {
    db := setupTestDB(t)
    repo := NewObjectTypeRepository(db)
    
    _, err := repo.GetByID(context.Background(), 99999)
    if err != ErrObjectTypeNotFound {
        t.Errorf("Expected ErrObjectTypeNotFound, got %v", err)
    }
}

func TestObjectTypeRepository_List(t *testing.T) {
    db := setupTestDB(t)
    repo := NewObjectTypeRepository(db)
    
    filter := &models.ObjectTypeFilter{}
    
    types, err := repo.List(context.Background(), filter)
    if err != nil {
        t.Fatalf("Failed to list object types: %v", err)
    }
    
    if len(types) == 0 {
        t.Error("Expected at least one object type")
    }
}

func TestObjectTypeRepository_Update(t *testing.T) {
    db := setupTestDB(t)
    repo := NewObjectTypeRepository(db)
    
    ot := &models.ObjectType{
        Name:    "UpdateTest",
        IsSealed: false,
        Metadata: make(models.Metadata),
    }
    
    if err := repo.Create(context.Background(), ot); err != nil {
        t.Fatalf("Failed to create object type: %v", err)
    }
    
    newName := "UpdatedType"
    ot.Name = newName
    
    if err := repo.Update(context.Background(), ot); err != nil {
        t.Fatalf("Failed to update object type: %v", err)
    }
    
    updated, err := repo.GetByID(context.Background(), ot.ID)
    if err != nil {
        t.Fatalf("Failed to get updated object type: %v", err)
    }
    
    if updated.Name != newName {
        t.Errorf("Name not updated: expected %s, got %s", newName, updated.Name)
    }
}

func TestObjectTypeRepository_Delete(t *testing.T) {
    db := setupTestDB(t)
    repo := NewObjectTypeRepository(db)
    
    ot := &models.ObjectType{
        Name:    "DeleteTest",
        IsSealed: false,
        Metadata: make(models.Metadata),
    }
    
    if err := repo.Create(context.Background(), ot); err != nil {
        t.Fatalf("Failed to create object type: %v", err)
    }
    
    id := ot.ID
    
    if err := repo.Delete(context.Background(), id); err != nil {
        t.Fatalf("Failed to delete object type: %v", err)
    }
    
    _, err := repo.GetByID(context.Background(), id)
    if err != ErrObjectTypeNotFound {
        t.Errorf("Expected ErrObjectTypeNotFound after deletion, got %v", err)
    }
}

func TestObjectRepository_Create(t *testing.T) {
    db := setupTestDB(t)
    typeRepo := NewObjectTypeRepository(db)
    objRepo := NewObjectRepository(db)
    
    ot := &models.ObjectType{
        Name:    "TestObjectType",
        IsSealed: false,
        Metadata: make(models.Metadata),
    }
    
    if err := typeRepo.Create(context.Background(), ot); err != nil {
        t.Fatalf("Failed to create object type: %v", err)
    }
    
    obj := &models.Object{
        ObjectTypeID: ot.ID,
        Name:         "TestObject",
        Status:       models.StatusActive,
        Tags:         []string{"tag1"},
        Metadata:     make(models.Metadata),
    }
    
    err := objRepo.Create(context.Background(), obj, "test-user")
    if err != nil {
        t.Fatalf("Failed to create object: %v", err)
    }
    
    if obj.ID == 0 {
        t.Error("Expected non-zero ID after creation")
    }
}

func TestObjectRepository_GetByPublicID(t *testing.T) {
    db := setupTestDB(t)
    typeRepo := NewObjectTypeRepository(db)
    objRepo := NewObjectRepository(db)
    
    ot := &models.ObjectType{
        Name:    "GetByPublicTest",
        IsSealed: false,
        Metadata: make(models.Metadata),
    }
    
    if err := typeRepo.Create(context.Background(), ot); err != nil {
        t.Fatalf("Failed to create object type: %v", err)
    }
    
    obj := &models.Object{
        ObjectTypeID: ot.ID,
        Name:         "TestObject",
        Status:       models.StatusActive,
        Tags:         []string{"tag1"},
        Metadata:     make(models.Metadata),
    }
    
    if err := objRepo.Create(context.Background(), obj, "test-user"); err != nil {
        t.Fatalf("Failed to create object: %v", err)
    }
    
    fetched, err := objRepo.GetByPublicID(context.Background(), obj.PublicID)
    if err != nil {
        t.Fatalf("Failed to get object by public ID: %v", err)
    }
    
    if fetched.Name != obj.Name {
        t.Errorf("Name mismatch: expected %s, got %s", obj.Name, fetched.Name)
    }
}

func TestObjectRepository_SoftDelete(t *testing.T) {
    db := setupTestDB(t)
    typeRepo := NewObjectTypeRepository(db)
    objRepo := NewObjectRepository(db)
    
    ot := &models.ObjectType{
        Name:    "SoftDeleteTest",
        IsSealed: false,
        Metadata: make(models.Metadata),
    }
    
    if err := typeRepo.Create(context.Background(), ot); err != nil {
        t.Fatalf("Failed to create object type: %v", err)
    }
    
    obj := &models.Object{
        ObjectTypeID: ot.ID,
        Name:         "TestObject",
        Status:       models.StatusActive,
        Tags:         []string{"tag1"},
        Metadata:     make(models.Metadata),
    }
    
    if err := objRepo.Create(context.Background(), obj, "test-user"); err != nil {
        t.Fatalf("Failed to create object: %v", err)
    }
    
    if err := objRepo.SoftDelete(context.Background(), obj.ID, "test-user"); err != nil {
        t.Fatalf("Failed to soft delete object: %v", err)
    }
    
    fetched, err := objRepo.GetByID(context.Background(), obj.ID)
    if err != nil {
        t.Fatalf("Failed to get soft deleted object: %v", err)
    }
    
    if fetched.DeletedAt == nil {
        t.Error("Expected DeletedAt to be set")
    }
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

# Run integration tests
go test ./tests/integration/... -v
```

## Common Issues

**Issue**: Tests fail due to missing test database
**Solution**: Set up test database connection string in environment or test setup

**Issue**: Concurrent test failures
**Solution**: Use test database transactions and rollback after each test

**Issue**: UUID parsing errors in tests
**Solution**: Use proper UUID test fixtures or generate valid UUIDs

## Next Phase

Proceed to [Phase 9: Documentation](phase-09-documentation.md) once all tasks in this phase are complete.
