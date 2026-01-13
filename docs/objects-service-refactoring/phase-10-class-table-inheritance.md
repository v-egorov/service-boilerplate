# Phase 10: Class Table Inheritance (Future Enhancement)

**Estimated Time**: 8-10 hours
**Status**: ⬜ Not Started
**Dependencies**: All previous phases completed

## Overview

Implement Class Table Inheritance (CTI) pattern to support type-specific database schemas. This provides SQL-level type safety and better query performance for objects with many type-specific fields, while maintaining compatibility with the existing foundation.

## When to Use CTI vs JSONB Metadata

### Use JSONB Metadata (Current Implementation)

✅ **Good for**:
- Types with few type-specific attributes (<5 fields)
- Rapidly evolving schemas
- MVP/early development
- Simple attribute queries
- Low volume of data per type

❌ **Not ideal for**:
- Complex queries on type-specific fields (JOINs needed)
- Type-specific constraints (UNIQUE, CHECK on specific fields)
- Heavy filtering on specific fields
- Performance-critical queries on specific fields

### Use CTI Pattern (This Phase)

✅ **Good for**:
- Types with many type-specific attributes (5+ fields)
- Complex queries on specific fields
- Type-specific constraints
- Performance-critical queries
- Stable schemas (low evolution rate)
- Need for SQL-level type safety

❌ **Not ideal for**:
- Rapidly evolving schemas
- Types with few specific fields
- Simple use cases (JSONB overhead is minimal)

## Architecture

### Table Structure

```sql
-- Generic table (already exists)
objects:
  - id BIGSERIAL PRIMARY KEY
  - public_id UUID
  - object_type_id FK
  -- Other generic fields (name, description, status, etc.)
  - metadata JSONB  -- Still useful for truly dynamic attributes

-- Concrete table (NEW)
products:  -- Named after type, not "product_objects"
  - object_id BIGINT PRIMARY KEY REFERENCES objects(id) ON DELETE CASCADE
  - sku VARCHAR(100) NOT NULL
  - price DECIMAL(10,2) NOT NULL
  - inventory_count INTEGER DEFAULT 0
  -- Other product-specific fields...

-- Another concrete table example
articles:
  - object_id BIGINT PRIMARY KEY REFERENCES objects(id) ON DELETE CASCADE
  - slug VARCHAR(255) NOT NULL
  - word_count INTEGER DEFAULT 0
  -- Other article-specific fields...
```

### Query Pattern with CTE

```sql
-- Get product with generic + specific fields
WITH products AS (
    SELECT 
        o.id,
        o.public_id,
        o.name,
        o.description,
        o.status,
        o.metadata,
        o.created_at,
        o.updated_at,
        -- Product-specific fields
        p.sku,
        p.price,
        p.inventory_count
    FROM objects o
    INNER JOIN products p ON o.id = p.object_id
    WHERE o.object_type_id = 1  -- Product type ID
)
SELECT * FROM products;
```

## Tasks

### 10.1 Create Concrete Tables Registry

**File**: `migrations/000010_create_concrete_tables_registry.up.sql`

```sql
-- Registry to track which object types have concrete tables
CREATE TABLE concrete_tables_registry (
    id BIGSERIAL PRIMARY KEY,
    object_type_id BIGINT NOT NULL REFERENCES object_types(id) ON DELETE CASCADE,
    table_name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT unique_type UNIQUE (object_type_id)
);

CREATE INDEX idx_concrete_tables_type ON concrete_tables_registry(object_type_id);
COMMENT ON TABLE concrete_tables_registry IS 'Registry of concrete tables for class table inheritance pattern';
```

### 10.2 Update ObjectType Model

**File**: `internal/models/object_type.go`

Add fields for concrete table tracking:

```go
type ObjectType struct {
    ID                 int64     `json:"id" db:"id"`
    Name               string    `json:"name" db:"name"`
    ParentTypeID       *int64    `json:"parent_type_id,omitempty" db:"parent_type_id"`
    ConcreteTableName  *string   `json:"concrete_table_name,omitempty" db:"concrete_table_name"`
    Description        *string   `json:"description,omitempty" db:"description"`
    IsSealed           bool      `json:"is_sealed" db:"is_sealed"`
    Metadata           Metadata  `json:"metadata" db:"metadata"`
    CreatedAt          time.Time `json:"created_at" db:"created_at"`
    UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
    
    // Relationships
    ParentType   *ObjectType `json:"parent_type,omitempty"`
    Children     []ObjectType `json:"children,omitempty"`
    Objects      []Object     `json:"objects,omitempty"`
    
    // NEW: Concrete table info
    HasConcreteTable bool      `json:"has_concrete_table" db:"-"`
    ConcreteSchema   *string   `json:"concrete_schema,omitempty" db:"-"`
}
```

