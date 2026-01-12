# Phase 4: Service Layer

**Estimated Time**: 4 hours
**Status**: ⬜ Not Started
**Dependencies**: Phase 3 (Repositories)

## Overview

Create service layer with business logic for Object Types and Objects, including validation, hierarchical operations, and error handling. Delete old Entity service files.

## Tasks

### 4.1 Create Object Type Service

**File**: `internal/services/object_type_service.go`

**Steps**:
1. Create ObjectTypeService struct
2. Implement business logic methods
3. Add validation for circular references
4. Add sealed type checking

```go
package services

import (
    "context"
    "errors"
    "fmt"
    
    "github.com/sirupsen/logrus"
    
    "github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
    "github.com/v-egorov/service-boilerplate/services/objects-service/internal/repository"
)

type ObjectTypeService struct {
    repo   *repository.ObjectTypeRepository
    logger *logrus.Logger
}

func NewObjectTypeService(repo *repository.ObjectTypeRepository, logger *logrus.Logger) *ObjectTypeService {
    return &ObjectTypeService{
        repo:   repo,
        logger: logger,
    }
}

// ObjectTypeRepositoryInterface defines repository operations needed
type ObjectTypeRepositoryInterface interface {
    Create(ctx context.Context, ot *models.ObjectType) error
    GetByID(ctx context.Context, id int64) (*models.ObjectType, error)
    GetByName(ctx context.Context, name string) (*models.ObjectType, error)
    List(ctx context.Context, filter *models.ObjectTypeFilter) ([]models.ObjectType, error)
    Update(ctx context.Context, ot *models.ObjectType) error
    Delete(ctx context.Context, id int64) error
    GetChildren(ctx context.Context, parentID int64) ([]models.ObjectType, error)
    GetTree(ctx context.Context) ([]models.ObjectType, error)
    Count(ctx context.Context, filter *models.ObjectTypeFilter) (int64, error)
}

// NewObjectTypeServiceWithInterface creates service with custom repository (for testing)
func NewObjectTypeServiceWithInterface(repo ObjectTypeRepositoryInterface, logger *logrus.Logger) *ObjectTypeService {
    return &ObjectTypeService{
        repo:   repo,
        logger: logger,
    }
}

func (s *ObjectTypeService) Create(ctx context.Context, req *models.CreateObjectTypeRequest) (*models.ObjectType, error) {
    if req.ParentTypeID != nil {
        parent, err := s.repo.GetByID(ctx, *req.ParentTypeID)
        if err != nil {
            s.logger.WithError(err).Error("Failed to get parent object type")
            return nil, fmt.Errorf("parent type not found: %w", err)
        }
        
        if parent.IsSealed {
            s.logger.Warn("Attempted to create child type for sealed parent type")
            return nil, errors.New("cannot create child type for sealed parent type")
        }
    }
    
    ot := &models.ObjectType{
        Name:              req.Name,
        ParentTypeID:      req.ParentTypeID,
        ConcreteTableName: req.ConcreteTableName,
        Description:       req.Description,
        IsSealed:          req.IsSealed != nil && *req.IsSealed,
        Metadata:          req.Metadata,
    }
    
    if ot.Metadata == nil {
        ot.Metadata = make(models.Metadata)
    }
    
    if err := s.repo.Create(ctx, ot); err != nil {
        s.logger.WithError(err).Error("Failed to create object type in service")
        
        // Check for constraint violations
        errMsg := err.Error()
        if containsString(errMsg, "duplicate key") ||
           containsString(errMsg, "unique constraint") ||
           containsString(errMsg, "already exists") {
            return nil, fmt.Errorf("object type already exists: %w", err)
        }
        
        return nil, fmt.Errorf("failed to create object type: %w", err)
    }
    
    s.logger.WithField("object_type_id", ot.ID).Info("Object type created successfully")
    return ot, nil
}

func (s *ObjectTypeService) GetByID(ctx context.Context, id int64) (*models.ObjectType, error) {
    if id == 0 {
        return nil, errors.New("object type ID is required")
    }
    
    return s.repo.GetByID(ctx, id)
}

func (s *ObjectTypeService) GetByName(ctx context.Context, name string) (*models.ObjectType, error) {
    if name == "" {
        return nil, errors.New("object type name is required")
    }
    
    return s.repo.GetByName(ctx, name)
}

// Helper function for string checking
func containsString(s, substr string) bool {
    return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && s[len(s)-len(substr):] == substr || len(s) > len(substr) && s[:len(substr)] == substr)
}

func (s *ObjectTypeService) List(ctx context.Context, filter *models.ObjectTypeFilter) (*models.ObjectTypeListResponse, error) {
    types, err := s.repo.List(ctx, filter)
    if err != nil {
        return nil, fmt.Errorf("failed to list object types: %w", err)
    }
    
    total, err := s.repo.Count(ctx, filter)
    if err != nil {
        return nil, fmt.Errorf("failed to count object types: %w", err)
    }
    
    return &models.ObjectTypeListResponse{
        ObjectTypes: types,
        Total:       total,
        Page:        1,
        PageSize:    len(types),
    }, nil
}

func (s *ObjectTypeService) Update(ctx context.Context, id int64, req *models.UpdateObjectTypeRequest) (*models.ObjectType, error) {
    ot, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("object type not found: %w", err)
    }
    
    if req.Name != nil {
        ot.Name = *req.Name
    }
    
    if req.ConcreteTableName != nil {
        ot.ConcreteTableName = req.ConcreteTableName
    }
    
    if req.Description != nil {
        ot.Description = req.Description
    }
    
    if req.IsSealed != nil {
        if *req.IsSealed && !ot.IsSealed {
            hasChildren, err := s.checkHasChildren(ctx, id)
            if err != nil {
                return nil, fmt.Errorf("failed to check children: %w", err)
            }
            if hasChildren {
                return nil, errors.New("cannot seal type with children")
            }
        }
        ot.IsSealed = *req.IsSealed
    }
    
    if req.Metadata != nil {
        ot.Metadata = *req.Metadata
    }
    
    if err := s.repo.Update(ctx, ot); err != nil {
        return nil, fmt.Errorf("failed to update object type: %w", err)
    }
    
    return ot, nil
}

func (s *ObjectTypeService) Delete(ctx context.Context, id int64) error {
    hasChildren, err := s.checkHasChildren(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to check children: %w", err)
    }
    
    if hasChildren {
        return errors.New("cannot delete object type with children")
    }
    
    if err := s.repo.Delete(ctx, id); err != nil {
        return fmt.Errorf("failed to delete object type: %w", err)
    }
    
    return nil
}

func (s *ObjectTypeService) GetTree(ctx context.Context) ([]models.ObjectType, error) {
    return s.repo.GetTree(ctx, nil)
}

func (s *ObjectTypeService) GetChildren(ctx context.Context, parentID int64) ([]models.ObjectType, error) {
    return s.repo.GetChildren(ctx, parentID)
}

func (s *ObjectTypeService) checkHasChildren(ctx context.Context, id int64) (bool, error) {
    children, err := s.repo.GetChildren(ctx, id)
    if err != nil {
        return false, err
    }
    return len(children) > 0, nil
}
```

