# Tracing Decorator Pattern for Repository Layer

## Overview

This document explains how to add OpenTelemetry tracing back to the repository layer using a decorator pattern that keeps repos and mocks simple while maintaining production observability.

## Problem: What We Avoided

By removing tracing from repositories, we solved several issues:
- **Simplified mocking**: Mocks don't need to handle `tracer` field
- **Cleaner interfaces**: `DBInterface` has only 3 methods instead of 6+
- **Easier tests**: No need to mock `tracer.Start()`, `span.End()` calls
- **Type safety**: No type assertion or `interface{}` casting

**But we lost:**
- Production observability for repository operations
- Query performance tracking
- Distributed tracing across service boundaries

## Solution: Decorator Pattern

The decorator pattern adds tracing at the interface boundary, keeping implementations simple.

## Implementation

### Step 1: Add Tracing to DBInterface

```go
// In internal/repository/interfaces.go, add to DBInterface:

// DBInterface defines minimal database operations needed for testing (user-service pattern)
type DBInterface interface {
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) Row
	Exec(ctx context.Context, sql string, args ...any) (CommandTag, error)
	// NEW: Optional tracing method for decorator pattern
	TraceQuery(ctx context.Context, name string, fn func() error) error
}
```

### Step 2: Create Tracing Decorator

```go
// In internal/repository/tracing_decorator.go:

package repository

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TracingConfig configures tracing behavior
type TracingConfig struct {
	Enabled        bool
	SampleRate     float64
	ServiceName    string
	TracerProvider *trace.Tracer
}

// DefaultTracingConfig returns sensible defaults
func DefaultTracingConfig() *TracingConfig {
	return &TracingConfig{
		Enabled:     true,
		SampleRate:  0.1, // 10% sampling
		ServiceName: "objects-service",
	}
}

// DBTraceWrapper wraps DBInterface with tracing
type DBTraceWrapper struct {
	db     DBInterface
	config *TracingConfig
	tracer trace.Tracer
}

// NewDBTraceWrapper creates a new traced database wrapper
func NewDBTraceWrapper(db DBInterface, config *TracingConfig) *DBTraceWrapper {
	return &DBTraceWrapper{
		db:     db,
		config: config,
		tracer: otel.Tracer(config.ServiceName),
	}
}

// Query implements DBInterface with tracing
func (w *DBTraceWrapper) Query(ctx context.Context, sql string, args ...any) (Rows, error) {
	if !w.config.Enabled {
		return w.db.Query(ctx, sql, args...)
	}

	ctx, span := w.tracer.Start(ctx, fmt.Sprintf("DB.Query: %s", sql))
	defer span.End()

	start := time.Now()
	result, err := w.db.Query(ctx, sql, args...)
	duration := time.Since(start)

	// Record metrics if query was slow
	if duration > 100*time.Millisecond {
		span.SetAttributes(attribute.String("slow.query", "true"))
	}

	return result, err
}

// QueryRow implements DBInterface with tracing
func (w *DBTraceWrapper) QueryRow(ctx context.Context, sql string, args ...any) Row {
	if !w.config.Enabled {
		return w.db.QueryRow(ctx, sql, args...)
	}

	ctx, span := w.tracer.Start(ctx, fmt.Sprintf("DB.QueryRow: %s", sql))
	defer span.End()

	start := time.Now()
	result := w.db.QueryRow(ctx, sql, args...)
	duration := time.Since(start)

	if duration > 100*time.Millisecond {
		span.SetAttributes(attribute.String("slow.query", "true"))
	}

	return result
}

// Exec implements DBInterface with tracing
func (w *DBTraceWrapper) Exec(ctx context.Context, sql string, args ...any) (CommandTag, error) {
	if !w.config.Enabled {
		return w.db.Exec(ctx, sql, args...)
	}

	ctx, span := w.tracer.Start(ctx, fmt.Sprintf("DB.Exec: %s", sql))
	defer span.End()

	start := time.Now()
	result, err := w.db.Exec(ctx, sql, args...)
	duration := time.Since(start)

	if err != nil {
		span.SetStatus(codes.Error)
		span.RecordError(err)
	}

	return result, err
}

// TraceQuery provides explicit tracing for complex queries
func (w *DBTraceWrapper) TraceQuery(ctx context.Context, name string, fn func() error) error {
	if !w.config.Enabled {
		return fn()
	}

	ctx, span := w.tracer.Start(ctx, name)
	defer span.End()

	start := time.Now()
	err := fn()
	duration := time.Since(start)

	if err != nil {
		span.SetStatus(codes.Error)
		span.RecordError(err)
	}

	span.SetAttributes(
		attribute.String("db.duration", duration.String()),
		attribute.String("db.slow", fmt.Sprintf("%v", duration > 100*time.Millisecond)),
	)

	return err
}
```

### Step 3: Update Repositories to Use Traced DBInterface

