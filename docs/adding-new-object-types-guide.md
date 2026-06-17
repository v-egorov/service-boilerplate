# How to Add a New Object Type

This guide provides step-by-step instructions for adding new object types to the objects-service. There are two scenarios depending on whether your type needs its own concrete table (CTI) or can use the base `objects` table with JSONB metadata only.

## Prerequisites

- Understand the unified ownership model: all resources check `created_by == userID` for `:own` scope
- Permission middleware uses data-driven construction from `RouteConfig{TypeKey, HTTPMethod}` — no hardcoded permission strings in routes
- Shared handler utilities are available in `services/objects-service/internal/handlers/base.go`

---

## Scenario A: No CTI table needed (simple types with JSONB metadata only)

Use this scenario when your object type's data fits entirely in the base `objects` table plus JSONB metadata fields. Example: a "Document" type that needs name, description, status, and arbitrary metadata stored as JSONB.

### Step 1: Register permissions in auth-service

Insert permission records for each CRUD action into `auth_service.permissions`:

```sql
-- In services/auth-service/migrations/development/YYYYNN_new_document_permissions.up.sql
INSERT INTO auth_service.permissions (name, resource, action) VALUES
    ('document:create', 'documents', 'create'),
    ('document:read:all', 'documents', 'read'),
    ('document:read:own', 'documents', 'read'),
    ('document:update:all', 'documents', 'update'),
    ('document:update:own', 'documents', 'update'),
    ('document:delete:all', 'documents', 'delete'),
    ('document:delete:own', 'documents', 'delete')
ON CONFLICT (name) DO NOTHING;

-- Assign permissions to roles via auth_service.role_permissions
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
CROSS JOIN auth_service.permissions p
WHERE r.name = 'admin' AND p.resource = 'documents'
ON CONFLICT DO NOTHING;

-- Grant user role basic access to own documents
INSERT INTO auth_service.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM auth_service.roles r
JOIN auth_service.permissions p ON p.name IN ('document:read:own', 'document:create', 'document:update:own', 'document:delete:own')
WHERE r.name = 'user'
ON CONFLICT DO NOTHING;
```

**Note:** `create` uses flat permission names on objects/types (no scoped variants) — you always own what you create. Only relationships use scoped `create` variants (`create:own`, `create:all`).

### Step 2: Add type_key to object_types table

Register the new type in the object type registry:

```sql
-- In services/objects-service/migrations/development/YYYYNN_add_document_type.up.sql
INSERT INTO objects_service.object_types (name, description, type_key, is_sealed, metadata, created_at, updated_at)
VALUES ('Document', 'Simple document type with JSONB metadata', 'documents', false, '{}', NOW(), NOW());
```

### Step 3: Create model layer

Create `services/objects-service/internal/models/document.go`:

```go
package models

import "time"

// Document represents a simple document object stored in the base objects table.
type Document struct {
    ObjectID     int64             `json:"id"`
    PublicID     string            `json:"public_id"`
    Name         string            `json:"name"`
    Description  string            `json:"description"`
    Status       string            `json:"status"`
    Metadata     map[string]any    `json:"metadata,omitempty"`
    CreatedBy    string            `json:"created_by"`
    UpdatedBy    string            `json:"updated_by,omitempty"`
    CreatedAt    time.Time         `json:"created_at"`
    UpdatedAt    *time.Time        `json:"updated_at,omitempty"`
}

// CreateDocumentRequest is the DTO for creating a document.
type CreateDocumentRequest struct {
    Name        string            `json:"name" binding:"required,min=1,max=255"`
    Description string            `json:"description"`
    Metadata    map[string]any    `json:"metadata,omitempty"`
}

// UpdateDocumentRequest is the DTO for updating a document.
type UpdateDocumentRequest struct {
    Name        string  `json:"name" binding:"omitempty,min=1,max=255"`
    Description *string `json:"description"`
    Metadata    map[string]any `json:"metadata,omitempty"`
}
```

