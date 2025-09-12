package monitoring

import (
	"fmt"
	"sync"
	"time"

	"github.com/v-egorov/service-boilerplate/cli/internal/client"
	"github.com/v-egorov/service-boilerplate/cli/internal/config"
	"github.com/v-egorov/service-boilerplate/cli/internal/discovery"
)

// ServiceMetrics holds performance metrics for a service
type ServiceMetrics struct {
	ServiceName  string        `json:"service_name"`
	ResponseTime time.Duration `json:"response_time"`
	StatusCode   int           `json:"status_code"`
	Success      bool          `json:"success"`
	ErrorMessage string        `json:"error_message,omitempty"`
	Timestamp    time.Time     `json:"timestamp"`
	Endpoint     string        `json:"endpoint"`
}

// HealthStatus represents the health status of a service
type HealthStatus struct {
	ServiceName   string                 `json:"service_name"`
	Status        string                 `json:"status"` // healthy, unhealthy, degraded
	ResponseTime  time.Duration          `json:"response_time"`
	LastChecked   time.Time              `json:"last_checked"`
	Uptime        float64                `json:"uptime_percentage"`
	ErrorCount    int                    `json:"error_count"`
	TotalRequests int                    `json:"total_requests"`
	Details       map[string]interface{} `json:"details,omitempty"`
}

// SystemMetrics holds overall system metrics
type SystemMetrics struct {
	TotalServices       int                      `json:"total_services"`
	HealthyServices     int                      `json:"healthy_services"`
	UnhealthyServices   int                      `json:"unhealthy_services"`
	AverageResponseTime time.Duration            `json:"average_response_time"`
	UptimePercentage    float64                  `json:"uptime_percentage"`
	ServiceMetrics      map[string]*HealthStatus `json:"service_metrics"`
	Timestamp           time.Time                `json:"timestamp"`
}

// Monitor provides monitoring and metrics collection
type Monitor struct {
	config       *config.Config
	serviceReg   *discovery.ServiceRegistry
	apiClient    *client.APIClient
	metrics      map[string][]*ServiceMetrics
	healthStatus map[string]*HealthStatus
	mutex        sync.RWMutex
}

// NewMonitor creates a new monitoring instance
func NewMonitor(cfg *config.Config, serviceReg *discovery.ServiceRegistry, apiClient *client.APIClient) *Monitor {
	return &Monitor{
		config:       cfg,
		serviceReg:   serviceReg,
		apiClient:    apiClient,
		metrics:      make(map[string][]*ServiceMetrics),
		healthStatus: make(map[string]*HealthStatus),
	}
}

