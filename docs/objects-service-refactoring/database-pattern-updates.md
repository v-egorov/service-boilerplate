# Database Pattern Updates

This document summarizes the changes made to align objects-service refactoring phases with the existing codebase's pgx/v5 patterns.

## Overview

The codebase uses **pgx/v5** (not sqlx) as the PostgreSQL driver. All documentation has been updated to reflect this pattern.

## Key Changes

### 1. Repository Layer (Phase 3)

**Changed from sqlx to pgx/v5:**

```go
// ❌ OLD (sqlx)
import "github.com/jmoiron/sqlx"

type Repository struct {
    db *sqlx.DB
}

// ✅ NEW (pgx/v5)
import (
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/jackc/pgx/v5/pgconn"
)

type Repository struct {
    db     DBInterface  // For testability
    logger *logrus.Logger
}
```

**Added DBInterface pattern:**

```go
type DBInterface interface {
    Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
    QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
    Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
    Begin(ctx context.Context) (pgx.Tx, error)
}
```

**Two constructors for testability:**

```go
// Production constructor
func NewRepository(db *pgxpool.Pool, logger *logrus.Logger) *Repository

// Test constructor (accepts mock DBInterface)
func NewRepositoryWithInterface(db DBInterface, logger *logrus.Logger) *Repository
```

### 2. OpenTelemetry Tracing

**All database operations wrapped in tracing:**

```go
import "github.com/v-egorov/service-boilerplate/common/database"

func (r *Repository) Create(ctx context.Context, ...) error {
    query := `INSERT INTO ... VALUES ($1, $2) ...`
    
    return database.TraceDBInsert(ctx, "table_name", query, func(ctx context.Context) error {
        return r.db.QueryRow(ctx, query, ...).Scan(...)
    })
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*Model, error) {
    query := `SELECT ... FROM table WHERE id = $1`
    
    return database.TraceDBQuery(ctx, "table_name", query, func(ctx context.Context) error {
        return r.db.QueryRow(ctx, query, id).Scan(...)
    })
}

func (r *Repository) Update(ctx context.Context, ...) error {
    query := `UPDATE table SET ... WHERE id = $1`
    
    return database.TraceDBUpdate(ctx, "table_name", query, func(ctx context.Context) error {
        return r.db.QueryRow(ctx, query, ...).Scan(...)
    })
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
    query := `DELETE FROM table WHERE id = $1`
    
    return database.TraceDBDelete(ctx, "table_name", query, func(ctx context.Context) error {
        var result pgconn.CommandTag
        var execErr error
        result, execErr = r.db.Exec(ctx, query, id)
        return execErr
    })
}
```

### 3. Error Handling

**Using pgx.ErrNoRows instead of sql.ErrNoRows:**

```go
import (
    "errors"
    "github.com/jackc/pgx/v5"
)

func (r *Repository) GetByID(ctx context.Context, id int64) (*Model, error) {
    // ... query logic ...
    
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("failed to get: %w", err)
    }
    
    return &model, nil
}
```

**Constraint violation detection:**

```go
import (
    "errors"
    "github.com/jackc/pgx/v5/pgconn"
)

func isUniqueViolation(err error) bool {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        return pgErr.Code == "23505" // unique_violation
    }
    return false
}
```

### 4. Logging

**Structured logging with logrus:**

```go
import "github.com/sirupsen/logrus"

func (r *Repository) Create(ctx context.Context, ...) error {
    // ... create logic ...
    
    if err != nil {
        r.logger.WithError(err).Error("Failed to create model")
        return fmt.Errorf("failed to create: %w", err)
    }
    
    r.logger.WithField("id", model.ID).Info("Model created successfully")
    return nil
}
```

### 5. Transaction Support

**Using common/database.WithTx helper:**

```go
import "github.com/v-egorov/service-boilerplate/common/database"

func (r *Repository) CreateBatch(ctx context.Context, items []Model) ([]Model, error) {
    var created []Model
    
    err := database.WithTx(ctx, "table_name", r.db, func(tx pgx.Tx) error {
        for i := range items {
            query := `INSERT INTO table (...) VALUES (...) RETURNING id, ...`
            err := tx.QueryRow(ctx, query, ...).Scan(...)
            if err != nil {
                return err
            }
            created = append(created, item)
        }
        return nil
    })
    
    if err != nil {
        r.logger.WithError(err).Error("Failed to create batch")
        return nil, err
    }
    
    return created, nil
}
```

### 6. Service Layer (Phase 4)

**Added logging and proper error handling:**

```go
type Service struct {
    repo   RepositoryInterface
    logger *logrus.Logger
}

func NewService(repo RepositoryInterface, logger *logrus.Logger) *Service {
    return &Service{
        repo:   repo,
        logger: logger,
    }
}

// Two constructors for testability
func NewServiceWithInterface(repo RepositoryInterface, logger *logrus.Logger) *Service {
    return &Service{
        repo:   repo,
        logger: logger,
    }
}
```

### 7. Main Application (Phase 6)

**Using common/database.PostgresDB:**

