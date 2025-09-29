package database

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// DBOperation represents a database operation type
type DBOperation string

const (
	DBOpSelect DBOperation = "SELECT"
	DBOpInsert DBOperation = "INSERT"
	DBOpUpdate DBOperation = "UPDATE"
	DBOpDelete DBOperation = "DELETE"
)

// TraceDBOperation creates a span for database operations and handles error recording
func TraceDBOperation(ctx context.Context, operation DBOperation, table string, query string, fn func(ctx context.Context) error) error {
	tracer := otel.Tracer("database")

	// Create span with database operation attributes
	ctx, span := tracer.Start(ctx, fmt.Sprintf("db.%s", operation),
		trace.WithAttributes(
			semconv.DBSystemPostgreSQL,
			semconv.DBOperationKey.String(string(operation)),
			semconv.DBNameKey.String("service_db"),
			attribute.String("db.table", table),
			attribute.String("db.statement", query),
		))
	defer span.End()

	// Execute the database operation
	err := fn(ctx)

	// Record error if occurred
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "Database operation completed successfully")
	}

	return err
}

// TraceDBQuery creates a span specifically for SELECT queries with result count
func TraceDBQuery(ctx context.Context, table string, query string, fn func(ctx context.Context) error) error {
	return TraceDBOperation(ctx, DBOpSelect, table, query, fn)
}

// TraceDBInsert creates a span specifically for INSERT operations
func TraceDBInsert(ctx context.Context, table string, query string, fn func(ctx context.Context) error) error {
	return TraceDBOperation(ctx, DBOpInsert, table, query, fn)
}

// TraceDBUpdate creates a span specifically for UPDATE operations
func TraceDBUpdate(ctx context.Context, table string, query string, fn func(ctx context.Context) error) error {
	return TraceDBOperation(ctx, DBOpUpdate, table, query, fn)
}

// TraceDBDelete creates a span specifically for DELETE operations
func TraceDBDelete(ctx context.Context, table string, query string, fn func(ctx context.Context) error) error {
	return TraceDBOperation(ctx, DBOpDelete, table, query, fn)
}
