package logging

import (
	"strings"

	"github.com/sirupsen/logrus"
)

// ColoredFormatter provides enhanced visual formatting for logs
type ColoredFormatter struct {
	logrus.TextFormatter
}

// Format renders a single log entry with enhanced colors for HTTP-related fields
func (f *ColoredFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Get the standard text formatter output first
	formatted, err := f.TextFormatter.Format(entry)
	if err != nil {
		return formatted, err
	}

	// Convert to string for manipulation
	logLine := string(formatted)

	// Apply color highlighting for specific fields
	logLine = f.highlightHTTPMethod(logLine)
	logLine = f.highlightStatusCodeInFormattedText(logLine)
	logLine = f.highlightDuration(logLine)
	logLine = f.highlightErrors(logLine)
	logLine = f.highlightServices(logLine)
	logLine = f.highlightRequestIDs(logLine)

	return []byte(logLine), nil
}

// highlightStatusCodeInFormattedText highlights status codes in the already formatted text
func (f *ColoredFormatter) highlightStatusCodeInFormattedText(line string) string {
	// Look for status values in the formatted text and add background colors
	// This handles cases where the status appears as: status=401

	statusPatterns := []struct {
		pattern     string
		replacement string
	}{
		// 2xx Success - Bold green foreground
		{"status=200", "status=\033[1;32m200\033[0m"},
		{"status=201", "status=\033[1;32m201\033[0m"},
		{"status=202", "status=\033[1;32m202\033[0m"},
		{"status=204", "status=\033[1;32m204\033[0m"},
		{"status=206", "status=\033[1;32m206\033[0m"},
		{"status 200", "status \033[1;32m200\033[0m"},
		{"status 201", "status \033[1;32m201\033[0m"},
		{"status 202", "status \033[1;32m202\033[0m"},
		{"status 204", "status \033[1;32m204\033[0m"},
		{"status 206", "status \033[1;32m206\033[0m"},
		{"status_code=200", "status_code=\033[1;32m200\033[0m"},
		{"status_code=201", "status_code=\033[1;32m201\033[0m"},
		{"status_code=202", "status_code=\033[1;32m202\033[0m"},
		{"status_code=204", "status_code=\033[1;32m204\033[0m"},
		{"status_code=206", "status_code=\033[1;32m206\033[0m"},
		{"status_code 200", "status_code \033[1;32m200\033[0m"},
		{"status_code 201", "status_code \033[1;32m201\033[0m"},
		{"status_code 202", "status_code \033[1;32m202\033[0m"},
		{"status_code 204", "status_code \033[1;32m204\033[0m"},
		{"status_code 206", "status_code \033[1;32m206\033[0m"},

		// 3xx Redirection - Bold yellow foreground
		{"status=301", "status=\033[1;33m301\033[0m"},
		{"status=302", "status=\033[1;33m302\033[0m"},
		{"status=303", "status=\033[1;33m303\033[0m"},
		{"status=304", "status=\033[1;33m304\033[0m"},
		{"status=307", "status=\033[1;33m307\033[0m"},
		{"status=308", "status=\033[1;33m308\033[0m"},
		{"status 301", "status \033[1;33m301\033[0m"},
		{"status 302", "status \033[1;33m302\033[0m"},
		{"status 303", "status \033[1;33m303\033[0m"},
		{"status 304", "status \033[1;33m304\033[0m"},
		{"status 307", "status \033[1;33m307\033[0m"},
		{"status 308", "status \033[1;33m308\033[0m"},
		{"status_code=301", "status_code=\033[1;33m301\033[0m"},
		{"status_code=302", "status_code=\033[1;33m302\033[0m"},
		{"status_code=303", "status_code=\033[1;33m303\033[0m"},
		{"status_code=304", "status_code=\033[1;33m304\033[0m"},
		{"status_code=307", "status_code=\033[1;33m307\033[0m"},
		{"status_code=308", "status_code=\033[1;33m308\033[0m"},
		{"status_code 301", "status_code \033[1;33m301\033[0m"},
		{"status_code 302", "status_code \033[1;33m302\033[0m"},
		{"status_code 303", "status_code \033[1;33m303\033[0m"},
		{"status_code 304", "status_code \033[1;33m304\033[0m"},
		{"status_code 307", "status_code \033[1;33m307\033[0m"},
		{"status_code 308", "status_code \033[1;33m308\033[0m"},

		// 4xx Client Error - Bold red foreground
		{"status=400", "status=\033[1;31m400\033[0m"},
		{"status=401", "status=\033[1;31m401\033[0m"},
		{"status=403", "status=\033[1;31m403\033[0m"},
		{"status=404", "status=\033[1;31m404\033[0m"},
		{"status=405", "status=\033[1;31m405\033[0m"},
		{"status=409", "status=\033[1;31m409\033[0m"},
		{"status=422", "status=\033[1;31m422\033[0m"},
		{"status=429", "status=\033[1;31m429\033[0m"},
		{"status 400", "status \033[1;31m400\033[0m"},
		{"status 401", "status \033[1;31m401\033[0m"},
		{"status 403", "status \033[1;31m403\033[0m"},
		{"status 404", "status \033[1;31m404\033[0m"},
		{"status 405", "status \033[1;31m405\033[0m"},
		{"status 409", "status \033[1;31m409\033[0m"},
		{"status 422", "status \033[1;31m422\033[0m"},
		{"status 429", "status \033[1;31m429\033[0m"},
		{"status_code=400", "status_code=\033[1;31m400\033[0m"},
		{"status_code=401", "status_code=\033[1;31m401\033[0m"},
		{"status_code=403", "status_code=\033[1;31m403\033[0m"},
		{"status_code=404", "status_code=\033[1;31m404\033[0m"},
		{"status_code=405", "status_code=\033[1;31m405\033[0m"},
		{"status_code=409", "status_code=\033[1;31m409\033[0m"},
		{"status_code=422", "status_code=\033[1;31m422\033[0m"},
		{"status_code=429", "status_code=\033[1;31m429\033[0m"},
		{"status_code 400", "status_code \033[1;31m400\033[0m"},
		{"status_code 401", "status_code \033[1;31m401\033[0m"},
		{"status_code 403", "status_code \033[1;31m403\033[0m"},
		{"status_code 404", "status_code \033[1;31m404\033[0m"},
		{"status_code 405", "status_code \033[1;31m405\033[0m"},
		{"status_code 409", "status_code \033[1;31m409\033[0m"},
		{"status_code 422", "status_code \033[1;31m422\033[0m"},
		{"status_code 429", "status_code \033[1;31m429\033[0m"},

		// 5xx Server Error - Bold red foreground
		{"status=500", "status=\033[1;31m500\033[0m"},
		{"status=502", "status=\033[1;31m502\033[0m"},
		{"status=503", "status=\033[1;31m503\033[0m"},
		{"status=504", "status=\033[1;31m504\033[0m"},
		{"status=505", "status=\033[1;31m505\033[0m"},
		{"status 500", "status \033[1;31m500\033[0m"},
		{"status 502", "status \033[1;31m502\033[0m"},
		{"status 503", "status \033[1;31m503\033[0m"},
		{"status 504", "status \033[1;31m504\033[0m"},
		{"status 505", "status \033[1;31m505\033[0m"},
		{"status_code=500", "status_code=\033[1;31m500\033[0m"},
		{"status_code=502", "status_code=\033[1;31m502\033[0m"},
		{"status_code=503", "status_code=\033[1;31m503\033[0m"},
		{"status_code=504", "status_code=\033[1;31m504\033[0m"},
		{"status_code=505", "status_code=\033[1;31m505\033[0m"},
		{"status_code 500", "status_code \033[1;31m500\033[0m"},
		{"status_code 502", "status_code \033[1;31m502\033[0m"},
		{"status_code 503", "status_code \033[1;31m503\033[0m"},
		{"status_code 504", "status_code \033[1;31m504\033[0m"},
		{"status_code 505", "status_code \033[1;31m505\033[0m"},
	}

	for _, pattern := range statusPatterns {
		line = strings.ReplaceAll(line, pattern.pattern, pattern.replacement)
	}

	return line
}

