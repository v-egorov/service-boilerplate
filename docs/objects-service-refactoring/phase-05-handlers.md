# Phase 5: Handlers Layer

**Estimated Time**: 4 hours
**Status**: ⬜ Not Started
**Dependencies**: Phase 4 (Services)

## Overview

Create HTTP handlers for Object Types and Objects with RESTful endpoints, request validation, and error handling. Delete old Entity handler files.

## Tasks

### 5.1 Create Object Type Handler

**File**: `internal/handlers/object_type_handler.go`

**Steps**:
1. Create ObjectTypeHandler struct
2. Implement HTTP handlers for all CRUD operations
3. Add request validation
4. Add error response formatting
5. Add authentication/authorization checks

```go
package handlers

import (
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    
    "your-project/services/objects-service/internal/models"
    "your-project/services/objects-service/internal/services"
)

type ObjectTypeHandler struct {
    service *services.ObjectTypeService
}

func NewObjectTypeHandler(service *services.ObjectTypeService) *ObjectTypeHandler {
    return &ObjectTypeHandler{service: service}
}

// RegisterRoutes registers all object type routes
func (h *ObjectTypeHandler) RegisterRoutes(router *gin.Engine) {
    types := router.Group("/api/v1/object-types")
    {
        types.POST("", h.Create)
        types.GET("", h.List)
        types.GET("/tree", h.GetTree)
        types.GET("/:id", h.GetByID)
        types.GET("/name/:name", h.GetByName)
        types.PUT("/:id", h.Update)
        types.DELETE("/:id", h.Delete)
        types.GET("/:id/children", h.GetChildren)
    }
}

// Create creates a new object type
// @Summary Create object type
// @Tags object-types
// @Accept json
// @Produce json
// @Param request body models.CreateObjectTypeRequest true "Object type data"
// @Success 201 {object} models.ObjectType
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/object-types [post]
func (h *ObjectTypeHandler) Create(c *gin.Context) {
    var req models.CreateObjectTypeRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    ot, err := h.service.Create(c.Request.Context(), &req)
    if err != nil {
        handleError(c, err)
        return
    }
    
    c.JSON(http.StatusCreated, ot)
}

// List lists all object types
// @Summary List object types
// @Tags object-types
// @Produce json
// @Param parent_type_id query int false "Filter by parent type ID"
// @Param is_sealed query bool false "Filter by sealed status"
// @Param search query string false "Search in name and description"
// @Success 200 {object} models.ObjectTypeListResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/object-types [get]
func (h *ObjectTypeHandler) List(c *gin.Context) {
    filter := &models.ObjectTypeFilter{}
    
    if parentTypeIDStr := c.Query("parent_type_id"); parentTypeIDStr != "" {
        if parentTypeID, err := strconv.ParseInt(parentTypeIDStr, 10, 64); err == nil {
            filter.ParentTypeID = &parentTypeID
        }
    }
    
    if isSealedStr := c.Query("is_sealed"); isSealedStr != "" {
        if isSealed, err := strconv.ParseBool(isSealedStr); err == nil {
            filter.IsSealed = &isSealed
        }
    }
    
    if search := c.Query("search"); search != "" {
        filter.Search = &search
    }
    
    response, err := h.service.List(c.Request.Context(), filter)
    if err != nil {
        handleError(c, err)
        return
    }
    
    c.JSON(http.StatusOK, response)
}

// GetTree retrieves the full object type tree
// @Summary Get object type tree
// @Tags object-types
// @Produce json
// @Success 200 {array} models.ObjectType
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/object-types/tree [get]
func (h *ObjectTypeHandler) GetTree(c *gin.Context) {
    types, err := h.service.GetTree(c.Request.Context())
    if err != nil {
        handleError(c, err)
        return
    }
    
    c.JSON(http.StatusOK, types)
}

// GetByID retrieves an object type by ID
// @Summary Get object type by ID
// @Tags object-types
// @Produce json
// @Param id path int true "Object type ID"
// @Success 200 {object} models.ObjectType
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/object-types/{id} [get]
func (h *ObjectTypeHandler) GetByID(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    
    ot, err := h.service.GetByID(c.Request.Context(), id)
    if err != nil {
        handleError(c, err)
        return
    }
    
    c.JSON(http.StatusOK, ot)
}

// GetByName retrieves an object type by name
// @Summary Get object type by name
// @Tags object-types
// @Produce json
// @Param name path string true "Object type name"
// @Success 200 {object} models.ObjectType
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/object-types/name/{name} [get]
func (h *ObjectTypeHandler) GetByName(c *gin.Context) {
    name := c.Param("name")
    
    ot, err := h.service.GetByName(c.Request.Context(), name)
    if err != nil {
        handleError(c, err)
        return
    }
    
    c.JSON(http.StatusOK, ot)
}

// Update updates an object type
// @Summary Update object type
// @Tags object-types
// @Accept json
// @Produce json
// @Param id path int true "Object type ID"
// @Param request body models.UpdateObjectTypeRequest true "Object type data"
// @Success 200 {object} models.ObjectType
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/object-types/{id} [put]
func (h *ObjectTypeHandler) Update(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    
    var req models.UpdateObjectTypeRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    ot, err := h.service.Update(c.Request.Context(), id, &req)
    if err != nil {
        handleError(c, err)
        return
    }
    
    c.JSON(http.StatusOK, ot)
}

// Delete deletes an object type
// @Summary Delete object type
// @Tags object-types
// @Produce json
// @Param id path int true "Object type ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/object-types/{id} [delete]
func (h *ObjectTypeHandler) Delete(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    
    if err := h.service.Delete(c.Request.Context(), id); err != nil {
        handleError(c, err)
        return
    }
    
    c.Status(http.StatusNoContent)
}

// GetChildren retrieves child object types
// @Summary Get child object types
// @Tags object-types
// @Produce json
// @Param id path int true "Parent object type ID"
// @Success 200 {array} models.ObjectType
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/object-types/{id}/children [get]
func (h *ObjectTypeHandler) GetChildren(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    
    types, err := h.service.GetChildren(c.Request.Context(), id)
    if err != nil {
        handleError(c, err)
        return
    }
    
    c.JSON(http.StatusOK, types)
}

func handleError(c *gin.Context, err error) {
    switch {
    case err == repository.ErrObjectTypeNotFound:
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
    case err == repository.ErrCircularReference:
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    case err.Error() == "cannot seal type with children":
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    case err.Error() == "cannot create child type for sealed parent type":
        c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
    case err.Error() == "cannot delete object type with children":
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    default:
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
    }
}
```

