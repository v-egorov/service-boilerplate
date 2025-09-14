package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/common/config"
)

// HealthHandler provides health check endpoints for the user service
type HealthHandler struct {
	db        *pgxpool.Pool
	logger    *logrus.Logger
	config    *config.Config
	startTime time.Time
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
	Version   string    `json:"version,omitempty"`
}

// StatusResponse represents the comprehensive status response
type StatusResponse struct {
	Status    string         `json:"status"`
	Timestamp time.Time      `json:"timestamp"`
	Service   ServiceInfo    `json:"service"`
	Database  DatabaseHealth `json:"database"`
	Checks    CheckSummary   `json:"checks"`
}

// ServiceInfo represents service information
type ServiceInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
	Uptime      string `json:"uptime"`
}

// DatabaseHealth represents database health information
type DatabaseHealth struct {
	Status       string          `json:"status"`
	ResponseTime string          `json:"response_time"`
	Connections  ConnectionStats `json:"connections,omitempty"`
	Error        string          `json:"error,omitempty"`
}

// ConnectionStats represents database connection statistics
type ConnectionStats struct {
	Active int `json:"active"`
	Idle   int `json:"idle"`
	Total  int `json:"total"`
}

// CheckSummary represents the summary of health checks
type CheckSummary struct {
	Total  int `json:"total"`
	Passed int `json:"passed"`
	Failed int `json:"failed"`
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *pgxpool.Pool, logger *logrus.Logger, cfg *config.Config) *HealthHandler {
	return &HealthHandler{
		db:        db,
		logger:    logger,
		config:    cfg,
		startTime: time.Now(),
	}
}

// LivenessHandler provides basic liveness check
func (h *HealthHandler) LivenessHandler(c *gin.Context) {
	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
		Service:   h.config.App.Name,
		Version:   h.config.App.Version,
	}

	c.JSON(http.StatusOK, response)
}

// ReadinessHandler checks if the service is ready to accept traffic
func (h *HealthHandler) ReadinessHandler(c *gin.Context) {
	// Check database connectivity if database is available
	if h.db == nil {
		// No database connection, but service is still ready for basic operations
		response := HealthResponse{
			Status:    "ok",
			Timestamp: time.Now().UTC(),
			Service:   h.config.App.Name,
			Version:   h.config.App.Version,
		}
		c.JSON(http.StatusOK, response)
		return
	}

	dbHealth := h.checkDatabaseHealth()

	if dbHealth.Status == "healthy" {
		response := HealthResponse{
			Status:    "ok",
			Timestamp: time.Now().UTC(),
			Service:   h.config.App.Name,
			Version:   h.config.App.Version,
		}
		c.JSON(http.StatusOK, response)
	} else {
		response := HealthResponse{
			Status:    "error",
			Timestamp: time.Now().UTC(),
			Service:   h.config.App.Name,
			Version:   h.config.App.Version,
		}
		c.JSON(http.StatusServiceUnavailable, response)
	}
}

// PingHandler provides simple ping/pong response
func (h *HealthHandler) PingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "pong",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   h.config.App.Name,
		"version":   h.config.App.Version,
	})
}

// StatusHandler provides comprehensive service status
func (h *HealthHandler) StatusHandler(c *gin.Context) {
	// Check database health if database is available
	var dbHealth DatabaseHealth
	var totalChecks int
	var failedChecks int

	if h.db == nil {
		// No database connection
		dbHealth = DatabaseHealth{
			Status: "unavailable",
			Error:  "Database not configured",
		}
		totalChecks = 0 // No database check
	} else {
		dbHealth = h.checkDatabaseHealth()
		totalChecks = 1 // Database check

		if dbHealth.Status != "healthy" {
			failedChecks++
		}
	}

	// Calculate overall status
	overallStatus := "healthy"
	if failedChecks > 0 {
		overallStatus = "unhealthy"
	}

	// Prepare response
	response := StatusResponse{
		Status:    overallStatus,
		Timestamp: time.Now().UTC(),
		Service: ServiceInfo{
			Name:        h.config.App.Name,
			Version:     h.config.App.Version,
			Environment: h.config.App.Environment,
			Uptime:      h.calculateUptime(),
		},
		Database: dbHealth,
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

	c.JSON(statusCode, response)
}

// checkDatabaseHealth performs database health check
func (h *HealthHandler) checkDatabaseHealth() DatabaseHealth {
	if h.db == nil {
		return DatabaseHealth{
			Status: "unavailable",
			Error:  "Database not configured",
		}
	}

	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := h.db.Ping(ctx)
	responseTime := time.Since(start)

	if err != nil {
		h.logger.WithError(err).Warn("Database health check failed")
		return DatabaseHealth{
			Status:       "unhealthy",
			ResponseTime: responseTime.Round(time.Millisecond).String(),
			Error:        err.Error(),
		}
	}

	// Get connection statistics
	stats := h.db.Stat()

	return DatabaseHealth{
		Status:       "healthy",
		ResponseTime: responseTime.Round(time.Millisecond).String(),
		Connections: ConnectionStats{
			Active: int(stats.AcquiredConns()),
			Idle:   int(stats.IdleConns()),
			Total:  int(stats.TotalConns()),
		},
	}
}

// calculateUptime returns formatted uptime string
func (h *HealthHandler) calculateUptime() string {
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
