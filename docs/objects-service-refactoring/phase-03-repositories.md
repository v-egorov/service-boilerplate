# Phase 3: Repository Layer

**Estimated Time**: 4 hours
**Status**: ⬜ Not Started
**Dependencies**: Phase 2 (Models)

## Overview

Create repository implementations for Object Types and Objects with full CRUD operations, hierarchical queries, and search capabilities. Delete old Entity repository files.

## Tasks

### 3.1 Create Object Type Repository

**File**: `internal/repository/object_type_repository.go`

**Steps**:
1. Create ObjectTypeRepository struct
2. Implement CRUD methods
3. Implement hierarchical queries
4. Add transaction support

```go
package repository

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    
    "your-project/services/objects-service/internal/models"
    
    "github.com/jmoiron/sqlx"
)

var (
    ErrObjectTypeNotFound = errors.New("object type not found")
    ErrObjectTypeExists   = errors.New("object type already exists")
    ErrCircularReference  = errors.New("circular reference detected")
)

type ObjectTypeRepository struct {
    db *sqlx.DB
}

func NewObjectTypeRepository(db *sqlx.DB) *ObjectTypeRepository {
    return &ObjectTypeRepository{db: db}
}

func (r *ObjectTypeRepository) Create(ctx context.Context, ot *models.ObjectType) error {
    query := `
        INSERT INTO object_types (name, parent_type_id, concrete_table_name, description, is_sealed, metadata)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, created_at, updated_at
    `
    
    err := r.db.QueryRowContext(ctx, query,
        ot.Name,
        ot.ParentTypeID,
        ot.ConcreteTableName,
        ot.Description,
        ot.IsSealed,
        ot.Metadata,
    ).Scan(&ot.ID, &ot.CreatedAt, &ot.UpdatedAt)
    
    if err != nil {
        if isUniqueViolation(err) {
            return ErrObjectTypeExists
        }
        return err
    }
    
    return nil
}

func (r *ObjectTypeRepository) GetByID(ctx context.Context, id int64) (*models.ObjectType, error) {
    query := `
        SELECT id, name, parent_type_id, concrete_table_name, description, is_sealed, metadata, created_at, updated_at
        FROM object_types
        WHERE id = $1
    `
    
    var ot models.ObjectType
    err := r.db.GetContext(ctx, &ot, query, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrObjectTypeNotFound
        }
        return nil, err
    }
    
    return &ot, nil
}

func (r *ObjectTypeRepository) GetByName(ctx context.Context, name string) (*models.ObjectType, error) {
    query := `
        SELECT id, name, parent_type_id, concrete_table_name, description, is_sealed, metadata, created_at, updated_at
        FROM object_types
        WHERE name = $1
    `
    
    var ot models.ObjectType
    err := r.db.GetContext(ctx, &ot, query, name)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrObjectTypeNotFound
        }
        return nil, err
    }
    
    return &ot, nil
}

func (r *ObjectTypeRepository) List(ctx context.Context, filter *models.ObjectTypeFilter) ([]models.ObjectType, error) {
    query := `
        SELECT id, name, parent_type_id, concrete_table_name, description, is_sealed, metadata, created_at, updated_at
        FROM object_types
        WHERE 1=1
    `
    
    args := []interface{}{}
    argPos := 1
    
    if filter.ParentTypeID != nil {
        query += fmt.Sprintf(" AND parent_type_id = $%d", argPos)
        args = append(args, *filter.ParentTypeID)
        argPos++
    }
    
    if filter.IsSealed != nil {
        query += fmt.Sprintf(" AND is_sealed = $%d", argPos)
        args = append(args, *filter.IsSealed)
        argPos++
    }
    
    if filter.Search != nil {
        query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argPos, argPos+1)
        searchTerm := "%" + *filter.Search + "%"
        args = append(args, searchTerm, searchTerm)
        argPos += 2
    }
    
    query += " ORDER BY name ASC"
    
    var types []models.ObjectType
    err := r.db.SelectContext(ctx, &types, query, args...)
    if err != nil {
        return nil, err
    }
    
    return types, nil
}

func (r *ObjectTypeRepository) Update(ctx context.Context, ot *models.ObjectType) error {
    query := `
        UPDATE object_types
        SET name = $2, concrete_table_name = $3, description = $4, is_sealed = $5, metadata = $6
        WHERE id = $1
        RETURNING updated_at
    `
    
    err := r.db.QueryRowContext(ctx, query,
        ot.ID,
        ot.Name,
        ot.ConcreteTableName,
        ot.Description,
        ot.IsSealed,
        ot.Metadata,
    ).Scan(&ot.UpdatedAt)
    
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return ErrObjectTypeNotFound
        }
        if isUniqueViolation(err) {
            return ErrObjectTypeExists
        }
        return err
    }
    
    return nil
}

func (r *ObjectTypeRepository) Delete(ctx context.Context, id int64) error {
    query := `DELETE FROM object_types WHERE id = $1`
    
    result, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        return err
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    
    if rowsAffected == 0 {
        return ErrObjectTypeNotFound
    }
    
    return nil
}

func (r *ObjectTypeRepository) GetChildren(ctx context.Context, parentID int64) ([]models.ObjectType, error) {
    query := `
        SELECT id, name, parent_type_id, concrete_table_name, description, is_sealed, metadata, created_at, updated_at
        FROM object_types
        WHERE parent_type_id = $1
        ORDER BY name ASC
    `
    
    var types []models.ObjectType
    err := r.db.SelectContext(ctx, &types, query, parentID)
    if err != nil {
        return nil, err
    }
    
    return types, nil
}

func (r *ObjectTypeRepository) GetTree(ctx context.Context, rootID *int64) ([]models.ObjectType, error) {
    query := `
        WITH RECURSIVE type_tree AS (
            SELECT id, name, parent_type_id, concrete_table_name, description, is_sealed, metadata, created_at, updated_at, 0 as level
            FROM object_types
            WHERE parent_type_id IS NULL
            UNION ALL
            SELECT ot.id, ot.name, ot.parent_type_id, ot.concrete_table_name, ot.description, ot.is_sealed, ot.metadata, ot.created_at, ot.updated_at, tt.level + 1
            FROM object_types ot
            INNER JOIN type_tree tt ON ot.parent_type_id = tt.id
        )
        SELECT id, name, parent_type_id, concrete_table_name, description, is_sealed, metadata, created_at, updated_at
        FROM type_tree
        ORDER BY name ASC
    `
    
    var types []models.ObjectType
    err := r.db.SelectContext(ctx, &types, query)
    if err != nil {
        return nil, err
    }
    
    return types, nil
}

func (r *ObjectTypeRepository) Count(ctx context.Context, filter *models.ObjectTypeFilter) (int64, error) {
    query := `SELECT COUNT(*) FROM object_types WHERE 1=1`
    args := []interface{}{}
    argPos := 1
    
    if filter.ParentTypeID != nil {
        query += fmt.Sprintf(" AND parent_type_id = $%d", argPos)
        args = append(args, *filter.ParentTypeID)
        argPos++
    }
    
    if filter.IsSealed != nil {
        query += fmt.Sprintf(" AND is_sealed = $%d", argPos)
        args = append(args, *filter.IsSealed)
        argPos++
    }
    
    var count int64
    err := r.db.GetContext(ctx, &count, query, args...)
    return count, err
}

func isUniqueViolation(err error) bool {
    var pgErr *pq.Error
    if errors.As(err, &pgErr) {
        return pgErr.Code == "23505"
    }
    return false
}
```