---

### 5.2 Create Object Handler

**File**: `internal/handlers/object_handler.go`

**Steps**:
1. Create ObjectHandler struct
2. Implement HTTP handlers for all CRUD operations
3. Add batch operation endpoints
4. Add request validation
5. Add error response formatting

```go
package handlers

import (
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    
    "your-project/services/objects-service/internal/models"
    "your-project/services/objects-service/internal/repository"
    "your-project/services/objects-service/internal/services"
)

type ObjectHandler struct {
    service *services.ObjectService
}

func NewObjectHandler(service *services.ObjectService) *ObjectHandler {
    return &ObjectHandler{service: service}
}

// RegisterRoutes registers all object routes
func (h *ObjectHandler) RegisterRoutes(router *gin.Engine) {
    objects := router.Group("/api/v1/objects")
    {
        objects.POST("", h.Create)
        objects.GET("", h.List)
        objects.POST("/batch", h.CreateBatch)
        objects.PATCH("/batch", h.UpdateBatch)
        objects.GET("/:public_id", h.GetByPublicID)
        objects.PUT("/:public_id", h.Update)
        objects.DELETE("/:public_id", h.SoftDelete)
        objects.DELETE("/:public_id/hard", h.HardDelete)
        objects.GET("/:public_id/children", h.GetChildren)
    }
}

// Create creates a new object
// @Summary Create object
// @Tags objects
// @Accept json
// @Produce json
// @Param request body models.CreateObjectRequest true "Object data"
// @Success 201 {object} models.Object
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/objects [post]
func (h *ObjectHandler) Create(c *gin.Context) {
    var req models.CreateObjectRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    userID := getUserID(c)
    
    obj, err := h.service.Create(c.Request.Context(), &req, userID)
    if err != nil {
        handleObjectError(c, err)
        return
    }
    
    c.JSON(http.StatusCreated, obj)
}

// List lists objects
// @Summary List objects
// @Tags objects
// @Produce json
// @Param object_type_id query int false "Filter by object type ID"
// @Param parent_object_id query int false "Filter by parent object ID"
// @Param status query string false "Filter by status"
// @Param include_deleted query bool false "Include deleted objects"
// @Param search query string false "Search in name and description"
// @Param tags query string false "Filter by tags (comma-separated)"
// @Param tags_mode query string false "Tags match mode (any/all)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param sort_by query string false "Sort field" default(created_at)
// @Param sort_order query string false "Sort order" default(desc)
// @Success 200 {object} models.ObjectListResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/objects [get]
func (h *ObjectHandler) List(c *gin.Context) {
    filter := &models.ObjectFilter{
        Page:     1,
        PageSize: 20,
        TagsMode: "any",
    }
    
    if objectTypeIDStr := c.Query("object_type_id"); objectTypeIDStr != "" {
        if objectTypeID, err := strconv.ParseInt(objectTypeIDStr, 10, 64); err == nil {
            filter.ObjectTypeID = &objectTypeID
        }
    }
    
    if parentObjectIDStr := c.Query("parent_object_id"); parentObjectIDStr != "" {
        if parentObjectID, err := strconv.ParseInt(parentObjectIDStr, 10, 64); err == nil {
            filter.ParentObjectID = &parentObjectID
        }
    }
    
    if status := c.Query("status"); status != "" {
        filter.Status = &status
    }
    
    if includeDeletedStr := c.Query("include_deleted"); includeDeletedStr != "" {
        if includeDeleted, err := strconv.ParseBool(includeDeletedStr); err == nil {
            filter.IncludeDeleted = &includeDeleted
        }
    }
    
    if search := c.Query("search"); search != "" {
        filter.Search = &search
    }
    
    if tags := c.Query("tags"); tags != "" {
        filter.Tags = parseTags(tags)
    }
    
    if tagsMode := c.Query("tags_mode"); tagsMode != "" {
        filter.TagsMode = tagsMode
    }
    
    if page := c.Query("page"); page != "" {
        if pageNum, err := strconv.Atoi(page); err == nil && pageNum > 0 {
            filter.Page = pageNum
        }
    }
    
    if pageSize := c.Query("page_size"); pageSize != "" {
        if size, err := strconv.Atoi(pageSize); err == nil && size > 0 && size <= 100 {
            filter.PageSize = size
        }
    }
    
    if sortBy := c.Query("sort_by"); sortBy != "" {
        filter.SortBy = sortBy
    }
    
    if sortOrder := c.Query("sort_order"); sortOrder != "" {
        filter.SortOrder = sortOrder
    }
    
    response, err := h.service.List(c.Request.Context(), filter)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, response)
}

// CreateBatch creates multiple objects
// @Summary Create objects in batch
// @Tags objects
// @Accept json
// @Produce json
// @Param request body models.BatchCreateRequest true "Objects data"
// @Success 201 {object} models.BatchCreateResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/objects/batch [post]
func (h *ObjectHandler) CreateBatch(c *gin.Context) {
    var req models.BatchCreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    userID := getUserID(c)
    
    response, err := h.service.CreateBatch(c.Request.Context(), &req, userID)
    if err != nil {
        handleObjectError(c, err)
        return
    }
    
    c.JSON(http.StatusCreated, response)
}

// UpdateBatch updates multiple objects
// @Summary Update objects in batch
// @Tags objects
// @Accept json
// @Produce json
// @Param request body models.BatchUpdateRequest true "Updates data"
// @Success 200 {object} models.BatchUpdateResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/objects/batch [patch]
func (h *ObjectHandler) UpdateBatch(c *gin.Context) {
    var req models.BatchUpdateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    userID := getUserID(c)
    
    response, err := h.service.UpdateBatch(c.Request.Context(), &req, userID)
    if err != nil {
        handleObjectError(c, err)
        return
    }
    
    c.JSON(http.StatusOK, response)
}

// GetByPublicID retrieves an object by public ID
// @Summary Get object by public ID
// @Tags objects
// @Produce json
// @Param public_id path string true "Object public ID (UUID)"
// @Success 200 {object} models.Object
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/objects/{public_id} [get]
func (h *ObjectHandler) GetByPublicID(c *gin.Context) {
    publicID, err := uuid.Parse(c.Param("public_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid public_id"})
        return
    }
    
    obj, err := h.service.GetByPublicID(c.Request.Context(), publicID)
    if err != nil {
        handleObjectError(c, err)
        return
    }
    
    c.JSON(http.StatusOK, obj)
}

// Update updates an object
// @Summary Update object
// @Tags objects
// @Accept json
// @Produce json
// @Param public_id path string true "Object public ID (UUID)"
// @Param request body models.UpdateObjectRequest true "Object data"
// @Success 200 {object} models.Object
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/objects/{public_id} [put]
func (h *ObjectHandler) Update(c *gin.Context) {
    publicID, err := uuid.Parse(c.Param("public_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid public_id"})
        return
    }
    
    var req models.UpdateObjectRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    obj, err := h.service.GetByPublicID(c.Request.Context(), publicID)
    if err != nil {
        handleObjectError(c, err)
        return
    }
    
    userID := getUserID(c)
    
    updated, err := h.service.Update(c.Request.Context(), obj.ID, &req, userID)
    if err != nil {
        handleObjectError(c, err)
        return
    }
    
    c.JSON(http.StatusOK, updated)
}

// SoftDelete soft deletes an object
// @Summary Soft delete object
// @Tags objects
// @Produce json
// @Param public_id path string true "Object public ID (UUID)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/objects/{public_id} [delete]
func (h *ObjectHandler) SoftDelete(c *gin.Context) {
    publicID, err := uuid.Parse(c.Param("public_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid public_id"})
        return
    }
    
    obj, err := h.service.GetByPublicID(c.Request.Context(), publicID)
    if err != nil {
        handleObjectError(c, err)
        return
    }
    
    userID := getUserID(c)
    
    if err := h.service.SoftDelete(c.Request.Context(), obj.ID, userID); err != nil {
        handleObjectError(c, err)
        return
    }
    
    c.Status(http.StatusNoContent)
}

// HardDelete permanently deletes an object
// @Summary Hard delete object
// @Tags objects
// @Produce json
// @Param public_id path string true "Object public ID (UUID)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/objects/{public_id}/hard [delete]
func (h *ObjectHandler) HardDelete(c *gin.Context) {
    publicID, err := uuid.Parse(c.Param("public_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid public_id"})
        return
    }
    
    obj, err := h.service.GetByPublicID(c.Request.Context(), publicID)
    if err != nil {
        handleObjectError(c, err)
        return
    }
    
    if err := h.service.HardDelete(c.Request.Context(), obj.ID); err != nil {
        handleObjectError(c, err)
        return
    }
    
    c.Status(http.StatusNoContent)
}

// GetChildren retrieves child objects
// @Summary Get child objects
// @Tags objects
// @Produce json
// @Param public_id path string true "Parent object public ID (UUID)"
// @Success 200 {array} models.Object
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/objects/{public_id}/children [get]
func (h *ObjectHandler) GetChildren(c *gin.Context) {
    publicID, err := uuid.Parse(c.Param("public_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid public_id"})
        return
    }
    
    obj, err := h.service.GetByPublicID(c.Request.Context(), publicID)
    if err != nil {
        handleObjectError(c, err)
        return
    }
    
    objects, err := h.service.GetChildren(c.Request.Context(), obj.ID)
    if err != nil {
        handleObjectError(c, err)
        return
    }
    
    c.JSON(http.StatusOK, objects)
}

func handleObjectError(c *gin.Context, err error) {
    switch {
    case err == repository.ErrObjectNotFound:
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
    case err == repository.ErrObjectVersionConflict:
        c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
    default:
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
    }
}

func getUserID(c *gin.Context) string {
    userID, _ := c.Get("user_id")
    if userID == nil {
        return "system"
    }
    if uid, ok := userID.(string); ok {
        return uid
    }
    return "system"
}

func parseTags(tagsStr string) []string {
    // Implement tag parsing logic
    return []string{}
}

type ErrorResponse struct {
    Error string `json:"error"`
}
```

