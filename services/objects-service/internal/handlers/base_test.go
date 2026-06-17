package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCheckOwnership(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("admin role bypasses ownership check", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		c.Request = nil
		c.Set("user_id", "user-123")
		c.Set("user_roles", []string{"admin"})

		result := CheckOwnership(c, "other-user")
		assert.True(t, result, "admin should bypass ownership check")
	})

	t.Run("object-type-admin does NOT bypass ownership check", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		c.Request = nil
		c.Set("user_id", "user-123")
		c.Set("user_roles", []string{"object-type-admin"})

		result := CheckOwnership(c, "other-user")
		assert.False(t, result, "object-type-admin should NOT bypass ownership check for non-types resources")
	})

	t.Run(":all permission scope grants access regardless of owner", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		c.Request = nil
		c.Set("user_id", "user-123")
		c.Set("user_roles", []string{"user"})
		c.Set("matched_permissions", []string{"objects:read:own", "relationships:update:all"})

		result := CheckOwnership(c, "other-user")
		assert.True(t, result, ":all permission should grant access even when not the owner")
	})

	t.Run(":own scope requires matching created_by", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		c.Request = nil
		c.Set("user_id", "user-123")
		c.Set("user_roles", []string{"user"})
		c.Set("matched_permissions", []string{"objects:read:own"})

		result := CheckOwnership(c, "other-user")
		assert.False(t, result, ":own scope should deny access when not the owner")
	})

	t.Run(":own scope allows access when created_by matches userID", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		c.Request = nil
		c.Set("user_id", "user-123")
		c.Set("user_roles", []string{"user"})
		c.Set("matched_permissions", []string{"objects:read:own"})

		result := CheckOwnership(c, "user-123")
		assert.True(t, result, ":own scope should allow access when created_by matches userID")
	})

	t.Run("owner matches userID without matched_permissions", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		c.Request = nil
		c.Set("user_id", "user-123")
		c.Set("user_roles", []string{"user"})

		result := CheckOwnership(c, "user-123")
		assert.True(t, result, "owner should pass ownership check without matched_permissions context")
	})

	t.Run("non-owner fails without matched_permissions", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		c.Request = nil
		c.Set("user_id", "user-123")
		c.Set("user_roles", []string{"user"})

		result := CheckOwnership(c, "other-user")
		assert.False(t, result, "non-owner should fail ownership check without matched_permissions context")
	})

	t.Run("empty userID returns false", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		c.Request = nil
		c.Set("user_id", "")
		c.Set("user_roles", []string{"user"})

		result := CheckOwnership(c, "other-user")
		assert.False(t, result, "empty userID should fail ownership check")
	})

	t.Run("empty createdByID returns false", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		c.Request = nil
		c.Set("user_id", "user-123")
		c.Set("user_roles", []string{"user"})

		result := CheckOwnership(c, "")
		assert.False(t, result, "empty createdByID should fail ownership check")
	})

	t.Run("nil matched_permissions treated as empty array", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		c.Request = nil
		c.Set("user_id", "user-123")
		c.Set("user_roles", []string{"user"})
		c.Set("matched_permissions", nil)

		result := CheckOwnership(c, "other-user")
		assert.False(t, result, "nil matched_permissions should be treated as empty array and fail ownership check")
	})
}

func TestHandleOwnershipViolation(t *testing.T) {
	t.Run("sends 403 with correct response format", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = nil

		HandleOwnershipViolation(c, "test-request-id")

		assert.Equal(t, http.StatusForbidden, w.Code, "HTTP status should be 403")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "response body should be valid JSON")
		assert.Equal(t, "permission_denied", response["type"], "error type should be permission_denied")
	})
}