// CheckServiceHealth performs a comprehensive health check on a service
func (m *Monitor) CheckServiceHealth(serviceName string) (*HealthStatus, error) {
	service, err := m.serviceReg.GetService(serviceName)
	if err != nil {
		return nil, fmt.Errorf("service not found: %w", err)
	}

	start := time.Now()

	// Perform multiple health checks
	healthChecks := []struct {
		name     string
		endpoint string
		method   string
		required bool
	}{
		{"health", "/health", "GET", true},
		{"readiness", "/api/v1/users?limit=1", "GET", false},
		{"liveness", "/api/v1/users", "GET", false},
	}

	var totalResponseTime time.Duration
	var successfulChecks int
	var errors []string

	for _, check := range healthChecks {
		resp, err := m.apiClient.CallService(serviceName, check.method, check.endpoint, nil, nil)
		responseTime := time.Since(start)

		metric := &ServiceMetrics{
			ServiceName:  serviceName,
			ResponseTime: responseTime,
			Timestamp:    time.Now(),
			Endpoint:     check.endpoint,
		}

		if err != nil {
			metric.Success = false
			metric.ErrorMessage = err.Error()
			if check.required {
				errors = append(errors, fmt.Sprintf("%s: %v", check.name, err))
			}
		} else {
			metric.StatusCode = resp.StatusCode
			metric.Success = resp.StatusCode >= 200 && resp.StatusCode < 300
			if !metric.Success {
				metric.ErrorMessage = fmt.Sprintf("HTTP %d", resp.StatusCode)
				if check.required {
					errors = append(errors, fmt.Sprintf("%s: HTTP %d", check.name, resp.StatusCode))
				}
			}
		}

		if metric.Success {
			successfulChecks++
		}

		totalResponseTime += responseTime
		m.recordMetric(metric)
	}

	// Determine overall health status
	var status string
	var uptime float64

	m.mutex.Lock()
	health := m.healthStatus[serviceName]
	if health == nil {
		health = &HealthStatus{
			ServiceName: serviceName,
			Uptime:      100.0,
		}
	}

	health.LastChecked = time.Now()
	health.ResponseTime = totalResponseTime / time.Duration(len(healthChecks))
	health.TotalRequests++

	if len(errors) == 0 {
		status = "healthy"
		health.Uptime = (health.Uptime*float64(health.TotalRequests-1) + 100.0) / float64(health.TotalRequests)
	} else if successfulChecks > 0 {
		status = "degraded"
		health.Uptime = (health.Uptime*float64(health.TotalRequests-1) + 50.0) / float64(health.TotalRequests)
		health.ErrorCount++
	} else {
		status = "unhealthy"
		health.Uptime = (health.Uptime*float64(health.TotalRequests-1) + 0.0) / float64(health.TotalRequests)
		health.ErrorCount++
	}

	health.Status = status
	uptime = health.Uptime

	// Add details
	health.Details = map[string]interface{}{
		"successful_checks": successfulChecks,
		"total_checks":      len(healthChecks),
		"errors":            errors,
		"url":               service.URL,
	}

	m.healthStatus[serviceName] = health
	m.mutex.Unlock()

	return &HealthStatus{
		ServiceName:   serviceName,
		Status:        status,
		ResponseTime:  totalResponseTime / time.Duration(len(healthChecks)),
		LastChecked:   time.Now(),
		Uptime:        uptime,
		ErrorCount:    health.ErrorCount,
		TotalRequests: health.TotalRequests,
		Details:       health.Details,
	}, nil
}

// GetSystemMetrics returns comprehensive system metrics
func (m *Monitor) GetSystemMetrics() (*SystemMetrics, error) {
	services := m.serviceReg.GetAllServices()

	var totalResponseTime time.Duration
	healthyCount := 0
	unhealthyCount := 0
	serviceMetrics := make(map[string]*HealthStatus)

	for _, service := range services {
		health, err := m.CheckServiceHealth(service.Name)
		if err != nil {
			unhealthyCount++
			continue
		}

		serviceMetrics[service.Name] = health
		totalResponseTime += health.ResponseTime

		if health.Status == "healthy" {
			healthyCount++
		} else {
			unhealthyCount++
		}
	}

	var avgResponseTime time.Duration
	var uptimePercentage float64

	if len(services) > 0 {
		avgResponseTime = totalResponseTime / time.Duration(len(services))
		uptimePercentage = float64(healthyCount) / float64(len(services)) * 100.0
	}

	return &SystemMetrics{
		TotalServices:       len(services),
		HealthyServices:     healthyCount,
		UnhealthyServices:   unhealthyCount,
		AverageResponseTime: avgResponseTime,
		UptimePercentage:    uptimePercentage,
		ServiceMetrics:      serviceMetrics,
		Timestamp:           time.Now(),
	}, nil
}

// GetServiceMetrics returns metrics for a specific service
func (m *Monitor) GetServiceMetrics(serviceName string, limit int) ([]*ServiceMetrics, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	metrics, exists := m.metrics[serviceName]
	if !exists {
		return nil, fmt.Errorf("no metrics found for service: %s", serviceName)
	}

	if limit <= 0 || limit > len(metrics) {
		limit = len(metrics)
	}

	// Return the most recent metrics
	start := len(metrics) - limit
	if start < 0 {
		start = 0
	}

	return metrics[start:], nil
}

// recordMetric records a service metric
func (m *Monitor) recordMetric(metric *ServiceMetrics) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.metrics[metric.ServiceName] == nil {
		m.metrics[metric.ServiceName] = make([]*ServiceMetrics, 0)
	}

	// Keep only the last 100 metrics per service
	metrics := m.metrics[metric.ServiceName]
	if len(metrics) >= 100 {
		metrics = metrics[1:]
	}

	m.metrics[metric.ServiceName] = append(metrics, metric)
}

// GetHealthStatus returns the current health status of all services
func (m *Monitor) GetHealthStatus() map[string]*HealthStatus {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string]*HealthStatus)
	for name, status := range m.healthStatus {
		result[name] = status
	}
	return result
}
