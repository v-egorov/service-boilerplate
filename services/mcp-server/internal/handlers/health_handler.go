package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	mcpclient "github.com/v-egorov/service-boilerplate/services/mcp-server/internal/client"
)

// HealthResponse represents the health check response body.
type HealthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Version   string `json:"version,omitempty"`
	Timestamp string `json:"timestamp"`
}

// StatusResponse represents the comprehensive status response.
type StatusResponse struct {
	Status         string           `json:"status"`
	Timestamp      string           `json:"timestamp"`
	Service        ServiceInfo      `json:"service"`
	ObjectsService ObjectsSvcHealth `json:"objects_service"`
	Checks         CheckSummary     `json:"checks"`
}

// ServiceInfo represents service information.
type ServiceInfo struct {
	Name   string `json:"name"`
	Version string `json:"version"`
	Uptime string `json:"uptime"`
}

// ObjectsSvcHealth represents objects-service connectivity status.
type ObjectsSvcHealth struct {
	Status       string `json:"status"`
	ResponseTime string `json:"response_time,omitempty"`
	Error        string `json:"error,omitempty"`
}

// CheckSummary represents the summary of health checks.
type CheckSummary struct {
	Total  int `json:"total"`
	Passed int `json:"passed"`
	Failed int `json:"failed"`
}

// HealthHandler provides health check endpoints for the MCP server.
type HealthHandler struct {
	objClient *mcpclient.ObjectsClient
	logger    *logrus.Logger
	name      string
	version   string
	startTime time.Time
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(objClient *mcpclient.ObjectsClient, logger *logrus.Logger, name, version string) *HealthHandler {
	return &HealthHandler{
		objClient: objClient,
		logger:    logger,
		name:      name,
		version:   version,
		startTime: time.Now(),
	}
}

// LivenessHandler provides basic liveness check.
func (h *HealthHandler) LivenessHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get("X-Request-ID")

	resp := HealthResponse{
		Status:    "ok",
		Service:   h.name,
		Version:   h.version,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"data":%s,"meta":{"request_id":"%s"}}`, toJSON(resp), requestID)
}

// ReadinessHandler checks if the service is ready to accept traffic.
func (h *HealthHandler) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get("X-Request-ID")

	if h.objClient != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		statusCode, err := h.checkObjectsServiceHealth(ctx)
		if statusCode != http.StatusOK || err != nil {
			h.logger.WithError(err).Warn("Readiness check failed: objects-service unreachable")
			resp := HealthResponse{
				Status:    "error",
				Service:   h.name,
				Version:   h.version,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"data":%s,"meta":{"request_id":"%s"}}`, toJSON(resp), requestID)
			return
		}
	}

	resp := HealthResponse{
		Status:    "ok",
		Service:   h.name,
		Version:   h.version,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"data":%s,"meta":{"request_id":"%s"}}`, toJSON(resp), requestID)
}

// PingHandler provides simple ping/pong response.
func (h *HealthHandler) PingHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get("X-Request-ID")
	resp := map[string]interface{}{
		"status":    "pong",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   h.name,
		"version":   h.version,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"data":%s,"meta":{"request_id":"%s"}}`, toJSON(resp), requestID)
}

// StatusHandler provides comprehensive service status.
func (h *HealthHandler) StatusHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get("X-Request-ID")

	var objSvcHealth ObjectsSvcHealth
	totalChecks := 0
	failedChecks := 0

	if h.objClient != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		statusCode, err := h.checkObjectsServiceHealth(ctx)
		totalChecks++
		objSvcHealth.Status = "unhealthy"
		if statusCode == http.StatusOK && err == nil {
			objSvcHealth.Status = "healthy"
		} else {
			objSvcHealth.Error = err.Error()
			failedChecks++
		}
	}

	overallStatus := "healthy"
	if failedChecks > 0 {
		overallStatus = "unhealthy"
	}

	resp := StatusResponse{
		Status:    overallStatus,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Service: ServiceInfo{
			Name:   h.name,
			Version: h.version,
			Uptime:  calculateUptime(h.startTime),
		},
		ObjectsService: objSvcHealth,
		Checks: CheckSummary{
			Total:  totalChecks,
			Passed: totalChecks - failedChecks,
			Failed: failedChecks,
		},
	}

	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, `{"data":%s,"meta":{"request_id":"%s"}}`, toJSON(resp), requestID)
}

// checkObjectsServiceHealth pings the objects-service health endpoint.
func (h *HealthHandler) checkObjectsServiceHealth(ctx context.Context) (int, error) {
	healthURL := h.objClient.HealthCheckURL() + "/health"

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("objects-service unreachable: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}

// toJSON marshals v to JSON for inline response formatting.
func toJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

// calculateUptime returns formatted uptime string.
func calculateUptime(start time.Time) string {
	uptime := time.Since(start)
	hours := int(uptime.Hours())
	minutes := int(uptime.Minutes()) % 60
	seconds := int(uptime.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh%dm%ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
