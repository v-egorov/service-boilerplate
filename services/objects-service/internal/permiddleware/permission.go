package permiddleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/common/middleware"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/client"
)

type PermissionMiddlewareConfig struct {
	AuthClient client.AuthClient
	Logger     *logrus.Logger
}

type RequirePermissionFunc func(requiredPermissions ...string) gin.HandlerFunc

func NewPermissionMiddleware(cfg PermissionMiddlewareConfig) RequirePermissionFunc {
	return func(requiredPermissions ...string) gin.HandlerFunc {
		return func(c *gin.Context) {
			userID := middleware.GetAuthenticatedUserID(c)
			if userID == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
				c.Abort()
				return
			}

			for _, permission := range requiredPermissions {
				allowed, err := cfg.AuthClient.CheckPermission(c.Request.Context(), userID, permission)
				if err != nil {
					if cfg.Logger != nil {
						cfg.Logger.WithError(err).Error("Permission check failed")
					}
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Permission check failed"})
					c.Abort()
					return
				}

				if allowed {
					c.Next()
					return
				}
			}

			if cfg.Logger != nil {
				cfg.Logger.WithFields(logrus.Fields{
					"user_id":  userID,
					"required": requiredPermissions,
				}).Warn("Permission denied")
			}

			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
		}
	}
}
