package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/common/logging"
	"github.com/v-egorov/service-boilerplate/common/middleware"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/services"
)

// ObjectTypeServiceInterface defines the service operations needed for handlers
type ObjectTypeServiceInterface interface {
	Create(ctx context.Context, req *models.CreateObjectTypeRequest) (*models.ObjectType, error)
	GetByID(ctx context.Context, id int64) (*models.ObjectType, error)
	GetByName(ctx context.Context, name string) (*models.ObjectType, error)
	Update(ctx context.Context, id int64, req *models.UpdateObjectTypeRequest) (*models.ObjectType, error)
	Delete(ctx context.Context, id int64) error
	GetTree(ctx context.Context, rootID *int64) ([]*models.ObjectType, error)
	GetChildren(ctx context.Context, parentID int64) ([]*models.ObjectType, error)
	GetDescendants(ctx context.Context, rootID int64, maxDepth *int) ([]*models.ObjectType, error)
	GetAncestors(ctx context.Context, id int64) ([]*models.ObjectType, error)
	GetPath(ctx context.Context, id int64) ([]*models.ObjectType, error)
	List(ctx context.Context, filter *models.ObjectTypeFilter) ([]*models.ObjectType, error)
	Search(ctx context.Context, query string, limit int) ([]*models.ObjectType, error)
	ValidateMove(ctx context.Context, id int64, newParentID *int64) error
	GetSubtreeObjectCount(ctx context.Context, id int64) (int64, error)
}

// ObjectTypeHandler handles HTTP requests for object types
type ObjectTypeHandler struct {
	service        ObjectTypeServiceInterface
	logger         *logrus.Logger
	auditLogger    *logging.AuditLogger
	standardLogger *logging.StandardLogger
}

// NewObjectTypeHandler creates a new ObjectTypeHandler
func NewObjectTypeHandler(service services.ObjectTypeService, logger *logrus.Logger) *ObjectTypeHandler {
	return &ObjectTypeHandler{
		service:        service,
		logger:         logger,
		auditLogger:    logging.NewAuditLogger(logger, "objects-service"),
		standardLogger: logging.NewStandardLogger(logger, "objects-service"),
	}
}

// NewObjectTypeHandlerWithInterface creates a handler with a service interface (for testing)
func NewObjectTypeHandlerWithInterface(service ObjectTypeServiceInterface, logger *logrus.Logger) *ObjectTypeHandler {
	return &ObjectTypeHandler{
		service:        service,
		logger:         logger,
		auditLogger:    logging.NewAuditLogger(logger, "objects-service"),
		standardLogger: logging.NewStandardLogger(logger, "objects-service"),
	}
}

func (h *ObjectTypeHandler) handleServiceError(c *gin.Context, err error, operation string, requestID string) {
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

func (h *ObjectTypeHandler) Create(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	var req models.CreateObjectTypeRequest
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

	// Set created_by and updated_by from authenticated user
	userID := middleware.GetAuthenticatedUserID(c)
	req.CreatedBy = userID
	req.UpdatedBy = userID

	objectType, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to create object type", requestID)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"object_type_id": objectType.ID,
		"request_id":     requestID,
	}).Info("Object type created successfully")

	c.JSON(http.StatusCreated, gin.H{
		"data":    objectType,
		"message": "Object type created successfully",
		"meta":    gin.H{"request_id": requestID},
	})
}

func (h *ObjectTypeHandler) GetByID(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"id":         idStr,
		}).WithError(err).Error("Invalid object type ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid id format: id must be a positive integer",
			"type":  "validation_error",
			"field": "id",
			"meta":  gin.H{"request_id": requestID},
		})
		return
	}

	objectType, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get object type", requestID)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id":     requestID,
		"object_type_id": objectType.ID,
	}).Debug("Object type retrieved successfully")

	c.JSON(http.StatusOK, gin.H{
		"data":    objectType,
		"meta":    gin.H{"request_id": requestID},
	})
}

func (h *ObjectTypeHandler) GetByName(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	name := c.Param("name")
	if name == "" {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
		}).Error("Missing name parameter")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing name parameter: name is required",
			"type":  "validation_error",
			"field": "name",
			"meta":  gin.H{"request_id": requestID},
		})
		return
	}

	objectType, err := h.service.GetByName(c.Request.Context(), name)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get object type by name", requestID)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id":       requestID,
		"object_type_id":   objectType.ID,
		"object_type_name": objectType.Name,
	}).Debug("Object type retrieved by name successfully")

	c.JSON(http.StatusOK, gin.H{
		"data":    objectType,
		"meta":    gin.H{"request_id": requestID},
	})
}

