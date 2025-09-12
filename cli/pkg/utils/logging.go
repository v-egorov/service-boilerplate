package utils

import (
	"fmt"
	"os"
	"time"
)

// Logger provides structured logging for CLI operations
type Logger struct {
	verbose bool
}

// NewLogger creates a new CLI logger
func NewLogger(verbose bool) *Logger {
	return &Logger{
		verbose: verbose,
	}
}

// Info logs informational messages
func (l *Logger) Info(message string, args ...interface{}) {
	if l.verbose {
		timestamp := time.Now().Format("15:04:05")
		fmt.Fprintf(os.Stderr, "[%s] ℹ️  %s\n", timestamp, fmt.Sprintf(message, args...))
	}
}

// Success logs success messages
func (l *Logger) Success(message string, args ...interface{}) {
	fmt.Printf("✅ %s\n", fmt.Sprintf(message, args...))
}

// Warning logs warning messages
func (l *Logger) Warning(message string, args ...interface{}) {
	fmt.Printf("⚠️  %s\n", fmt.Sprintf(message, args...))
}

// Error logs error messages
func (l *Logger) Error(message string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "❌ %s\n", fmt.Sprintf(message, args...))
}

// Debug logs debug messages (only in verbose mode)
func (l *Logger) Debug(message string, args ...interface{}) {
	if l.verbose {
		timestamp := time.Now().Format("15:04:05")
		fmt.Fprintf(os.Stderr, "[%s] 🔍 %s\n", timestamp, fmt.Sprintf(message, args...))
	}
}

// Progress shows progress for long-running operations
func (l *Logger) Progress(current, total int, message string) {
	if l.verbose {
		percentage := float64(current) / float64(total) * 100
		fmt.Fprintf(os.Stderr, "\r📊 Progress: %d/%d (%.1f%%) - %s", current, total, percentage, message)
		if current == total {
			fmt.Fprintf(os.Stderr, "\n")
		}
	}
}
