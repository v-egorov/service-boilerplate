package logging

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/common/metrics"
	"go.opentelemetry.io/otel/trace"
)

// ServiceRequestLogger provides comprehensive request/response logging for services
type ServiceRequestLogger struct {
	logger           *logrus.Logger
	serviceName      string
	metricsCollector *metrics.MetricsCollector
}

// NewServiceRequestLogger creates a new service request logger
func NewServiceRequestLogger(logger *logrus.Logger, serviceName string) *ServiceRequestLogger {
	return &ServiceRequestLogger{
		logger:           logger,
		serviceName:      serviceName,
		metricsCollector: metrics.NewMetricsCollector(logger, serviceName),
	}
}

// GetMetricsCollector returns the metrics collector for external access
func (srl *ServiceRequestLogger) GetMetricsCollector() *metrics.MetricsCollector {
	return srl.metricsCollector
}

// RequestResponseLogger middleware logs HTTP requests and responses for services
func (srl *ServiceRequestLogger) RequestResponseLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Capture request body for logging (if needed and reasonable size)
		var requestBody []byte
		if c.Request.Body != nil && c.Request.ContentLength > 0 && c.Request.ContentLength < 1024*1024 { // Limit to 1MB
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Get request ID from context (set by requestIDMiddleware)
		requestID := ""
		if rid, ok := c.Request.Context().Value("request_id").(string); ok {
			requestID = rid
		}
		// Fallback to header if not in context
		if requestID == "" {
			requestID = c.GetHeader("X-Request-ID")
		}

		// Exctract trace information from context
		span := trace.SpanFromContext(c.Request.Context())
		traceID := span.SpanContext().TraceID().String()
		spanID := span.SpanContext().SpanID().String()

		// Get user ID from context (if authenticated)
		userID := ""
		if uid, exists := c.Get("user_id"); exists {
			if uuid, ok := uid.(string); ok {
				userID = uuid
			} else {
				srl.logger.WithFields(logrus.Fields{
					"request_id":    requestID,
					"path":          c.Request.URL.Path,
					"method":        c.Request.Method,
					"user_id_type":  fmt.Sprintf("%T", uid),
					"user_id_value": fmt.Sprintf("%v", uid),
				}).Warn("Logging middleware: user_id exists in context but wrong type")
			}
		} else {
			srl.logger.WithFields(logrus.Fields{
				"request_id": requestID,
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
			}).Debug("Logging middleware: user_id not found in context")
		}

		// Create response writer wrapper to capture response size and status
		responseWriter := &responseWriter{ResponseWriter: c.Writer, status: 200}
		c.Writer = responseWriter

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Create log entry with standardized fields
		logEntry := srl.logger.WithFields(logrus.Fields{
			"service":       srl.serviceName,
			"request_id":    requestID,
			"trace_id":      traceID,
			"span_id":       spanID,
			"user_id":       userID,
			"method":        c.Request.Method,
			"path":          c.Request.URL.Path,
			"status":        responseWriter.status,
			"duration_ms":   duration.Milliseconds(),
			"user_agent":    c.GetHeader("User-Agent"),
			"ip":            c.ClientIP(),
			"request_size":  c.Request.ContentLength,
			"response_size": int64(responseWriter.size),
		})

		// Add operation context if available
		if operation, exists := c.Get("operation"); exists {
			logEntry = logEntry.WithField("operation", operation.(string))
		}

		// Add error information if status >= 400
		if responseWriter.status >= 400 {
			if len(c.Errors) > 0 {
				logEntry = logEntry.WithField("error", c.Errors.Last().Error())
			}
		}

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
		srl.metricsCollector.RecordRequest(requestMetrics)

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
