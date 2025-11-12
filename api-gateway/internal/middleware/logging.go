package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/common/metrics"
)

type RequestLogger struct {
	logger           *logrus.Logger
	serviceName      string
	metricsCollector *metrics.MetricsCollector
}

type LogFields struct {
	Timestamp    string        `json:"timestamp"`
	Level        string        `json:"level"`
	Service      string        `json:"service"`
	RequestID    string        `json:"request_id,omitempty"`
	UserID       string        `json:"user_id,omitempty"`
	Method       string        `json:"method"`
	Path         string        `json:"path"`
	Status       int           `json:"status"`
	Duration     time.Duration `json:"duration_ms"`
	UserAgent    string        `json:"user_agent,omitempty"`
	IP           string        `json:"ip"`
	RequestSize  int64         `json:"request_size,omitempty"`
	ResponseSize int64         `json:"response_size,omitempty"`
	Error        string        `json:"error,omitempty"`
}

func NewRequestLogger(logger *logrus.Logger) *RequestLogger {
	return &RequestLogger{
		logger:           logger,
		serviceName:      "api-gateway",
		metricsCollector: metrics.NewMetricsCollector(logger, "api-gateway"),
	}
}

// GetMetricsCollector returns the metrics collector for external access
func (rl *RequestLogger) GetMetricsCollector() *metrics.MetricsCollector {
	return rl.metricsCollector
}

// RequestResponseLogger middleware logs HTTP requests and responses
func (rl *RequestLogger) RequestResponseLogger() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/health", "/ready", "/live", "/ping", "/status"},
	})
}

// DetailedRequestLogger provides comprehensive request/response logging
func (rl *RequestLogger) DetailedRequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Capture request body for logging (if needed)
		var requestBody []byte
		if c.Request.Body != nil && c.Request.ContentLength > 0 && c.Request.ContentLength < 1024*1024 { // Limit to 1MB
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Get request ID from context
		requestID := c.GetString("request_id")

		// Get user ID from context (if authenticated)
		userID := ""
		if uid, exists := c.Get("user_id"); exists {
			if uuid, ok := uid.(string); ok {
				userID = uuid
			}
		}

		// Create response writer wrapper to capture response size
		responseWriter := &responseWriter{ResponseWriter: c.Writer, status: 200}
		c.Writer = responseWriter

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Record metrics
		requestMetrics := metrics.RequestMetrics{
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			Status:    responseWriter.status,
			Duration:  duration,
			UserID:    userID,
			RequestID: requestID,
			Error:     responseWriter.status >= 400,
		}
		rl.metricsCollector.RecordRequest(requestMetrics)

		// Create log entry
		fields := LogFields{
			Timestamp:    start.UTC().Format(time.RFC3339),
			Service:      rl.serviceName,
			RequestID:    requestID,
			UserID:       userID,
			Method:       c.Request.Method,
			Path:         c.Request.URL.Path,
			Status:       responseWriter.status,
			Duration:     duration,
			UserAgent:    c.GetHeader("User-Agent"),
			IP:           c.ClientIP(),
			RequestSize:  c.Request.ContentLength,
			ResponseSize: int64(responseWriter.size),
		}

		// Add error information if status >= 400
		if responseWriter.status >= 400 {
			if len(c.Errors) > 0 {
				fields.Error = c.Errors.Last().Error()
			}
		}

		// Log based on status code
		logEntry := rl.logger.WithFields(logrus.Fields{
			"timestamp":     fields.Timestamp,
			"service":       fields.Service,
			"request_id":    fields.RequestID,
			"user_id":       fields.UserID,
			"method":        fields.Method,
			"path":          fields.Path,
			"status":        fields.Status,
			"duration_ms":   fields.Duration.Milliseconds(),
			"user_agent":    fields.UserAgent,
			"ip":            fields.IP,
			"request_size":  fields.RequestSize,
			"response_size": fields.ResponseSize,
		})

		if fields.Error != "" {
			logEntry = logEntry.WithField("error", fields.Error)
		}

		// Log level based on status code
		switch {
		case responseWriter.status >= 500:
			logEntry.Error("Request completed with server error")
		case responseWriter.status >= 400:
			logEntry.Warn("Request completed with client error")
		case responseWriter.status >= 300:
			logEntry.Info("Request completed with redirect")
		default:
			logEntry.Info("Request completed successfully")
		}
	}
}

// responseWriter wraps gin.ResponseWriter to capture response size and status
type responseWriter struct {
	gin.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(data)
	rw.size += size
	return size, err
}
