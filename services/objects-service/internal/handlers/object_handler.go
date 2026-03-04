package handlers

import (
	"context"
	"net/http"
	"slices"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/common/logging"
	"github.com/v-egorov/service-boilerplate/common/middleware"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/repository"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/services"
)

// ObjectServiceInterface defines the service operations needed for handlers
type ObjectServiceInterface interface {
	Create(ctx context.Context, req *models.CreateObjectRequest) (*models.Object, error)
	GetByID(ctx context.Context, id int64) (*models.Object, error)
	GetByPublicID(ctx context.Context, publicID uuid.UUID) (*models.Object, error)
	GetByName(ctx context.Context, name string) (*models.Object, error)
	Update(ctx context.Context, id int64, req *models.UpdateObjectRequest) (*models.Object, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, filter *models.ObjectFilter) ([]*models.Object, int64, error)
	Search(ctx context.Context, query string, limit int) ([]*models.Object, error)
	FindByMetadata(ctx context.Context, key, value string) ([]*models.Object, error)
	FindByTags(ctx context.Context, tags []string, matchAll bool) ([]*models.Object, error)
	UpdateMetadata(ctx context.Context, id int64, metadata map[string]interface{}) error
	AddTags(ctx context.Context, id int64, tags []string) error
	RemoveTags(ctx context.Context, id int64, tags []string) error
	GetChildren(ctx context.Context, parentID int64) ([]*models.Object, error)
	GetDescendants(ctx context.Context, rootID int64, maxDepth *int) ([]*models.Object, error)
	GetAncestors(ctx context.Context, id int64) ([]*models.Object, error)
	GetPath(ctx context.Context, id int64) ([]*models.Object, error)
	BulkCreate(ctx context.Context, objects []*models.CreateObjectRequest) ([]*models.Object, error)
	BulkUpdate(ctx context.Context, ids []int64, updates *models.UpdateObjectRequest) ([]*models.Object, error)
	BulkDelete(ctx context.Context, ids []int64) error
	ValidateParentChild(ctx context.Context, parentID, childID int64) error
	GetObjectStats(ctx context.Context, filter *models.ObjectFilter) (*repository.ObjectStats, error)
}

// ObjectHandler handles HTTP requests for objects
type ObjectHandler struct {
	service        ObjectServiceInterface
	logger         *logrus.Logger
	auditLogger    *logging.AuditLogger
	standardLogger *logging.StandardLogger
}

// NewObjectHandler creates a new ObjectHandler
func NewObjectHandler(service services.ObjectService, logger *logrus.Logger) *ObjectHandler {
	return &ObjectHandler{
		service:        service,
		logger:         logger,
		auditLogger:    logging.NewAuditLogger(logger, "objects-service"),
		standardLogger: logging.NewStandardLogger(logger, "objects-service"),
	}
}

// NewObjectHandlerWithInterface creates a handler with a service interface (for testing)
func NewObjectHandlerWithInterface(service ObjectServiceInterface, logger *logrus.Logger) *ObjectHandler {
	return &ObjectHandler{
		service:        service,
		logger:         logger,
		auditLogger:    logging.NewAuditLogger(logger, "objects-service"),
		standardLogger: logging.NewStandardLogger(logger, "objects-service"),
	}
}

func (h *ObjectHandler) handleServiceError(c *gin.Context, err error, operation string, requestID string) {
	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
	}).WithError(err).Error(operation)

	switch err {
	case nil:
		return
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
			"type":  "internal_error",
		})
	}
}

func (h *ObjectHandler) checkOwnership(c *gin.Context, object *models.Object, allPermission string) bool {
	userID := middleware.GetAuthenticatedUserID(c)
	if userID == "" {
		return false
	}

	userRoles := middleware.GetAuthenticatedUserRoles(c)
	for _, role := range userRoles {
		if role == "admin" || role == "object-type-admin" {
			return true
		}
	}

	matchedPermissions, exists := c.Get("matched_permissions")
	if !exists {
		return object.CreatedBy == userID
	}

	perms, ok := matchedPermissions.([]string)
	if !ok {
		return object.CreatedBy == userID
	}

	if slices.Contains(perms, allPermission) {
		return true
	}

	return object.CreatedBy == userID
}

