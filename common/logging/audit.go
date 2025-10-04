package logging

import (
	"time"

	"github.com/sirupsen/logrus"
)

// AuditEvent represents an auditable security event
type AuditEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	EventType string                 `json:"event_type"`
	Service   string                 `json:"service"`
	UserID    string                 `json:"user_id,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	Resource  string                 `json:"resource,omitempty"`
	Action    string                 `json:"action"`
	Result    string                 `json:"result"` // success, failure, blocked, etc.
	TraceID   string                 `json:"trace_id,omitempty"`
	SpanID    string                 `json:"span_id,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// AuditLogger provides structured audit logging for security events
type AuditLogger struct {
	logger      *logrus.Logger
	serviceName string
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logger *logrus.Logger, serviceName string) *AuditLogger {
	return &AuditLogger{
		logger:      logger,
		serviceName: serviceName,
	}
}

// LogAuthAttempt logs authentication attempts
func (al *AuditLogger) LogAuthAttempt(requestID, ipAddress, userAgent, email, traceID, spanID string, success bool, errorMsg string) {
	event := AuditEvent{
		Timestamp: time.Now().UTC(),
		EventType: "auth_attempt",
		Service:   al.serviceName,
		RequestID: requestID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Resource:  "authentication",
		Action:    "login",
		TraceID:   traceID,
		SpanID:    spanID,
		Details: map[string]interface{}{
			"email": email,
		},
	}

	if success {
		event.Result = "success"
	} else {
		event.Result = "failure"
		event.Error = errorMsg
	}

	al.logEvent(event)
}

// LogUserCreation logs user account creation events
func (al *AuditLogger) LogUserCreation(requestID, userID, ipAddress, userAgent, traceID, spanID string, success bool, errorMsg string) {
	event := AuditEvent{
		Timestamp: time.Now().UTC(),
		EventType: "user_creation",
		Service:   al.serviceName,
		UserID:    userID,
		RequestID: requestID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Resource:  "user",
		Action:    "create",
		TraceID:   traceID,
		SpanID:    spanID,
	}

	if success {
		event.Result = "success"
	} else {
		event.Result = "failure"
		event.Error = errorMsg
	}

	al.logEvent(event)
}

// LogTokenOperation logs JWT token operations
func (al *AuditLogger) LogTokenOperation(requestID, userID, ipAddress, userAgent, operation, traceID, spanID string, success bool, errorMsg string) {
	event := AuditEvent{
		Timestamp: time.Now().UTC(),
		EventType: "token_operation",
		Service:   al.serviceName,
		UserID:    userID,
		RequestID: requestID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Resource:  "token",
		Action:    operation,
		TraceID:   traceID,
		SpanID:    spanID,
	}

	if success {
		event.Result = "success"
	} else {
		event.Result = "failure"
		event.Error = errorMsg
	}

	al.logEvent(event)
}

// LogPasswordChange logs password change attempts
func (al *AuditLogger) LogPasswordChange(requestID, userID, ipAddress, userAgent, traceID, spanID string, success bool, errorMsg string) {
	event := AuditEvent{
		Timestamp: time.Now().UTC(),
		EventType: "password_change",
		Service:   al.serviceName,
		UserID:    userID,
		RequestID: requestID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Resource:  "password",
		Action:    "change",
		TraceID:   traceID,
		SpanID:    spanID,
	}

	if success {
		event.Result = "success"
	} else {
		event.Result = "failure"
		event.Error = errorMsg
	}

	al.logEvent(event)
}

// LogEntityCreation logs entity creation events
func (al *AuditLogger) LogEntityCreation(requestID, entityID, ipAddress, userAgent, traceID, spanID string, success bool, errorMsg string) {
	event := AuditEvent{
		Timestamp: time.Now().UTC(),
		EventType: "entity_creation",
		Service:   al.serviceName,
		UserID:    entityID,
		RequestID: requestID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Resource:  "entity",
		Action:    "create",
		TraceID:   traceID,
		SpanID:    spanID,
	}

	if success {
		event.Result = "success"
	} else {
		event.Result = "failure"
		event.Error = errorMsg
	}

	al.logEvent(event)
}

// LogEntityUpdate logs entity update events
func (al *AuditLogger) LogEntityUpdate(requestID, entityID, ipAddress, userAgent, traceID, spanID string, success bool, errorMsg string) {
	event := AuditEvent{
		Timestamp: time.Now().UTC(),
		EventType: "entity_update",
		Service:   al.serviceName,
		UserID:    entityID,
		RequestID: requestID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Resource:  "entity",
		Action:    "update",
		TraceID:   traceID,
		SpanID:    spanID,
	}

	if success {
		event.Result = "success"
	} else {
		event.Result = "failure"
		event.Error = errorMsg
	}

	al.logEvent(event)
}

// LogEntityDeletion logs entity deletion events
func (al *AuditLogger) LogEntityDeletion(requestID, entityID, ipAddress, userAgent, traceID, spanID string, success bool, errorMsg string) {
	event := AuditEvent{
		Timestamp: time.Now().UTC(),
		EventType: "entity_deletion",
		Service:   al.serviceName,
		UserID:    entityID,
		RequestID: requestID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Resource:  "entity",
		Action:    "delete",
		TraceID:   traceID,
		SpanID:    spanID,
	}

	if success {
		event.Result = "success"
	} else {
		event.Result = "failure"
		event.Error = errorMsg
	}

	al.logEvent(event)
}

// LogSuspiciousActivity logs potentially suspicious activities
func (al *AuditLogger) LogSuspiciousActivity(requestID, ipAddress, userAgent, activityType, traceID, spanID string, details map[string]interface{}) {
	event := AuditEvent{
		Timestamp: time.Now().UTC(),
		EventType: "suspicious_activity",
		Service:   al.serviceName,
		RequestID: requestID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Resource:  "security",
		Action:    activityType,
		Result:    "flagged",
		TraceID:   traceID,
		SpanID:    spanID,
		Details:   details,
	}

	al.logEvent(event)
}

// logEvent logs the audit event with appropriate log level
func (al *AuditLogger) logEvent(event AuditEvent) {
	logEntry := al.logger.WithFields(logrus.Fields{
		"audit_event": true,
		"event_type":  event.EventType,
		"service":     event.Service,
		"user_id":     event.UserID,
		"request_id":  event.RequestID,
		"ip_address":  event.IPAddress,
		"user_agent":  event.UserAgent,
		"resource":    event.Resource,
		"action":      event.Action,
		"result":      event.Result,
		"trace_id":    event.TraceID,
		"span_id":     event.SpanID,
		"timestamp":   event.Timestamp.Format(time.RFC3339),
	})

	// Add details if present
	if event.Details != nil {
		for k, v := range event.Details {
			logEntry = logEntry.WithField("detail_"+k, v)
		}
	}

	// Add error if present
	if event.Error != "" {
		logEntry = logEntry.WithField("error", event.Error)
	}

	// Log based on result
	switch event.Result {
	case "failure", "blocked":
		logEntry.Warn("Security event: " + event.EventType)
	case "flagged":
		logEntry.Error("Suspicious activity detected: " + event.EventType)
	default:
		logEntry.Info("Security event: " + event.EventType)
	}
}