### 10.3 Create Concrete Table Repository

**File**: `internal/repository/concrete_table_repository.go`

```go
package repository

import (
    "context"
    "fmt"
    
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/sirupsen/logrus"
    
    "github.com/v-egorov/service-boilerplate/common/database"
)

type ConcreteTableRepository struct {
    db     *pgxpool.Pool
    logger *logrus.Logger
}

func NewConcreteTableRepository(db *pgxpool.Pool, logger *logrus.Logger) *ConcreteTableRepository {
    return &ConcreteTableRepository{
        db:     db,
        logger: logger,
    }
}

// RegisterConcreteTable registers a concrete table for an object type
func (r *ConcreteTableRepository) RegisterConcreteTable(ctx context.Context, objectTypeID int64, tableName string, schema string) error {
    query := `
        INSERT INTO concrete_tables_registry (object_type_id, table_name)
        VALUES ($1, $2)
        ON CONFLICT (object_type_id) 
        DO UPDATE SET table_name = EXCLUDED.table_name
    `
    
    return database.TraceDBInsert(ctx, "concrete_tables_registry", query, func(ctx context.Context) error {
        _, err := r.db.Exec(ctx, query, objectTypeID, tableName)
        return err
    })
}

// GetConcreteTableName returns the concrete table name for an object type
func (r *ConcreteTableRepository) GetConcreteTableName(ctx context.Context, objectTypeID int64) (*string, error) {
    query := `
        SELECT table_name
        FROM concrete_tables_registry
        WHERE object_type_id = $1
    `
    
    var tableName string
    err := database.TraceDBQuery(ctx, "concrete_tables_registry", query, func(ctx context.Context) error {
        return r.db.QueryRow(ctx, query, objectTypeID).Scan(&tableName)
    })
    
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, nil // No concrete table exists
        }
        return nil, fmt.Errorf("failed to get concrete table name: %w", err)
    }
    
    return &tableName, nil
}

// ListConcreteTables returns all registered concrete tables
func (r *ConcreteTableRepository) ListConcreteTables(ctx context.Context) (map[int64]string, error) {
    query := `
        SELECT object_type_id, table_name
        FROM concrete_tables_registry
    `
    
    rows, err := database.TraceDBQuery(ctx, "concrete_tables_registry", query, func(ctx context.Context) error {
        return r.db.Query(ctx, query)
    })
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    result := make(map[int64]string)
    for rows.Next() {
        var typeID int64
        var tableName string
        if err := rows.Scan(&typeID, &tableName); err != nil {
            return err
        }
        result[typeID] = tableName
    }
    
    return result, rows.Err()
}

// UnregisterConcreteTable removes a concrete table registration
func (r *ConcreteTableRepository) UnregisterConcreteTable(ctx context.Context, objectTypeID int64) error {
    query := `DELETE FROM concrete_tables_registry WHERE object_type_id = $1`
    
    return database.TraceDBDelete(ctx, "concrete_tables_registry", query, func(ctx context.Context) error {
        _, err := r.db.Exec(ctx, query, objectTypeID)
        return err
    })
}
```

### 10.4 Enhance ObjectRepository with CTI Support

**File**: `internal/repository/object_repository.go` (enhancement)

Add methods for CTI queries:

```go
// GetWithConcreteFields retrieves an object with type-specific fields using CTE
func (r *ObjectRepository) GetWithConcreteFields(ctx context.Context, publicID uuid.UUID, concreteTable string, concreteTableType interface{}) (*ConcreteObject, error) {
    // Build CTE query dynamically based on concrete table structure
    query := fmt.Sprintf(`
        WITH object_data AS (
            SELECT 
                o.id, o.public_id, o.object_type_id, o.parent_object_id,
                o.name, o.description, o.created_at, o.updated_at, 
                o.deleted_at, o.version, o.created_by, o.updated_by,
                o.status, o.tags,
                -- Dynamic concrete fields (example, would use reflection)
                %s
            FROM objects o
            INNER JOIN %s c ON o.id = c.object_id
            WHERE o.public_id = $1
        )
        SELECT * FROM object_data
    `, r.buildConcreteFieldsSelect(concreteTableType), concreteTable)
    
    var obj ConcreteObject
    err := database.TraceDBQuery(ctx, "objects", query, func(ctx context.Context) error {
        rows, err := r.db.Query(ctx, query, publicID)
        if err != nil {
            return err
        }
        defer rows.Close()
        
        if !rows.Next() {
            return pgx.ErrNoRows
        }
        
        return r.scanConcreteObject(rows, concreteTableType, &obj)
    })
    
    if err != nil {
        return nil, err
    }
    
    return &obj, nil
}

// ListWithConcreteFields retrieves objects with type-specific fields
func (r *ObjectRepository) ListWithConcreteFields(ctx context.Context, filter *models.ObjectFilter, concreteTable string) ([]ConcreteObject, error) {
    // Similar implementation with WHERE clauses from filter
    query := fmt.Sprintf(`
        WITH object_data AS (
            SELECT 
                o.id, o.public_id, o.object_type_id, o.name, o.status, o.tags,
                %s
            FROM objects o
            INNER JOIN %s c ON o.id = c.object_id
            WHERE o.deleted_at IS NULL
            -- Add filter conditions here
            ORDER BY o.created_at DESC
        )
        SELECT * FROM object_data
        LIMIT $1 OFFSET $2
    `, r.buildConcreteFieldsSelect(nil), concreteTable)
    
    var objects []ConcreteObject
    err := database.TraceDBQuery(ctx, "objects", query, func(ctx context.Context) error {
        rows, err := r.db.Query(ctx, query, filter.PageSize, (filter.Page-1)*filter.PageSize)
        if err != nil {
            return err
        }
        defer rows.Close()
        
        for rows.Next() {
            var obj ConcreteObject
            if err := r.scanConcreteObject(rows, nil, &obj); err != nil {
                return err
            }
            objects = append(objects, obj)
        }
        
        return rows.Err()
    })
    
    return objects, err
}

// buildConcreteFieldsSelect dynamically builds SELECT clause for concrete fields
func (r *ObjectRepository) buildConcreteFieldsSelect(concreteType interface{}) string {
    // Use reflection to dynamically build field list
    // Example: "c.sku, c.price, c.inventory_count"
    // This is a simplified example - production would need proper reflection
    if concreteType == nil {
        return "c.*"
    }
    // Implementation would inspect concreteType and build field list
    return "c.*"
}

func (r *ObjectRepository) scanConcreteObject(rows pgx.Rows, concreteType interface{}, obj *ConcreteObject) error {
    // Dynamic scanning based on concrete type
    // This is simplified - production would need proper reflection
    return rows.Scan(
        &obj.ID, &obj.PublicID, &obj.ObjectTypeID,
        &obj.Name, &obj.Description, &obj.Status, &obj.Tags,
        &obj.CreatedAt, &obj.UpdatedAt,
        &obj.Metadata, // Still includes generic metadata
        &obj.ConcreteFields, // Type-specific fields as map or interface
    )
}
```

### 10.5 Create Concrete Object Models

**File**: `internal/models/concrete_objects.go`

```go
package models

import (
    "time"
    
    "github.com/google/uuid"
)

// ConcreteObject represents an object with type-specific fields
type ConcreteObject struct {
    // Generic object fields
    ID             int64     `json:"id" db:"id"`
    PublicID       uuid.UUID `json:"public_id" db:"public_id"`
    ObjectTypeID   int64     `json:"object_type_id" db:"object_type_id"`
    Name           string    `json:"name" db:"name"`
    Description    *string   `json:"description,omitempty" db:"description"`
    Status         string    `json:"status" db:"status"`
    Tags           []string  `json:"tags" db:"tags"`
    CreatedAt      time.Time `json:"created_at" db:"created_at"`
    UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
    
    // Generic metadata (still useful for truly dynamic fields)
    Metadata       Metadata  `json:"metadata" db:"metadata"`
    
    // Concrete fields (dynamic)
    ConcreteFields interface{} `json:"concrete_fields" db:"-"`
}

// ProductConcreteObject extends ConcreteObject with product-specific fields
type ProductConcreteObject struct {
    ConcreteObject
    SKU            string  `json:"sku"`
    Price          float64 `json:"price"`
    InventoryCount int     `json:"inventory_count"`
}

// ArticleConcreteObject extends ConcreteObject with article-specific fields
type ArticleConcreteObject struct {
    ConcreteObject
    Slug      string `json:"slug"`
    WordCount int    `json:"word_count"`
}

// ToConcrete converts ConcreteObject to type-specific struct
func ToConcreteObject[T any](base *ConcreteObject, concrete T) interface{} {
    // Helper for type conversion
    return struct {
        *T
        ConcreteObject
    }{
        ConcreteObject: *base,
    }
}
```

