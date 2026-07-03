package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/v-egorov/service-boilerplate/api-gateway/internal/services"
	"github.com/v-egorov/service-boilerplate/common/config"
)

func newTestGatewayHandler(t *testing.T, registry *services.ServiceRegistry) (*GatewayHandler, *logrus.Logger) {
	t.Helper()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetOutput(&discardWriter{})
	cfg := &config.Config{
		App:           config.AppConfig{Name: "test-gateway", Version: "0.1.0"},
		SystemAccount: config.SystemAccountConfig{McpAgentUserID: "00000000-0000-4000-8000-000000000001"},
	}
	handler := NewGatewayHandler(registry, logger, cfg)
	return handler, logger
}

type discardWriter struct{}

func (d *discardWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func TestLivenessHandler(t *testing.T) {
	t.Run("returns ok status with metadata", func(t *testing.T) {
		handler, _ := newTestGatewayHandler(t, services.NewServiceRegistry(logrus.New()))

		w := httptest.NewRecorder()
		c := createGinContext(w, "GET", "/health")
		handler.LivenessHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var body map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &body)
		assert.NoError(t, err)

		assert.Equal(t, "ok", body["status"])
		assert.Equal(t, "test-gateway", body["gateway"])
		assert.Equal(t, "0.1.0", body["version"])
		assert.NotNil(t, body["timestamp"]) // just check it exists, don't assert exact value
	})
}

func TestPingHandler(t *testing.T) {
	t.Run("returns pong status with metadata", func(t *testing.T) {
		handler, _ := newTestGatewayHandler(t, services.NewServiceRegistry(logrus.New()))

		w := httptest.NewRecorder()
		c := createGinContext(w, "GET", "/ping")
		handler.PingHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var body map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &body)
		assert.NoError(t, err)

		assert.Equal(t, "pong", body["status"])
		assert.Equal(t, "test-gateway", body["gateway"])
		assert.Equal(t, "0.1.0", body["version"])
	})
}

func TestReadinessHandler(t *testing.T) {
	t.Run("returns ok when services are registered in backend registry", func(t *testing.T) {
		reg := services.NewServiceRegistry(logrus.New())
		reg.RegisterService("auth-service", "http://auth:8083")
		handler, _ := newTestGatewayHandler(t, reg)

		w := httptest.NewRecorder()
		c := createGinContext(w, "GET", "/ready")
		handler.ReadinessHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var body map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &body)
		assert.NoError(t, err)
		assert.Equal(t, "ok", body["status"])
	})

	t.Run("returns error when no services are registered", func(t *testing.T) {
		reg := services.NewServiceRegistry(logrus.New())
		handler, _ := newTestGatewayHandler(t, reg)

		w := httptest.NewRecorder()
		c := createGinContext(w, "GET", "/ready")
		handler.ReadinessHandler(c)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		var body map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &body)
		assert.NoError(t, err)
		assert.Equal(t, "error", body["status"])
		assert.Contains(t, body["message"], "No services registered")
	})
}

func createGinContext(w *httptest.ResponseRecorder, method, path string) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(method, path, nil)
	c.Request = req
	return c
}