// highlightHTTPMethod colors HTTP methods with background colors
func (f *ColoredFormatter) highlightHTTPMethod(line string) string {
	methods := map[string]string{
		"GET":     "\033[42mGET\033[0m",     // Green background
		"POST":    "\033[44mPOST\033[0m",    // Blue background
		"PUT":     "\033[43mPUT\033[0m",     // Yellow background
		"PATCH":   "\033[45mPATCH\033[0m",   // Magenta background
		"DELETE":  "\033[41mDELETE\033[0m",  // Red background
		"HEAD":    "\033[46mHEAD\033[0m",    // Cyan background
		"OPTIONS": "\033[47mOPTIONS\033[0m", // White background
	}

	for method, colored := range methods {
		line = strings.ReplaceAll(line, method, colored)
	}

	return line
}

// highlightDuration colors response duration based on performance thresholds
func (f *ColoredFormatter) highlightDuration(line string) string {
	// Simple approach: just highlight the duration_ms field name
	// For more sophisticated coloring based on actual values,
	// we would need to parse the JSON structure properly
	line = strings.ReplaceAll(line, "duration_ms", "\033[36mduration_ms\033[0m")
	return line
}

// highlightErrors highlights error-related content with background colors
func (f *ColoredFormatter) highlightErrors(line string) string {
	// Highlight error level with background colors
	line = strings.ReplaceAll(line, "level=error", "level=\033[41merror\033[0m") // Red background
	line = strings.ReplaceAll(line, "level=warn", "level=\033[43mwarn\033[0m")   // Yellow background

	// Highlight common error patterns with background colors
	line = strings.ReplaceAll(line, "error", "\033[41merror\033[0m") // Red background

	return line
}

// highlightServices highlights service names for easy identification with background colors
func (f *ColoredFormatter) highlightServices(line string) string {
	services := map[string]string{
		"api-gateway":  "\033[45mapi-gateway\033[0m",  // Magenta background
		"user-service": "\033[46muser-service\033[0m", // Cyan background
		"auth-service": "\033[44mauth-service\033[0m", // Blue background
	}

	for service, colored := range services {
		line = strings.ReplaceAll(line, service, colored)
	}

	return line
}

// highlightRequestIDs highlights request IDs
func (f *ColoredFormatter) highlightRequestIDs(line string) string {
	// Look for UUID patterns in request_id field
	// This is a simple heuristic - could be improved
	if strings.Contains(line, "request_id=") {
		// For now, just highlight the field name
		line = strings.ReplaceAll(line, "request_id", "\033[36mrequest_id\033[0m")
	}

	return line
}

// NewColoredFormatter creates a new colored formatter
func NewColoredFormatter() *ColoredFormatter {
	return &ColoredFormatter{
		TextFormatter: logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02T15:04:05Z07:00",
			ForceColors:     true,
			DisableColors:   false,
		},
	}
}
