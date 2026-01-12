# Phase 2: Models Layer

**Estimated Time**: 2.5 hours
**Status**: ⬜ Not Started
**Dependencies**: Phase 1 (Migrations)

## Overview

Create new model files for Object Types and Objects, including DTOs for API requests. Delete old Entity model files.

## Tasks

### 2.1 Create Object Type Model

**File**: `internal/models/object_type.go`

**Steps**:
1. Create ObjectType struct with all database fields
2. Add TableName() method
3. Add JSON tags for API serialization
4. Add validation tags

```go
package models

import (
    "database/sql/driver"
    "encoding/json"
    "time"
)

// Metadata type for JSONB storage
type Metadata map[string]interface{}

func (m *Metadata) Scan(value interface{}) error {
    if value == nil {
        *m = make(Metadata)
        return nil
    }
    bytes, ok := value.([]byte)
    if !ok {
        return nil
    }
    return json.Unmarshal(bytes, m)
}

func (m Metadata) Value() (driver.Value, error) {
    if m == nil {
        return nil, nil
    }
    return json.Marshal(m)
}

type ObjectType struct {
    ID                 int64     `json:"id" db:"id"`
    Name               string    `json:"name" db:"name" validate:"required,min=1,max=255"`
    ParentTypeID       *int64    `json:"parent_type_id,omitempty" db:"parent_type_id"`
    ConcreteTableName  *string   `json:"concrete_table_name,omitempty" db:"concrete_table_name"`
    Description        *string   `json:"description,omitempty" db:"description"`
    IsSealed           bool      `json:"is_sealed" db:"is_sealed"`
    Metadata           Metadata  `json:"metadata" db:"metadata"`
    CreatedAt          time.Time `json:"created_at" db:"created_at"`
    UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
    
    // Relationships (populated when needed)
    ParentType   *ObjectType `json:"parent_type,omitempty"`
    Children     []ObjectType `json:"children,omitempty"`
    Objects      []Object     `json:"objects,omitempty"`
}

func (ObjectType) TableName() string {
    return "object_types"
}
```

---

### 2.2 Create Object Model

**File**: `internal/models/object.go`

**Steps**:
1. Create Object struct with all database fields
2. Add TableName() method
3. Add JSON tags
4. Add validation tags

```go
package models

import (
    "github.com/google/uuid"
    "time"
)

const (
    StatusActive   = "active"
    StatusInactive = "inactive"
    StatusArchived = "archived"
    StatusDeleted  = "deleted"
    StatusPending  = "pending"
)

type Object struct {
    ID             int64              `json:"id" db:"id"`
    PublicID       uuid.UUID          `json:"public_id" db:"public_id"`
    ObjectTypeID   int64              `json:"object_type_id" db:"object_type_id" validate:"required"`
    ParentObjectID *int64             `json:"parent_object_id,omitempty" db:"parent_object_id"`
    Name           string             `json:"name" db:"name" validate:"required,min=1,max=255"`
    Description    *string            `json:"description,omitempty" db:"description"`
    CreatedAt      time.Time          `json:"created_at" db:"created_at"`
    UpdatedAt      time.Time          `json:"updated_at" db:"updated_at"`
    DeletedAt      *time.Time         `json:"deleted_at,omitempty" db:"deleted_at"`
    Version        int64              `json:"version" db:"version"`
    CreatedBy      string             `json:"created_by" db:"created_by"`
    UpdatedBy      string             `json:"updated_by" db:"updated_by"`
    Metadata       Metadata           `json:"metadata" db:"metadata"`
    Status         string             `json:"status" db:"status" validate:"oneof=active inactive archived deleted pending"`
    Tags           []string           `json:"tags" db:"tags"`
    
    // Relationships
    ObjectType     *ObjectType `json:"object_type,omitempty"`
    ParentObject   *Object     `json:"parent_object,omitempty"`
    Children       []Object    `json:"children,omitempty"`
}

func (Object) TableName() string {
    return "objects"
}

// IsSoftDeleted returns true if the object has been soft deleted
func (o *Object) IsSoftDeleted() bool {
    return o.DeletedAt != nil
}

// IsActive returns true if object status is active and not deleted
func (o *Object) IsActive() bool {
    return o.Status == StatusActive && !o.IsSoftDeleted()
}
```

---

### 2.3 Create Object Type Request DTOs

**File**: `internal/models/object_type_request.go`

**Steps**:
1. Create CreateObjectTypeRequest
2. Create UpdateObjectTypeRequest
3. Add validation tags

```go
package models

type CreateObjectTypeRequest struct {
    Name               string   `json:"name" validate:"required,min=1,max=255"`
    ParentTypeID       *int64   `json:"parent_type_id,omitempty" validate:"omitempty,min=1"`
    ConcreteTableName  *string  `json:"concrete_table_name,omitempty" validate:"omitempty,min=1,max=255"`
    Description        *string  `json:"description,omitempty"`
    IsSealed           *bool    `json:"is_sealed,omitempty"`
    Metadata           Metadata `json:"metadata,omitempty"`
}

type UpdateObjectTypeRequest struct {
    Name              *string  `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
    ConcreteTableName *string  `json:"concrete_table_name,omitempty" validate:"omitempty,min=1,max=255"`
    Description       *string  `json:"description,omitempty"`
    IsSealed          *bool    `json:"is_sealed,omitempty"`
    Metadata          *Metadata `json:"metadata,omitempty"`
}