---

### 3.2 Create Object Repository

**File**: `internal/repository/object_repository.go`

**Steps**:
1. Create ObjectRepository struct
2. Implement CRUD methods with public_id support
3. Implement filtering and search
4. Add version checking for optimistic locking
5. Implement soft delete

```go
package repository

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    
    "github.com/jmoiron/sqlx"
    "github.com/google/uuid"
    
    "your-project/services/objects-service/internal/models"
)

var (
    ErrObjectNotFound        = errors.New("object not found")
    ErrObjectVersionConflict = errors.New("object version conflict")
)

type ObjectRepository struct {
    db *sqlx.DB
}

func NewObjectRepository(db *sqlx.DB) *ObjectRepository {
    return &ObjectRepository{db: db}
}

func (r *ObjectRepository) Create(ctx context.Context, obj *models.Object, userID string) error {
    query := `
        INSERT INTO objects (public_id, object_type_id, parent_object_id, name, description, created_by, updated_by, metadata, status, tags)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING id, created_at, updated_at, version
    `
    
    err := r.db.QueryRowContext(ctx, query,
        obj.PublicID,
        obj.ObjectTypeID,
        obj.ParentObjectID,
        obj.Name,
        obj.Description,
        userID,
        userID,
        obj.Metadata,
        obj.Status,
        obj.Tags,
    ).Scan(&obj.ID, &obj.CreatedAt, &obj.UpdatedAt, &obj.Version)
    
    return err
}

func (r *ObjectRepository) GetByID(ctx context.Context, id int64) (*models.Object, error) {
    query := `
        SELECT id, public_id, object_type_id, parent_object_id, name, description, created_at, updated_at, deleted_at, version, created_by, updated_by, metadata, status, tags
        FROM objects
        WHERE id = $1
    `
    
    var obj models.Object
    err := r.db.GetContext(ctx, &obj, query, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrObjectNotFound
        }
        return nil, err
    }
    
    return &obj, nil
}

func (r *ObjectRepository) GetByPublicID(ctx context.Context, publicID uuid.UUID) (*models.Object, error) {
    query := `
        SELECT id, public_id, object_type_id, parent_object_id, name, description, created_at, updated_at, deleted_at, version, created_by, updated_by, metadata, status, tags
        FROM objects
        WHERE public_id = $1
    `
    
    var obj models.Object
    err := r.db.GetContext(ctx, &obj, query, publicID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrObjectNotFound
        }
        return nil, err
    }
    
    return &obj, nil
}

func (r *ObjectRepository) List(ctx context.Context, filter *models.ObjectFilter) ([]models.Object, int64, error) {
    var query string
    var countQuery string
    var args []interface{}
    argPos := 1
    
    if !*filter.IncludeDeleted {
        query = `
            SELECT id, public_id, object_type_id, parent_object_id, name, description, created_at, updated_at, deleted_at, version, created_by, updated_by, metadata, status, tags
            FROM objects
            WHERE deleted_at IS NULL
        `
        countQuery = `SELECT COUNT(*) FROM objects WHERE deleted_at IS NULL`
    } else {
        query = `
            SELECT id, public_id, object_type_id, parent_object_id, name, description, created_at, updated_at, deleted_at, version, created_by, updated_by, metadata, status, tags
            FROM objects
            WHERE 1=1
        `
        countQuery = `SELECT COUNT(*) FROM objects WHERE 1=1`
    }
    
    if filter.ObjectTypeID != nil {
        query += fmt.Sprintf(" AND object_type_id = $%d", argPos)
        countQuery += fmt.Sprintf(" AND object_type_id = $%d", argPos)
        args = append(args, *filter.ObjectTypeID)
        argPos++
    }
    
    if filter.ParentObjectID != nil {
        query += fmt.Sprintf(" AND parent_object_id = $%d", argPos)
        countQuery += fmt.Sprintf(" AND parent_object_id = $%d", argPos)
        args = append(args, *filter.ParentObjectID)
        argPos++
    }
    
    if filter.Status != nil {
        query += fmt.Sprintf(" AND status = $%d", argPos)
        countQuery += fmt.Sprintf(" AND status = $%d", argPos)
        args = append(args, *filter.Status)
        argPos++
    }
    
    if filter.Search != nil {
        query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argPos, argPos+1)
        countQuery += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argPos, argPos+1)
        searchTerm := "%" + *filter.Search + "%"
        args = append(args, searchTerm, searchTerm)
        argPos += 2
    }
    
    if len(filter.Tags) > 0 {
        if filter.TagsMode == "all" {
            query += fmt.Sprintf(" AND tags @> $%d", argPos)
            countQuery += fmt.Sprintf(" AND tags @> $%d", argPos)
            args = append(args, filter.Tags)
            argPos++
        } else {
            query += fmt.Sprintf(" AND tags && $%d", argPos)
            countQuery += fmt.Sprintf(" AND tags && $%d", argPos)
            args = append(args, filter.Tags)
            argPos++
        }
    }
    
    var total int64
    err := r.db.GetContext(ctx, &total, countQuery, args...)
    if err != nil {
        return nil, 0, err
    }
    
    sortBy := "created_at"
    if filter.SortBy != "" {
        sortBy = filter.SortBy
    }
    
    sortOrder := "DESC"
    if filter.SortOrder == "asc" {
        sortOrder = "ASC"
    }
    
    query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)
    
    offset := (filter.Page - 1) * filter.PageSize
    query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
    args = append(args, filter.PageSize, offset)
    
    var objects []models.Object
    err = r.db.SelectContext(ctx, &objects, query, args...)
    if err != nil {
        return nil, 0, err
    }
    
    return objects, total, nil
}

func (r *ObjectRepository) Update(ctx context.Context, obj *models.Object, userID string) error {
    query := `
        UPDATE objects
        SET name = $2, description = $3, metadata = $4, status = $5, tags = $6, updated_by = $7, version = version + 1
        WHERE id = $1 AND version = $8
        RETURNING updated_at, version
    `
    
    err := r.db.QueryRowContext(ctx, query,
        obj.ID,
        obj.Name,
        obj.Description,
        obj.Metadata,
        obj.Status,
        obj.Tags,
        userID,
        obj.Version,
    ).Scan(&obj.UpdatedAt, &obj.Version)
    
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return ErrObjectVersionConflict
        }
        return err
    }
    
    return nil
}

func (r *ObjectRepository) SoftDelete(ctx context.Context, id int64, userID string) error {
    query := `
        UPDATE objects
        SET deleted_at = NOW(), status = 'deleted', updated_by = $2
        WHERE id = $1 AND deleted_at IS NULL
    `
    
    result, err := r.db.ExecContext(ctx, query, id, userID)
    if err != nil {
        return err
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    
    if rowsAffected == 0 {
        return ErrObjectNotFound
    }
    
    return nil
}

func (r *ObjectRepository) HardDelete(ctx context.Context, id int64) error {
    query := `DELETE FROM objects WHERE id = $1`
    
    result, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        return err
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    
    if rowsAffected == 0 {
        return ErrObjectNotFound
    }
    
    return nil
}

func (r *ObjectRepository) GetChildren(ctx context.Context, parentID int64) ([]models.Object, error) {
    query := `
        SELECT id, public_id, object_type_id, parent_object_id, name, description, created_at, updated_at, deleted_at, version, created_by, updated_by, metadata, status, tags
        FROM objects
        WHERE parent_object_id = $1 AND deleted_at IS NULL
        ORDER BY name ASC
    `
    
    var objects []models.Object
    err := r.db.SelectContext(ctx, &objects, query, parentID)
    if err != nil {
        return nil, err
    }
    
    return objects, nil
}

func (r *ObjectRepository) CreateBatch(ctx context.Context, objects []models.Object, userID string) ([]models.Object, []error) {
    tx, err := r.db.BeginTxx(ctx, nil)
    if err != nil {
        return nil, []error{err}
    }
    defer tx.Rollback()
    
    created := make([]models.Object, 0, len(objects))
    errors := make([]error, 0)
    
    query := `
        INSERT INTO objects (public_id, object_type_id, parent_object_id, name, description, created_by, updated_by, metadata, status, tags)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING id, created_at, updated_at, version
    `
    
    for i := range objects {
        obj := &objects[i]
        err := tx.QueryRow(query,
            obj.PublicID,
            obj.ObjectTypeID,
            obj.ParentObjectID,
            obj.Name,
            obj.Description,
            userID,
            userID,
            obj.Metadata,
            obj.Status,
            obj.Tags,
        ).Scan(&obj.ID, &obj.CreatedAt, &obj.UpdatedAt, &obj.Version)
        
        if err != nil {
            errors = append(errors, fmt.Errorf("object %d: %w", i, err))
            continue
        }
        
        created = append(created, *obj)
    }
    
    if err := tx.Commit(); err != nil {
        return created, append(errors, err)
    }
    
    return created, errors
}
```