### 10.6 Update Object Service with CTI Support

**File**: `internal/services/object_service.go` (enhancement)

```go
// Add new method to ObjectService
func (s *ObjectService) GetWithConcreteFields(ctx context.Context, publicID uuid.UUID) (interface{}, error) {
    obj, err := s.objectRepo.GetByPublicID(ctx, publicID)
    if err != nil {
        return nil, err
    }
    
    objType, err := s.objectTypeRepo.GetByID(ctx, obj.ObjectTypeID)
    if err != nil {
        return nil, err
    }
    
    // Check if this type has a concrete table
    if objType.ConcreteTableName == nil || *objType.ConcreteTableName == "" {
        // No concrete table, return standard object
        return s.toResponse(obj), nil
    }
    
    // Has concrete table - fetch with CTE query
    concreteObj, err := s.objectRepo.GetWithConcreteFields(ctx, publicID, *objType.ConcreteTableName, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to get object with concrete fields: %w", err)
    }
    
    // Return based on type
    switch objType.Name {
    case "Product":
        return &ProductConcreteObject{
            ConcreteObject: ConcreteObject{
                ID:          concreteObj.ID,
                PublicID:    concreteObj.PublicID,
                ObjectTypeID: concreteObj.ObjectTypeID,
                Name:        concreteObj.Name,
                Description: concreteObj.Description,
                Status:      concreteObj.Status,
                Tags:        concreteObj.Tags,
                CreatedAt:   concreteObj.CreatedAt,
                UpdatedAt:   concreteObj.UpdatedAt,
                Metadata:    concreteObj.Metadata,
            },
        }, nil
    
    case "Article":
        return &ArticleConcreteObject{
            ConcreteObject: ConcreteObject{
                ID:          concreteObj.ID,
                PublicID:    concreteObj.PublicID,
                ObjectTypeID: concreteObj.ObjectTypeID,
                Name:        concreteObj.Name,
                Description: concreteObj.Description,
                Status:      concreteObj.Status,
                Tags:        concreteObj.Tags,
                CreatedAt:   concreteObj.CreatedAt,
                UpdatedAt:   concreteObj.UpdatedAt,
                Metadata:    concreteObj.Metadata,
            },
        }, nil
    
    default:
        // Unknown concrete type, return generic concrete object
        return concreteObj, nil
    }
}
```

### 10.7 Update Handlers for Concrete Objects

**File**: `internal/handlers/object_handler.go` (enhancement)

Add endpoint parameter to include concrete fields:

```go
// GET /api/v1/objects/{public_id}?include_concrete=true
func (h *ObjectHandler) GetByPublicID(c *gin.Context) {
    publicID, err := uuid.Parse(c.Param("public_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid public_id"})
        return
    }
    
    includeConcrete := c.Query("include_concrete") == "true"
    
    if includeConcrete {
        obj, err := h.service.GetWithConcreteFields(c.Request.Context(), publicID)
        if err != nil {
            handleObjectError(c, err)
            return
        }
        c.JSON(http.StatusOK, obj)
        return
    }
    
    // Standard object retrieval
    obj, err := h.service.GetByPublicID(c.Request.Context(), publicID)
    if err != nil {
        handleObjectError(c, err)
        return
    }
    c.JSON(http.StatusOK, obj)
}

// GET /api/v1/object-types/{id}/concrete-table
func (h *ObjectHandler) GetConcreteTableInfo(c *gin.Context) {
    typeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    
    tableName, err := h.concreteRepo.GetConcreteTableName(c.Request.Context(), typeID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get concrete table"})
        return
    }
    
    if tableName == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "no concrete table for this type"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "object_type_id": typeID,
        "concrete_table": *tableName,
    })
}
```

## Migration: From JSONB to CTI

When migrating an existing object type from JSONB to CTI:

1. **Create concrete table** with type-specific columns
2. **Copy data** from `metadata` JSONB to concrete columns
3. **Register** concrete table in registry
4. **Keep** metadata field for any remaining dynamic attributes
5. **Update API** to use concrete fields

