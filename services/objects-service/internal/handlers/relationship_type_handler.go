package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/common/logging"
	"github.com/v-egorov/service-boilerplate/common/middleware"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
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
			"type":    "validation_error",
			"meta":    gin.H{"request_id": requestID},
		})
		return
	}

	// Set created_by and updated_by from authenticated user
	userID := middleware.GetAuthenticatedUserID(c)
	req.CreatedBy = userID
	req.UpdatedBy = userID

	rt, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		HandleError(c, err, requestID)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"type_key":   rt.TypeKey,
	}).Info("Relationship type created")

	c.JSON(http.StatusCreated, gin.H{
		"data":    rt.ToResponse(),
		"meta":    gin.H{"request_id": requestID},
	})
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
			"type":    "validation_error",
			"meta":    gin.H{"request_id": requestID},
		})
		return
	}

	// Parse boolean for required filter
	if requiredStr := c.Query("required"); requiredStr != "" {
		required := requiredStr == "true"
		filter.Required = &required
	}

	// Set defaults
	if filter.Limit < 1 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	rts, err := h.service.List(c.Request.Context(), &filter)
	if err != nil {
		HandleError(c, err, requestID)
		return
	}

	// Convert to response format
	responses := make([]models.RelationshipTypeResponse, 0, len(rts))
	for _, rt := range rts {
		responses = append(responses, *rt.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"data": responses,
		"pagination": models.PaginationResponse{
			Limit:  filter.Limit,
			Offset: filter.Offset,
			Total:  int64(len(rts)),
		},
		"meta": gin.H{"request_id": requestID},
	})
}

// GetByTypeKey handles GET /api/v1/relationship-types/:type_key
func (h *RelationshipTypeHandler) GetByTypeKey(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")
	typeKey := c.Param("type_key")

	if typeKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing type_key: type_key parameter is required",
			"type":  "validation_error",
			"field": "type_key",
			"meta":  gin.H{"request_id": requestID},
		})
		return
	}

	rt, err := h.service.GetByTypeKey(c.Request.Context(), typeKey)
	if err != nil {
		HandleError(c, err, requestID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    rt.ToResponse(),
		"meta":    gin.H{"request_id": requestID},
	})
}

// Update handles PUT /api/v1/relationship-types/:type_key
func (h *RelationshipTypeHandler) Update(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")
	typeKey := c.Param("type_key")

	if typeKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "type_key is required",
			"type":  "validation_error",
			"meta":  gin.H{"request_id": requestID},
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
			"type":    "validation_error",
			"meta":    gin.H{"request_id": requestID},
		})
		return
	}

	// Set updated_by from authenticated user
	userID := middleware.GetAuthenticatedUserID(c)
	req.UpdatedBy = userID

	rt, err := h.service.Update(c.Request.Context(), typeKey, &req)
	if err != nil {
		HandleError(c, err, requestID)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"type_key":   typeKey,
	}).Info("Relationship type updated")

	c.JSON(http.StatusOK, gin.H{
		"data":    rt.ToResponse(),
		"meta":    gin.H{"request_id": requestID},
	})
}

// Delete handles DELETE /api/v1/relationship-types/:type_key
func (h *RelationshipTypeHandler) Delete(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")
	typeKey := c.Param("type_key")

	if typeKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing type_key: type_key parameter is required",
			"type":  "validation_error",
			"field": "type_key",
			"meta":  gin.H{"request_id": requestID},
		})
		return
	}

	err := h.service.Delete(c.Request.Context(), typeKey)
	if err != nil {
		HandleError(c, err, requestID)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"type_key":   typeKey,
	}).Info("Relationship type deleted")

	c.Status(http.StatusNoContent)
}

