package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func newTestRequestLogger(t *testing.T) *RequestLogger {
	t.Helper()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	return NewRequestLogger(logger)
}

// --- RequestResponseLogger skip-paths tests ---

func TestRequestResponseLoggerSkipPaths(t *testing.T) {
	t.Run("does not panic on skipped paths", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		rl := newTestRequestLogger(t)
		router := gin.New()
		router.Use(rl.RequestResponseLogger())
		router.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		assert.NotPanics(t, func() {
			router.ServeHTTP(w, req)
		})
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify multiple skipped paths don't cause issues
		skippedPaths := []string{"/ready", "/live", "/ping", "/status"}
		for _, path := range skippedPaths {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", path, nil)
			assert.NotPanics(t, func() {
				router.ServeHTTP(w, req)
			}, "should not panic on skipped path: %s", path)
		}
	})

	t.Run("does not panic on skipped paths", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		logger := logrus.New()
		var buf bytes.Buffer
		logger.SetOutput(&buf)

		rl := NewRequestLogger(logger)
		router := gin.New()
		router.Use(rl.RequestResponseLogger())
		router.GET("/ready", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/ready", nil)
		assert.NotPanics(t, func() {
			router.ServeHTTP(w, req)
		})
		assert.Equal(t, http.StatusOK, w.Code)

		logOutput := buf.String()
		assert.NotContains(t, logOutput, "/ready")
	})
}

func TestRequestLoggerGetMetricsCollector(t *testing.T) {
	t.Run("returns the metrics collector", func(t *testing.T) {
		rl := newTestRequestLogger(t)
		mc := rl.GetMetricsCollector()
		assert.NotNil(t, mc)
	})
}

// --- DetailedRequestLogger basic smoke test ---

func TestDetailedRequestLoggerSmoke(t *testing.T) {
	t.Run("does not panic on a normal request", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		logger := logrus.New()
		var buf bytes.Buffer
		logger.SetOutput(&buf)

		rl := NewRequestLogger(logger)
		router := gin.New()
		router.Use(rl.DetailedRequestLogger())
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", strings.NewReader(`{"key":"value"}`))
		req.Header.Set("Content-Type", "application/json")
		assert.NotPanics(t, func() {
			router.ServeHTTP(w, req)
		})

		assert.Equal(t, http.StatusOK, w.Code)

		logOutput := buf.String()
		assert.Contains(t, logOutput, "/test")
		assert.Contains(t, logOutput, "GET")
	})

	t.Run("captures status >= 400 as warn/error", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		logger := logrus.New()
		var buf bytes.Buffer
		logger.SetOutput(&buf)

		rl := NewRequestLogger(logger)
		router := gin.New()
		router.Use(rl.DetailedRequestLogger())
		router.GET("/error", func(c *gin.Context) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/error", nil)
		assert.NotPanics(t, func() {
			router.ServeHTTP(w, req)
		})

		logOutput := buf.String()
		assert.Contains(t, logOutput, "400")
	})
}