type ObjectTypeFilter struct {
    ParentTypeID *int64  `form:"parent_type_id"`
    IsSealed     *bool   `form:"is_sealed"`
    Search       *string `form:"search"`
    IncludeTree  *bool   `form:"include_tree"`
}

type ObjectTypeListResponse struct {
    ObjectTypes []ObjectType `json:"object_types"`
    Total       int64        `json:"total"`
    Page        int          `json:"page"`
    PageSize    int          `json:"page_size"`
}
```

---

### 2.4 Create Object Request DTOs

**File**: `internal/models/object_request.go`

**Steps**:
1. Create CreateObjectRequest
2. Create UpdateObjectRequest
3. Create ObjectFilter for search
4. Add validation tags

```go
package models

import "github.com/google/uuid"

type CreateObjectRequest struct {
    ObjectTypeID   int64     `json:"object_type_id" validate:"required,min=1"`
    ParentObjectID *int64    `json:"parent_object_id,omitempty" validate:"omitempty,min=1"`
    Name           string    `json:"name" validate:"required,min=1,max=255"`
    Description    *string   `json:"description,omitempty"`
    Metadata       Metadata  `json:"metadata,omitempty"`
    Status         *string  `json:"status,omitempty" validate:"omitempty,oneof=active inactive archived deleted pending"`
    Tags           *[]string `json:"tags,omitempty" validate:"omitempty,dive,min=1,max=100"`
}

type UpdateObjectRequest struct {
    Name        *string   `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
    Description *string   `json:"description,omitempty"`
    Metadata    *Metadata `json:"metadata,omitempty"`
    Status      *string   `json:"status,omitempty" validate:"omitempty,oneof=active inactive archived deleted pending"`
    Tags        *[]string `json:"tags,omitempty" validate:"omitempty,dive,min=1,max=100"`
    Version     *int64    `json:"version,omitempty" validate:"required,min=0"` // For optimistic locking
}

type ObjectFilter struct {
    ObjectTypeID    *int64          `form:"object_type_id"`
    ParentObjectID  *int64          `form:"parent_object_id"`
    Status          *string         `form:"status" validate:"omitempty,oneof=active inactive archived deleted pending"`
    IncludeDeleted  *bool           `form:"include_deleted"`
    Search          *string         `form:"search"`
    Tags            []string        `form:"tags"`
    TagsMode        string          `form:"tags_mode" validate:"omitempty,oneof=any all"`
    Metadata        map[string]any  `form:"metadata"`
    Page            int             `form:"page" validate:"min=1"`
    PageSize        int             `form:"page_size" validate:"min=1,max=100"`
    SortBy          string          `form:"sort_by" validate:"omitempty,oneof=name created_at updated_at"`
    SortOrder       string          `form:"sort_order" validate:"omitempty,oneof=asc desc"`
    IncludeTree     *bool           `form:"include_tree"`
    IncludeType     *bool           `form:"include_type"`
}

type ObjectListResponse struct {
    Objects []Object `json:"objects"`
    Total   int64    `json:"total"`
    Page    int      `json:"page"`
    PageSize int     `json:"page_size"`
}

type BatchCreateRequest struct {
    Objects []CreateObjectRequest `json:"objects" validate:"required,min=1,max=100"`
}

type BatchCreateResponse struct {
    Created []Object `json:"created"`
    Errors  []struct {
        Index   int    `json:"index"`
        Message string `json:"message"`
    } `json:"errors"`
}

type BatchUpdateRequest struct {
    Updates []struct {
        ID      int64             `json:"id" validate:"required,min=1"`
        Version int64             `json:"version" validate:"required,min=0"`
        Changes UpdateObjectRequest `json:"changes"`
    } `json:"updates" validate:"required,min=1,max=100"`
}

type BatchUpdateResponse struct {
    Updated []Object `json:"updated"`
    Errors  []struct {
        Index   int    `json:"index"`
        Message string `json:"message"`
    } `json:"errors"`
}
```

---

### 2.5 Delete Old Entity Model Files

**Files**:
- `internal/models/entity.go`
- `internal/models/entity_test.go`

**Steps**:
1. Delete `internal/models/entity.go`
2. Delete `internal/models/entity_test.go`

```bash
rm services/objects-service/internal/models/entity.go
rm services/objects-service/internal/models/entity_test.go
```

---

## Checklist

- [ ] Create `internal/models/object_type.go`
- [ ] Create `internal/models/object.go`
- [ ] Create `internal/models/object_type_request.go`
- [ ] Create `internal/models/object_request.go`
- [ ] Delete `internal/models/entity.go`
- [ ] Delete `internal/models/entity_test.go`
- [ ] Verify no compilation errors: `go build ./internal/models/...`
- [ ] Create basic unit tests for models (optional)
- [ ] Update progress.md

## Testing

```bash
# Verify models compile
cd services/objects-service
go build ./internal/models/...

# Run model tests if created
go test ./internal/models/... -v

# Verify no old entity references remain
grep -r "Entity" internal/models/ || echo "No Entity references found"
```

## Common Issues

**Issue**: JSONB Scan/Value methods not working
**Solution**: Ensure Metadata type implements sql.Scanner and driver.Valuer interfaces correctly

**Issue**: UUID type not recognized
**Solution**: Import `github.com/google/uuid` package

**Issue**: Validation tags not compiling
**Solution**: Ensure `github.com/go-playground/validator/v10` is imported

## Next Phase

Proceed to [Phase 3: Repository Layer](phase-03-repositories.md) once all tasks in this phase are complete.
