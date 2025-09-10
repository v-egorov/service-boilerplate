package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	// SERVICE_IMPORT_SERVICES
)

type ServiceHandler struct {
	service Service
	logger  *logrus.Logger
}

func NewServiceHandler(service Service, logger *logrus.Logger) *ServiceHandler {
	return &ServiceHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ServiceHandler) CreateService(c *gin.Context) {
	var req CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create service"})
		return
	}

	h.logger.WithField("id", service.ID).Info("Service created")
	c.JSON(http.StatusCreated, service)
}

func (h *ServiceHandler) GetService(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.WithError(err).Error("Invalid service ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	service, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get service")
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	c.JSON(http.StatusOK, service)
}

func (h *ServiceHandler) UpdateService(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.WithError(err).Error("Invalid service ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	var req UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update service"})
		return
	}

	h.logger.WithField("id", id).Info("Service updated")
	c.JSON(http.StatusOK, service)
}

func (h *ServiceHandler) DeleteService(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.WithError(err).Error("Invalid service ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		h.logger.WithError(err).Error("Failed to delete service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete service"})
		return
	}

	h.logger.WithField("id", id).Info("Service deleted")
	c.JSON(http.StatusNoContent, nil)
}

func (h *ServiceHandler) ListServices(c *gin.Context) {
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

	services, err := h.service.List(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list services")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list services"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"services": services,
		"limit":    limit,
		"offset":   offset,
	})
}

// Service interface for dependency injection
type Service interface {
	Create(ctx context.Context, req CreateServiceRequest) (*ServiceResponse, error)
	GetByID(ctx context.Context, id int64) (*ServiceResponse, error)
	Update(ctx context.Context, id int64, req UpdateServiceRequest) (*ServiceResponse, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int) ([]*ServiceResponse, error)
}

// Request/Response types
type CreateServiceRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	// Add more fields as needed
}

type UpdateServiceRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	// Add more fields as needed
}

type ServiceResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	// Add more fields as needed
}
