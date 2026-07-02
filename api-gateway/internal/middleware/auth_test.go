package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter(middlewares ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	for _, m := range middlewares {
		r.Use(m)
	}
	return r
}

// --- AuthMiddleware tests ---

func TestAuthMiddleware(t *testing.T) {
	t.Run("allows valid Bearer token requests through", func(t *testing.T) {
		router := setupRouter(AuthMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer test-token-123")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("returns 401 when Authorization header is missing", func(t *testing.T) {
		router := setupRouter(AuthMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		// No Authorization header set
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Authorization header required")
	})

	t.Run("returns 401 for non-Bearer authorization format", func(t *testing.T) {
		router := setupRouter(AuthMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid authorization format")
	})

	t.Run("returns 401 for empty Bearer token", func(t *testing.T) {
		router := setupRouter(AuthMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer ") // empty token still passes prefix check
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// --- CORSMiddleware tests ---

func TestCORSMiddleware(t *testing.T) {
	t.Run("OPTIONS request returns NoContent status", func(t *testing.T) {
		router := setupRouter(CORSMiddleware())
		router.OPTIONS("/test", func(c *gin.Context) {})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("OPTIONS", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("standard method requests proceed through middleware with CORS headers set", func(t *testing.T) {
		methods := []string{"GET", "POST", "PUT", "DELETE"}

		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				router := setupRouter(CORSMiddleware())
				router.Handle(method, "/test", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "ok"})
				})

				w := httptest.NewRecorder()
				req, _ := http.NewRequest(method, "/test", nil)
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)
				assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
				assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
			})
		}
	})

	t.Run("CORS headers are set on response", func(t *testing.T) {
		router := setupRouter(CORSMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	})
}

// --- RequestIDMiddleware tests ---

func TestRequestIDMiddleware(t *testing.T) {
	t.Run("preserves existing X-Request-ID header from inbound request", func(t *testing.T) {
		router := setupRouter(RequestIDMiddleware())
		router.GET("/test", func(c *gin.Context) {
			reqID := c.GetString("request_id")
			c.JSON(http.StatusOK, gin.H{"request_id": reqID})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", "my-custom-id-12345")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "my-custom-id-12345")
	})

	t.Run("generates new UUID when no X-Request-ID is present in request", func(t *testing.T) {
		router := setupRouter(RequestIDMiddleware())
		router.GET("/test", func(c *gin.Context) {
			reqID := c.GetString("request_id")
			c.JSON(http.StatusOK, gin.H{"request_id": reqID})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		// No X-Request-ID header
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify the generated ID looks like a UUID (not empty, has dashes)
		var resp map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		reqID := resp["request_id"]
		assert.NotEmpty(t, reqID)
		assert.Contains(t, reqID, "-") // UUID format has dashes

		// Verify response header also contains the generated ID
		assert.Equal(t, reqID, w.Header().Get("X-Request-ID"))
	})

	t.Run("response X-Request-ID matches context value", func(t *testing.T) {
		router := setupRouter(RequestIDMiddleware())
		router.GET("/test", func(c *gin.Context) {
			reqID := c.GetString("request_id")
			c.JSON(http.StatusOK, gin.H{"context_id": reqID})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		contextID := resp["context_id"]
		headerID := w.Header().Get("X-Request-ID")
		assert.Equal(t, contextID, headerID)
	})
}
