package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/common/logging"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/services"
)

type RelationshipHandler struct {
	service        services.RelationshipService
	logger         *logrus.Logger
	standardLogger *logging.StandardLogger
}

func NewRelationshipHandler(service services.RelationshipService, logger *logrus.Logger) *RelationshipHandler {
	return &RelationshipHandler{
		service:        service,
		logger:         logger,
		standardLogger: logging.NewStandardLogger(logger, "objects-service"),
	}
}

func (h *RelationshipHandler) Create(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	var req models.CreateRelationshipRequest
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

	rel, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, requestID, err, "create relationship")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id":      requestID,
		"relationship_id": rel.ObjectID,
	}).Info("Relationship created successfully")

	c.JSON(http.StatusCreated, rel.ToResponse())
}

func (h *RelationshipHandler) GetByPublicID(c *gin.Context) {
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

	rel, err := h.service.GetByPublicID(c.Request.Context(), publicID)
	if err != nil {
		h.handleError(c, requestID, err, "get relationship")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id":      requestID,
		"relationship_id": rel.ObjectID,
	}).Info("Relationship retrieved successfully")

	c.JSON(http.StatusOK, rel.ToResponse())
}

func (h *RelationshipHandler) Update(c *gin.Context) {
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

	var req models.UpdateRelationshipRequest
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

	rel, err := h.service.Update(c.Request.Context(), publicID, &req)
	if err != nil {
		h.handleError(c, requestID, err, "update relationship")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id":      requestID,
		"relationship_id": rel.ObjectID,
	}).Info("Relationship updated successfully")

	c.JSON(http.StatusOK, rel.ToResponse())
}

func (h *RelationshipHandler) Delete(c *gin.Context) {
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

	err = h.service.Delete(c.Request.Context(), publicID)
	if err != nil {
		h.handleError(c, requestID, err, "delete relationship")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"public_id":  publicIDStr,
	}).Info("Relationship deleted successfully")

	c.Status(http.StatusNoContent)
}

func (h *RelationshipHandler) List(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	var filter models.RelationshipFilter
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

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}

	rels, err := h.service.List(c.Request.Context(), &filter)
	if err != nil {
		h.handleError(c, requestID, err, "list relationships")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"count":      len(rels),
	}).Info("Relationships listed successfully")

	response := models.RelationshipListResponse{
		Data: make([]models.RelationshipResponse, len(rels)),
		Pagination: models.PaginationResponse{
			Limit:  filter.PageSize,
			Offset: (filter.Page - 1) * filter.PageSize,
		},
	}

	for i, rel := range rels {
		resp := rel.ToResponse()
		response.Data[i] = *resp
	}
	response.Pagination.Total = int64(len(rels))

	c.JSON(http.StatusOK, response)
}

func (h *RelationshipHandler) GetForObject(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")
	objectPublicIDStr := c.Param("public_id")

	objectPublicID, err := uuid.Parse(objectPublicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid public ID format",
			"details": "Public ID must be a valid UUID",
			"type":    "validation_error",
			"field":   "public_id",
		})
		return
	}

	var filter models.RelationshipFilterForType
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

	rels, err := h.service.GetForObject(c.Request.Context(), objectPublicID, &filter)
	if err != nil {
		h.handleError(c, requestID, err, "get relationships for object")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"object_id":  objectPublicIDStr,
		"count":      len(rels),
	}).Info("Relationships for object retrieved successfully")

	response := make([]models.RelationshipResponse, len(rels))
	for i, rel := range rels {
		resp := rel.ToResponse()
		response[i] = *resp
	}

	c.JSON(http.StatusOK, gin.H{
		"relationships": response,
	})
}

func (h *RelationshipHandler) GetForObjectByType(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")
	objectPublicIDStr := c.Param("public_id")
	typeKey := c.Param("type_key")

	objectPublicID, err := uuid.Parse(objectPublicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid public ID format",
			"details": "Public ID must be a valid UUID",
			"type":    "validation_error",
			"field":   "public_id",
		})
		return
	}

	rels, err := h.service.GetForObjectByType(c.Request.Context(), objectPublicID, typeKey)
	if err != nil {
		h.handleError(c, requestID, err, "get relationships for object by type")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"object_id":  objectPublicIDStr,
		"type_key":   typeKey,
		"count":      len(rels),
	}).Info("Relationships for object by type retrieved successfully")

	response := make([]models.RelationshipResponse, len(rels))
	for i, rel := range rels {
		resp := rel.ToResponse()
		response[i] = *resp
	}

	c.JSON(http.StatusOK, response)
}

func (h *RelationshipHandler) handleError(c *gin.Context, requestID string, err error, operation string) {
	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"operation":  operation,
	}).WithError(err).Error("Relationship operation failed")

	statusCode := http.StatusInternalServerError
	errorMessage := "Internal server error"
	errorType := "internal_error"

	switch {
	case errors.Is(err, services.ErrRelationshipNotFound):
		statusCode = http.StatusNotFound
		errorMessage = "Relationship not found"
		errorType = "not_found"
	case errors.Is(err, services.ErrDuplicateRelationship):
		statusCode = http.StatusConflict
		errorMessage = "Relationship already exists"
		errorType = "conflict"
	case errors.Is(err, services.ErrSourceObjectNotFound):
		statusCode = http.StatusNotFound
		errorMessage = "Source object not found"
		errorType = "not_found"
	case errors.Is(err, services.ErrTargetObjectNotFound):
		statusCode = http.StatusNotFound
		errorMessage = "Target object not found"
		errorType = "not_found"
	case errors.Is(err, services.ErrRelationshipTypeNotFound):
		statusCode = http.StatusNotFound
		errorMessage = "Relationship type not found"
		errorType = "not_found"
	case errors.Is(err, services.ErrCircularRelationship):
		statusCode = http.StatusUnprocessableEntity
		errorMessage = "Cannot create circular relationship"
		errorType = "validation_error"
	case errors.Is(err, services.ErrCardinalityViolation):
		statusCode = http.StatusUnprocessableEntity
		errorMessage = "Cardinality constraint violated"
		errorType = "validation_error"
	case errors.Is(err, services.ErrSourceTargetSame):
		statusCode = http.StatusBadRequest
		errorMessage = "Source and target cannot be the same"
		errorType = "validation_error"
	}

	c.JSON(statusCode, gin.H{
		"error":   errorMessage,
		"details": err.Error(),
		"type":    errorType,
	})
}