func (h *ObjectHandler) Create(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	var req models.CreateObjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
		}).WithError(err).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"type":    "validation_error",
		})
		return
	}

	userID := middleware.GetAuthenticatedUserID(c)
	if userID != "" {
		req.CreatedBy = userID
	}

	object, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to create object", requestID)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"object_id":  object.ID,
		"request_id": requestID,
	}).Info("Object created successfully")

	c.JSON(http.StatusCreated, gin.H{
		"data":    object,
		"message": "Object created successfully",
	})
}

func (h *ObjectHandler) GetByID(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID format",
			"details": "ID must be a positive integer",
			"type":    "validation_error",
			"field":   "id",
		})
		return
	}

	object, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get object", requestID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": object,
	})
}

func (h *ObjectHandler) GetByPublicID(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	publicIDStr := c.Param("public_id")
	publicID, err := uuid.Parse(publicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid public ID format",
			"details": "Public ID must be a valid UUID",
			"type":    "validation_error",
			"field":   "public_id",
		})
		return
	}

	object, err := h.service.GetByPublicID(c.Request.Context(), publicID)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get object by public ID", requestID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": object,
	})
}

func (h *ObjectHandler) GetByName(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing name parameter",
			"type":  "validation_error",
			"field": "name",
		})
		return
	}

	object, err := h.service.GetByName(c.Request.Context(), name)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get object by name", requestID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": object,
	})
}

func (h *ObjectHandler) Update(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID format",
			"details": "ID must be a positive integer",
			"type":    "validation_error",
			"field":   "id",
		})
		return
	}

	existingObj, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get object", requestID)
		return
	}

	if !h.checkOwnership(c, existingObj, "objects:update:all") {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You can only update your own objects",
			"type":  "ownership_error",
		})
		return
	}

	var req models.UpdateObjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
		}).WithError(err).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"type":    "validation_error",
		})
		return
	}

	object, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to update object", requestID)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"object_id":  object.ID,
		"request_id": requestID,
	}).Info("Object updated successfully")

	c.JSON(http.StatusOK, gin.H{
		"data":    object,
		"message": "Object updated successfully",
	})
}

func (h *ObjectHandler) Delete(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID format",
			"details": "ID must be a positive integer",
			"type":    "validation_error",
			"field":   "id",
		})
		return
	}

	existingObj, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get object", requestID)
		return
	}

	if !h.checkOwnership(c, existingObj, "objects:delete:all") {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You can only delete your own objects",
			"type":  "ownership_error",
		})
		return
	}

	err = h.service.Delete(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err, "Failed to delete object", requestID)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"object_id":  id,
		"request_id": requestID,
	}).Info("Object deleted successfully")

	c.JSON(http.StatusNoContent, nil)
}

func (h *ObjectHandler) List(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	filter := &models.ObjectFilter{}

	if objectTypeIDStr := c.Query("object_type_id"); objectTypeIDStr != "" {
		if objectTypeID, err := strconv.ParseInt(objectTypeIDStr, 10, 64); err == nil && objectTypeID > 0 {
			filter.ObjectTypeID = &objectTypeID
		}
	}

	if name := c.Query("name"); name != "" {
		filter.Name = name
	}

	if status := c.Query("status"); status != "" {
		filter.Status = status
	}

	filter.Limit = 50
	filter.Offset = 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	objects, total, err := h.service.List(c.Request.Context(), filter)
	if err != nil {
		h.handleServiceError(c, err, "Failed to list objects", requestID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": objects,
		"pagination": gin.H{
			"limit":  filter.Limit,
			"offset": filter.Offset,
			"total":  total,
			"count":  len(objects),
		},
	})
}

func (h *ObjectHandler) Search(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing search query",
			"details": "Query parameter 'q' is required",
			"type":    "validation_error",
			"field":   "q",
		})
		return
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	results, err := h.service.Search(c.Request.Context(), query, limit)
	if err != nil {
		h.handleServiceError(c, err, "Failed to search objects", requestID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  results,
		"query": query,
	})
}

func (h *ObjectHandler) UpdateMetadata(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID format",
			"details": "ID must be a positive integer",
			"type":    "validation_error",
			"field":   "id",
		})
		return
	}

	var metadata map[string]interface{}
	if err := c.ShouldBindJSON(&metadata); err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
		}).WithError(err).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"type":    "validation_error",
		})
		return
	}

	err = h.service.UpdateMetadata(c.Request.Context(), id, metadata)
	if err != nil {
		h.handleServiceError(c, err, "Failed to update metadata", requestID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Metadata updated successfully",
	})
}

