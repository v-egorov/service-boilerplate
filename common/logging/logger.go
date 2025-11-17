package logging

import (
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
}

type Config struct {
	Level              string
	Format             string
	Output             string
	FilePath           string
	DualOutput         bool
	ServiceName        string
	StripANSIFromFiles bool
}

// stripANSIWriter wraps a writer and strips ANSI escape codes from writes
type stripANSIWriter struct {
	io.Writer
}

func (w *stripANSIWriter) Write(p []byte) (n int, err error) {
	stripped := stripANSI(string(p))
	_, err = w.Writer.Write([]byte(stripped))
	if err != nil {
		return 0, err
	}
	return len(p), nil // Return len(p) since we consumed all input
}

// stripANSI removes ANSI escape codes from a string
func stripANSI(str string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(str, "")
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
		logger.SetFormatter(NewPrioritizedJSONFormatter())
	default:
		logger.SetFormatter(NewColoredFormatter())
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

		// Always strip ANSI from file outputs for clean logs
		strippedLumberjackWriter := &stripANSIWriter{Writer: lumberjackWriter}

		if config.DualOutput {
			// Output to both file and stdout for Docker logging
			logger.SetOutput(io.MultiWriter(strippedLumberjackWriter, os.Stdout))
		} else {
			// Output to file only
			logger.SetOutput(strippedLumberjackWriter)
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