func (h *ObjectTypeHandler) Update(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"id":         idStr,
		}).WithError(err).Error("Invalid object type ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid id format: id must be a positive integer",
			"type":  "validation_error",
			"field": "id",
			"meta":  gin.H{"request_id": requestID},
		})
		return
	}

	var req models.UpdateObjectTypeRequest
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

	// Set updated_by from authenticated user
	userID := middleware.GetAuthenticatedUserID(c)
	req.UpdatedBy = userID

	objectType, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to update object type", requestID)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id":     requestID,
		"object_type_id": objectType.ID,
	}).Info("Object type updated successfully")

	c.JSON(http.StatusOK, gin.H{
		"data":    objectType,
		"message": "Object type updated successfully",
		"meta":    gin.H{"request_id": requestID},
	})
}

func (h *ObjectTypeHandler) Delete(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"id":         idStr,
		}).WithError(err).Error("Invalid object type ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid id format: id must be a positive integer",
			"type":  "validation_error",
			"field": "id",
			"meta":  gin.H{"request_id": requestID},
		})
		return
	}

	err = h.service.Delete(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err, "Failed to delete object type", requestID)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id":     requestID,
		"object_type_id": id,
	}).Info("Object type deleted successfully")

	c.JSON(http.StatusNoContent, nil)
}

func (h *ObjectTypeHandler) GetTree(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	rootIDStr := c.Query("root_id")
	var rootID *int64
	if rootIDStr != "" {
		id, err := strconv.ParseInt(rootIDStr, 10, 64)
		if err == nil && id > 0 {
			rootID = &id
		}
	}

	tree, err := h.service.GetTree(c.Request.Context(), rootID)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get object type tree", requestID)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"root_id":    rootID,
		"count":      len(tree),
	}).Debug("Object type tree retrieved successfully")

	c.JSON(http.StatusOK, gin.H{
		"data":    tree,
		"meta":    gin.H{"request_id": requestID},
	})
}

func (h *ObjectTypeHandler) GetChildren(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

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

	children, err := h.service.GetChildren(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get children", requestID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    children,
		"meta":    gin.H{"request_id": requestID},
	})
}

func (h *ObjectTypeHandler) GetDescendants(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

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
		"data":    descendants,
		"meta":    gin.H{"request_id": requestID},
	})
}

func (h *ObjectTypeHandler) GetAncestors(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

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

	ancestors, err := h.service.GetAncestors(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get ancestors", requestID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    ancestors,
		"meta":    gin.H{"request_id": requestID},
	})
}

func (h *ObjectTypeHandler) GetPath(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

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

	path, err := h.service.GetPath(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get path", requestID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    path,
		"meta":    gin.H{"request_id": requestID},
	})
}

func (h *ObjectTypeHandler) List(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	filter := &models.ObjectTypeFilter{
		Name:   c.Query("name"),
		Limit:  50,
		Offset: 0,
	}

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

	objectTypes, err := h.service.List(c.Request.Context(), filter)
	if err != nil {
		h.handleServiceError(c, err, "Failed to list object types", requestID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": objectTypes,
		"pagination": gin.H{
			"limit":  filter.Limit,
			"offset": filter.Offset,
			"count":  len(objectTypes),
		},
		"meta": gin.H{"request_id": requestID},
	})
}

func (h *ObjectTypeHandler) Search(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing search query: query parameter 'q' is required",
			"type":  "validation_error",
			"field": "q",
			"meta":  gin.H{"request_id": requestID},
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
		h.handleServiceError(c, err, "Failed to search object types", requestID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  results,
		"query": query,
		"meta":  gin.H{"request_id": requestID},
	})
}

func (h *ObjectTypeHandler) ValidateMove(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

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

	newParentIDStr := c.Query("new_parent_id")
	if newParentIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing new_parent_id: query parameter 'new_parent_id' is required",
			"type":  "validation_error",
			"field": "new_parent_id",
			"meta":  gin.H{"request_id": requestID},
		})
		return
	}

	newParentID, err := strconv.ParseInt(newParentIDStr, 10, 64)
	if err != nil || newParentID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid new parent id format: new parent id must be a positive integer",
			"type":  "validation_error",
			"field": "new_parent_id",
			"meta":  gin.H{"request_id": requestID},
		})
		return
	}

	err = h.service.ValidateMove(c.Request.Context(), id, &newParentID)
	if err != nil {
		h.handleServiceError(c, err, "Failed to validate move", requestID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":   true,
		"message": "Move is valid",
		"meta":    gin.H{"request_id": requestID},
	})
}

func (h *ObjectTypeHandler) GetSubtreeObjectCount(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

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

	count, err := h.service.GetSubtreeObjectCount(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get subtree object count", requestID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"object_type_id": id,
		"count":          count,
		"meta":           gin.H{"request_id": requestID},
	})
}
