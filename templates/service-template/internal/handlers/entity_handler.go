package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

func (h *EntityHandler) CreateEntity(c *gin.Context) {
	var req services.CreateEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create entity")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create entity"})
		return
	}

	h.logger.WithField("id", entity.ID).Info("Entity created")
	c.JSON(http.StatusCreated, entity)
}

func (h *EntityHandler) GetEntity(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.WithError(err).Error("Invalid entity ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get entity")
		c.JSON(http.StatusNotFound, gin.H{"error": "Entity not found"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

func (h *EntityHandler) UpdateEntity(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.WithError(err).Error("Invalid entity ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

	var req services.UpdateEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update entity")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update entity"})
		return
	}

	h.logger.WithField("id", id).Info("Entity updated")
	c.JSON(http.StatusOK, entity)
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
	Update(ctx context.Context, id int64, req services.UpdateEntityRequest) (*services.EntityResponse, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int) ([]*services.EntityResponse, error)
}