---

### 4.2 Create Object Service

**File**: `internal/services/object_service.go`

**Steps**:
1. Create ObjectService struct
2. Implement business logic methods
3. Add validation
4. Add version checking
5. Implement batch operations

```go
package services

import (
    "context"
    "errors"
    "fmt"
    
    "github.com/google/uuid"
    
    "your-project/services/objects-service/internal/models"
    "your-project/services/objects-service/internal/repository"
)

type ObjectService struct {
    objectRepo      *repository.ObjectRepository
    objectTypeRepo  *repository.ObjectTypeRepository
}

func NewObjectService(objectRepo *repository.ObjectRepository, objectTypeRepo *repository.ObjectTypeRepository) *ObjectService {
    return &ObjectService{
        objectRepo:     objectRepo,
        objectTypeRepo: objectTypeRepo,
    }
}

func (s *ObjectService) Create(ctx context.Context, req *models.CreateObjectRequest, userID string) (*models.Object, error) {
    _, err := s.objectTypeRepo.GetByID(ctx, req.ObjectTypeID)
    if err != nil {
        return nil, fmt.Errorf("object type not found: %w", err)
    }
    
    if req.ParentObjectID != nil {
        _, err := s.objectRepo.GetByID(ctx, *req.ParentObjectID)
        if err != nil {
            return nil, fmt.Errorf("parent object not found: %w", err)
        }
    }
    
    obj := &models.Object{
        PublicID:       uuid.New(),
        ObjectTypeID:   req.ObjectTypeID,
        ParentObjectID: req.ParentObjectID,
        Name:           req.Name,
        Description:    req.Description,
        Metadata:       req.Metadata,
        Status:         models.StatusActive,
        Tags:           []string{},
    }
    
    if req.Status != nil {
        obj.Status = *req.Status
    }
    
    if req.Tags != nil {
        obj.Tags = *req.Tags
    }
    
    if obj.Metadata == nil {
        obj.Metadata = make(models.Metadata)
    }
    
    if err := s.objectRepo.Create(ctx, obj, userID); err != nil {
        return nil, fmt.Errorf("failed to create object: %w", err)
    }
    
    return obj, nil
}

func (s *ObjectService) GetByID(ctx context.Context, id int64) (*models.Object, error) {
    return s.objectRepo.GetByID(ctx, id)
}

func (s *ObjectService) GetByPublicID(ctx context.Context, publicID uuid.UUID) (*models.Object, error) {
    return s.objectRepo.GetByPublicID(ctx, publicID)
}

func (s *ObjectService) List(ctx context.Context, filter *models.ObjectFilter) (*models.ObjectListResponse, error) {
    if filter.Page < 1 {
        filter.Page = 1
    }
    if filter.PageSize < 1 || filter.PageSize > 100 {
        filter.PageSize = 20
    }
    
    objects, total, err := s.objectRepo.List(ctx, filter)
    if err != nil {
        return nil, fmt.Errorf("failed to list objects: %w", err)
    }
    
    return &models.ObjectListResponse{
        Objects: objects,
        Total:   total,
        Page:    filter.Page,
        PageSize: filter.PageSize,
    }, nil
}

func (s *ObjectService) Update(ctx context.Context, id int64, req *models.UpdateObjectRequest, userID string) (*models.Object, error) {
    obj, err := s.objectRepo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("object not found: %w", err)
    }
    
    if req.Version != nil && *req.Version != obj.Version {
        return nil, repository.ErrObjectVersionConflict
    }
    
    if req.Name != nil {
        obj.Name = *req.Name
    }
    
    if req.Description != nil {
        obj.Description = req.Description
    }
    
    if req.Metadata != nil {
        obj.Metadata = *req.Metadata
    }
    
    if req.Status != nil {
        obj.Status = *req.Status
    }
    
    if req.Tags != nil {
        obj.Tags = *req.Tags
    }
    
    if err := s.objectRepo.Update(ctx, obj, userID); err != nil {
        return nil, fmt.Errorf("failed to update object: %w", err)
    }
    
    return obj, nil
}

func (s *ObjectService) SoftDelete(ctx context.Context, id int64, userID string) error {
    if err := s.objectRepo.SoftDelete(ctx, id, userID); err != nil {
        return fmt.Errorf("failed to delete object: %w", err)
    }
    return nil
}

func (s *ObjectService) HardDelete(ctx context.Context, id int64) error {
    if err := s.objectRepo.HardDelete(ctx, id); err != nil {
        return fmt.Errorf("failed to permanently delete object: %w", err)
    }
    return nil
}

func (s *ObjectService) GetChildren(ctx context.Context, parentID int64) ([]models.Object, error) {
    return s.objectRepo.GetChildren(ctx, parentID)
}

func (s *ObjectService) CreateBatch(ctx context.Context, req *models.BatchCreateRequest, userID string) (*models.BatchCreateResponse, error) {
    objects := make([]models.Object, len(req.Objects))
    
    for i, objReq := range req.Objects {
        _, err := s.objectTypeRepo.GetByID(ctx, objReq.ObjectTypeID)
        if err != nil {
            return nil, fmt.Errorf("object type not found for object %d: %w", i, err)
        }
        
        objects[i] = models.Object{
            PublicID:       uuid.New(),
            ObjectTypeID:   objReq.ObjectTypeID,
            ParentObjectID: objReq.ParentObjectID,
            Name:           objReq.Name,
            Description:    objReq.Description,
            Metadata:       objReq.Metadata,
            Status:         models.StatusActive,
            Tags:           []string{},
        }
        
        if objReq.Status != nil {
            objects[i].Status = *objReq.Status
        }
        
        if objReq.Tags != nil {
            objects[i].Tags = *objReq.Tags
        }
        
        if objects[i].Metadata == nil {
            objects[i].Metadata = make(models.Metadata)
        }
    }
    
    created, errs := s.objectRepo.CreateBatch(ctx, objects, userID)
    
    response := &models.BatchCreateResponse{
        Created: created,
    }
    
    for _, err := range errs {
        response.Errors = append(response.Errors, struct {
            Index   int    `json:"index"`
            Message string `json:"message"`
        }{
            Message: err.Error(),
        })
    }
    
    return response, nil
}

func (s *ObjectService) UpdateBatch(ctx context.Context, req *models.BatchUpdateRequest, userID string) (*models.BatchUpdateResponse, error) {
    updated := make([]models.Object, 0)
    errors := make([]struct {
        Index   int    `json:"index"`
        Message string `json:"message"`
    }, 0)
    
    for i, updateReq := range req.Updates {
        obj, err := s.objectRepo.GetByID(ctx, updateReq.ID)
        if err != nil {
            errors = append(errors, struct {
                Index   int    `json:"index"`
                Message string `json:"message"`
            }{
                Index:   i,
                Message: fmt.Sprintf("object not found: %v", err),
            })
            continue
        }
        
        if updateReq.Version != obj.Version {
            errors = append(errors, struct {
                Index   int    `json:"index"`
                Message string `json:"message"`
            }{
                Index:   i,
                Message: "version conflict",
            })
            continue
        }
        
        if updateReq.Changes.Name != nil {
            obj.Name = *updateReq.Changes.Name
        }
        
        if updateReq.Changes.Description != nil {
            obj.Description = updateReq.Changes.Description
        }
        
        if updateReq.Changes.Metadata != nil {
            obj.Metadata = *updateReq.Changes.Metadata
        }
        
        if updateReq.Changes.Status != nil {
            obj.Status = *updateReq.Changes.Status
        }
        
        if updateReq.Changes.Tags != nil {
            obj.Tags = *updateReq.Changes.Tags
        }
        
        if err := s.objectRepo.Update(ctx, obj, userID); err != nil {
            errors = append(errors, struct {
                Index   int    `json:"index"`
                Message string `json:"message"`
            }{
                Index:   i,
                Message: fmt.Sprintf("update failed: %v", err),
            })
            continue
        }
        
        updated = append(updated, *obj)
    }
    
    return &models.BatchUpdateResponse{
        Updated: updated,
        Errors:  errors,
    }, nil
}
```

