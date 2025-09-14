package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/api-gateway/internal/services"
	"github.com/v-egorov/service-boilerplate/common/config"
)

type GatewayHandler struct {
	registry  *services.ServiceRegistry
	logger    *logrus.Logger
	config    *config.Config
	startTime time.Time
}

func NewGatewayHandler(registry *services.ServiceRegistry, logger *logrus.Logger, cfg *config.Config) *GatewayHandler {
	return &GatewayHandler{
		registry:  registry,
		logger:    logger,
		config:    cfg,
		startTime: time.Now(),
	}
}

func (h *GatewayHandler) ProxyRequest(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get service URL
		serviceURL, err := h.registry.GetServiceURL(serviceName)
		if err != nil {
			h.logger.WithError(err).Error("Service not found")
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Service unavailable"})
			return
		}

		// Parse service URL
		targetURL, err := url.Parse(serviceURL)
		if err != nil {
			h.logger.WithError(err).Error("Invalid service URL")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Create reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		// Modify the request
		c.Request.Host = targetURL.Host
		c.Request.URL.Scheme = targetURL.Scheme
		c.Request.URL.Host = targetURL.Host

		// Add request ID to headers
		if requestID, exists := c.Get("request_id"); exists {
			c.Request.Header.Set("X-Request-ID", requestID.(string))
		}

		// Log the proxy request
		h.logger.WithFields(logrus.Fields{
			"service":    serviceName,
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"request_id": c.GetString("request_id"),
		}).Info("Proxying request")

		// Custom director to handle request body properly
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)

			// Read request body
			if req.Body != nil {
				bodyBytes, err := io.ReadAll(req.Body)
				if err != nil {
					h.logger.WithError(err).Error("Failed to read request body")
					return
				}
				req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// Custom error handler
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			h.logger.WithError(err).Error("Proxy error")
			c.JSON(http.StatusBadGateway, gin.H{"error": "Service unavailable"})
		}

		// Serve the request
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// LivenessHandler provides basic liveness check
func (h *GatewayHandler) LivenessHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"gateway":   h.config.App.Name,
		"version":   h.config.App.Version,
	})
}

// ReadinessHandler checks if the gateway is ready to accept traffic
func (h *GatewayHandler) ReadinessHandler(c *gin.Context) {
	// Check if services are available
	services := h.registry.ListServices()
	if len(services) == 0 {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "error",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"gateway":   h.config.App.Name,
			"message":   "No services registered",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"gateway":   h.config.App.Name,
		"version":   h.config.App.Version,
	})
}

// PingHandler provides a simple liveness check
func (h *GatewayHandler) PingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "pong",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"gateway":   h.config.App.Name,
		"version":   h.config.App.Version,
	})
}

// StatusHandler provides comprehensive gateway and service status
func (h *GatewayHandler) StatusHandler(c *gin.Context) {
	services := h.registry.ListServices()

	// Check service health concurrently
	serviceHealth := h.checkAllServicesHealth(services)

	// Calculate overall status
	overallStatus := h.calculateOverallStatus(serviceHealth)

	// Get system information
	systemInfo := h.getSystemInfo()

	response := gin.H{
		"status":    overallStatus,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"gateway":   systemInfo,
		"services":  serviceHealth,
		"checks": gin.H{
			"total":  len(services) + 1,                      // +1 for gateway itself
			"passed": h.countPassedChecks(serviceHealth) + 1, // +1 for gateway
			"failed": h.countFailedChecks(serviceHealth),
		},
	}

	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

// ServiceHealth represents the health status of a service
type ServiceHealth struct {
	Status       string `json:"status"`
	URL          string `json:"url"`
	ResponseTime string `json:"response_time,omitempty"`
	LastChecked  string `json:"last_checked"`
	Error        string `json:"error,omitempty"`
}

// GatewayInfo represents gateway system information
type GatewayInfo struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Environment string            `json:"environment"`
	Uptime      string            `json:"uptime"`
	MemoryUsage map[string]uint64 `json:"memory_usage,omitempty"`
}