```go
import (
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/sirupsen/logrus"
    
    "github.com/v-egorov/service-boilerplate/common/database"
)

func main() {
    cfg, err := config.Load()
    
    logger := logrus.New()
    logger.SetLevel(logrus.InfoLevel)
    if cfg.Environment == "development" {
        logger.SetLevel(logrus.DebugLevel)
    }
    
    db, err := database.NewPostgresDB(database.Config{
        Host:        cfg.DB.Host,
        Port:        cfg.DB.Port,
        User:        cfg.DB.User,
        Password:    cfg.DB.Password,
        Database:    cfg.DB.Database,
        SSLMode:     cfg.DB.SSLMode,
        MaxConns:    cfg.DB.MaxConns,
        MinConns:    cfg.DB.MinConns,
        MaxConnIdle: time.Hour,
        MaxConnLife: 24 * time.Hour,
    }, logger)
    
    pool := db.GetPool()
    repositories := initRepositories(pool, logger)
    services := initServices(repositories, logger)
    handlers := initHandlers(services)
    
    router := setupRouter(handlers, cfg)
    // ... rest of main
}
```

### 8. Testing (Phase 8)

**Mock implementations for pgx/v5:**

```go
import (
    "context"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/stretchr/testify/assert"
)

// MockDBPool implements DBInterface
type MockDBPool struct {
    QueryFunc    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
    QueryRowFunc func(ctx context.Context, sql string, args ...any) pgx.Row
    ExecFunc     func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
    BeginFunc    func(ctx context.Context) (pgx.Tx, error)
}

// Implement all DBInterface methods
func (m *MockDBPool) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
    if m.QueryFunc != nil {
        return m.QueryFunc(ctx, sql, args...)
    }
    return nil, nil
}

// ... implement other methods

// MockRow for testing QueryRow responses
type MockRow struct {
    scanFunc func(dest ...any) error
}

func (m *MockRow) Scan(dest ...any) error {
    if m.scanFunc != nil {
        return m.scanFunc(dest...)
    }
    return nil
}
```

## Updated Phase Files

All phase documentation files have been updated to use pgx/v5 patterns:

1. **phase-03-repositories.md** - Updated with pgx/v5, DBInterface, tracing, logging
2. **phase-04-services.md** - Updated with logger and proper error handling
3. **phase-06-main.md** - Updated with common/database.PostgresDB and pgxpool.Pool
4. **phase-08-tests.md** - Updated with pgx/v5 mock implementations

## Quick Reference

### Imports

```go
import (
    "github.com/jackc/pgx/v5"           // Core pgx
    "github.com/jackc/pgx/v5/pgxpool"   // Connection pooling
    "github.com/jackc/pgx/v5/pgconn"     // Command tags
    "github.com/jackc/pgx/v5/pgtype"     // Postgres types (if needed)
    "github.com/sirupsen/logrus"          // Logging
    "github.com/v-egorov/service-boilerplate/common/database"  // Tracing helpers
)
```

### Key Patterns

1. **Use `pgx.ErrNoRows`** for "not found" errors (not `sql.ErrNoRows`)
2. **Use `pgconn.CommandTag`** for Exec results
3. **Use `database.TraceDB*`** wrappers for all operations
4. **Use `database.WithTx()`** for transactions
5. **Use `pgxpool.Pool`** for connections (not `sqlx.DB`)
6. **Implement `DBInterface`** for testability
7. **Use `logrus`** for structured logging with fields

### Query Patterns

```go
// Query row with tracing
var model Model
err := database.TraceDBQuery(ctx, "table", query, func(ctx context.Context) error {
    return r.db.QueryRow(ctx, query, id).Scan(&model.ID, &model.Name, ...)
})

// Query multiple rows with tracing
var models []Model
err := database.TraceDBQuery(ctx, "table", query, func(ctx context.Context) error {
    rows, err := r.db.Query(ctx, query, args...)
    if err != nil {
        return err
    }
    defer rows.Close()
    
    for rows.Next() {
        var m Model
        err := rows.Scan(&m.ID, &m.Name, ...)
        if err != nil {
            return err
        }
        models = append(models, m)
    }
    return rows.Err()
})

// Exec with tracing and command tag
var result pgconn.CommandTag
err := database.TraceDBUpdate(ctx, "table", query, func(ctx context.Context) error {
    var execErr error
    result, execErr = r.db.Exec(ctx, query, args...)
    return execErr
})
if result.RowsAffected() == 0 {
    return ErrNotFound
}
```

## Next Steps

When implementing the phases:

1. **Phase 3 (Repositories)**: Use pgx/v5 patterns from this document
2. **Phase 4 (Services)**: Add logger to all services
3. **Phase 6 (Main)**: Use common/database.PostgresDB for connection
4. **Phase 8 (Tests)**: Create pgx/v5 mocks using patterns shown above

## Benefits

1. **Consistency**: Matches existing codebase patterns (user-service, auth-service)
2. **Observability**: All database operations are traced
3. **Testability**: DBInterface pattern allows easy mocking
4. **Performance**: pgx/v5 is faster than sqlx for PostgreSQL
5. **Logging**: Structured logging with logrus
6. **Connection Pooling**: Efficient connection management with pgxpool
