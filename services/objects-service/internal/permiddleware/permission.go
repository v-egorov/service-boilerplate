package permiddleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/common/middleware"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/client"
)

// RouteConfig holds runtime-per-route configuration. TypeKey is required — every object type has one.
type RouteConfig struct {
	TypeKey    string // e.g. "relationships", "objects" (required, never empty)
	HTTPMethod string // "GET", "POST", "PUT", "DELETE"
}

// PermissionMapping defines which permission strings to check for each HTTP method.
var PermissionMapping = map[string][]string{
	"POST": {""}, // build dynamically: {type_key}:create
	"GET":  {"read"},      // {type_key}:read:all, {type_key}:read:own
	"PUT":  {"update"},    // {type_key}:update:all, {type_key}:update:own
	"DELETE": {"delete"},  // {type_key}:delete:all, {type_key}:delete:own
}

func buildPermissionStrings(cfg RouteConfig) []string {
	switch cfg.HTTPMethod {
	case "POST":
		return []string{cfg.TypeKey + ":create"}
	case "GET":
		return []string{cfg.TypeKey + ":read:all", cfg.TypeKey + ":read:own"}
	case "PUT":
		return []string{cfg.TypeKey + ":update:all", cfg.TypeKey + ":update:own"}
	case "DELETE":
		return []string{cfg.TypeKey + ":delete:all", cfg.TypeKey + ":delete:own"}
	default:
		// Fallback for any other HTTP method — check flat permission
		return []string{cfg.TypeKey + ":" + strings.ToLower(cfg.HTTPMethod)}
	}
}

// NewPermissionMiddleware creates a gin.HandlerFunc from config using data-driven type_key routing.
func NewPermissionMiddleware(cfg RouteConfig, authClient client.AuthClient, logger *logrus.Logger) gin.HandlerFunc {
	requiredPermissions := buildPermissionStrings(cfg)

	return func(c *gin.Context) {
		userID := middleware.GetAuthenticatedUserID(c)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		jwtToken := c.GetHeader("Authorization")
		if strings.HasPrefix(jwtToken, "Bearer ") {
			jwtToken = strings.TrimPrefix(jwtToken, "Bearer ")
		} else {
			jwtToken = ""
		}

		var matchedPermissions []string

		for _, permission := range requiredPermissions {
			allowed, err := authClient.CheckPermission(c.Request.Context(), userID, permission, jwtToken)
			if err != nil {
				if logger != nil {
					logger.WithError(err).WithField("permission", permission).Error("Permission check failed")
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Permission check failed"})
				c.Abort()
				return
			}

			if allowed {
				matchedPermissions = append(matchedPermissions, permission)
			}
		}

		if len(matchedPermissions) == 0 {
			if logger != nil {
				logger.WithFields(logrus.Fields{
					"user_id":    userID,
					"type_key":   cfg.TypeKey,
					"method":     cfg.HTTPMethod,
					"required":   requiredPermissions,
				}).Warn("Permission denied")
			}
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Set("matched_permissions", matchedPermissions)
		c.Next()
	}
}

// MultiPermissionMiddleware checks multiple explicit permission strings (OR logic).
// Used for admin routes and bulk operations where multiple distinct permissions apply.
func MultiPermissionMiddleware(requiredPermissions []string, authClient client.AuthClient, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetAuthenticatedUserID(c)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		jwtToken := c.GetHeader("Authorization")
		if strings.HasPrefix(jwtToken, "Bearer ") {
			jwtToken = strings.TrimPrefix(jwtToken, "Bearer ")
		} else {
			jwtToken = ""
		}

		var matchedPermissions []string

		for _, permission := range requiredPermissions {
			allowed, err := authClient.CheckPermission(c.Request.Context(), userID, permission, jwtToken)
			if err != nil {
				if logger != nil {
					logger.WithError(err).WithField("permission", permission).Error("Permission check failed")
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Permission check failed"})
				c.Abort()
				return
			}

			if allowed {
				matchedPermissions = append(matchedPermissions, permission)
			}
		}

		if len(matchedPermissions) == 0 {
			if logger != nil {
				logger.WithFields(logrus.Fields{
					"user_id":    userID,
					"required":   requiredPermissions,
				}).Warn("Permission denied")
			}
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Set("matched_permissions", matchedPermissions)
		c.Next()
	}
}
