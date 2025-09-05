package logging

import (
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
}

type Config struct {
	Level       string
	Format      string
	Output      string
	ServiceName string
}

func NewLogger(config Config) *Logger {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set formatter
	switch strings.ToLower(config.Format) {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
	default:
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}

	// Set output
	switch strings.ToLower(config.Output) {
	case "stderr":
		logger.SetOutput(os.Stderr)
	default:
		logger.SetOutput(os.Stdout)
	}

	return &Logger{Logger: logger}
}

func (l *Logger) WithService(serviceName string) *Logger {
	return &Logger{
		Logger: l.Logger.WithField("service", serviceName).Logger,
	}
}

func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		Logger: l.Logger.WithField("request_id", requestID).Logger,
	}
}

func (l *Logger) WithFields(fields logrus.Fields) *Logger {
	return &Logger{
		Logger: l.Logger.WithFields(fields).Logger,
	}
}
