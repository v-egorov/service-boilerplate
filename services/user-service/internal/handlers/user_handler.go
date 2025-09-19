package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/services/user-service/internal/models"
	"github.com/v-egorov/service-boilerplate/services/user-service/internal/services"
)

type UserHandler struct {
	service *services.UserService
	logger  *logrus.Logger
}

func NewUserHandler(service *services.UserService, logger *logrus.Logger) *UserHandler {
	return &UserHandler{
		service: service,
		logger:  logger,
	}
}

// handleServiceError handles different types of service errors and returns appropriate HTTP responses
func (h *UserHandler) handleServiceError(c *gin.Context, err error, operation string) {
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

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"type":    "validation_error",
		})
		return
	}

	user, err := h.service.CreateUser(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to create user")
		return
	}

	h.logger.WithField("user_id", user.ID).Info("User created successfully")
	c.JSON(http.StatusCreated, gin.H{
		"data":    user,
		"message": "User created successfully",
	})
}

func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.WithError(err).Error("Invalid user ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID format",
			"details": "User ID must be a valid UUID",
			"type":    "validation_error",
			"field":   "id",
		})
		return
	}

	user, err := h.service.GetUser(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get user")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}

func (h *UserHandler) ReplaceUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.WithError(err).Error("Invalid user ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID format",
			"details": "User ID must be a valid UUID",
			"type":    "validation_error",
			"field":   "id",
		})
		return
	}

	var req models.ReplaceUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"type":    "validation_error",
		})
		return
	}

	user, err := h.service.ReplaceUser(c.Request.Context(), id, &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to replace user")
		return
	}

	h.logger.WithField("user_id", user.ID).Info("User replaced successfully")
	c.JSON(http.StatusOK, gin.H{
		"data":    user,
		"message": "User replaced successfully",
	})
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.WithError(err).Error("Invalid user ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID format",
			"details": "User ID must be a valid UUID",
			"type":    "validation_error",
			"field":   "id",
		})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"type":    "validation_error",
		})
		return
	}

	user, err := h.service.UpdateUser(c.Request.Context(), id, &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to update user")
		return
	}

	h.logger.WithField("user_id", user.ID).Info("User updated successfully")
	c.JSON(http.StatusOK, gin.H{
		"data":    user,
		"message": "User updated successfully",
	})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.WithError(err).Error("Invalid user ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID format",
			"details": "User ID must be a valid UUID",
			"type":    "validation_error",
			"field":   "id",
		})
		return
	}

	err = h.service.DeleteUser(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err, "Failed to delete user")
		return
	}

	h.logger.WithField("user_id", id).Info("User deleted successfully")
	c.JSON(http.StatusNoContent, nil)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
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

	// Validate limits
	if limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Limit too high",
			"details": "Maximum limit is 100",
			"type":    "validation_error",
			"field":   "limit",
		})
		return
	}

	users, err := h.service.ListUsers(c.Request.Context(), limit, offset)
	if err != nil {
		h.handleServiceError(c, err, "Failed to list users")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": users,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(users),
		},
	})
}

func (h *UserHandler) GetUserByEmail(c *gin.Context) {
	email := c.Param("email")
	if email == "" {
		h.logger.Error("Email parameter is required")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email parameter is required",
			"type":  "validation_error",
			"field": "email",
		})
		return
	}

	user, err := h.service.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get user by email")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}

func (h *UserHandler) GetUserWithPasswordByEmail(c *gin.Context) {
	email := c.Param("email")
	if email == "" {
		h.logger.Error("Email parameter is required")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email parameter is required",
			"type":  "validation_error",
			"field": "email",
		})
		return
	}

	userLogin, err := h.service.GetUserWithPasswordByEmail(c.Request.Context(), email)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get user with password by email")
		return
	}

	c.JSON(http.StatusOK, userLogin)
}