### Step 4: Create repository layer

Create `services/objects-service/internal/repository/document_repository.go`:

```go
package repository

import (
    "context"
    "database/sql"

    "github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
)

type documentRepository struct {
    db         *PGDatabase
    repoOptions RepositoryOptions
}

func NewDocumentRepository(db *PGDatabase, opts RepositoryOptions) *documentRepository {
    return &documentRepository{db: db, repoOptions: opts}
}

// Create inserts a new document into the base objects table.
func (r *documentRepository) Create(ctx context.Context, obj *models.Object, userID string) (*models.Document, error) {
    query := `
        INSERT INTO objects_service.objects 
            (name, description, status, metadata, object_type_id, created_by, updated_by)
        VALUES ($1, $2, $3, $4, 
                (SELECT id FROM objects_service.object_types WHERE type_key = 'documents'),
                $5, $6)
        RETURNING id, public_id, name, description, status, metadata, created_at`

    var obj models.Object
    err := r.db.QueryRow(ctx, query, obj.Name, obj.Description, "active", obj.Metadata, userID, userID).Scan(
        &obj.ID, &obj.PublicID, &obj.Name, &obj.Description, &obj.Status, &obj.Metadata, &obj.CreatedAt)

    if err != nil {
        return nil, err
    }

    return r.getByPublicID(ctx, obj.PublicID)
}

// GetByID retrieves a document by its ID.
func (r *documentRepository) GetByID(ctx context.Context, id int64) (*models.Document, error) {
    query := `
        SELECT id, public_id, name, description, status, metadata, created_by, updated_by, created_at
        FROM objects_service.objects
        WHERE id = $1 
          AND object_type_id = (SELECT id FROM objects_service.object_types WHERE type_key = 'documents')
          AND deleted_at IS NULL`

    return r.getObjectRow(ctx, query, id)
}

// List retrieves documents with optional filtering.
func (r *documentRepository) List(ctx context.Context, filter *models.DocumentFilter) ([]*models.Document, int64, error) {
    // Build dynamic query based on filter criteria
    qb := newQueryBuilder("objects_service.objects")
    qb.Where("object_type_id = (SELECT id FROM objects_service.object_types WHERE type_key = 'documents')")
    
    if filter.UserID != nil && *filter.UserID != "" {
        qb.Where("created_by = $1", *filter.UserID)
    }

    // ... apply other filters, pagination
    
    return r.getObjects(ctx, query, args)
}

// Update modifies a document.
func (r *documentRepository) Update(ctx context.Context, id int64, updates *models.UpdateDocumentRequest, userID string) (*models.Document, error) {
    // ... implementation similar to ObjectRepository.Update
}

// Delete soft-deletes a document.
func (r *documentRepository) Delete(ctx context.Context, id int64) error {
    query := `UPDATE objects_service.objects SET deleted_at = NOW() WHERE id = $1 AND object_type_id = (SELECT id FROM objects_service.object_types WHERE type_key = 'documents')`
    _, err := r.db.Exec(ctx, query, id)
    return err
}

// Helper: get a document by public ID.
func (r *documentRepository) getByPublicID(ctx context.Context, publicID string) (*models.Document, error) {
    // ... implementation
}

// Helper: scan a row into Document struct.
func (r *documentRepository) getObjectRow(ctx context.Context, query string, args ...any) (*models.Document, error) {
    var obj models.Object
    err := r.db.QueryRow(ctx, query, args...).Scan(
        &obj.ID, &obj.PublicID, &obj.Name, &obj.Description, &obj.Status, 
        &obj.Metadata, &obj.CreatedBy, &obj.UpdatedBy, &obj.CreatedAt)

    if err != nil {
        return nil, err
    }

    doc := &models.Document{ObjectID: obj.ID, PublicID: obj.PublicID, Name: obj.Name}
    doc.Status = obj.Status
    // ... map remaining fields
    return doc, nil
}
```