---

### 3.3 Delete Old Entity Repository Files

**Files**:
- `internal/repository/entity_repository.go`
- `internal/repository/entity_repository_test.go`

**Steps**:
1. Delete `internal/repository/entity_repository.go`
2. Delete `internal/repository/entity_repository_test.go`

```bash
rm services/objects-service/internal/repository/entity_repository.go
rm services/objects-service/internal/repository/entity_repository_test.go
```

---

## Checklist

- [ ] Create `internal/repository/object_type_repository.go`
- [ ] Create `internal/repository/object_repository.go`
- [ ] Delete `internal/repository/entity_repository.go`
- [ ] Delete `internal/repository/entity_repository_test.go`
- [ ] Verify no compilation errors: `go build ./internal/repository/...`
- [ ] Create basic unit tests for repositories
- [ ] Test database queries manually with psql
- [ ] Update progress.md

## Testing

```bash
# Verify repositories compile
cd services/objects-service
go build ./internal/repository/...

# Run repository tests
go test ./internal/repository/... -v

# Test queries manually
psql postgresql://postgres:password@localhost:5432/objects_service
```

## Common Issues

**Issue**: Recursive query syntax error
**Solution**: Ensure PostgreSQL version is 8.4+ for CTE support

**Issue**: JSONB array operators not working
**Solution**: Use `@>` (contains) or `&&` (overlap) for array operations

**Issue**: Transaction rollback not working
**Solution**: Ensure defer tx.Rollback() is called before checking err

## Next Phase

Proceed to [Phase 4: Service Layer](phase-04-services.md) once all tasks in this phase are complete.
