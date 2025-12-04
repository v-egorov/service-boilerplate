package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/common/logging"
	"github.com/v-egorov/service-boilerplate/common/middleware"
	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/models"
	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/services"
	"go.opentelemetry.io/otel/trace"
)

type AuthHandler struct {
	authService    services.AuthServiceInterface
	logger         *logrus.Logger
	auditLogger    *logging.AuditLogger
	standardLogger *logging.StandardLogger
}

func NewAuthHandler(authService services.AuthServiceInterface, logger *logrus.Logger) *AuthHandler {
	return &AuthHandler{
		authService:    authService,
		logger:         logger,
		auditLogger:    logging.NewAuditLogger(logger, "auth-service"),
		standardLogger: logging.NewStandardLogger(logger, "auth-service"),
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	// Extract trace information
	span := trace.SpanFromContext(c.Request.Context())
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid login request")
		h.auditLogger.LogAuthAttempt("", c.GetHeader("X-Request-ID"), c.ClientIP(), c.GetHeader("User-Agent"), req.Email, traceID, spanID, false, "Invalid request format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	requestID := c.GetHeader("X-Request-ID")

	response, err := h.authService.Login(c.Request.Context(), &req, ipAddress, userAgent)
	if err != nil {
		h.standardLogger.AuthOperation(requestID, "", req.Email, "login", false, err)
		h.auditLogger.LogAuthAttempt("", requestID, ipAddress, userAgent, req.Email, traceID, spanID, false, err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	h.standardLogger.AuthOperation(requestID, response.User.ID.String(), req.Email, "login", true, nil)
	h.auditLogger.LogAuthAttempt(response.User.ID.String(), requestID, ipAddress, userAgent, req.Email, traceID, spanID, true, "")
	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) Register(c *gin.Context) {
	// Extract trace information
	span := trace.SpanFromContext(c.Request.Context())
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid register request")
		h.auditLogger.LogUserCreation("", c.GetHeader("X-Request-ID"), "", c.ClientIP(), c.GetHeader("User-Agent"), traceID, spanID, false, "Invalid request format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	requestID := c.GetHeader("X-Request-ID")

	user, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		h.standardLogger.AuthOperation(requestID, "", req.Email, "register", false, err)
		h.auditLogger.LogUserCreation("", requestID, "", ipAddress, userAgent, traceID, spanID, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration failed"})
		return
	}

	h.standardLogger.AuthOperation(requestID, user.ID.String(), req.Email, "register", true, nil)
	h.auditLogger.LogUserCreation("", requestID, user.ID.String(), ipAddress, userAgent, traceID, spanID, true, "")
	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    user,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// Extract trace information
	span := trace.SpanFromContext(c.Request.Context())
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	// Get authenticated user ID
	actorUserID := middleware.GetAuthenticatedUserID(c)

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		h.auditLogger.LogTokenOperation(actorUserID, c.GetHeader("X-Request-ID"), "", c.ClientIP(), c.GetHeader("User-Agent"), "logout", traceID, spanID, false, "Authorization header required")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		h.auditLogger.LogTokenOperation(actorUserID, c.GetHeader("X-Request-ID"), "", c.ClientIP(), c.GetHeader("User-Agent"), "logout", traceID, spanID, false, "Invalid authorization header format")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	requestID := c.GetHeader("X-Request-ID")

	tokenString := tokenParts[1]
	if err := h.authService.Logout(c.Request.Context(), tokenString); err != nil {
		h.logger.WithError(err).Warn("Logout failed")
		h.auditLogger.LogTokenOperation(actorUserID, requestID, "", ipAddress, userAgent, "logout", traceID, spanID, false, err.Error())
		// Don't return error for logout failures to avoid leaking information
	} else {
		h.auditLogger.LogTokenOperation(actorUserID, requestID, "", ipAddress, userAgent, "logout", traceID, spanID, true, "")
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Extract trace information
	span := trace.SpanFromContext(c.Request.Context())
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	// Get authenticated user ID
	actorUserID := middleware.GetAuthenticatedUserID(c)

	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid refresh token request")
		h.auditLogger.LogTokenOperation(actorUserID, c.GetHeader("X-Request-ID"), "", c.ClientIP(), c.GetHeader("User-Agent"), "refresh", traceID, spanID, false, "Invalid request format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	requestID := c.GetHeader("X-Request-ID")

	response, err := h.authService.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Warn("Token refresh failed")
		h.auditLogger.LogTokenOperation(actorUserID, requestID, "", ipAddress, userAgent, "refresh", traceID, spanID, false, err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	h.auditLogger.LogTokenOperation(actorUserID, requestID, "", ipAddress, userAgent, "refresh", traceID, spanID, true, "")
	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDStr, ok := userIDValue.(string)
	if !ok {
		h.logger.Error("Invalid user ID type in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.logger.WithError(err).Error("Failed to parse user ID from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Get user email from context
	userEmailValue, exists := c.Get("user_email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User email not found"})
		return
	}

	userEmail, ok := userEmailValue.(string)
	if !ok {
		h.logger.Error("Invalid user email in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	user, err := h.authService.GetCurrentUser(c.Request.Context(), userID, userEmail)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get current user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user information"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *AuthHandler) ValidateToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
		return
	}

	tokenString := tokenParts[1]
	_, err := h.authService.ValidateToken(c.Request.Context(), tokenString)
	if err != nil {
		h.logger.WithError(err).Warn("Token validation failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or revoked token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": true})
}

func (h *AuthHandler) GetPublicKey(c *gin.Context) {
	publicKeyPEM, err := h.authService.GetPublicKeyPEM()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get public key")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get public key"})
		return
	}

	c.Header("Content-Type", "application/x-pem-file")
	c.String(http.StatusOK, string(publicKeyPEM))
}

func (h *AuthHandler) RotateKeys(c *gin.Context) {
	// Extract trace information
	span := trace.SpanFromContext(c.Request.Context())
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	// Get authenticated user ID
	actorUserID := middleware.GetAuthenticatedUserID(c)

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	requestID := c.GetHeader("X-Request-ID")

	// Check if user has admin role
	userRoles := middleware.GetAuthenticatedUserRoles(c)
	isAdmin := false
	for _, role := range userRoles {
		if role == "admin" {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		h.auditLogger.LogTokenOperation(actorUserID, requestID, "", ipAddress, userAgent, "admin_rotate_keys", traceID, spanID, false, "Insufficient permissions")
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required"})
		return
	}

	if err := h.authService.RotateKeys(c.Request.Context()); err != nil {
		h.logger.WithError(err).Error("Failed to rotate JWT keys")
		h.auditLogger.LogTokenOperation(actorUserID, requestID, "", ipAddress, userAgent, "admin_rotate_keys", traceID, spanID, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to rotate keys"})
		return
	}

	h.auditLogger.LogTokenOperation(actorUserID, requestID, "", ipAddress, userAgent, "admin_rotate_keys", traceID, spanID, true, "")
	c.JSON(http.StatusOK, gin.H{"message": "JWT keys rotated successfully"})
}

// Middleware function to validate JWT tokens
func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]
		claims, err := h.authService.ValidateToken(c.Request.Context(), tokenString)
		if err != nil {
			h.logger.WithError(err).Warn("Token validation failed")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_roles", claims.Roles)

		c.Next()
	}
}

// Role Management Handlers

// CreateRole creates a new role
func (h *AuthHandler) CreateRole(c *gin.Context) {
	// Extract trace information
	span := trace.SpanFromContext(c.Request.Context())
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	// Get authenticated user ID
	actorUserID := middleware.GetAuthenticatedUserID(c)

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), "", c.ClientIP(), c.GetHeader("User-Agent"), "create_role", traceID, spanID, false, "Invalid request data")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	role, err := h.authService.CreateRole(c.Request.Context(), req.Name, req.Description)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create role")

		// Check if this is a unique constraint violation (duplicate role name)
		if strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "23505") {
			h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), "", c.ClientIP(), c.GetHeader("User-Agent"), "create_role", traceID, spanID, false, "Role with this name already exists")
			c.JSON(http.StatusConflict, gin.H{"error": "Role with this name already exists"})
			return
		}

		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), "", c.ClientIP(), c.GetHeader("User-Agent"), "create_role", traceID, spanID, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role"})
		return
	}

	h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), role.ID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "create_role", traceID, spanID, true, "")
	c.JSON(http.StatusCreated, role)
}

