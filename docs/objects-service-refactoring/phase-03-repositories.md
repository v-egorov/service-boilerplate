# Phase 3: Repository Layer

**Estimated Time**: 4 hours
**Status**: ⬜ Not Started
**Dependencies**: Phase 2 (Models)

## Overview

Create repository implementations for Object Types and Objects with full CRUD operations, hierarchical queries, and search capabilities. Use pgx/v5 patterns consistent with existing services. Delete old Entity repository files.

## Key Patterns to Follow

1. **DBInterface Pattern** - For testability and mocking
2. **OpenTelemetry Tracing** - Use database.TraceDB* helpers for all operations
3. **pgx/v5 Direct Usage** - Not sqlx (inconsistent with codebase)
4. **Connection Pooling** - Use pgxpool.Pool
5. **Error Handling** - Check for pgx.ErrNoRows and constraint violations
6. **Transaction Support** - Use db.WithTx() helper

## Tasks

### 3.1 Create Object Type Repository

**File**: `internal/repository/object_type_repository.go`

**Steps**:
1. Create DBInterface for testability
2. Create ObjectTypeRepository struct with logger
3. Implement CRUD methods with tracing
4. Implement hierarchical queries
5. Add transaction support

```go
package repository

import (
    "context"
    "errors"
    "fmt"
    "strings"
    
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/sirupsen/logrus"
    
    "github.com/v-egorov/service-boilerplate/common/database"
    "github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
)

// DBInterface defines database operations needed for testing
type DBInterface interface {
    Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
    QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
    Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
    Begin(ctx context.Context) (pgx.Tx, error)
}

type ObjectTypeRepository struct {
    db     DBInterface
    logger *logrus.Logger
}

func NewObjectTypeRepository(db *pgxpool.Pool, logger *logrus.Logger) *ObjectTypeRepository {
    return &ObjectTypeRepository{
        db:     db,
        logger: logger,
    }
}

// NewObjectTypeRepositoryWithInterface creates repository with custom DB interface (for testing)
func NewObjectTypeRepositoryWithInterface(db DBInterface, logger *logrus.Logger) *ObjectTypeRepository {
    return &ObjectTypeRepository{
        db:     db,
        logger: logger,
    }
}

func (r *ObjectTypeRepository) Create(ctx context.Context, ot *models.ObjectType) error {
    query := `
        INSERT INTO object_types (name, parent_type_id, concrete_table_name, description, is_sealed, metadata)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, created_at, updated_at
    `
    
    err := database.TraceDBInsert(ctx, "object_types", query, func(ctx context.Context) error {
        return r.db.QueryRow(ctx, query,
            ot.Name,
            ot.ParentTypeID,
            ot.ConcreteTableName,
            ot.Description,
            ot.IsSealed,
            ot.Metadata,
        ).Scan(&ot.ID, &ot.CreatedAt, &ot.UpdatedAt)
    })
    
    if err != nil {
        if isUniqueViolation(err) {
            return ErrObjectTypeExists
        }
        r.logger.WithError(err).Error("Failed to create object type")
        return fmt.Errorf("failed to create object type: %w", err)
    }
    
    r.logger.WithField("object_type_id", ot.ID).Info("Object type created successfully")
    return nil
}

func (r *ObjectTypeRepository) GetByID(ctx context.Context, id int64) (*models.ObjectType, error) {
    query := `
        SELECT id, name, parent_type_id, concrete_table_name, description, is_sealed, metadata, created_at, updated_at
        FROM object_types
        WHERE id = $1
    `
    
    var ot models.ObjectType
    err := database.TraceDBQuery(ctx, "object_types", query, func(ctx context.Context) error {
        return r.db.QueryRow(ctx, query, id).Scan(
            &ot.ID, &ot.Name, &ot.ParentTypeID, &ot.ConcreteTableName,
            &ot.Description, &ot.IsSealed, &ot.Metadata, &ot.CreatedAt, &ot.UpdatedAt,
        )
    })
    
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, ErrObjectTypeNotFound
        }
        r.logger.WithError(err).Error("Failed to get object type by ID")
        return nil, fmt.Errorf("failed to get object type: %w", err)
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
    err := database.TraceDBQuery(ctx, "object_types", query, func(ctx context.Context) error {
        return r.db.QueryRow(ctx, query, name).Scan(
            &ot.ID, &ot.Name, &ot.ParentTypeID, &ot.ConcreteTableName,
            &ot.Description, &ot.IsSealed, &ot.Metadata, &ot.CreatedAt, &ot.UpdatedAt,
        )
    })
    
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, ErrObjectTypeNotFound
        }
        r.logger.WithError(err).Error("Failed to get object type by name")
        return nil, fmt.Errorf("failed to get object type: %w", err)
    }
    
    return &ot, nil
}

func (r *ObjectTypeRepository) List(ctx context.Context, filter *models.ObjectTypeFilter) ([]models.ObjectType, error) {
    query := `
        SELECT id, name, parent_type_id, concrete_table_name, description, is_sealed, metadata, created_at, updated_at
        FROM object_types
        WHERE 1=1
    `
    
    args := []any{}
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
    err := database.TraceDBQuery(ctx, "object_types", query, func(ctx context.Context) error {
        rows, err := r.db.Query(ctx, query, args...)
        if err != nil {
            return err
        }
        defer rows.Close()
        
        for rows.Next() {
            var ot models.ObjectType
            err := rows.Scan(
                &ot.ID, &ot.Name, &ot.ParentTypeID, &ot.ConcreteTableName,
                &ot.Description, &ot.IsSealed, &ot.Metadata, &ot.CreatedAt, &ot.UpdatedAt,
            )
            if err != nil {
                return fmt.Errorf("failed to scan object type: %w", err)
            }
            types = append(types, ot)
        }
        
        return rows.Err()
    })
    
    if err != nil {
        r.logger.WithError(err).Error("Failed to list object types")
        return nil, fmt.Errorf("failed to list object types: %w", err)
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
    
    err := database.TraceDBUpdate(ctx, "object_types", query, func(ctx context.Context) error {
        return r.db.QueryRow(ctx, query,
            ot.ID,
            ot.Name,
            ot.ConcreteTableName,
            ot.Description,
            ot.IsSealed,
            ot.Metadata,
        ).Scan(&ot.UpdatedAt)
    })
    
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return ErrObjectTypeNotFound
        }
        if isUniqueViolation(err) {
            return ErrObjectTypeExists
        }
        r.logger.WithError(err).Error("Failed to update object type")
        return fmt.Errorf("failed to update object type: %w", err)
    }
    
    r.logger.WithField("object_type_id", ot.ID).Info("Object type updated successfully")
    return nil
}

func (r *ObjectTypeRepository) Delete(ctx context.Context, id int64) error {
    query := `DELETE FROM object_types WHERE id = $1`
    
    var result pgconn.CommandTag
    err := database.TraceDBDelete(ctx, "object_types", query, func(ctx context.Context) error {
        var execErr error
        result, execErr = r.db.Exec(ctx, query, id)
        return execErr
    })
    
    if err != nil {
        r.logger.WithError(err).Error("Failed to delete object type")
        return fmt.Errorf("failed to delete object type: %w", err)
    }
    
    if result.RowsAffected() == 0 {
        return ErrObjectTypeNotFound
    }
    
    r.logger.WithField("object_type_id", id).Info("Object type deleted successfully")
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
    err := database.TraceDBQuery(ctx, "object_types", query, func(ctx context.Context) error {
        rows, err := r.db.Query(ctx, query, parentID)
        if err != nil {
            return err
        }
        defer rows.Close()
        
        for rows.Next() {
            var ot models.ObjectType
            err := rows.Scan(
                &ot.ID, &ot.Name, &ot.ParentTypeID, &ot.ConcreteTableName,
                &ot.Description, &ot.IsSealed, &ot.Metadata, &ot.CreatedAt, &ot.UpdatedAt,
            )
            if err != nil {
                return fmt.Errorf("failed to scan object type: %w", err)
            }
            types = append(types, ot)
        }
        
        return rows.Err()
    })
    
    if err != nil {
        r.logger.WithError(err).Error("Failed to get child object types")
        return nil, fmt.Errorf("failed to get child object types: %w", err)
    }
    
    return types, nil
}

func (r *ObjectTypeRepository) GetTree(ctx context.Context) ([]models.ObjectType, error) {
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
    err := database.TraceDBQuery(ctx, "object_types", query, func(ctx context.Context) error {
        rows, err := r.db.Query(ctx, query)
        if err != nil {
            return err
        }
        defer rows.Close()
        
        for rows.Next() {
            var ot models.ObjectType
            err := rows.Scan(
                &ot.ID, &ot.Name, &ot.ParentTypeID, &ot.ConcreteTableName,
                &ot.Description, &ot.IsSealed, &ot.Metadata, &ot.CreatedAt, &ot.UpdatedAt,
            )
            if err != nil {
                return fmt.Errorf("failed to scan object type: %w", err)
            }
            types = append(types, ot)
        }
        
        return rows.Err()
    })
    
    if err != nil {
        r.logger.WithError(err).Error("Failed to get object type tree")
        return nil, fmt.Errorf("failed to get object type tree: %w", err)
    }
    
    return types, nil
}

func (r *ObjectTypeRepository) Count(ctx context.Context, filter *models.ObjectTypeFilter) (int64, error) {
    query := `SELECT COUNT(*) FROM object_types WHERE 1=1`
    args := []any{}
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
    err := database.TraceDBQuery(ctx, "object_types", query, func(ctx context.Context) error {
        return r.db.QueryRow(ctx, query, args...).Scan(&count)
    })
    
    if err != nil {
        r.logger.WithError(err).Error("Failed to count object types")
        return 0, fmt.Errorf("failed to count object types: %w", err)
    }
    
    return count, nil
}

// isUniqueViolation checks if error is a unique constraint violation
func isUniqueViolation(err error) bool {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        return pgErr.Code == "23505" // unique_violation
    }
    return false
}
```