---

### 4.3 Delete Old Entity Service Files

**Files**:
- `internal/services/entity_service.go`
- `internal/services/entity_service_test.go`

**Steps**:
1. Delete `internal/services/entity_service.go`
2. Delete `internal/services/entity_service_test.go`

```bash
rm services/objects-service/internal/services/entity_service.go
rm services/objects-service/internal/services/entity_service_test.go
```

---

## Checklist

- [ ] Create `internal/services/object_type_service.go`
- [ ] Create `internal/services/object_service.go`
- [ ] Delete `internal/services/entity_service.go`
- [ ] Delete `internal/services/entity_service_test.go`
- [ ] Verify no compilation errors: `go build ./internal/services/...`
- [ ] Create unit tests for services
- [ ] Test business logic manually
- [ ] Update progress.md

## Testing

```bash
# Verify services compile
cd services/objects-service
go build ./internal/services/...

# Run service tests
go test ./internal/services/... -v

# Test validation manually
```

## Common Issues

**Issue**: Validation not working
**Solution**: Ensure validation struct tags are correct and validator is initialized

**Issue**: Circular reference check not implemented
**Solution**: Add recursive check in ObjectTypeService.Create for parent type hierarchy

**Issue**: Version conflict error not propagated
**Solution**: Ensure ErrObjectVersionConflict is properly returned and handled

## Next Phase

Proceed to [Phase 5: Handlers Layer](phase-05-handlers.md) once all tasks in this phase are complete.