// ListRoles returns all roles
func (h *AuthHandler) ListRoles(c *gin.Context) {
	roles, err := h.authService.ListRoles(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to list roles")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

// GetRole returns a specific role
func (h *AuthHandler) GetRole(c *gin.Context) {
	roleID, err := uuid.Parse(c.Param("role_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	role, err := h.authService.GetRole(c.Request.Context(), roleID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get role")
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	c.JSON(http.StatusOK, role)
}

// UpdateRole updates a role
func (h *AuthHandler) UpdateRole(c *gin.Context) {
	// Extract trace information
	span := trace.SpanFromContext(c.Request.Context())
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	// Get authenticated user ID
	actorUserID := middleware.GetAuthenticatedUserID(c)

	roleID, err := uuid.Parse(c.Param("role_id"))
	if err != nil {
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), c.Param("role_id"), c.ClientIP(), c.GetHeader("User-Agent"), "update_role", traceID, spanID, false, "Invalid role ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), roleID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "update_role", traceID, spanID, false, "Invalid request data")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	role, err := h.authService.UpdateRole(c.Request.Context(), roleID, req.Name, req.Description)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update role")
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), roleID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "update_role", traceID, spanID, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}

	h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), roleID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "update_role", traceID, spanID, true, "")
	c.JSON(http.StatusOK, role)
}