func (h *ObjectHandler) AddTags(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID format",
			"details": "ID must be a positive integer",
			"type":    "validation_error",
			"field":   "id",
		})
		return
	}

	var req struct {
		Tags []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"type":    "validation_error",
		})
		return
	}

	err = h.service.AddTags(c.Request.Context(), id, req.Tags)
	if err != nil {
		h.handleServiceError(c, err, "Failed to add tags", requestID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tags added successfully",
	})
}

func (h *ObjectHandler) RemoveTags(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID format",
			"details": "ID must be a positive integer",
			"type":    "validation_error",
			"field":   "id",
		})
		return
	}

	var req struct {
		Tags []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"type":    "validation_error",
		})
		return
	}

	err = h.service.RemoveTags(c.Request.Context(), id, req.Tags)
	if err != nil {
		h.handleServiceError(c, err, "Failed to remove tags", requestID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tags removed successfully",
	})
}

func (h *ObjectHandler) GetChildren(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID format",
			"details": "ID must be a positive integer",
			"type":    "validation_error",
			"field":   "id",
		})
		return
	}

	children, err := h.service.GetChildren(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get children", requestID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": children,
	})
}

func (h *ObjectHandler) GetDescendants(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID format",
			"details": "ID must be a positive integer",
			"type":    "validation_error",
			"field":   "id",
		})
		return
	}

	maxDepthStr := c.Query("max_depth")
	var maxDepth *int
	if maxDepthStr != "" {
		if depth, err := strconv.Atoi(maxDepthStr); err == nil && depth > 0 {
			maxDepth = &depth
		}
	}

	descendants, err := h.service.GetDescendants(c.Request.Context(), id, maxDepth)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get descendants", requestID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": descendants,
	})
}

func (h *ObjectHandler) GetAncestors(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID format",
			"details": "ID must be a positive integer",
			"type":    "validation_error",
			"field":   "id",
		})
		return
	}

	ancestors, err := h.service.GetAncestors(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get ancestors", requestID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": ancestors,
	})
}

func (h *ObjectHandler) GetPath(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID format",
			"details": "ID must be a positive integer",
			"type":    "validation_error",
			"field":   "id",
		})
		return
	}

	path, err := h.service.GetPath(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get path", requestID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": path,
	})
}

func (h *ObjectHandler) BulkCreate(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	var objects []*models.CreateObjectRequest
	if err := c.ShouldBindJSON(&objects); err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
		}).WithError(err).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"type":    "validation_error",
		})
		return
	}

	results, err := h.service.BulkCreate(c.Request.Context(), objects)
	if err != nil {
		h.handleServiceError(c, err, "Failed to bulk create objects", requestID)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"count":      len(results),
	}).Info("Objects bulk created successfully")

	c.JSON(http.StatusCreated, gin.H{
		"data":    results,
		"message": "Objects created successfully",
	})
}

func (h *ObjectHandler) BulkUpdate(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	var req struct {
		IDs     []int64                     `json:"ids"`
		Updates *models.UpdateObjectRequest `json:"updates"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"type":    "validation_error",
		})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing IDs",
			"details": "At least one ID is required",
			"type":    "validation_error",
			"field":   "ids",
		})
		return
	}

	results, err := h.service.BulkUpdate(c.Request.Context(), req.IDs, req.Updates)
	if err != nil {
		h.handleServiceError(c, err, "Failed to bulk update objects", requestID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    results,
		"message": "Objects updated successfully",
	})
}

func (h *ObjectHandler) BulkDelete(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	var req struct {
		IDs []int64 `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"type":    "validation_error",
		})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing IDs",
			"details": "At least one ID is required",
			"type":    "validation_error",
			"field":   "ids",
		})
		return
	}

	err := h.service.BulkDelete(c.Request.Context(), req.IDs)
	if err != nil {
		h.handleServiceError(c, err, "Failed to bulk delete objects", requestID)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"count":      len(req.IDs),
	}).Info("Objects bulk deleted successfully")

	c.JSON(http.StatusNoContent, nil)
}

func (h *ObjectHandler) GetStats(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	filter := &models.ObjectFilter{}
	if objectTypeIDStr := c.Query("object_type_id"); objectTypeIDStr != "" {
		if objectTypeID, err := strconv.ParseInt(objectTypeIDStr, 10, 64); err == nil && objectTypeID > 0 {
			filter.ObjectTypeID = &objectTypeID
		}
	}

	stats, err := h.service.GetObjectStats(c.Request.Context(), filter)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get object stats", requestID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}
