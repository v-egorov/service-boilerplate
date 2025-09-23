package metrics

import (
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// MetricsCollector provides basic metrics collection for services
type MetricsCollector struct {
	mu          sync.RWMutex
	logger      *logrus.Logger
	serviceName string

	// Request metrics
	requestCount  int64
	errorCount    int64
	totalDuration time.Duration

	// Response time percentiles (simple implementation)
	responseTimes []time.Duration

	// Endpoint-specific metrics
	endpointMetrics map[string]*EndpointMetrics // key: "METHOD /normalized-path"

	// Business metrics
	businessMetrics map[string]int64

	// Start time for uptime calculation
	startTime time.Time
}

// RequestMetrics represents metrics for a single request
type RequestMetrics struct {
	Method    string
	Path      string
	Status    int
	Duration  time.Duration
	UserID    string
	RequestID string
	Error     bool
}

// EndpointMetrics represents metrics for a specific endpoint
type EndpointMetrics struct {
	Path          string
	Method        string
	RequestCount  int64
	ErrorCount    int64
	TotalDuration time.Duration
	ResponseTimes []time.Duration
}

// EndpointMetricsSummary represents summarized metrics for an endpoint
type EndpointMetricsSummary struct {
	Path            string        `json:"path"`
	Method          string        `json:"method,omitempty"`
	Requests        int64         `json:"requests"`
	ErrorRate       float64       `json:"error_rate"`
	AvgResponseTime time.Duration `json:"avg_response_time"`
	P95ResponseTime time.Duration `json:"p95_response_time,omitempty"`
	P99ResponseTime time.Duration `json:"p99_response_time,omitempty"`
}

// ServiceMetrics represents aggregated metrics for the service
type ServiceMetrics struct {
	ServiceName     string                            `json:"service_name"`
	Uptime          time.Duration                     `json:"uptime"`
	RequestCount    int64                             `json:"request_count"`
	ErrorCount      int64                             `json:"error_count"`
	ErrorRate       float64                           `json:"error_rate"`
	AvgResponseTime time.Duration                     `json:"avg_response_time"`
	P95ResponseTime time.Duration                     `json:"p95_response_time,omitempty"`
	P99ResponseTime time.Duration                     `json:"p99_response_time,omitempty"`
	EndpointMetrics map[string]EndpointMetricsSummary `json:"endpoint_metrics,omitempty"`
	BusinessMetrics map[string]int64                  `json:"business_metrics,omitempty"`
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(logger *logrus.Logger, serviceName string) *MetricsCollector {
	return &MetricsCollector{
		logger:          logger,
		serviceName:     serviceName,
		businessMetrics: make(map[string]int64),
		endpointMetrics: make(map[string]*EndpointMetrics),
		responseTimes:   make([]time.Duration, 0, 1000), // Keep last 1000 response times
		startTime:       time.Now(),
	}
}

// RecordRequest records metrics for a completed request
func (mc *MetricsCollector) RecordRequest(metrics RequestMetrics) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Update counters
	mc.requestCount++
	if metrics.Error || metrics.Status >= 400 {
		mc.errorCount++
	}

	// Update response time tracking
	mc.totalDuration += metrics.Duration
	mc.responseTimes = append(mc.responseTimes, metrics.Duration)

	// Keep only last 1000 response times for percentile calculation
	if len(mc.responseTimes) > 1000 {
		mc.responseTimes = mc.responseTimes[1:]
	}

	// Log high latency requests
	if metrics.Duration > 5*time.Second {
		mc.logger.WithFields(logrus.Fields{
			"service":     mc.serviceName,
			"request_id":  metrics.RequestID,
			"user_id":     metrics.UserID,
			"method":      metrics.Method,
			"path":        metrics.Path,
			"duration_ms": metrics.Duration.Milliseconds(),
			"status":      metrics.Status,
		}).Warn("High latency request detected")
	}

	// Record endpoint-specific metrics
	mc.recordEndpointMetrics(metrics)

	// Log errors
	if metrics.Error || metrics.Status >= 500 {
		mc.logger.WithFields(logrus.Fields{
			"service":    mc.serviceName,
			"request_id": metrics.RequestID,
			"user_id":    metrics.UserID,
			"method":     metrics.Method,
			"path":       metrics.Path,
			"status":     metrics.Status,
		}).Error("Request error recorded")
	}
}

// recordEndpointMetrics records metrics for a specific endpoint
func (mc *MetricsCollector) recordEndpointMetrics(metrics RequestMetrics) {
	normalizedPath := normalizePath(metrics.Path)
	endpointKey := metrics.Method + " " + normalizedPath

	endpoint, exists := mc.endpointMetrics[endpointKey]
	if !exists {
		endpoint = &EndpointMetrics{
			Path:          normalizedPath,
			Method:        metrics.Method,
			ResponseTimes: make([]time.Duration, 0, 100), // Keep last 100 response times per endpoint
		}
		mc.endpointMetrics[endpointKey] = endpoint
	}

	// Update endpoint counters
	endpoint.RequestCount++
	if metrics.Error || metrics.Status >= 400 {
		endpoint.ErrorCount++
	}
	endpoint.TotalDuration += metrics.Duration
	endpoint.ResponseTimes = append(endpoint.ResponseTimes, metrics.Duration)

	// Keep only last 100 response times for percentile calculation
	if len(endpoint.ResponseTimes) > 100 {
		endpoint.ResponseTimes = endpoint.ResponseTimes[1:]
	}
}

// normalizePath normalizes parameterized paths for consistent grouping
func normalizePath(path string) string {
	if path == "" {
		return path
	}

	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "" {
			continue
		}
		if regexp.MustCompile(`^\d+$`).MatchString(part) {
			parts[i] = "{id}"
		} else if regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`).MatchString(part) {
			parts[i] = "{uuid}"
		} else if strings.Contains(part, "@") {
			parts[i] = "{email}"
		}
	}
	return strings.Join(parts, "/")
}

// IncrementBusinessMetric increments a business metric counter
func (mc *MetricsCollector) IncrementBusinessMetric(metricName string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.businessMetrics[metricName]++
}

// GetMetrics returns current aggregated metrics
func (mc *MetricsCollector) GetMetrics() ServiceMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	var avgResponseTime time.Duration
	var p95ResponseTime time.Duration
	var p99ResponseTime time.Duration
	var errorRate float64

	if mc.requestCount > 0 {
		avgResponseTime = mc.totalDuration / time.Duration(mc.requestCount)
		errorRate = float64(mc.errorCount) / float64(mc.requestCount)

		// Calculate percentiles (simple implementation)
		if len(mc.responseTimes) > 0 {
			sortedTimes := make([]time.Duration, len(mc.responseTimes))
			copy(sortedTimes, mc.responseTimes)

			// Simple sort for percentile calculation
			for i := 0; i < len(sortedTimes)-1; i++ {
				for j := i + 1; j < len(sortedTimes); j++ {
					if sortedTimes[i] > sortedTimes[j] {
						sortedTimes[i], sortedTimes[j] = sortedTimes[j], sortedTimes[i]
					}
				}
			}

			p95Index := int(float64(len(sortedTimes)) * 0.95)
			p99Index := int(float64(len(sortedTimes)) * 0.99)

			if p95Index < len(sortedTimes) {
				p95ResponseTime = sortedTimes[p95Index]
			}
			if p99Index < len(sortedTimes) {
				p99ResponseTime = sortedTimes[p99Index]
			}
		}
	}

	// Calculate endpoint metrics
	endpointMetrics := mc.calculateEndpointMetrics()

	// Copy business metrics
	businessMetrics := make(map[string]int64)
	for k, v := range mc.businessMetrics {
		businessMetrics[k] = v
	}

	return ServiceMetrics{
		ServiceName:     mc.serviceName,
		Uptime:          time.Since(mc.startTime),
		RequestCount:    mc.requestCount,
		ErrorCount:      mc.errorCount,
		ErrorRate:       errorRate,
		AvgResponseTime: avgResponseTime,
		P95ResponseTime: p95ResponseTime,
		P99ResponseTime: p99ResponseTime,
		EndpointMetrics: endpointMetrics,
		BusinessMetrics: businessMetrics,
	}
}

// calculateEndpointMetrics calculates summarized metrics for each endpoint
func (mc *MetricsCollector) calculateEndpointMetrics() map[string]EndpointMetricsSummary {
	endpointMetrics := make(map[string]EndpointMetricsSummary)

	for key, endpoint := range mc.endpointMetrics {
		if endpoint.RequestCount == 0 {
			continue
		}

		// Calculate error rate
		errorRate := float64(endpoint.ErrorCount) / float64(endpoint.RequestCount)

		// Calculate average response time
		avgResponseTime := endpoint.TotalDuration / time.Duration(endpoint.RequestCount)

		// Calculate percentiles for response times
		var p95ResponseTime, p99ResponseTime time.Duration
		if len(endpoint.ResponseTimes) > 0 {
			sortedTimes := make([]time.Duration, len(endpoint.ResponseTimes))
			copy(sortedTimes, endpoint.ResponseTimes)

			// Simple sort for percentile calculation
			sort.Slice(sortedTimes, func(i, j int) bool {
				return sortedTimes[i] < sortedTimes[j]
			})

			p95Index := int(float64(len(sortedTimes)) * 0.95)
			p99Index := int(float64(len(sortedTimes)) * 0.99)

			if p95Index < len(sortedTimes) {
				p95ResponseTime = sortedTimes[p95Index]
			}
			if p99Index < len(sortedTimes) {
				p99ResponseTime = sortedTimes[p99Index]
			}
		}

		endpointMetrics[key] = EndpointMetricsSummary{
			Path:            endpoint.Path,
			Method:          endpoint.Method,
			Requests:        endpoint.RequestCount,
			ErrorRate:       errorRate,
			AvgResponseTime: avgResponseTime,
			P95ResponseTime: p95ResponseTime,
			P99ResponseTime: p99ResponseTime,
		}
	}

	return endpointMetrics
}

// Reset resets all metrics (useful for testing)
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.requestCount = 0
	mc.errorCount = 0
	mc.totalDuration = 0
	mc.responseTimes = mc.responseTimes[:0]
	mc.endpointMetrics = make(map[string]*EndpointMetrics)
	mc.businessMetrics = make(map[string]int64)
	mc.startTime = time.Now()
}
