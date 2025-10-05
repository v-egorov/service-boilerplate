package handlers

import (
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
	authService    *services.AuthService
	logger         *logrus.Logger
	auditLogger    *logging.AuditLogger
	standardLogger *logging.StandardLogger
}

func NewAuthHandler(authService *services.AuthService, logger *logrus.Logger) *AuthHandler {
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