### Step 5: Create service layer

Create `services/objects-service/internal/services/document_service.go`:

```go
package services

import (
    "context"
    "fmt"

    "github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
    "github.com/v-egorov/service-boilerplate/services/objects-service/internal/repository"
)

type DocumentService struct {
    repo repository.DocumentRepository
}

func NewDocumentService(repo repository.DocumentRepository) *DocumentService {
    return &DocumentService{repo: repo}
}

// Create creates a new document with ownership tracking.
func (s *DocumentService) Create(ctx context.Context, req *models.CreateDocumentRequest, userID string) (*models.Document, error) {
    obj := &models.Object{Name: req.Name, Description: req.Description, Metadata: req.Metadata}
    
    doc, err := s.repo.Create(ctx, obj, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to create document: %w", err)
    }

    return doc, nil
}

// GetByID retrieves a document by ID.
func (s *DocumentService) GetByID(ctx context.Context, id int64) (*models.Document, error) {
    doc, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get document: %w", err)
    }

    return doc, nil
}

// List retrieves documents with filtering.
func (s *DocumentService) List(ctx context.Context, filter *models.DocumentFilter) ([]*models.Document, int64, error) {
    docs, total, err := s.repo.List(ctx, filter)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to list documents: %w", err)
    }

    return docs, total, nil
}

// Update modifies a document.
func (s *DocumentService) Update(ctx context.Context, id int64, req *models.UpdateDocumentRequest, userID string) (*models.Document, error) {
    doc, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get document: %w", err)
    }

    // Business validation: check ownership for :own scope
    if doc.CreatedBy != userID {
        return nil, fmt.Errorf("permission denied: you can only update your own documents")
    }

    updatedDoc, err := s.repo.Update(ctx, id, req, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to update document: %w", err)
    }

    return updatedDoc, nil
}

// Delete soft-deletes a document.
func (s *DocumentService) Delete(ctx context.Context, id int64, userID string) error {
    doc, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to get document: %w", err)
    }

    // Business validation: check ownership for :own scope
    if doc.CreatedBy != userID {
        return fmt.Errorf("permission denied: you can only delete your own documents")
    }

    return s.repo.Delete(ctx, id)
}
```

### Step 6: Create handler layer

Create `services/objects-service/internal/handlers/document_handler.go`:

```go
package handlers

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
    "github.com/v-egorov/service-boilerplate/common/middleware"
    "github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
    "github.com/v-egorov/service-boilerplate/services/objects-service/internal/services"
)

type DocumentHandler struct {
    service *services.DocumentService
    logger  *logrus.Logger
}

func NewDocumentHandler(service *services.DocumentService, logger *logrus.Logger) *DocumentHandler {
    return &DocumentHandler{service: service, logger: logger}
}

// Create handles POST /documents.
func (h *DocumentHandler) Create(c *gin.Context) {
    requestID := c.GetHeader("X-Request-ID")
    userID := middleware.GetAuthenticatedUserID(c)

    var req models.CreateDocumentRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.logger.WithFields(logrus.Fields{
            "request_id": requestID,
        }).WithError(err).Error("Invalid request body")
        
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid request format: failed to parse request body",
            "type":  "validation_error",
            "meta":  gin.H{"request_id": requestID},
        })
        return
    }

    doc, err := h.service.Create(c.Request.Context(), &req, userID)
    if err != nil {
        handleServiceError(c, err, "Failed to create document", requestID)
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "data":    doc,
        "message": "Document created successfully",
        "meta":    gin.H{"request_id": requestID},
    })
}

// GetByID handles GET /documents/:id.
func (h *DocumentHandler) GetByID(c *gin.Context) {
    requestID := c.GetHeader("X-Request-ID")
    userID := middleware.GetAuthenticatedUserID(c)

    idStr := c.Param("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil || id <= 0 {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid id format: id must be a positive integer",
            "type":  "validation_error",
            "field": "id",
            "meta":  gin.H{"request_id": requestID},
        })
        return
    }

    doc, err := h.service.GetByID(c.Request.Context(), id)
    if err != nil {
        handleServiceError(c, err, "Failed to get document", requestID)
        return
    }

    // Check ownership for :own scope (admin and :all bypass this check)
    matchedPermissions := c.GetStringSlice("matched_permissions")
    if !checkOwnership(c, doc.CreatedBy) {
        c.JSON(http.StatusForbidden, gin.H{
            "error": "You can only access your own documents",
            "type":  "permission_denied",
            "meta":  gin.H{"request_id": requestID},
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "data":   doc,
        "meta":   gin.H{"request_id": requestID},
    })
}

// List handles GET /documents with optional filters.
func (h *DocumentHandler) List(c *gin.Context) {
    requestID := c.GetHeader("X-Request-ID")

    filter := &models.DocumentFilter{}

    if status := c.Query("status"); status != "" {
        filter.Status = status
    }

    // Pass authenticated user ID for :own scope filtering at repo level
    userID := middleware.GetAuthenticatedUserID(c)
    if userID != "" {
        filter.UserID = &userID
    }

    docs, total, err := h.service.List(c.Request.Context(), filter)
    if err != nil {
        handleServiceError(c, err, "Failed to list documents", requestID)
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "data":       docs,
        "pagination": gin.H{"limit": 50, "offset": 0, "total": total, "count": len(docs)},
        "meta":       gin.H{"request_id": requestID},
    })
}

// Update handles PUT /documents/:id.
func (h *DocumentHandler) Update(c *gin.Context) {
    requestID := c.GetHeader("X-Request-ID")
    userID := middleware.GetAuthenticatedUserID(c)

    idStr := c.Param("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil || id <= 0 {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid id format: id must be a positive integer",
            "type":  "validation_error",
            "field": "id",
            "meta":  gin.H{"request_id": requestID},
        })
        return
    }

    var req models.UpdateDocumentRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid request format: failed to parse request body",
            "type":  "validation_error",
            "meta":  gin.H{"request_id": requestID},
        })
        return
    }

    doc, err := h.service.Update(c.Request.Context(), id, &req, userID)
    if err != nil {
        handleServiceError(c, err, "Failed to update document", requestID)
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "data":    doc,
        "message": "Document updated successfully",
        "meta":    gin.H{"request_id": requestID},
    })
}

// Delete handles DELETE /documents/:id.
func (h *DocumentHandler) Delete(c *gin.Context) {
    requestID := c.GetHeader("X-Request-ID")
    userID := middleware.GetAuthenticatedUserID(c)

    idStr := c.Param("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil || id <= 0 {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid id format: id must be a positive integer",
            "type":  "validation_error",
            "field": "id",
            "meta":  gin.H{"request_id": requestID},
        })
        return
    }

    err = h.service.Delete(c.Request.Context(), id, userID)
    if err != nil {
        handleServiceError(c, err, "Failed to delete document", requestID)
        return
    }

    c.JSON(http.StatusNoContent, nil)
}
```

### Step 7: Wire in main.go

Add the new service and route groups using the current `perm()` helper + RouteConfig pattern from `services/objects-service/cmd/main.go`:

```go
// In the initialization block (around line 100-120):
documentRepo := repository.NewDocumentRepository(pgDatabase, repoOptions)
documentService := services.NewDocumentService(documentRepo)
documentHandler := handlers.NewDocumentHandler(documentService, logger.Logger)

// In the route groups section:
documentsRead := v1.Group("/documents")
documentsRead.Use(perm(permiddleware.RouteConfig{TypeKey: "documents", HTTPMethod: "GET"}))
{
    documentsRead.GET("/:id", documentHandler.GetByID)
    documentsRead.GET("", documentHandler.List)
}

documentsCreate := v1.Group("/documents")
documentsCreate.Use(perm(permiddleware.RouteConfig{TypeKey: "documents", HTTPMethod: "POST"}))
{
    documentsCreate.POST("", documentHandler.Create)
}

documentsUpdate := v1.Group("/documents")
documentsUpdate.Use(perm(permiddleware.RouteConfig{TypeKey: "documents", HTTPMethod: "PUT"}))
{
    documentsUpdate.PUT("/:id", documentHandler.Update)
}

documentsDelete := v1.Group("/documents")
documentsDelete.Use(perm(permiddleware.RouteConfig{TypeKey: "documents", HTTPMethod: "DELETE"}))
{
    documentsDelete.DELETE("/:id", documentHandler.Delete)
}
```

### Step 8: Add migration rollback (.down.sql)

Create the corresponding `.down.sql` file to clean up permissions and type registration when rolling back:

```sql
-- In services/auth-service/migrations/development/YYYYNN_new_document_permissions.down.sql
DELETE FROM auth_service.role_permissions rp
JOIN auth_service.permissions p ON rp.permission_id = p.id
WHERE p.resource = 'documents';

DELETE FROM auth_service.permissions WHERE resource = 'documents';

-- In services/objects-service/migrations/development/YYYYNN_add_document_type.down.sql
DELETE FROM objects_service.object_types WHERE type_key = 'documents';
```

---

## Scenario B: CTI table needed (types with type-specific columns)

Use this scenario when your object type needs its own concrete table alongside the base `objects` table, following the same pattern as relationships. Example: a "Product" type needing `sku`, `price`, `inventory_count` as native SQL columns instead of JSONB metadata.

### Step 1: Create CTI table migration

Create `services/objects-service/migrations/{development,staging,production}/NNNNN_create_products_cti.up.sql`:

```sql
CREATE TABLE objects_service.products (
    object_id BIGINT PRIMARY KEY REFERENCES objects_service.objects(id) ON DELETE CASCADE,
    sku VARCHAR(100) NOT NULL UNIQUE,
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    inventory_count INTEGER DEFAULT 0 CHECK (inventory_count >= 0),
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE objects_service.products IS 'CTI concrete table for Product instances';

CREATE INDEX idx_products_sku ON objects_service.products(sku);

-- Add trigger to update the updated_at timestamp
CREATE OR REPLACE FUNCTION objects_service.update_product_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_product_updated_at
    BEFORE UPDATE ON objects_service.products
    FOR EACH ROW EXECUTE FUNCTION objects_service.update_product_updated_at_column();
```

Create the corresponding `.down.sql`:

```sql
DROP TRIGGER IF EXISTS update_product_updated_at ON objects_service.products;
DROP FUNCTION IF EXISTS objects_service.update_product_updated_at_column();
DROP TABLE IF EXISTS objects_service.products;
```

### Step 2: Register permissions in auth-service

Same as Scenario A, step 1 — insert permission records into `auth_service.permissions` and assign to roles via `auth_service.role_permissions`. Use the type's `type_key` as the resource value (e.g., `'products'`).

### Step 3: Add type_key to object_types

Same as Scenario A, step 2 — register the new type in `objects_service.object_types`:

```sql
INSERT INTO objects_service.object_types (name, description, type_key, is_sealed, metadata, created_at, updated_at)
VALUES ('Product', 'Products with SKU, price, and inventory tracking', 'products', false, '{}', NOW(), NOW());
```

Optionally add a `concrete_table_name = "products"` field if you extend the object_types schema to track CTI tables.

### Step 4: Create model layer

Create `services/objects-service/internal/models/product.go`:

```go
package models

import (
    "database/sql"
    "time"
)

// Product represents a product with both base object fields and CTI-specific fields.
type Product struct {
    ObjectID        int64             `json:"id"`
    PublicID        string            `json:"public_id"`
    Name            string            `json:"name"`
    Description     string            `json:"description"`
    Status          string            `json:"status"`
    Metadata        map[string]any    `json:"metadata,omitempty"`
    SKU             string            `json:"sku" binding:"required,min=1,max=100"`
    Price           float64           `json:"price" binding:"required,gt=0"`
    InventoryCount  int               `json:"inventory_count"`
    CreatedBy       string            `json:"created_by"`
    UpdatedBy       *string           `json:"updated_by,omitempty"`
    CreatedAt       time.Time         `json:"created_at"`
    UpdatedAt       *time.Time        `json:"updated_at,omitempty"`
}

// CreateProductRequest is the DTO for creating a product.
type CreateProductRequest struct {
    Name             string  `json:"name" binding:"required,min=1,max=255"`
    Description      string  `json:"description"`
    SKU              string  `json:"sku" binding:"required,min=1,max=100"`
    Price            float64 `json:"price" binding:"required,gt=0"`
    InventoryCount   int     `json:"inventory_count"`
    Metadata         map[string]any `json:"metadata,omitempty"`
}

// UpdateProductRequest is the DTO for updating a product.
type UpdateProductRequest struct {
    Name             *string  `json:"name" binding:"omitempty,min=1,max=255"`
    Description      *string  `json:"description"`
    SKU              *string  `json:"sku" binding:"omitempty,min=1,max=100"`
    Price            *float64 `json:"price" binding:"omitempty,gt=0"`
    InventoryCount   *int     `json:"inventory_count"`
    Metadata         map[string]any `json:"metadata,omitempty"`
}
```

### Step 5: Create repository layer

Create `services/objects-service/internal/repository/product_repository.go`:

```go
package repository

import (
    "context"
    "fmt"
    "strings"

    "github.com/jackc/pgx/v5"
    "github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
)

type productRepository struct {
    db          *PGDatabase
    repoOptions RepositoryOptions
}

func NewProductRepository(db *PGDatabase, opts RepositoryOptions) *productRepository {
    return &productRepository{db: db, repoOptions: opts}
}

// Create inserts a product into both objects table and CTI products table.
func (r *productRepository) Create(ctx context.Context, req *models.CreateProductRequest, userID string) (*models.Product, error) {
    query := `
        WITH inserted_object AS (
            INSERT INTO objects_service.objects 
                (name, description, status, metadata, object_type_id, created_by, updated_by)
            VALUES ($1, $2, 'active', $3, 
                    (SELECT id FROM objects_service.object_types WHERE type_key = 'products'),
                    $4, $4)
            RETURNING id, public_id
        ),
        inserted_product AS (
            INSERT INTO objects_service.products 
                (object_id, sku, price, inventory_count, created_by, updated_by)
            SELECT io.id, $5, $6, COALESCE($7, 0), $4, $4
            FROM inserted_object io
            RETURNING object_id, sku, price, inventory_count, created_at
        )
        SELECT op.object_id, op.public_id, o.name, o.description, o.status, o.metadata,
               op.sku, op.price, op.inventory_count, op.created_by, op.updated_by, 
               op.created_at, op.updated_at
        FROM inserted_product op
        JOIN objects_service.objects o ON o.id = op.object_id`

    var product models.Product
    err := r.db.QueryRow(ctx, query, req.Name, req.Description, req.Metadata, userID, req.SKU, req.Price, req.InventoryCount).Scan(
        &product.ObjectID, &product.PublicID, &product.Name, &product.Description, &product.Status,
        &product.Metadata, &product.SKU, &product.Price, &product.InventoryCount,
        &product.CreatedBy, &product.UpdatedBy, &product.CreatedAt, &product.UpdatedAt)

    if err != nil {
        return nil, fmt.Errorf("failed to create product: %w", err)
    }

    return &product, nil
}

// GetByID retrieves a product by its ID using CTE pattern.
func (r *productRepository) GetByID(ctx context.Context, id int64) (*models.Product, error) {
    query := `
        WITH product_data AS (
            SELECT 
                o.id as object_id, o.public_id, o.name, o.description, o.status, o.metadata,
                p.sku, p.price, p.inventory_count, p.created_by, p.updated_by, p.created_at, p.updated_at
            FROM objects_service.objects o
            INNER JOIN objects_service.products p ON o.id = p.object_id
            WHERE o.id = $1 
              AND o.object_type_id = (SELECT id FROM objects_service.object_types WHERE type_key = 'products')
              AND o.deleted_at IS NULL
        )
        SELECT * FROM product_data`

    var product models.Product
    err := r.db.QueryRow(ctx, query, id).Scan(
        &product.ObjectID, &product.PublicID, &product.Name, &product.Description, &product.Status,
        &product.Metadata, &product.SKU, &product.Price, &product.InventoryCount,
        &product.CreatedBy, &product.UpdatedBy, &product.CreatedAt, &product.UpdatedAt)

    if err != nil {
        return nil, fmt.Errorf("failed to get product: %w", err)
    }

    return &product, nil
}

// List retrieves products with optional filtering (includes CTI columns).
func (r *productRepository) List(ctx context.Context, filter *models.ProductFilter) ([]*models.Product, int64, error) {
    qb := newQueryBuilder("objects_service.objects")
    qb.Where("o.object_type_id = (SELECT id FROM objects_service.object_types WHERE type_key = 'products')")
    
    if filter.UserID != nil && *filter.UserID != "" {
        qb.Where("o.created_by = $1", *filter.UserID)
    }

    // Apply additional filters (status, sku search, etc.)

    where := strings.Join(qb.whereClauses, " AND ")
    query := fmt.Sprintf(`
        SELECT o.id as object_id, o.public_id, o.name, o.description, o.status, o.metadata,
               p.sku, p.price, p.inventory_count, o.created_by, o.updated_by, o.created_at
        FROM objects_service.objects o
        INNER JOIN objects_service.products p ON o.id = p.object_id
        WHERE %s
          AND o.deleted_at IS NULL`, where)

    // ... execute query and scan results
    
    return products, total, nil
}

// Update modifies a product (updates both base object and CTI table).
func (r *productRepository) Update(ctx context.Context, id int64, req *models.UpdateProductRequest, userID string) (*models.Product, error) {
    // ... implementation using CTE to update objects_service.objects AND products table
}

// Delete soft-deletes a product (deletes from CTI table first due to FK cascade).
func (r *productRepository) Delete(ctx context.Context, id int64) error {
    query := `DELETE FROM objects_service.products WHERE object_id = $1`
    _, err := r.db.Exec(ctx, query, id)
    if err != nil {
        return fmt.Errorf("failed to delete product CTI record: %w", err)
    }

    // Then soft-delete from base objects table (handled by ObjectRepository.Delete)
    return nil
}
```

### Step 6: Create service layer

Create `services/objects-service/internal/services/product_service.go`:

```go
package services

import (
    "context"
    "fmt"

    "github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
    "github.com/v-egorov/service-boilerplate/services/objects-service/internal/repository"
)

type ProductService struct {
    productRepo  repository.ProductRepository
    objectRepo   repository.ObjectRepository
}

func NewProductService(productRepo repository.ProductRepository, objectRepo repository.ObjectRepository) *ProductService {
    return &ProductService{productRepo: productRepo, objectRepo: objectRepo}
}

// Create creates a new product with ownership tracking.
func (s *ProductService) Create(ctx context.Context, req *models.CreateProductRequest, userID string) (*models.Product, error) {
    product, err := s.productRepo.Create(ctx, req, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to create product: %w", err)
    }

    return product, nil
}

// GetByID retrieves a product by ID.
func (s *ProductService) GetByID(ctx context.Context, id int64) (*models.Product, error) {
    product, err := s.productRepo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get product: %w", err)
    }

    return product, nil
}

// List retrieves products with filtering.
func (s *ProductService) List(ctx context.Context, filter *models.ProductFilter) ([]*models.Product, int64, error) {
    products, total, err := s.productRepo.List(ctx, filter)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to list products: %w", err)
    }

    return products, total, nil
}

// Update modifies a product with ownership check.
func (s *ProductService) Update(ctx context.Context, id int64, req *models.UpdateProductRequest, userID string) (*models.Product, error) {
    product, err := s.productRepo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get product: %w", err)
    }

    // Business validation: check ownership for :own scope
    if product.CreatedBy != userID {
        return nil, fmt.Errorf("permission denied: you can only update your own products")
    }

    updatedProduct, err := s.productRepo.Update(ctx, id, req, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to update product: %w", err)
    }

    return updatedProduct, nil
}

// Delete soft-deletes a product with ownership check.
func (s *ProductService) Delete(ctx context.Context, id int64, userID string) error {
    product, err := s.productRepo.GetByID(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to get product: %w", err)
    }

    // Business validation: check ownership for :own scope
    if product.CreatedBy != userID {
        return fmt.Errorf("permission denied: you can only delete your own products")
    }

    // Delete from CTI table first (FK constraint), then soft-delete base object
    err = s.productRepo.Delete(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to delete product CTI record: %w", err)
    }

    return s.objectRepo.Delete(ctx, id) // Soft-deletes the base objects row
}
```

### Step 7: Create handler layer

Create `services/objects-service/internal/handlers/product_handler.go` using the same pattern as Scenario A's document handler — use shared `checkOwnership()` utility from `handlers/base.go` and `handleServiceError()` for error mapping.

### Step 8: Wire in main.go

Same structure as Scenario A, but with CTI-aware service initialization:

```go
productRepo := repository.NewProductRepository(pgDatabase, repoOptions)
objectRepo := repository.NewObjectRepository(pgDatabase, repoOptions) // Needed for base object operations
productService := services.NewProductService(productRepo, objectRepo)
productHandler := handlers.NewProductHandler(productService, logger.Logger)

productsRead := v1.Group("/products")
productsRead.Use(perm(permiddleware.RouteConfig{TypeKey: "products", HTTPMethod: "GET"}))
{
    productsRead.GET("/:id", productHandler.GetByID)
    productsRead.GET("", productHandler.List)
}

// ... more route groups for create/update/delete using RouteConfig pattern
```

---

## Common Patterns & Tips

### Ownership Checking

All object types use the same unified ownership model: `created_by == userID`. The shared utility functions in `handlers/base.go` handle this automatically:

- `checkOwnership(c, createdByID)` — returns true if admin role or user has `:all` permission; otherwise checks `created_by == userID`
- `handleOwnershipViolation(c, requestID)` — standard 403 response for ownership violations

### Error Mapping

Use the shared `handleServiceError()` function to map service-layer errors to proper HTTP status codes. It uses `errors.Is()` for wrapped error comparison (see Tasks 6+8).

### Permission Middleware

Always use data-driven permission construction via `perm(permiddleware.RouteConfig{TypeKey: "...", HTTPMethod: "..."})`. This builds the correct permission strings (`type_key:create`, `type_key:read:all`, etc.) dynamically — no hardcoded permission names in route definitions.

### List Filtering by Owner

When implementing `List()`, pass the authenticated user ID to the repository filter so it can apply `WHERE created_by = $1` for `:own` scope users. This prevents leaking objects to users with only `read:own` permissions (see Task 7).

### Migration Best Practices

- Each migration must have both `.up.sql` and `.down.sql` files
- Use `ON CONFLICT DO NOTHING` for idempotent inserts
- Number migrations sequentially within each environment directory
- Test migrations by running them in dev, then verify rollback works before applying to staging/prod