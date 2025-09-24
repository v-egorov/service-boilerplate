package logging

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
}

type Config struct {
	Level       string
	Format      string
	Output      string
	FilePath    string
	DualOutput  bool
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
	case "file":
		if config.FilePath == "" {
			config.FilePath = "/app/logs/" + config.ServiceName + ".log"
		}
		// Use lumberjack for log rotation
		lumberjackWriter := &lumberjack.Logger{
			Filename:   config.FilePath,
			MaxSize:    10, // megabytes
			MaxBackups: 3,  // number of backups
			MaxAge:     28, // days
			Compress:   true,
		}

		if config.DualOutput {
			// Output to both file and stdout for Docker logging
			logger.SetOutput(io.MultiWriter(lumberjackWriter, os.Stdout))
		} else {
			// Output to file only
			logger.SetOutput(lumberjackWriter)
		}
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
