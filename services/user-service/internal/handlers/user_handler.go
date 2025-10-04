package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/common/logging"
	"github.com/v-egorov/service-boilerplate/services/user-service/internal/models"
	"github.com/v-egorov/service-boilerplate/services/user-service/internal/services"
	"go.opentelemetry.io/otel/trace"
)

type UserHandler struct {
	service        *services.UserService
	logger         *logrus.Logger
	auditLogger    *logging.AuditLogger
	standardLogger *logging.StandardLogger
}

func NewUserHandler(service *services.UserService, logger *logrus.Logger) *UserHandler {
	return &UserHandler{
		service:        service,
		logger:         logger,
		auditLogger:    logging.NewAuditLogger(logger, "user-service"),
		standardLogger: logging.NewStandardLogger(logger, "user-service"),
	}
}

// handleServiceError handles different types of service errors and returns appropriate HTTP responses
func (h *UserHandler) handleServiceError(c *gin.Context, err error, operation string, requestID string) {
	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
	}).WithError(err).Error(operation)

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
	// Extract trace information
	span := trace.SpanFromContext(c.Request.Context())
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		requestID := c.GetHeader("X-Request-ID")
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
		}).WithError(err).Error("Invalid request body")
		h.auditLogger.LogUserCreation(requestID, "", c.ClientIP(), c.GetHeader("User-Agent"), traceID, spanID, false, "Invalid request format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"type":    "validation_error",
		})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	requestID := c.GetHeader("X-Request-ID")

	user, err := h.service.CreateUser(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to create user", requestID)
		h.auditLogger.LogUserCreation(requestID, "", ipAddress, userAgent, traceID, spanID, false, err.Error())
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":    user.ID,
		"request_id": requestID,
	}).Info("User created successfully")
	h.standardLogger.UserOperation(requestID, user.ID.String(), "create", true, nil)
	h.auditLogger.LogUserCreation(requestID, user.ID.String(), ipAddress, userAgent, traceID, spanID, true, "")
	c.JSON(http.StatusCreated, gin.H{
		"data":    user,
		"message": "User created successfully",
	})
}

func (h *UserHandler) GetUser(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"user_id":    idStr,
		}).WithError(err).Error("Invalid user ID format")
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
		h.handleServiceError(c, err, "Failed to get user", requestID)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"user_id":    user.ID,
	}).Debug("User retrieved successfully")
	h.standardLogger.UserOperation(requestID, user.ID.String(), "get", true, nil)

	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}

func (h *UserHandler) ReplaceUser(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"user_id":    idStr,
		}).WithError(err).Error("Invalid user ID format")
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
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"user_id":    id.String(),
		}).WithError(err).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"type":    "validation_error",
		})
		return
	}

	user, err := h.service.ReplaceUser(c.Request.Context(), id, &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to replace user", requestID)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"user_id":    user.ID,
	}).Info("User replaced successfully")
	h.standardLogger.UserOperation(requestID, user.ID.String(), "replace", true, nil)
	c.JSON(http.StatusOK, gin.H{
		"data":    user,
		"message": "User replaced successfully",
	})
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"user_id":    idStr,
		}).WithError(err).Error("Invalid user ID format")
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
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"user_id":    id.String(),
		}).WithError(err).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"type":    "validation_error",
		})
		return
	}

	user, err := h.service.UpdateUser(c.Request.Context(), id, &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to update user", requestID)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"user_id":    user.ID,
	}).Info("User updated successfully")
	h.standardLogger.UserOperation(requestID, user.ID.String(), "update", true, nil)
	c.JSON(http.StatusOK, gin.H{
		"data":    user,
		"message": "User updated successfully",
	})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"user_id":    idStr,
		}).WithError(err).Error("Invalid user ID format")
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
		h.handleServiceError(c, err, "Failed to delete user", requestID)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"user_id":    id,
	}).Info("User deleted successfully")
	h.standardLogger.UserOperation(requestID, id.String(), "delete", true, nil)
	c.JSON(http.StatusNoContent, nil)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

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
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"limit":      limit,
		}).Warn("Limit too high, capping at 100")
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
		h.handleServiceError(c, err, "Failed to list users", requestID)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"limit":      limit,
		"offset":     offset,
		"count":      len(users),
	}).Debug("Users listed successfully")

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
	requestID := c.GetHeader("X-Request-ID")

	email := c.Param("email")
	if email == "" {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
		}).Error("Email parameter is required")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email parameter is required",
			"type":  "validation_error",
			"field": "email",
		})
		return
	}

	user, err := h.service.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get user by email", requestID)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"user_id":    user.ID,
		"email":      user.Email,
	}).Debug("User retrieved by email successfully")
	h.standardLogger.UserOperation(requestID, user.ID.String(), "get_by_email", true, nil)

	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}

func (h *UserHandler) GetUserWithPasswordByEmail(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	email := c.Param("email")
	if email == "" {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
		}).Error("Email parameter is required")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email parameter is required",
			"type":  "validation_error",
			"field": "email",
		})
		return
	}

	userLogin, err := h.service.GetUserWithPasswordByEmail(c.Request.Context(), email)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get user with password by email", requestID)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"user_id":    userLogin.User.ID,
		"email":      userLogin.User.Email,
	}).Debug("User with password retrieved by email successfully")
	h.standardLogger.UserOperation(requestID, userLogin.User.ID.String(), "get_with_password_by_email", true, nil)

	c.JSON(http.StatusOK, userLogin)
}
