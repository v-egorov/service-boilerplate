package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	// ENTITY_IMPORT_MODELS
	// ENTITY_IMPORT_SERVICES
)

type EntityHandler struct {
	service Service
	logger  *logrus.Logger
}

func NewEntityHandler(service Service, logger *logrus.Logger) *EntityHandler {
	return &EntityHandler{
		service: service,
		logger:  logger,
	}
}

// handleServiceError handles different types of service errors and returns appropriate HTTP responses
func (h *EntityHandler) handleServiceError(c *gin.Context, err error, operation string) {
	h.logger.WithError(err).Error(operation)

	// Handle specific error types
	switch e := err.(type) {
	case models.ValidationError:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": e.Error(),
			"type":  "validation_error",
			"field": e.Field,
		})
		return

	case models.ConflictError:
		c.JSON(http.StatusConflict, gin.H{
			"error":    e.Error(),
			"type":     "conflict_error",
			"resource": e.Resource,
			"field":    e.Field,
			"value":    e.Value,
		})
		return

	case models.NotFoundError:
		c.JSON(http.StatusNotFound, gin.H{
			"error":    e.Error(),
			"type":     "not_found_error",
			"resource": e.Resource,
			"field":    e.Field,
			"value":    e.Value,
		})
		return

	case models.InternalError:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     "Internal server error",
			"type":      "internal_error",
			"operation": e.Operation,
		})
		return

	default:
		// Fallback for unknown error types
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "An unexpected error occurred",
			"type":  "unknown_error",
		})
		return
	}
}

func (h *EntityHandler) CreateEntity(c *gin.Context) {
	var req services.CreateEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"type":    "validation_error",
		})
		return
	}

	entity, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to create entity")
		return
	}

	h.logger.WithField("id", entity.ID).Info("Entity created")
	c.JSON(http.StatusCreated, gin.H{
		"data":    entity,
		"message": "Entity created successfully",
	})
}

func (h *EntityHandler) GetEntity(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.WithError(err).Error("Invalid entity ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid entity ID format",
			"details": "Entity ID must be a valid integer",
			"type":    "validation_error",
			"field":   "id",
		})
		return
	}

	entity, err := h.service.GetByID(c.Request.Context(), int64(id))
	if err != nil {
		h.handleServiceError(c, err, "Failed to get entity")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": entity,
	})
}

func (h *EntityHandler) ReplaceEntity(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.WithError(err).Error("Invalid entity ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid entity ID format",
			"details": "Entity ID must be a valid integer",
			"type":    "validation_error",
			"field":   "id",
		})
		return
	}

	var req services.ReplaceEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"type":    "validation_error",
		})
		return
	}

	entity, err := h.service.Replace(c.Request.Context(), int64(id), req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to replace entity")
		return
	}

	h.logger.WithField("id", entity.ID).Info("Entity replaced successfully")
	c.JSON(http.StatusOK, gin.H{
		"data":    entity,
		"message": "Entity replaced successfully",
	})
}

func (h *EntityHandler) UpdateEntity(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.WithError(err).Error("Invalid entity ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid entity ID format",
			"details": "Entity ID must be a valid integer",
			"type":    "validation_error",
			"field":   "id",
		})
		return
	}

	var req services.UpdateEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"type":    "validation_error",
		})
		return
	}

	entity, err := h.service.Update(c.Request.Context(), int64(id), req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to update entity")
		return
	}

	h.logger.WithField("id", entity.ID).Info("Entity updated successfully")
	c.JSON(http.StatusOK, gin.H{
		"data":    entity,
		"message": "Entity updated successfully",
	})
}

func (h *EntityHandler) DeleteEntity(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.WithError(err).Error("Invalid entity ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		h.logger.WithError(err).Error("Failed to delete entity")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete entity"})
		return
	}

	h.logger.WithField("id", id).Info("Entity deleted")
	c.JSON(http.StatusNoContent, nil)
}

func (h *EntityHandler) ListEntities(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	entities, err := h.service.List(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list entities")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list entities"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entities": entities,
		"limit":    limit,
		"offset":   offset,
	})
}

// Service interface for dependency injection
type Service interface {
	Create(ctx context.Context, req services.CreateEntityRequest) (*services.EntityResponse, error)
	GetByID(ctx context.Context, id int64) (*services.EntityResponse, error)
	Replace(ctx context.Context, id int64, req services.ReplaceEntityRequest) (*services.EntityResponse, error)
	Update(ctx context.Context, id int64, req services.UpdateEntityRequest) (*services.EntityResponse, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int) ([]*services.EntityResponse, error)
}
