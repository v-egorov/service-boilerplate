package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/cache"
)

type PermissionServiceInterface interface {
	CheckPermission(ctx context.Context, userID, permission string) (bool, error)
	GetUserPermissions(ctx context.Context, userID string) ([]string, error)
}

type PermissionHandler struct {
	authService PermissionServiceInterface
	cache       cache.PermissionCache
	logger      *logrus.Logger
}

type CheckPermissionRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	Permission string `json:"permission" binding:"required"`
}

type CheckPermissionResponse struct {
	Allowed    bool   `json:"allowed"`
	UserID     string `json:"user_id"`
	Permission string `json:"permission"`
}

type UserPermissionsResponse struct {
	UserID      string   `json:"user_id"`
	Permissions []string `json:"permissions"`
}

func NewPermissionHandler(authService PermissionServiceInterface, permCache cache.PermissionCache, logger *logrus.Logger) *PermissionHandler {
	return &PermissionHandler{
		authService: authService,
		cache:       permCache,
		logger:      logger,
	}
}

func (h *PermissionHandler) CheckPermission(c *gin.Context) {
	var req CheckPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"type":    "validation_error",
			"meta":    gin.H{"request_id": c.GetHeader("X-Request-ID")},
		})
		return
	}

	allowed, err := h.authService.CheckPermission(c.Request.Context(), req.UserID, req.Permission)
	if err != nil {
		h.logger.WithError(err).Error("Failed to check permission")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "permission check failed",
			"type":    "internal_error",
			"meta":    gin.H{"request_id": c.GetHeader("X-Request-ID")},
		})
		return
	}

	c.JSON(http.StatusOK, CheckPermissionResponse{
		Allowed:    allowed,
		UserID:     req.UserID,
		Permission: req.Permission,
	})
}

func (h *PermissionHandler) GetUserPermissions(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "user_id required",
			"type":    "validation_error",
			"meta":    gin.H{"request_id": c.GetHeader("X-Request-ID")},
		})
		return
	}

	permissions, err := h.authService.GetUserPermissions(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user permissions")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get permissions",
			"type":    "internal_error",
			"meta":    gin.H{"request_id": c.GetHeader("X-Request-ID")},
		})
		return
	}

	c.JSON(http.StatusOK, UserPermissionsResponse{
		UserID:      userID,
		Permissions: permissions,
	})
}
