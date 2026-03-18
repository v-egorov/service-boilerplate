package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/common/logging"
	"github.com/v-egorov/service-boilerplate/common/middleware"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/repository"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/services"
)

// RelationshipTypeServiceInterface defines the service operations needed for handlers
type RelationshipTypeServiceInterface interface {
	Create(ctx context.Context, req *models.CreateRelationshipTypeRequest) (*models.RelationshipType, error)
	GetByTypeKey(ctx context.Context, typeKey string) (*models.RelationshipType, error)
	Update(ctx context.Context, typeKey string, req *models.UpdateRelationshipTypeRequest) (*models.RelationshipType, error)
	Delete(ctx context.Context, typeKey string) error
	List(ctx context.Context, filter *models.RelationshipTypeFilter) ([]*models.RelationshipType, error)
}

// RelationshipTypeHandler handles HTTP requests for relationship types
type RelationshipTypeHandler struct {
	service        RelationshipTypeServiceInterface
	logger         *logrus.Logger
	standardLogger *logging.StandardLogger
}

// NewRelationshipTypeHandler creates a new RelationshipTypeHandler
func NewRelationshipTypeHandler(service services.RelationshipTypeService, logger *logrus.Logger) *RelationshipTypeHandler {
	return &RelationshipTypeHandler{
		service:        service,
		logger:         logger,
		standardLogger: logging.NewStandardLogger(logger, "objects-service"),
	}
}

// NewRelationshipTypeHandlerWithInterface creates a handler with a service interface (for testing)
func NewRelationshipTypeHandlerWithInterface(service RelationshipTypeServiceInterface, logger *logrus.Logger) *RelationshipTypeHandler {
	return &RelationshipTypeHandler{
		service:        service,
		logger:         logger,
		standardLogger: logging.NewStandardLogger(logger, "objects-service"),
	}
}

// Create handles POST /api/v1/relationship-types
func (h *RelationshipTypeHandler) Create(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	var req models.CreateRelationshipTypeRequest
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

	// Set created_by and updated_by from authenticated user
	userID := middleware.GetAuthenticatedUserID(c)
	req.CreatedBy = userID
	req.UpdatedBy = userID

	rt, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, requestID, err, "create relationship type")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"type_key":   rt.TypeKey,
	}).Info("Relationship type created")

	c.JSON(http.StatusCreated, rt.ToResponse())
}

// List handles GET /api/v1/relationship-types
func (h *RelationshipTypeHandler) List(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	var filter models.RelationshipTypeFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
		}).WithError(err).Error("Invalid query parameters")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid query parameters",
			"details": err.Error(),
			"type":    "validation_error",
		})
		return
	}

	// Parse boolean for required filter
	if requiredStr := c.Query("required"); requiredStr != "" {
		required := requiredStr == "true"
		filter.Required = &required
	}

	// Set defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}

	rts, err := h.service.List(c.Request.Context(), &filter)
	if err != nil {
		h.handleError(c, requestID, err, "list relationship types")
		return
	}

	// Convert to response format
	responses := make([]models.RelationshipTypeResponse, 0, len(rts))
	for _, rt := range rts {
		responses = append(responses, *rt.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"data": responses,
		"pagination": gin.H{
			"page":      filter.Page,
			"page_size": filter.PageSize,
		},
	})
}

// GetByTypeKey handles GET /api/v1/relationship-types/:type_key
func (h *RelationshipTypeHandler) GetByTypeKey(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")
	typeKey := c.Param("type_key")

	if typeKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "type_key is required",
			"type":  "validation_error",
		})
		return
	}

	rt, err := h.service.GetByTypeKey(c.Request.Context(), typeKey)
	if err != nil {
		h.handleError(c, requestID, err, "get relationship type")
		return
	}

	c.JSON(http.StatusOK, rt.ToResponse())
}

// Update handles PUT /api/v1/relationship-types/:type_key
func (h *RelationshipTypeHandler) Update(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")
	typeKey := c.Param("type_key")

	if typeKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "type_key is required",
			"type":  "validation_error",
		})
		return
	}

	var req models.UpdateRelationshipTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"type_key":   typeKey,
		}).WithError(err).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"type":    "validation_error",
		})
		return
	}

	// Set updated_by from authenticated user
	userID := middleware.GetAuthenticatedUserID(c)
	req.UpdatedBy = userID

	rt, err := h.service.Update(c.Request.Context(), typeKey, &req)
	if err != nil {
		h.handleError(c, requestID, err, "update relationship type")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"type_key":   typeKey,
	}).Info("Relationship type updated")

	c.JSON(http.StatusOK, rt.ToResponse())
}

// Delete handles DELETE /api/v1/relationship-types/:type_key
func (h *RelationshipTypeHandler) Delete(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")
	typeKey := c.Param("type_key")

	if typeKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "type_key is required",
			"type":  "validation_error",
		})
		return
	}

	err := h.service.Delete(c.Request.Context(), typeKey)
	if err != nil {
		h.handleError(c, requestID, err, "delete relationship type")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"type_key":   typeKey,
	}).Info("Relationship type deleted")

	c.Status(http.StatusNoContent)
}

// handleError maps service errors to HTTP status codes
func (h *RelationshipTypeHandler) handleError(c *gin.Context, requestID string, err error, operation string) {
	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
	}).WithError(err).Error(operation)

	// Check for specific error types
	switch {
	case errors.Is(err, services.ErrRelationshipTypeNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Relationship type not found",
			"type":  "not_found",
		})
	case errors.Is(err, services.ErrDuplicateRelationshipType):
		c.JSON(http.StatusConflict, gin.H{
			"error": "Relationship type already exists",
			"type":  "conflict",
		})
	case errors.Is(err, services.ErrInvalidCardinality):
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "Invalid cardinality value",
			"type":  "validation_error",
		})
	case errors.Is(err, services.ErrInvalidReverseType):
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "Invalid reverse type key",
			"type":  "validation_error",
		})
	case errors.Is(err, services.ErrRelationshipTypeInUse):
		c.JSON(http.StatusConflict, gin.H{
			"error": "Relationship type is in use and cannot be deleted",
			"type":  "conflict",
		})
	case errors.Is(err, services.ErrInvalidCountConstraint):
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "Invalid count constraint: min_count cannot exceed max_count",
			"type":  "validation_error",
		})
	case errors.Is(err, repository.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input",
			"type":  "validation_error",
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
			"type":  "internal_error",
		})
	}
}