```sql
-- Migration example: Product type
BEGIN;

-- Create concrete table
CREATE TABLE products (
    object_id BIGINT PRIMARY KEY REFERENCES objects(id) ON DELETE CASCADE,
    sku VARCHAR(100) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    inventory_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Migrate data from JSONB
INSERT INTO products (object_id, sku, price, inventory_count)
SELECT 
    o.id,
    o.metadata->>'sku',
    (o.metadata->>'price')::DECIMAL(10,2),
    (o.metadata->>'inventory_count')::INTEGER
FROM objects o
JOIN object_types ot ON o.object_type_id = ot.id
WHERE ot.name = 'Product'
  AND o.metadata ? 'sku'  -- Only migrate objects with relevant metadata
  AND o.metadata ? 'price';

-- Register concrete table
INSERT INTO concrete_tables_registry (object_type_id, table_name)
SELECT id, 'products'
FROM object_types
WHERE name = 'Product';

COMMIT;
```

## Decision Criteria for CTI

Migrate an object type to CTI when:
- ✅ Has 5+ type-specific fields
- ✅ Requires complex queries on specific fields
- ✅ Needs SQL-level constraints on specific fields
- ✅ Performance-critical queries on specific fields
- ✅ Schema is stable (low evolution rate)

Stay with JSONB when:
- ✅ Has <5 type-specific fields
- ✅ Schema rapidly evolving
- ✅ Simple query patterns
- ✅ Low data volume
- ✅ MVP/early development

## Checklist

- [ ] Create `migrations/000010_create_concrete_tables_registry.up.sql`
- [ ] Create `internal/models/concrete_objects.go`
- [ ] Create `internal/repository/concrete_table_repository.go`
- [ ] Update `internal/repository/object_repository.go` with CTI methods
- [ ] Update `internal/services/object_service.go` with concrete field support
- [ ] Update `internal/handlers/object_handler.go` with concrete endpoints
- [ ] Create example concrete table migration (products)
- [ ] Add CTI query builder/reflection utilities
- [ ] Write migration tool: JSONB → CTI
- [ ] Update documentation with CTI pattern
- [ ] Create decision criteria document
- [ ] Add tests for CTI queries
- [ ] Add tests for JSONB → CTI migration
- [ ] Update progress.md

## Testing

```bash
# Test CTI registry
go test ./internal/repository/concrete_table_repository_test.go -v

# Test concrete object queries
go test ./internal/repository/object_repository_test.go -v -run TestCTI

# Test CTI endpoints
curl "http://localhost:8085/api/v1/objects/{uuid}?include_concrete=true"

# Test concrete table info endpoint
curl "http://localhost:8085/api/v1/object-types/1/concrete-table"
```

## Common Issues

**Issue**: Reflection-based query building is complex
**Solution**: Consider code generation or explicit type registration

**Issue**: JSONB to CTI migration loses data
**Solution**: Keep metadata field, use selective extraction, verify before commit

**Issue**: Dynamic type marshaling is error-prone
**Solution**: Use code generation or strict type mapping registry

**Issue**: CTE performance with many joins
**Solution**: Add proper indexes on concrete tables, limit columns in SELECT

## Benefits

1. **Performance**: Native column access is faster than JSONB parsing
2. **Type Safety**: Compile-time type checking for concrete fields
3. **Constraints**: Can add UNIQUE, CHECK, FK constraints on specific fields
4. **Queries**: Standard SQL queries (no JSONB operators needed)
5. **Incremental**: Can migrate types to CTI as needed (no all-or-nothing)

## Backward Compatibility

- ✅ JSONB metadata field retained
- ✅ Generic Object API still works
- ✅ Concrete fields are additive (via query parameter)
- ✅ Existing data not affected until migration runs

## Estimated Effort

- Initial implementation: 8-10 hours
- First type migration: 2-4 hours per type
- Testing: 2-3 hours
- Documentation: 1-2 hours

**Total for first type: ~12-17 hours**

## Next Steps

1. Review this phase document
2. Identify which object types benefit from CTI
3. Create concrete table(s) for identified types
4. Run JSONB → CTI migration for those types
5. Update API handlers to use concrete fields
6. Monitor performance improvements
7. Expand CTI to more types as needed

## Integration with Previous Phases

This phase builds on the solid foundation from Phases 1-9:
- **Phase 1** - Schema already has `concrete_table_name` field
- **Phase 2** - Models support dynamic fields
- **Phase 3** - Repository pattern supports CTE queries
- **Phase 4** - Service layer handles type selection
- **Phase 5** - Handlers already have parameter support

No changes needed to previous phases - this is a forward-looking enhancement.
