package alerting

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/common/config"
	"github.com/v-egorov/service-boilerplate/common/metrics"
)

// Alert represents an alert that has been triggered
type Alert struct {
	ID          string    `json:"id"`
	ServiceName string    `json:"service_name"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	Acked       bool      `json:"acked"`
}

// AlertManager manages alerting for a service
type AlertManager struct {
	mu               sync.RWMutex
	logger           *logrus.Logger
	serviceName      string
	config           *config.AlertingConfig
	metricsCollector *metrics.MetricsCollector
	alerts           []Alert
	lastAlertTimes   map[string]time.Time
}

// NewAlertManager creates a new alert manager
func NewAlertManager(logger *logrus.Logger, serviceName string, config *config.AlertingConfig, metricsCollector *metrics.MetricsCollector) *AlertManager {
	return &AlertManager{
		logger:           logger,
		serviceName:      serviceName,
		config:           config,
		metricsCollector: metricsCollector,
		alerts:           make([]Alert, 0),
		lastAlertTimes:   make(map[string]time.Time),
	}
}

// CheckMetrics checks current metrics and triggers alerts if thresholds are exceeded
func (am *AlertManager) CheckMetrics() {
	if !am.config.Enabled {
		return
	}

	metrics := am.metricsCollector.GetMetrics()

	// Check error rate threshold
	if metrics.ErrorRate > am.config.ErrorRateThreshold {
		alertKey := "high_error_rate"
		if am.shouldTriggerAlert(alertKey) {
			alert := Alert{
				ID:          fmt.Sprintf("%s_%s_%d", am.serviceName, alertKey, time.Now().Unix()),
				ServiceName: am.serviceName,
				Type:        "error_rate",
				Severity:    "warning",
				Message:     fmt.Sprintf("High error rate: %.2f%% (threshold: %.2f%%)", metrics.ErrorRate*100, am.config.ErrorRateThreshold*100),
				Timestamp:   time.Now(),
				Acked:       false,
			}
			am.triggerAlert(alert)
		}
	}

	// Check response time threshold
	if metrics.AvgResponseTime > time.Duration(am.config.ResponseTimeThreshold)*time.Millisecond {
		alertKey := "high_response_time"
		if am.shouldTriggerAlert(alertKey) {
			alert := Alert{
				ID:          fmt.Sprintf("%s_%s_%d", am.serviceName, alertKey, time.Now().Unix()),
				ServiceName: am.serviceName,
				Type:        "response_time",
				Severity:    "warning",
				Message:     fmt.Sprintf("High average response time: %v (threshold: %dms)", metrics.AvgResponseTime, am.config.ResponseTimeThreshold),
				Timestamp:   time.Now(),
				Acked:       false,
			}
			am.triggerAlert(alert)
		}
	}

	// Check for service unavailability (no requests in last 5 minutes)
	if metrics.RequestCount == 0 && metrics.Uptime > 5*time.Minute {
		alertKey := "service_unavailable"
		if am.shouldTriggerAlert(alertKey) {
			alert := Alert{
				ID:          fmt.Sprintf("%s_%s_%d", am.serviceName, alertKey, time.Now().Unix()),
				ServiceName: am.serviceName,
				Type:        "availability",
				Severity:    "critical",
				Message:     "Service appears unavailable - no requests processed in the last 5 minutes",
				Timestamp:   time.Now(),
				Acked:       false,
			}
			am.triggerAlert(alert)
		}
	}
}

// shouldTriggerAlert checks if an alert should be triggered based on the alert interval
func (am *AlertManager) shouldTriggerAlert(alertKey string) bool {
	am.mu.RLock()
	lastAlert, exists := am.lastAlertTimes[alertKey]
	am.mu.RUnlock()

	if !exists {
		return true
	}

	// Check if enough time has passed since the last alert
	timeSinceLastAlert := time.Since(lastAlert)
	minInterval := time.Duration(am.config.AlertIntervalMinutes) * time.Minute

	return timeSinceLastAlert >= minInterval
}

// triggerAlert triggers an alert and logs it
func (am *AlertManager) triggerAlert(alert Alert) {
	am.mu.Lock()
	am.alerts = append(am.alerts, alert)
	am.lastAlertTimes[alert.Type] = alert.Timestamp
	am.mu.Unlock()

	// Log the alert
	am.logger.WithFields(logrus.Fields{
		"alert_id":   alert.ID,
		"alert_type": alert.Type,
		"severity":   alert.Severity,
		"service":    alert.ServiceName,
	}).Warn("Alert triggered: " + alert.Message)
}

// GetActiveAlerts returns all active (unacked) alerts
func (am *AlertManager) GetActiveAlerts() []Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	activeAlerts := make([]Alert, 0)
	for _, alert := range am.alerts {
		if !alert.Acked {
			activeAlerts = append(activeAlerts, alert)
		}
	}

	return activeAlerts
}

// AckAlert acknowledges an alert by ID
func (am *AlertManager) AckAlert(alertID string) bool {
	am.mu.Lock()
	defer am.mu.Unlock()

	for i, alert := range am.alerts {
		if alert.ID == alertID {
			am.alerts[i].Acked = true
			am.logger.WithField("alert_id", alertID).Info("Alert acknowledged")
			return true
		}
	}

	return false
}

// GetAllAlerts returns all alerts (active and acked)
func (am *AlertManager) GetAllAlerts() []Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	alerts := make([]Alert, len(am.alerts))
	copy(alerts, am.alerts)
	return alerts
}