---

### 3.2 Create Object Repository

**File**: `internal/repository/object_repository.go`

**Steps**:
1. Create DBInterface (can share with ObjectTypeRepository)
2. Create ObjectRepository struct with logger
3. Implement CRUD methods with public_id support
4. Implement filtering and search
5. Add version checking for optimistic locking
6. Implement soft delete with WithTx for transactions

```go
package repository

import (
    "context"
    "errors"
    "fmt"
    
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/google/uuid"
    "github.com/sirupsen/logrus"
    
    "github.com/v-egorov/service-boilerplate/common/database"
    "github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
)

type ObjectRepository struct {
    db     *pgxpool.Pool  // Need pool for WithTx
    logger *logrus.Logger
}

func NewObjectRepository(db *pgxpool.Pool, logger *logrus.Logger) *ObjectRepository {
    return &ObjectRepository{
        db:     db,
        logger: logger,
    }
}

func (r *ObjectRepository) Create(ctx context.Context, obj *models.Object, userID string) error {
    query := `
        INSERT INTO objects (public_id, object_type_id, parent_object_id, name, description, created_by, updated_by, metadata, status, tags)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING id, created_at, updated_at, version
    `
    
    err := database.TraceDBInsert(ctx, "objects", query, func(ctx context.Context) error {
        return r.db.QueryRow(ctx, query,
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
    })
    
    if err != nil {
        r.logger.WithError(err).Error("Failed to create object")
        return fmt.Errorf("failed to create object: %w", err)
    }
    
    r.logger.WithField("object_id", obj.ID).Info("Object created successfully")
    return nil
}

func (r *ObjectRepository) GetByID(ctx context.Context, id int64) (*models.Object, error) {
    query := `
        SELECT id, public_id, object_type_id, parent_object_id, name, description, created_at, updated_at, deleted_at, version, created_by, updated_by, metadata, status, tags
        FROM objects
        WHERE id = $1
    `
    
    var obj models.Object
    err := database.TraceDBQuery(ctx, "objects", query, func(ctx context.Context) error {
        return r.db.QueryRow(ctx, query, id).Scan(
            &obj.ID, &obj.PublicID, &obj.ObjectTypeID, &obj.ParentObjectID,
            &obj.Name, &obj.Description, &obj.CreatedAt, &obj.UpdatedAt, &obj.DeletedAt,
            &obj.Version, &obj.CreatedBy, &obj.UpdatedBy, &obj.Metadata, &obj.Status, &obj.Tags,
        )
    })
    
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, ErrObjectNotFound
        }
        r.logger.WithError(err).Error("Failed to get object by ID")
        return nil, fmt.Errorf("failed to get object: %w", err)
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
    err := database.TraceDBQuery(ctx, "objects", query, func(ctx context.Context) error {
        return r.db.QueryRow(ctx, query, publicID).Scan(
            &obj.ID, &obj.PublicID, &obj.ObjectTypeID, &obj.ParentObjectID,
            &obj.Name, &obj.Description, &obj.CreatedAt, &obj.UpdatedAt, &obj.DeletedAt,
            &obj.Version, &obj.CreatedBy, &obj.UpdatedBy, &obj.Metadata, &obj.Status, &obj.Tags,
        )
    })
    
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, ErrObjectNotFound
        }
        r.logger.WithError(err).Error("Failed to get object by public ID")
        return nil, fmt.Errorf("failed to get object: %w", err)
    }
    
    return &obj, nil
}

func (r *ObjectRepository) List(ctx context.Context, filter *models.ObjectFilter) ([]models.Object, int64, error) {
    var query string
    var countQuery string
    args := []any{}
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
    err := database.TraceDBQuery(ctx, "objects", countQuery, func(ctx context.Context) error {
        return r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
    })
    if err != nil {
        r.logger.WithError(err).Error("Failed to count objects")
        return nil, 0, fmt.Errorf("failed to count objects: %w", err)
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
    err = database.TraceDBQuery(ctx, "objects", query, func(ctx context.Context) error {
        rows, err := r.db.Query(ctx, query, args...)
        if err != nil {
            return err
        }
        defer rows.Close()
        
        for rows.Next() {
            var obj models.Object
            err := rows.Scan(
                &obj.ID, &obj.PublicID, &obj.ObjectTypeID, &obj.ParentObjectID,
                &obj.Name, &obj.Description, &obj.CreatedAt, &obj.UpdatedAt, &obj.DeletedAt,
                &obj.Version, &obj.CreatedBy, &obj.UpdatedBy, &obj.Metadata, &obj.Status, &obj.Tags,
            )
            if err != nil {
                return fmt.Errorf("failed to scan object: %w", err)
            }
            objects = append(objects, obj)
        }
        
        return rows.Err()
    })
    
    if err != nil {
        r.logger.WithError(err).Error("Failed to list objects")
        return nil, 0, fmt.Errorf("failed to list objects: %w", err)
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
    
    err := database.TraceDBUpdate(ctx, "objects", query, func(ctx context.Context) error {
        return r.db.QueryRow(ctx, query,
            obj.ID,
            obj.Name,
            obj.Description,
            obj.Metadata,
            obj.Status,
            obj.Tags,
            userID,
            obj.Version,
        ).Scan(&obj.UpdatedAt, &obj.Version)
    })
    
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return ErrObjectVersionConflict
        }
        r.logger.WithError(err).Error("Failed to update object")
        return fmt.Errorf("failed to update object: %w", err)
    }
    
    r.logger.WithField("object_id", obj.ID).Info("Object updated successfully")
    return nil
}

func (r *ObjectRepository) SoftDelete(ctx context.Context, id int64, userID string) error {
    query := `
        UPDATE objects
        SET deleted_at = NOW(), status = 'deleted', updated_by = $2
        WHERE id = $1 AND deleted_at IS NULL
    `
    
    var result pgconn.CommandTag
    err := database.TraceDBUpdate(ctx, "objects", query, func(ctx context.Context) error {
        var execErr error
        result, execErr = r.db.Exec(ctx, query, id, userID)
        return execErr
    })
    
    if err != nil {
        r.logger.WithError(err).Error("Failed to soft delete object")
        return fmt.Errorf("failed to soft delete object: %w", err)
    }
    
    if result.RowsAffected() == 0 {
        return ErrObjectNotFound
    }
    
    r.logger.WithField("object_id", id).Info("Object soft deleted successfully")
    return nil
}

func (r *ObjectRepository) HardDelete(ctx context.Context, id int64) error {
    query := `DELETE FROM objects WHERE id = $1`
    
    var result pgconn.CommandTag
    err := database.TraceDBDelete(ctx, "objects", query, func(ctx context.Context) error {
        var execErr error
        result, execErr = r.db.Exec(ctx, query, id)
        return execErr
    })
    
    if err != nil {
        r.logger.WithError(err).Error("Failed to hard delete object")
        return fmt.Errorf("failed to hard delete object: %w", err)
    }
    
    if result.RowsAffected() == 0 {
        return ErrObjectNotFound
    }
    
    r.logger.WithField("object_id", id).Info("Object hard deleted successfully")
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
    err := database.TraceDBQuery(ctx, "objects", query, func(ctx context.Context) error {
        rows, err := r.db.Query(ctx, query, parentID)
        if err != nil {
            return err
        }
        defer rows.Close()
        
        for rows.Next() {
            var obj models.Object
            err := rows.Scan(
                &obj.ID, &obj.PublicID, &obj.ObjectTypeID, &obj.ParentObjectID,
                &obj.Name, &obj.Description, &obj.CreatedAt, &obj.UpdatedAt, &obj.DeletedAt,
                &obj.Version, &obj.CreatedBy, &obj.UpdatedBy, &obj.Metadata, &obj.Status, &obj.Tags,
            )
            if err != nil {
                return fmt.Errorf("failed to scan object: %w", err)
            }
            objects = append(objects, obj)
        }
        
        return rows.Err()
    })
    
    if err != nil {
        r.logger.WithError(err).Error("Failed to get child objects")
        return nil, fmt.Errorf("failed to get child objects: %w", err)
    }
    
    return objects, nil
}

func (r *ObjectRepository) CreateBatch(ctx context.Context, objects []models.Object, userID string) ([]models.Object, []error) {
    // Use WithTx for transaction support with tracing
    var created []models.Object
    var errs []error
    
    err := database.WithTx(ctx, "objects", r.db, func(tx pgx.Tx) error {
        query := `
            INSERT INTO objects (public_id, object_type_id, parent_object_id, name, description, created_by, updated_by, metadata, status, tags)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
            RETURNING id, created_at, updated_at, version
        `
        
        for i := range objects {
            obj := &objects[i]
            
            err := tx.QueryRow(ctx, query,
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
                errs = append(errs, fmt.Errorf("object %d: %w", i, err))
                continue
            }
            
            created = append(created, *obj)
        }
        
        return nil
    })
    
    if err != nil {
        r.logger.WithError(err).Error("Failed to create objects in batch")
        return created, append(errs, err)
    }
    
    r.logger.WithFields(map[string]interface{}{
        "count":      len(created),
        "total":      len(objects),
        "errors":     len(errs),
    }).Info("Objects created in batch")
    
    return created, errs
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

- [ ] Create `internal/repository/object_type_repository.go` with pgx/v5
- [ ] Create `internal/repository/object_repository.go` with pgx/v5
- [ ] Add DBInterface pattern for testability
- [ ] Use database.TraceDB* helpers for all operations
- [ ] Implement proper error handling (pgx.ErrNoRows, constraint violations)
- [ ] Add logging with logrus
- [ ] Use WithTx for batch operations
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
psql postgresql://postgres:password@localhost:5432/objects_service -c "\dt"
psql postgresql://postgres:password@localhost:5432/objects_service -c "SELECT * FROM object_types LIMIT 5;"
```

## Common Issues

**Issue**: pgx.ErrNoRows not recognized
**Solution**: Import `github.com/jackc/pgx/v5` and use `errors.Is(err, pgx.ErrNoRows)`

**Issue**: Recursive query syntax error
**Solution**: Ensure PostgreSQL version is 8.4+ for CTE support

**Issue**: JSONB array operators not working
**Solution**: Use `@>` (contains) or `&&` (overlap) for array operations

**Issue**: Transaction not committing
**Solution**: Use `database.WithTx()` helper which handles commit/rollback automatically

**Issue**: Tracing spans not appearing
**Solution**: Ensure `database.TraceDB*` wrappers are used for all operations

**Issue**: Connection pool errors
**Solution**: Verify pgxpool.Pool is passed to constructors, not pgx.Conn

## Next Phase

Proceed to [Phase 4: Service Layer](phase-04-services.md) once all tasks in this phase are complete.