// checkAllServicesHealth checks health of all registered services concurrently
func (h *GatewayHandler) checkAllServicesHealth(services map[string]string) map[string]ServiceHealth {
	result := make(map[string]ServiceHealth)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for name, url := range services {
		wg.Add(1)
		go func(serviceName, serviceURL string) {
			defer wg.Done()
			health := h.checkServiceHealth(serviceName, serviceURL)

			mu.Lock()
			result[serviceName] = health
			mu.Unlock()
		}(name, url)
	}

	wg.Wait()
	return result
}

// checkServiceHealth performs health check on a single service
func (h *GatewayHandler) checkServiceHealth(serviceName, serviceURL string) ServiceHealth {
	healthURL := serviceURL + "/health"
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(h.config.Monitoring.HealthCheckTimeout)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return ServiceHealth{
			Status:      "unhealthy",
			URL:         serviceURL,
			LastChecked: time.Now().UTC().Format(time.RFC3339),
			Error:       err.Error(),
		}
	}

	client := &http.Client{Timeout: time.Duration(h.config.Monitoring.HealthCheckTimeout) * time.Second}
	resp, err := client.Do(req)
	responseTime := time.Since(start)

	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"service": serviceName,
			"url":     healthURL,
			"error":   err.Error(),
		}).Warn("Service health check failed")

		return ServiceHealth{
			Status:       "unhealthy",
			URL:          serviceURL,
			ResponseTime: responseTime.Round(time.Millisecond).String(),
			LastChecked:  time.Now().UTC().Format(time.RFC3339),
			Error:        err.Error(),
		}
	}
	defer resp.Body.Close()

	status := "healthy"
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		status = "unhealthy"
	}

	h.logger.WithFields(logrus.Fields{
		"service":       serviceName,
		"url":           healthURL,
		"status_code":   resp.StatusCode,
		"response_time": responseTime.Round(time.Millisecond).String(),
	}).Debug("Service health check completed")

	return ServiceHealth{
		Status:       status,
		URL:          serviceURL,
		ResponseTime: responseTime.Round(time.Millisecond).String(),
		LastChecked:  time.Now().UTC().Format(time.RFC3339),
	}
}

// calculateOverallStatus determines overall system health
func (h *GatewayHandler) calculateOverallStatus(serviceHealth map[string]ServiceHealth) string {
	if len(serviceHealth) == 0 {
		return "healthy"
	}

	unhealthyCount := 0
	for _, health := range serviceHealth {
		if health.Status == "unhealthy" {
			unhealthyCount++
		}
	}

	if unhealthyCount == 0 {
		return "healthy"
	} else if unhealthyCount < len(serviceHealth) {
		return "degraded"
	}
	return "unhealthy"
}

// getSystemInfo returns gateway system information
func (h *GatewayHandler) getSystemInfo() GatewayInfo {
	info := GatewayInfo{
		Name:        h.config.App.Name,
		Version:     h.config.App.Version,
		Environment: h.config.App.Environment,
		Uptime:      h.calculateUptime(),
	}

	if h.config.Monitoring.EnableDetailedMetrics {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		info.MemoryUsage = map[string]uint64{
			"alloc":       m.Alloc,
			"total_alloc": m.TotalAlloc,
			"sys":         m.Sys,
		}
	}

	return info
}

// calculateUptime returns formatted uptime string
func (h *GatewayHandler) calculateUptime() string {
	uptime := time.Since(h.startTime)
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

// countPassedChecks counts services with healthy status
func (h *GatewayHandler) countPassedChecks(serviceHealth map[string]ServiceHealth) int {
	count := 0
	for _, health := range serviceHealth {
		if health.Status == "healthy" {
			count++
		}
	}
	return count
}

// countFailedChecks counts services with unhealthy status
func (h *GatewayHandler) countFailedChecks(serviceHealth map[string]ServiceHealth) int {
	count := 0
	for _, health := range serviceHealth {
		if health.Status == "unhealthy" {
			count++
		}
	}
	return count
}