---

### 5.3 Delete Old Entity Handler Files

**Files**:
- `internal/handlers/entity_handler.go`
- `internal/handlers/entity_handler_test.go`

**Steps**:
1. Delete `internal/handlers/entity_handler.go`
2. Delete `internal/handlers/entity_handler_test.go`

```bash
rm services/objects-service/internal/handlers/entity_handler.go
rm services/objects-service/internal/handlers/entity_handler_test.go
```

---

## Checklist

- [ ] Create `internal/handlers/object_type_handler.go`
- [ ] Create `internal/handlers/object_handler.go`
- [ ] Delete `internal/handlers/entity_handler.go`
- [ ] Delete `internal/handlers/entity_handler_test.go`
- [ ] Verify no compilation errors: `go build ./internal/handlers/...`
- [ ] Create handler tests
- [ ] Test API endpoints manually with curl/postman
- [ ] Update progress.md

## Testing

```bash
# Verify handlers compile
cd services/objects-service
go build ./internal/handlers/...

# Run handler tests
go test ./internal/handlers/... -v

# Test API manually
curl http://localhost:8085/api/v1/object-types
curl http://localhost:8085/api/v1/objects
```

## Common Issues

**Issue**: Route registration not working
**Solution**: Ensure handlers are registered in cmd/main.go after service initialization

**Issue**: JSON binding errors
**Solution**: Add proper struct tags and validation tags to request models

**Issue**: User ID extraction fails
**Solution**: Ensure JWT middleware sets user_id in context before handlers

## Next Phase

Proceed to [Phase 6: Main Application](phase-06-main.md) once all tasks in this phase are complete.