// DeleteRole deletes a role
func (h *AuthHandler) DeleteRole(c *gin.Context) {
	// Extract trace information
	span := trace.SpanFromContext(c.Request.Context())
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	// Get authenticated user ID
	actorUserID := middleware.GetAuthenticatedUserID(c)

	roleID, err := uuid.Parse(c.Param("role_id"))
	if err != nil {
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), c.Param("role_id"), c.ClientIP(), c.GetHeader("User-Agent"), "delete_role", traceID, spanID, false, "Invalid role ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	err = h.authService.DeleteRole(c.Request.Context(), roleID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to delete role")
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), roleID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "delete_role", traceID, spanID, false, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), roleID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "delete_role", traceID, spanID, true, "")
	c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
}

// Permission Management Handlers

// CreatePermission creates a new permission
func (h *AuthHandler) CreatePermission(c *gin.Context) {
	// Extract trace information
	span := trace.SpanFromContext(c.Request.Context())
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	// Get authenticated user ID
	actorUserID := middleware.GetAuthenticatedUserID(c)

	var req struct {
		Name     string `json:"name" binding:"required"`
		Resource string `json:"resource" binding:"required"`
		Action   string `json:"action" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), "", c.ClientIP(), c.GetHeader("User-Agent"), "create_permission", traceID, spanID, false, "Invalid request data")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	permission, err := h.authService.CreatePermission(c.Request.Context(), req.Name, req.Resource, req.Action)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create permission")

		// Check if this is a unique constraint violation (duplicate permission name)
		if strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "23505") {
			h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), "", c.ClientIP(), c.GetHeader("User-Agent"), "create_permission", traceID, spanID, false, "Permission with this name already exists")
			c.JSON(http.StatusConflict, gin.H{"error": "Permission with this name already exists"})
			return
		}

		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), "", c.ClientIP(), c.GetHeader("User-Agent"), "create_permission", traceID, spanID, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create permission"})
		return
	}

	h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), permission.ID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "create_permission", traceID, spanID, true, "")
	c.JSON(http.StatusCreated, permission)
}

// ListPermissions returns all permissions
func (h *AuthHandler) ListPermissions(c *gin.Context) {
	permissions, err := h.authService.ListPermissions(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to list permissions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"permissions": permissions})
}

// GetPermission returns a specific permission
func (h *AuthHandler) GetPermission(c *gin.Context) {
	permissionID, err := uuid.Parse(c.Param("permission_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission ID"})
		return
	}

	permission, err := h.authService.GetPermission(c.Request.Context(), permissionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get permission")
		c.JSON(http.StatusNotFound, gin.H{"error": "Permission not found"})
		return
	}

	c.JSON(http.StatusOK, permission)
}

// UpdatePermission updates a permission
func (h *AuthHandler) UpdatePermission(c *gin.Context) {
	// Extract trace information
	span := trace.SpanFromContext(c.Request.Context())
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	// Get authenticated user ID
	actorUserID := middleware.GetAuthenticatedUserID(c)

	permissionID, err := uuid.Parse(c.Param("permission_id"))
	if err != nil {
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), c.Param("permission_id"), c.ClientIP(), c.GetHeader("User-Agent"), "update_permission", traceID, spanID, false, "Invalid permission ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission ID"})
		return
	}

	var req struct {
		Name     string `json:"name" binding:"required"`
		Resource string `json:"resource" binding:"required"`
		Action   string `json:"action" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), permissionID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "update_permission", traceID, spanID, false, "Invalid request data")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	permission, err := h.authService.UpdatePermission(c.Request.Context(), permissionID, req.Name, req.Resource, req.Action)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update permission")
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), permissionID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "update_permission", traceID, spanID, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update permission"})
		return
	}

	h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), permissionID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "update_permission", traceID, spanID, true, "")
	c.JSON(http.StatusOK, permission)
}

