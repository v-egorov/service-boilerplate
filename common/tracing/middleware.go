package tracing

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
)

// HTTPMiddleware creates a Gin middleware for tracing HTTP requests
func HTTPMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract trace context from incoming request headers
		ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// Get tracer for this service
		tracer := otel.Tracer(serviceName)

		// Create span name from HTTP method and path
		spanName := fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)

		// Start span
		ctx, span := tracer.Start(ctx, spanName)
		defer span.End()

		// Add HTTP attributes
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.user_agent", c.Request.UserAgent()),
			attribute.String("http.scheme", c.Request.URL.Scheme),
			attribute.String("http.host", c.Request.Host),
		)

		// Add custom attributes
		span.SetAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("http.route", c.Request.URL.Path),
		)

		// Store span in context for later use
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()

		// Add response attributes
		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
		)

		// Record errors if any
		if len(c.Errors) > 0 {
			span.SetStatus(codes.Error, c.Errors.String())
			for _, err := range c.Errors {
				span.RecordError(err)
			}
		} else {
			span.SetStatus(codes.Ok, "")
		}
	}
}