```go
// In internal/repository/object_type_repository.go:

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/models"
)

type objectTypeRepository struct {
	db      DBInterface // Changed from Database to DBInterface
	options *RepositoryOptions
	metrics *RepositoryMetrics
	// No tracer field needed anymore
}

func NewObjectTypeRepository(db DBInterface, options *RepositoryOptions) ObjectTypeRepository {
	if options == nil {
		options = DefaultRepositoryOptions()
	}

	return &objectTypeRepository{
		db:      db,
		options: options,
		metrics: &RepositoryMetrics{LastResetAt: time.Now()},
	}
}

// Example of using traced queries in repository methods:

func (r *objectTypeRepository) GetByID(ctx context.Context, id int64) (*models.ObjectType, error) {
	// For one-off queries, tracing happens at DB interface level
	// No need for manual span management
	query := `SELECT id, name, parent_type_id ... FROM object_types WHERE id = $1`
	
	row := r.db.QueryRow(ctx, query, id)
	// Tracing is handled by DBTraceWrapper automatically
}

// For complex operations, use explicit tracing:
func (r *objectTypeRepository) GetTree(ctx context.Context, rootID *int64) ([]*models.ObjectType, error) {
	// Use explicit tracing for complex operations
	if wrapper, ok := r.db.(*DBTraceWrapper); ok && r.options.EnableTracing {
		return wrapper.TraceQuery(ctx, "object_type_repository.GetTree", func() error {
		// Tracing is managed by the wrapper
		query := `WITH RECURSIVE ...`
		rows, err := r.db.Query(ctx, query, rootID)
		// Process rows and build hierarchy
		 return nil // or return hierarchy
		})
	}
	
	// When not using traced wrapper, execute normally
	query := `...`
	return r.db.Query(ctx, query, rootID)
}
```

### Step 4: Update main.go to Use Traced Database

```go
// In cmd/main.go:

import (
	// ... other imports ...
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/repository"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/repository/tracing_decorator"
)

if db != nil {
	// Wrap database with tracing
	tracingConfig := &repository.TracingConfig{
		Enabled:     cfg.Tracing.Enabled,
		SampleRate: cfg.Tracing.SampleRate,
		ServiceName: "objects-service-repository",
	}
	
	pgDatabase := repository.NewPGDatabase(db.GetPool())
	dbTraced := repository.NewDBTraceWrapper(pgDatabase, tracingConfig)
	
	// Use traced database in repositories
	objectTypeRepo := repository.NewObjectTypeRepository(dbTraced, repoOptions)
	objectRepo := repository.NewObjectRepository(dbTraced, repoOptions)
}
```

### Step 5: Update Tests (No Changes Needed!)

```go
// In internal/repository/repository_test.go:

// MockDBPool remains exactly the same - no tracing needed!
type MockDBPool struct {
	QueryFunc    func(ctx context.Context, sql string, args ...any) (Rows, error)
	QueryRowFunc func(ctx context.Context, sql string, args ...any) Row
	ExecFunc     func(ctx context.Context, sql string, args ...any) (CommandTag, error)
}

// Tests work the same - no need to mock tracing
func TestObjectTypeRepository_Create(t *testing.T) {
	mockDB := &MockDBPool{
		QueryRowFunc: func(ctx context.Context, sql string, args ...any) Row {
			return &MockRow{}
		},
	}

	repo := NewObjectTypeRepository(mockDB, DefaultRepositoryOptions())
	// Test logic unchanged
}
```

## Benefits of This Approach

### For Production Code:
- **Automatic Tracing**: Every DB operation is traced without code changes
- **Configurable**: Can enable/disable via environment
- **Sampled**: Reduces overhead with sampling
- **Structured**: Consistent span naming and attributes

### For Testing:
- **Simple Mocks**: No `tracer` field to mock
- **Fast Tests**: No span lifecycle to manage
- **Type Safe**: Clean interface with minimal surface area

### For Developers:
- **Separation of Concerns**: DB logic vs observability
- **Easy Toggle**: Can disable tracing in test mode
- **Clear Ownership**: Repository owns business logic, decorator owns tracing

## Usage Examples

### Simple Query (Automatic Tracing):
```go
// Tracing happens automatically at DB interface level
objectType, err := objectTypeRepo.GetByID(ctx, id)
// Span: "DB.QueryRow: object_type_repository.GetByID" (auto-generated from query)
// Attributes: db.duration, db.slow (if > 100ms)
```

### Complex Operation (Explicit Tracing):
```go
// Use explicit tracing for multi-step operations
return wrapper.TraceQuery(ctx, "object_type_repository.GetTree", func() error {
	// Span: "object_type_repository.GetTree"
	// Can add custom attributes as needed
	// Entire operation is traced as one unit
})
```

### Conditional Tracing:
```go
// Disable tracing in tests
repo := NewObjectTypeRepository(mockDB, &RepositoryOptions{
	EnableTracing: false, // Explicitly disable for tests
})

// Environment-based
tracingConfig := &TracingConfig{
	Enabled: os.Getenv("ENABLE_TRACING") == "true",
	SampleRate: 0.1,
	ServiceName: "objects-service-repository",
}
```

## Migration Path

**Current State**: No tracing in repositories (clean, simple, testable)  
**Future State**: Add decorator wrapper with automatic + explicit tracing patterns

**When to Add**: 
- After service/handler layers are implemented and validated
- Before production deployment
- When observability requirements become clear
- When performance optimization is needed

**Priority**: Medium (not blocking current work)

**Estimated Effort**: 4-6 hours for full implementation including tests

---

## Summary

The decorator pattern gives us the best of both worlds:
- Clean, simple repositories (current state)
- Easy to test (current state)
- Production-ready tracing when needed (future state)

This approach defers the complexity while documenting a clear migration path for when observability becomes important.