// DeletePermission deletes a permission
func (h *AuthHandler) DeletePermission(c *gin.Context) {
	// Extract trace information
	span := trace.SpanFromContext(c.Request.Context())
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	// Get authenticated user ID
	actorUserID := middleware.GetAuthenticatedUserID(c)

	permissionID, err := uuid.Parse(c.Param("permission_id"))
	if err != nil {
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), c.Param("permission_id"), c.ClientIP(), c.GetHeader("User-Agent"), "delete_permission", traceID, spanID, false, "Invalid permission ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission ID"})
		return
	}

	err = h.authService.DeletePermission(c.Request.Context(), permissionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to delete permission")
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), permissionID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "delete_permission", traceID, spanID, false, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), permissionID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "delete_permission", traceID, spanID, true, "")
	c.JSON(http.StatusOK, gin.H{"message": "Permission deleted successfully"})
}

// Role-Permission Management Handlers

// AssignPermissionToRole assigns a permission to a role
func (h *AuthHandler) AssignPermissionToRole(c *gin.Context) {
	// Extract trace information
	span := trace.SpanFromContext(c.Request.Context())
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	// Get authenticated user ID
	actorUserID := middleware.GetAuthenticatedUserID(c)

	roleID, err := uuid.Parse(c.Param("role_id"))
	if err != nil {
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), c.Param("role_id"), c.ClientIP(), c.GetHeader("User-Agent"), "assign_permission_to_role", traceID, spanID, false, "Invalid role ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	var req struct {
		PermissionID string `json:"permission_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), roleID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "assign_permission_to_role", traceID, spanID, false, "Invalid request data")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	permissionID, err := uuid.Parse(req.PermissionID)
	if err != nil {
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), roleID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "assign_permission_to_role", traceID, spanID, false, "Invalid permission ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission ID"})
		return
	}

	err = h.authService.AssignPermissionToRole(c.Request.Context(), roleID, permissionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to assign permission to role")
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), roleID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "assign_permission_to_role", traceID, spanID, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign permission to role"})
		return
	}

	h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), roleID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "assign_permission_to_role", traceID, spanID, true, fmt.Sprintf("permission_id: %s", permissionID.String()))
	c.JSON(http.StatusOK, gin.H{"message": "Permission assigned to role successfully"})
}

// RemovePermissionFromRole removes a permission from a role
func (h *AuthHandler) RemovePermissionFromRole(c *gin.Context) {
	// Extract trace information
	span := trace.SpanFromContext(c.Request.Context())
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	// Get authenticated user ID
	actorUserID := middleware.GetAuthenticatedUserID(c)

	roleID, err := uuid.Parse(c.Param("role_id"))
	if err != nil {
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), c.Param("role_id"), c.ClientIP(), c.GetHeader("User-Agent"), "remove_permission_from_role", traceID, spanID, false, "Invalid role ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	permissionID, err := uuid.Parse(c.Param("perm_id"))
	if err != nil {
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), roleID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "remove_permission_from_role", traceID, spanID, false, "Invalid permission ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission ID"})
		return
	}

	err = h.authService.RemovePermissionFromRole(c.Request.Context(), roleID, permissionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to remove permission from role")
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), roleID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "remove_permission_from_role", traceID, spanID, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove permission from role"})
		return
	}

	h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), roleID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "remove_permission_from_role", traceID, spanID, true, fmt.Sprintf("permission_id: %s", permissionID.String()))
	c.JSON(http.StatusOK, gin.H{"message": "Permission removed from role successfully"})
}

// GetRolePermissions returns permissions for a role
func (h *AuthHandler) GetRolePermissions(c *gin.Context) {
	roleID, err := uuid.Parse(c.Param("role_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	permissions, err := h.authService.GetRolePermissions(c.Request.Context(), roleID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get role permissions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get role permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"permissions": permissions})
}

// User Role Management Handlers

// AssignRoleToUser assigns a role to a user
func (h *AuthHandler) AssignRoleToUser(c *gin.Context) {
	// Extract trace information
	span := trace.SpanFromContext(c.Request.Context())
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	// Get authenticated user ID
	actorUserID := middleware.GetAuthenticatedUserID(c)

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), c.Param("user_id"), c.ClientIP(), c.GetHeader("User-Agent"), "assign_role_to_user", traceID, spanID, false, "Invalid user ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		RoleID string `json:"role_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), userID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "assign_role_to_user", traceID, spanID, false, "Invalid request data")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), userID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "assign_role_to_user", traceID, spanID, false, "Invalid role ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	err = h.authService.AssignRoleToUser(c.Request.Context(), userID, roleID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to assign role to user")
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), userID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "assign_role_to_user", traceID, spanID, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign role to user"})
		return
	}

	h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), userID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "assign_role_to_user", traceID, spanID, true, fmt.Sprintf("role_id: %s", roleID.String()))
	c.JSON(http.StatusOK, gin.H{"message": "Role assigned to user successfully"})
}

