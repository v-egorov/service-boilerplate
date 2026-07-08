package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/common/logging"
	"github.com/v-egorov/service-boilerplate/common/middleware"
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
			"type":    "validation_error",
			"meta":    gin.H{"request_id": requestID},
		})
		return
	}

	userID := middleware.GetAuthenticatedUserID(c)
	if userID != "" {
		req.CreatedBy = userID
	}

	rel, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		HandleError(c, err, requestID)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id":      requestID,
		"relationship_id": rel.ObjectID,
	}).Info("Relationship created successfully")

	c.JSON(http.StatusCreated, gin.H{
		"data":    rel.ToResponse(),
		"meta":    gin.H{"request_id": requestID},
	})
}

func (h *RelationshipHandler) GetByPublicID(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")
	publicIDStr := c.Param("public_id")

	publicID, err := uuid.Parse(publicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid public ID format",
			"type":    "validation_error",
			"field":   "public_id",
			"meta":    gin.H{"request_id": requestID},
		})
		return
	}

	rel, err := h.service.GetByPublicID(c.Request.Context(), publicID)
	if err != nil {
		HandleError(c, err, requestID)
		return
	}

	if !CheckOwnership(c, *rel.CreatedBy) {
		HandleOwnershipViolation(c, requestID)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id":      requestID,
		"relationship_id": rel.ObjectID,
	}).Info("Relationship retrieved successfully")

	c.JSON(http.StatusOK, gin.H{
		"data":    rel.ToResponse(),
		"meta":    gin.H{"request_id": requestID},
	})
}

func (h *RelationshipHandler) Update(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")
	publicIDStr := c.Param("public_id")

	publicID, err := uuid.Parse(publicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid public ID format",
			"type":    "validation_error",
			"field":   "public_id",
			"meta":    gin.H{"request_id": requestID},
		})
		return
	}

	existingRel, err := h.service.GetByPublicID(c.Request.Context(), publicID)
	if err != nil {
		HandleError(c, err, requestID)
		return
	}

	if !CheckOwnership(c, *existingRel.CreatedBy) {
		HandleOwnershipViolation(c, requestID)
		return
	}

	var req models.UpdateRelationshipRequest
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

	userID := middleware.GetAuthenticatedUserID(c)
	if userID != "" {
		req.UpdatedBy = userID
	}

	rel, err := h.service.Update(c.Request.Context(), publicID, &req)
	if err != nil {
		HandleError(c, err, requestID)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id":      requestID,
		"relationship_id": rel.ObjectID,
	}).Info("Relationship updated successfully")

	c.JSON(http.StatusOK, gin.H{
		"data":    rel.ToResponse(),
		"meta":    gin.H{"request_id": requestID},
	})
}

func (h *RelationshipHandler) Delete(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")
	publicIDStr := c.Param("public_id")

	publicID, err := uuid.Parse(publicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid public ID format",
			"type":    "validation_error",
			"field":   "public_id",
			"meta":    gin.H{"request_id": requestID},
		})
		return
	}

	existingRel, err := h.service.GetByPublicID(c.Request.Context(), publicID)
	if err != nil {
		HandleError(c, err, requestID)
		return
	}

	if !CheckOwnership(c, *existingRel.CreatedBy) {
		HandleOwnershipViolation(c, requestID)
		return
	}

	err = h.service.Delete(c.Request.Context(), publicID)
	if err != nil {
		HandleError(c, err, requestID)
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
			"type":    "validation_error",
			"meta":    gin.H{"request_id": requestID},
		})
		return
	}

	if filter.Limit < 1 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	// NOTE: No self-filtering by user ID for List(). Permission checks are handled
	// at the gateway level (via permiddleware). If a client needs ownership filtering,
	// it can pass an explicit query parameter (future enhancement).

	rels, err := h.service.List(c.Request.Context(), &filter)
	if err != nil {
		HandleError(c, err, requestID)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"count":      len(rels),
	}).Info("Relationships listed successfully")

	response := models.RelationshipListResponse{
		Data: make([]models.RelationshipResponse, len(rels)),
		Pagination: models.PaginationResponse{
			Limit:  filter.Limit,
			Offset: filter.Offset,
		},
	}

	for i, rel := range rels {
		resp := rel.ToResponse()
		response.Data[i] = *resp
	}
	response.Pagination.Total = int64(len(rels))

	c.JSON(http.StatusOK, gin.H{
		"data":     response.Data,
		"pagination": response.Pagination,
		"meta":     gin.H{"request_id": requestID},
	})
}

func (h *RelationshipHandler) GetForObject(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")
	objectPublicIDStr := c.Param("public_id")

	objectPublicID, err := uuid.Parse(objectPublicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid public ID format",
			"type":    "validation_error",
			"field":   "public_id",
			"meta":    gin.H{"request_id": requestID},
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
			"type":    "validation_error",
			"meta":    gin.H{"request_id": requestID},
		})
		return
	}

	rels, err := h.service.GetForObject(c.Request.Context(), objectPublicID, &filter)
	if err != nil {
		HandleError(c, err, requestID)
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
		"data":            response,
		"meta":            gin.H{"request_id": requestID},
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
			"type":    "validation_error",
			"field":   "public_id",
			"meta":    gin.H{"request_id": requestID},
		})
		return
	}

	rels, err := h.service.GetForObjectByType(c.Request.Context(), objectPublicID, typeKey)
	if err != nil {
		HandleError(c, err, requestID)
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

	c.JSON(http.StatusOK, gin.H{
		"data":    response,
		"meta":    gin.H{"request_id": requestID},
	})
}

