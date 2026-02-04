package repository

import (
	"context"
	"fmt"
	"time"
)

// DBInterface defines the minimal database operations needed for testing (user-service pattern)
type DBInterface interface {
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) Row
	Exec(ctx context.Context, sql string, args ...any) (CommandTag, error)
}

// CommandTag represents the result of an Exec operation
// TODO: - this is a different pattern compared with users-service: CommandTag in user-service
// is a 'real' pgconn.CommandTag. There were issues when pgconn was imported here - need to resolve
// and make it a ptoper type instead of generic interface
type CommandTag interface{}

// Row represents a single row returned from QueryRow
type Row interface {
	Scan(dest ...any) error
}

// Rows represents multiple rows returned from Query
type Rows interface {
	Close()
	Next() bool
	Scan(dest ...any) error
	Err() error
}

// RepositoryOptions for configuring repository behavior
type RepositoryOptions struct {
	// Query options
	DefaultPageSize  int
	MaxPageSize      int
	EnableSoftDelete bool
	EnableVersioning bool

	// Performance options
	EnableQueryCache        bool
	EnableConnectionPooling bool
	QueryTimeout            time.Duration

	// Tracing options
	EnableTracing     bool
	TracingSampleRate float64

	// Logging options
	LogSlowQueries     bool
	SlowQueryThreshold time.Duration
}

// DefaultRepositoryOptions returns sensible defaults
func DefaultRepositoryOptions() *RepositoryOptions {
	return &RepositoryOptions{
		DefaultPageSize:         50,
		MaxPageSize:             1000,
		EnableSoftDelete:        true,
		EnableVersioning:        true,
		EnableQueryCache:        false,
		EnableConnectionPooling: true,
		QueryTimeout:            30 * time.Second,
		EnableTracing:           true,
		TracingSampleRate:       0.1,
		LogSlowQueries:          true,
		SlowQueryThreshold:      100 * time.Millisecond,
	}
}

// RepositoryMetrics for performance monitoring
type RepositoryMetrics struct {
	// Query metrics
	QueryCount     int64
	SlowQueryCount int64
	ErrorCount     int64

	// Timing metrics
	TotalQueryTime   time.Duration
	AverageQueryTime time.Duration

	// Cache metrics (if enabled)
	CacheHitCount  int64
	CacheMissCount int64

	// Last reset time
	LastResetAt time.Time
}

// Reset resets all metrics
func (m *RepositoryMetrics) Reset() {
	m.QueryCount = 0
	m.SlowQueryCount = 0
	m.ErrorCount = 0
	m.TotalQueryTime = 0
	m.AverageQueryTime = 0
	m.CacheHitCount = 0
	m.CacheMissCount = 0
	m.LastResetAt = time.Now()
}

// UpdateAverageQueryTime recalculates average query time
func (m *RepositoryMetrics) UpdateAverageQueryTime() {
	if m.QueryCount > 0 {
		m.AverageQueryTime = time.Duration(int64(m.TotalQueryTime) / m.QueryCount)
	}
}

// Repository base interface
type Repository interface {
	// Database access
	DB() DBInterface
	Options() *RepositoryOptions

	// Metrics
	Metrics() *RepositoryMetrics
	ResetMetrics()

	// Health check
	Healthy(ctx context.Context) error
}

// Error types for repository operations
var (
	ErrNotFound        = fmt.Errorf("resource not found")
	ErrAlreadyExists   = fmt.Errorf("resource already exists")
	ErrInvalidInput    = fmt.Errorf("invalid input")
	ErrOptimisticLock  = fmt.Errorf("optimistic lock failed")
	ErrVersionConflict = fmt.Errorf("version conflict")
	ErrSoftDeleted     = fmt.Errorf("resource is soft deleted")
	ErrLimitExceeded   = fmt.Errorf("query limit exceeded")
	ErrTimeout         = fmt.Errorf("query timeout")
)
