package logging

import "github.com/sirupsen/logrus"

// Standardized field names for consistent logging across services
const (
	FieldTimestamp    = "timestamp"
	FieldService      = "service"
	FieldRequestID    = "request_id"
	FieldUserID       = "user_id"
	FieldResourceID   = "resource_id"
	FieldOperation    = "operation"
	FieldMethod       = "method"
	FieldPath         = "path"
	FieldStatus       = "status"
	FieldDuration     = "duration_ms"
	FieldUserAgent    = "user_agent"
	FieldIPAddress    = "ip_address"
	FieldRequestSize  = "request_size"
	FieldResponseSize = "response_size"
	FieldError        = "error"
	FieldEmail        = "email"
	FieldCount        = "count"
	FieldAuditEvent   = "audit_event"
	FieldEventType    = "event_type"
	FieldAction       = "action"
	FieldResult       = "result"
)

// StandardLogger provides consistent logging methods with standardized field names
type StandardLogger struct {
	logger      *logrus.Logger
	serviceName string
}

// NewStandardLogger creates a new standard logger
func NewStandardLogger(logger *logrus.Logger, serviceName string) *StandardLogger {
	return &StandardLogger{
		logger:      logger,
		serviceName: serviceName,
	}
}

// WithRequestID adds request_id to the log entry
func (sl *StandardLogger) WithRequestID(requestID string) *logrus.Entry {
	return sl.logger.WithField(FieldRequestID, requestID)
}

// WithUserID adds user_id to the log entry
func (sl *StandardLogger) WithUserID(userID string) *logrus.Entry {
	return sl.logger.WithField(FieldUserID, userID)
}

// WithResourceID adds resource_id to the log entry (for entities, records, etc.)
func (sl *StandardLogger) WithResourceID(resourceID string) *logrus.Entry {
	return sl.logger.WithField(FieldResourceID, resourceID)
}

// WithOperation adds operation to the log entry
func (sl *StandardLogger) WithOperation(operation string) *logrus.Entry {
	return sl.logger.WithField(FieldOperation, operation)
}

// WithCount adds count to the log entry
func (sl *StandardLogger) WithCount(count int) *logrus.Entry {
	return sl.logger.WithField(FieldCount, count)
}

// WithEmail adds email to the log entry
func (sl *StandardLogger) WithEmail(email string) *logrus.Entry {
	return sl.logger.WithField(FieldEmail, email)
}

// UserOperation logs a user-related operation with standardized fields
func (sl *StandardLogger) UserOperation(requestID, userID, operation string, success bool, err error) {
	entry := sl.logger.WithFields(logrus.Fields{
		FieldService:   sl.serviceName,
		FieldRequestID: requestID,
		FieldUserID:    userID,
		FieldOperation: operation,
	})

	if err != nil {
		entry = entry.WithField(FieldError, err.Error())
		entry.Error("User operation failed")
	} else {
		entry.Info("User operation completed")
	}
}

// EntityOperation logs an entity-related operation with standardized fields
func (sl *StandardLogger) EntityOperation(requestID, userID, entityID, operation string, success bool, err error) {
	entry := sl.logger.WithFields(logrus.Fields{
		FieldService:    sl.serviceName,
		FieldRequestID:  requestID,
		FieldUserID:     userID,
		FieldResourceID: entityID,
		FieldOperation:  operation,
	})

	if err != nil {
		entry = entry.WithField(FieldError, err.Error())
		entry.Error("Entity operation failed")
	} else {
		entry.Info("Entity operation completed")
	}
}

// AuthOperation logs an authentication operation with standardized fields
func (sl *StandardLogger) AuthOperation(requestID, userID, email, operation string, success bool, err error) {
	entry := sl.logger.WithFields(logrus.Fields{
		FieldService:   sl.serviceName,
		FieldRequestID: requestID,
		FieldUserID:    userID,
		FieldEmail:     email,
		FieldOperation: operation,
	})

	if err != nil {
		entry = entry.WithField(FieldError, err.Error())
		entry.Warn("Authentication operation failed")
	} else {
		entry.Info("Authentication operation completed")
	}
}

// DatabaseOperation logs a database operation with standardized fields
func (sl *StandardLogger) DatabaseOperation(operation, table string, recordID string, success bool, err error) {
	entry := sl.logger.WithFields(logrus.Fields{
		FieldService:    sl.serviceName,
		FieldOperation:  operation,
		FieldResourceID: recordID,
		"table":         table,
	})

	if err != nil {
		entry = entry.WithField(FieldError, err.Error())
		entry.Error("Database operation failed")
	} else {
		entry.Debug("Database operation completed")
	}
}
