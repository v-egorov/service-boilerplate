package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	mcpclient "github.com/v-egorov/service-boilerplate/services/mcp-server/internal/client"
)

func TestLivenessHandler(t *testing.T) {
	objClient := mcpclient.NewObjectsClient("http://objects-service:8085", 0)
	handler := NewHealthHandler(objClient, &logrus.Logger{}, "mcp-test", "1.0.0")

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.LivenessHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	data, ok := body["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'data' field in response")
	}
	if data["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%v'", data["status"])
	}
	if data["service"] != "mcp-test" {
		t.Errorf("Expected service 'mcp-test', got '%v'", data["service"])
	}

	meta, ok := body["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'meta' field in response")
	}
	// request_id may be empty when called without X-Request-ID middleware (e.g. docker healthcheck)
	_ = meta["request_id"]
}

func TestLivenessHandler_WithRequestID(t *testing.T) {
	objClient := mcpclient.NewObjectsClient("http://objects-service:8085", 0)
	handler := NewHealthHandler(objClient, &logrus.Logger{}, "mcp-test", "1.0.0")

	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("X-Request-ID", "test-request-id-123")
	w := httptest.NewRecorder()

	handler.LivenessHandler(w, req)

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	meta := body["meta"].(map[string]interface{})
	if meta["request_id"] != "test-request-id-123" {
		t.Errorf("Expected request_id 'test-request-id-123', got '%v'", meta["request_id"])
	}
}

func TestReadinessHandler_Healthy(t *testing.T) {
	// Create mock objects-service that returns healthy
	objServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"status":"ok"}}`))
	}))
	defer objServer.Close()

	objClient := mcpclient.NewObjectsClient(objServer.URL, 0)
	handler := NewHealthHandler(objClient, &logrus.Logger{}, "mcp-test", "1.0.0")

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	handler.ReadinessHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	data := body["data"].(map[string]interface{})
	if data["status"] != "ok" {
		t.Errorf("Expected readiness 'ok', got '%v'", data["status"])
	}
}

func TestReadinessHandler_Unhealthy(t *testing.T) {
	// Create mock objects-service that returns error (non-existent server)
	objClient := mcpclient.NewObjectsClient("http://127.0.0.1:1", 0) // invalid port, will fail
	handler := NewHealthHandler(objClient, &logrus.Logger{}, "mcp-test", "1.0.0")

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	handler.ReadinessHandler(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503 (unhealthy), got %d", w.Code)
	}

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	data := body["data"].(map[string]interface{})
	if data["status"] != "error" {
		t.Errorf("Expected readiness 'error', got '%v'", data["status"])
	}
}

func TestReadinessHandler_NoClient(t *testing.T) {
	handler := NewHealthHandler(nil, &logrus.Logger{}, "mcp-test", "1.0.0")

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	handler.ReadinessHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 (no client = ok), got %d", w.Code)
	}
}

func TestPingHandler(t *testing.T) {
	objClient := mcpclient.NewObjectsClient("http://objects-service:8085", 0)
	handler := NewHealthHandler(objClient, &logrus.Logger{}, "mcp-test", "1.0.0")

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	handler.PingHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	data := body["data"].(map[string]interface{})
	if data["status"] != "pong" {
		t.Errorf("Expected ping 'pong', got '%v'", data["status"])
	}
	if data["service"] != "mcp-test" {
		t.Errorf("Expected service 'mcp-test', got '%v'", data["service"])
	}
}

func TestStatusHandler_Healthy(t *testing.T) {
	objServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"status":"ok"}}`))
	}))
	defer objServer.Close()

	objClient := mcpclient.NewObjectsClient(objServer.URL, 0)
	handler := NewHealthHandler(objClient, &logrus.Logger{}, "mcp-test", "1.0.0")

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	handler.StatusHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	data := body["data"].(map[string]interface{})
	status := data["status"]
	if status != "healthy" {
		t.Errorf("Expected overall status 'healthy', got '%v'", status)
	}

	checks := data["checks"].(map[string]interface{})
	if int(checks["passed"].(float64)) != 1 {
		t.Errorf("Expected 1 passed check, got %v", checks["passed"])
	}
}

func TestStatusHandler_Unhealthy(t *testing.T) {
	objClient := mcpclient.NewObjectsClient("http://127.0.0.1:1", 0) // invalid port
	handler := NewHealthHandler(objClient, &logrus.Logger{}, "mcp-test", "1.0.0")

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	handler.StatusHandler(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503 (unhealthy), got %d", w.Code)
	}

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	data := body["data"].(map[string]interface{})
	status := data["status"]
	if status != "unhealthy" {
		t.Errorf("Expected overall status 'unhealthy', got '%v'", status)
	}

	objSvc := data["objects_service"].(map[string]interface{})
	if objSvc["status"] != "unhealthy" {
		t.Errorf("Expected objects_service status 'unhealthy', got '%v'", objSvc["status"])
	}
}

func TestStatusHandler_NoClient(t *testing.T) {
	handler := NewHealthHandler(nil, &logrus.Logger{}, "mcp-test", "1.0.0")

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	handler.StatusHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 (no client = ok), got %d", w.Code)
	}

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	data := body["data"].(map[string]interface{})
	status := data["status"]
	if status != "healthy" {
		t.Errorf("Expected overall status 'healthy', got '%v'", status)
	}
	checks := data["checks"].(map[string]interface{})
	if int(checks["total"].(float64)) != 0 {
		t.Errorf("Expected 0 total checks (no client), got %v", checks["total"])
	}
}

func TestHealthHandler_Uptime(t *testing.T) {
	objClient := mcpclient.NewObjectsClient("http://objects-service:8085", 0)
	startTime := time.Now().Add(-3 * time.Hour).Add(-30 * time.Minute).Add(-45 * time.Second)
	handler := &HealthHandler{
		objClient: objClient,
		logger:    &logrus.Logger{},
		name:      "mcp-test",
		version:   "1.0.0",
		startTime: startTime,
	}

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	handler.StatusHandler(w, req)

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	data := body["data"].(map[string]interface{})
	service := data["service"].(map[string]interface{})
	uptime := service["uptime"].(string)

	if !strings.Contains(uptime, "3h") {
		t.Errorf("Expected uptime to contain '3h', got '%s'", uptime)
	}
}
