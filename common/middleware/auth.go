package middleware

import (
	"crypto/rsa"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// JWTClaims represents the JWT token claims
type JWTClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Roles     []string  `json:"roles"`
	TokenType string    `json:"token_type"`
	jwt.RegisteredClaims
}

// TokenRevocationChecker interface for checking if a token has been revoked
type TokenRevocationChecker interface {
	IsTokenRevoked(tokenString string) bool
}

// JWTMiddleware creates JWT authentication middleware
func JWTMiddleware(jwtSecret interface{}, logger *logrus.Logger, revocationChecker TokenRevocationChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = "unknown"
		}

		// If no JWT secret provided, skip authentication
		if jwtSecret == nil {
			logger.WithFields(logrus.Fields{
				"request_id": requestID,
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
			}).Debug("JWT middleware: Skipping authentication (no JWT secret)")
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth header - continue without authentication
			logger.WithFields(logrus.Fields{
				"request_id": requestID,
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
			}).Debug("JWT middleware: No Authorization header, continuing without authentication")
			c.Next()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			logger.WithFields(logrus.Fields{
				"request_id": requestID,
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
			}).Warn("JWT middleware: Invalid authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate JWT token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Support both HMAC and RSA signing methods
			switch token.Method {
			case jwt.SigningMethodHS256, jwt.SigningMethodHS384, jwt.SigningMethodHS512:
				// HMAC method - jwtSecret should be []byte
				if secret, ok := jwtSecret.([]byte); ok {
					return secret, nil
				}
			case jwt.SigningMethodRS256, jwt.SigningMethodRS384, jwt.SigningMethodRS512:
				// RSA method - jwtSecret should be *rsa.PublicKey
				if publicKey, ok := jwtSecret.(*rsa.PublicKey); ok {
					return publicKey, nil
				}
			}
			return nil, jwt.ErrSignatureInvalid
		})

		if err != nil {
			logger.WithFields(logrus.Fields{
				"request_id": requestID,
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
				"error":      err.Error(),
			}).Warn("JWT middleware: Failed to parse JWT token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if !token.Valid {
			logger.WithFields(logrus.Fields{
				"request_id": requestID,
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
			}).Warn("JWT middleware: Token validation failed")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Check if token has been revoked (if revocation checker is provided)
		if revocationChecker != nil && revocationChecker.IsTokenRevoked(tokenString) {
			logger.WithFields(logrus.Fields{
				"request_id": requestID,
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
			}).Warn("JWT middleware: Token has been revoked")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			logger.WithFields(logrus.Fields{
				"request_id": requestID,
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
			}).Warn("JWT middleware: Invalid token claims type")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Set user information in context for use by handlers and logging
		userID := claims.UserID.String()
		c.Set("user_id", userID)
		c.Set("user_email", claims.Email)
		c.Set("user_roles", claims.Roles)
		c.Set("token_type", claims.TokenType)

		logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"path":       c.Request.URL.Path,
			"method":     c.Request.Method,
			"user_id":    userID,
			"user_email": claims.Email,
			"token_type": claims.TokenType,
		}).Debug("JWT middleware: Successfully authenticated user, set context")

		c.Next()
	}
}

// GetAuthenticatedUserID extracts the authenticated user ID from the Gin context
func GetAuthenticatedUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			return uid
		}
	}
	return ""
}

// GetAuthenticatedUserEmail extracts the authenticated user email from the Gin context
func GetAuthenticatedUserEmail(c *gin.Context) string {
	if email, exists := c.Get("user_email"); exists {
		if e, ok := email.(string); ok {
			return e
		}
	}
	return ""
}

// GetAuthenticatedUserRoles extracts the authenticated user roles from the Gin context
func GetAuthenticatedUserRoles(c *gin.Context) []string {
	if roles, exists := c.Get("user_roles"); exists {
		if r, ok := roles.([]string); ok {
			return r
		}
	}
	return []string{}
}

// RequireAuth middleware requires authentication for the endpoint
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetAuthenticatedUserID(c)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireRole middleware requires specific role(s) for the endpoint
func RequireRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles := GetAuthenticatedUserRoles(c)
		if len(userRoles) == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		// Check if user has any of the required roles
		hasRequiredRole := false
		for _, requiredRole := range requiredRoles {
			for _, userRole := range userRoles {
				if userRole == requiredRole {
					hasRequiredRole = true
					break
				}
			}
			if hasRequiredRole {
				break
			}
		}

		if !hasRequiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}
