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

// JWTMiddleware creates JWT authentication middleware
func JWTMiddleware(jwtSecret interface{}, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// If no JWT secret provided, skip authentication
		if jwtSecret == nil {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth header - continue without authentication
			c.Next()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
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
			logger.WithError(err).Warn("Failed to parse JWT token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Set user information in context for use by handlers and logging
		c.Set("user_id", claims.UserID.String())
		c.Set("user_email", claims.Email)
		c.Set("user_roles", claims.Roles)
		c.Set("token_type", claims.TokenType)

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
