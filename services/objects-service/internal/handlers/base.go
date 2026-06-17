package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/v-egorov/service-boilerplate/common/middleware"
)

// CheckOwnership checks if the authenticated user owns a resource by comparing created_by against matched_permissions.
// Returns true if:
// - User has "admin" role (full system access)
// - Any matched permission ends with ":all" (e.g., objects:update:all) from permission middleware
// - Resource was created by this user (owner check)
func CheckOwnership(c *gin.Context, createdByID string) bool {
	userID := middleware.GetAuthenticatedUserID(c)
	if userID == "" || createdByID == "" {
		return false
	}

	userRoles := middleware.GetAuthenticatedUserRoles(c)
	for _, role := range userRoles {
		if role == "admin" {
			return true
		}
	}

	matchedPermissions, exists := c.Get("matched_permissions")
	if !exists {
		return createdByID == userID
	}

	perms, ok := matchedPermissions.([]string)
	if !ok {
		return false
	}

	for _, perm := range perms {
		if strings.HasSuffix(perm, ":all") {
			return true
		}
	}

	return createdByID == userID
}

// HandleOwnershipViolation sends a 403 response for ownership checks.
func HandleOwnershipViolation(c *gin.Context, requestID string) {
	c.JSON(http.StatusForbidden, gin.H{
		"error":  "You can only access your own resources",
		"type":   "permission_denied",
		"meta":   gin.H{"request_id": requestID},
	})
}
