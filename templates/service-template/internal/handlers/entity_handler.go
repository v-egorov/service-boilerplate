package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/common/logging"
	"github.com/v-egorov/service-boilerplate/common/middleware"
	"go.opentelemetry.io/otel/trace"
	// ENTITY_IMPORT_MODELS
	// ENTITY_IMPORT_SERVICES
)

type EntityHandler struct {
	service        Service
	logger         *logrus.Logger
	standardLogger *logging.StandardLogger
	auditLogger    *logging.AuditLogger
}

func NewEntityHandler(service Service, logger *logrus.Logger, standardLogger *logging.StandardLogger) *EntityHandler {
	return &EntityHandler{
		service:        service,
		logger:         logger,
		standardLogger: standardLogger,
		auditLogger:    logging.NewAuditLogger(logger, "SERVICE_NAME"),
	}
}

// handleServiceError handles different types of service errors and returns appropriate HTTP responses
func (h *EntityHandler) handleServiceError(c *gin.Context, err error, operation string) {
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = "unknown"
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"error":      err.Error(),
	}).Error(operation)

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
	// Extract trace information
	span := trace.SpanFromContext(c.Request.Context())
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = "unknown"
	}

	var req services.CreateEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Invalid request body")
		actorUserID := middleware.GetAuthenticatedUserID(c)
		h.auditLogger.LogEntityCreation(actorUserID, requestID, "", c.ClientIP(), c.GetHeader("User-Agent"), traceID, spanID, false, "Invalid request format")
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
		actorUserID := middleware.GetAuthenticatedUserID(c)
		h.auditLogger.LogEntityCreation(actorUserID, requestID, "", c.ClientIP(), c.GetHeader("User-Agent"), traceID, spanID, false, err.Error())
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"id":         entity.ID,
	}).Info("Entity created")

	actorUserID := middleware.GetAuthenticatedUserID(c)
	h.standardLogger.EntityOperation(requestID, actorUserID, fmt.Sprintf("%d", entity.ID), "create", true, nil)
	h.auditLogger.LogEntityCreation(actorUserID, requestID, fmt.Sprintf("%d", entity.ID), c.ClientIP(), c.GetHeader("User-Agent"), traceID, spanID, true, "")

	c.JSON(http.StatusCreated, gin.H{
		"data":    entity,
		"message": "Entity created successfully",
	})
}

func (h *EntityHandler) GetEntity(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = "unknown"
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Invalid entity ID format")
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

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"id":         entity.ID,
	}).Info("Entity retrieved")

	h.standardLogger.EntityOperation(requestID, "", fmt.Sprintf("%d", entity.ID), "get", true, nil)

	c.JSON(http.StatusOK, gin.H{
		"data": entity,
	})
}

func (h *EntityHandler) ReplaceEntity(c *gin.Context) {
	// Extract trace information
	span := trace.SpanFromContext(c.Request.Context())
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = "unknown"
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Invalid entity ID format")
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
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Invalid request body")
		h.auditLogger.LogEntityUpdate("", requestID, fmt.Sprintf("%d", id), c.ClientIP(), c.GetHeader("User-Agent"), traceID, spanID, false, "Invalid request format")
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
		h.auditLogger.LogEntityUpdate("", requestID, fmt.Sprintf("%d", id), c.ClientIP(), c.GetHeader("User-Agent"), traceID, spanID, false, err.Error())
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"id":         entity.ID,
	}).Info("Entity replaced successfully")

	h.standardLogger.EntityOperation(requestID, "", fmt.Sprintf("%d", entity.ID), "replace", true, nil)
	h.auditLogger.LogEntityUpdate("", requestID, fmt.Sprintf("%d", entity.ID), c.ClientIP(), c.GetHeader("User-Agent"), traceID, spanID, true, "")

	c.JSON(http.StatusOK, gin.H{
		"data":    entity,
		"message": "Entity replaced successfully",
	})
}

func (h *EntityHandler) UpdateEntity(c *gin.Context) {
	// Extract trace information
	span := trace.SpanFromContext(c.Request.Context())
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = "unknown"
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Invalid entity ID format")
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
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Invalid request body")
		h.auditLogger.LogEntityUpdate("", requestID, fmt.Sprintf("%d", id), c.ClientIP(), c.GetHeader("User-Agent"), traceID, spanID, false, "Invalid request format")
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
		h.auditLogger.LogEntityUpdate("", requestID, fmt.Sprintf("%d", id), c.ClientIP(), c.GetHeader("User-Agent"), traceID, spanID, false, err.Error())
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"id":         entity.ID,
	}).Info("Entity updated successfully")

	h.standardLogger.EntityOperation(requestID, "", fmt.Sprintf("%d", entity.ID), "update", true, nil)
	h.auditLogger.LogEntityUpdate("", requestID, fmt.Sprintf("%d", entity.ID), c.ClientIP(), c.GetHeader("User-Agent"), traceID, spanID, true, "")

	c.JSON(http.StatusOK, gin.H{
		"data":    entity,
		"message": "Entity updated successfully",
	})
}

func (h *EntityHandler) DeleteEntity(c *gin.Context) {
	// Extract trace information
	span := trace.SpanFromContext(c.Request.Context())
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = "unknown"
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Invalid entity ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		h.logger.WithError(err).Error("Failed to delete entity")
		h.auditLogger.LogEntityDeletion("", requestID, fmt.Sprintf("%d", id), c.ClientIP(), c.GetHeader("User-Agent"), traceID, spanID, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete entity"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"id":         id,
	}).Info("Entity deleted")

	h.standardLogger.EntityOperation(requestID, "", fmt.Sprintf("%d", id), "delete", true, nil)
	h.auditLogger.LogEntityDeletion("", requestID, fmt.Sprintf("%d", id), c.ClientIP(), c.GetHeader("User-Agent"), traceID, spanID, true, "")

	c.JSON(http.StatusNoContent, nil)
}

func (h *EntityHandler) ListEntities(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = "unknown"
	}

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
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to list entities")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list entities"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"count":      len(entities),
		"limit":      limit,
		"offset":     offset,
	}).Info("Entities listed")

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