// RemoveRoleFromUser removes a role from a user
func (h *AuthHandler) RemoveRoleFromUser(c *gin.Context) {
	// Extract trace information
	span := trace.SpanFromContext(c.Request.Context())
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	// Get authenticated user ID
	actorUserID := middleware.GetAuthenticatedUserID(c)

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), c.Param("user_id"), c.ClientIP(), c.GetHeader("User-Agent"), "remove_role_from_user", traceID, spanID, false, "Invalid user ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	roleID, err := uuid.Parse(c.Param("role_id"))
	if err != nil {
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), userID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "remove_role_from_user", traceID, spanID, false, "Invalid role ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	err = h.authService.RemoveRoleFromUser(c.Request.Context(), userID, roleID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to remove role from user")
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), userID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "remove_role_from_user", traceID, spanID, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove role from user"})
		return
	}

	h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), userID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "remove_role_from_user", traceID, spanID, true, fmt.Sprintf("role_id: %s", roleID.String()))
	c.JSON(http.StatusOK, gin.H{"message": "Role removed from user successfully"})
}

// GetUserRoles returns roles for a user
func (h *AuthHandler) GetUserRoles(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	roles, err := h.authService.GetUserRoles(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user roles")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

// UpdateUserRoles updates all roles for a user (bulk operation)
func (h *AuthHandler) UpdateUserRoles(c *gin.Context) {
	// Extract trace information
	span := trace.SpanFromContext(c.Request.Context())
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	// Get authenticated user ID
	actorUserID := middleware.GetAuthenticatedUserID(c)

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), c.Param("user_id"), c.ClientIP(), c.GetHeader("User-Agent"), "update_user_roles", traceID, spanID, false, "Invalid user ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		RoleIDs []string `json:"role_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), userID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "update_user_roles", traceID, spanID, false, "Invalid request data")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	roleIDs := make([]uuid.UUID, len(req.RoleIDs))
	for i, roleIDStr := range req.RoleIDs {
		roleID, err := uuid.Parse(roleIDStr)
		if err != nil {
			h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), userID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "update_user_roles", traceID, spanID, false, "Invalid role ID in list")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID in list"})
			return
		}
		roleIDs[i] = roleID
	}

	err = h.authService.UpdateUserRoles(c.Request.Context(), userID, roleIDs)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update user roles")
		h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), userID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "update_user_roles", traceID, spanID, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user roles"})
		return
	}

	h.auditLogger.LogAdminAction(actorUserID, c.GetHeader("X-Request-ID"), userID.String(), c.ClientIP(), c.GetHeader("User-Agent"), "update_user_roles", traceID, spanID, true, fmt.Sprintf("role_count: %d", len(roleIDs)))
	c.JSON(http.StatusOK, gin.H{"message": "User roles updated successfully"})
}
