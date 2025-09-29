# Database Tracing Instrumentation

This document explains how to use the database tracing instrumentation to monitor database operations and measure performance.

## Overview

The database tracing system uses OpenTelemetry to create spans for database operations, allowing you to:

- Measure the time spent in database calls
- Track database operation types (SELECT, INSERT, UPDATE, DELETE)
- Monitor database performance bottlenecks
- Correlate database operations with request traces

## Architecture

### Tracing Helper Functions

The `common/database/tracing.go` file provides helper functions for instrumenting database operations:

```go
// Trace database operations with automatic span creation
database.TraceDBQuery(ctx, "table_name", "SELECT ...", func(ctx context.Context) error {
    // Your database operation here
    return db.QueryRow(ctx, query, args...).Scan(&result)
})

database.TraceDBInsert(ctx, "table_name", "INSERT ...", func(ctx context.Context) error {
    // Your insert operation
})

database.TraceDBUpdate(ctx, "table_name", "UPDATE ...", func(ctx context.Context) error {
    // Your update operation
})

database.TraceDBDelete(ctx, "table_name", "DELETE ...", func(ctx context.Context) error {
    // Your delete operation
})
```

### Span Attributes

Each database span includes the following attributes:

- `db.system`: "postgresql"
- `db.operation`: "SELECT", "INSERT", "UPDATE", "DELETE", or "TRANSACTION"
- `db.name`: Database name (typically "service_db")
- `db.table`: Table name being operated on
- `db.statement`: The SQL query being executed

## Usage Examples

### Instrumenting Repository Methods

Here's how the user repository methods are instrumented:

```go
func (r *UserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
    query := `INSERT INTO user_service.users ...`

    err := database.TraceDBInsert(ctx, "user_service.users", query, func(ctx context.Context) error {
        return r.db.QueryRow(ctx, query, user.Email, user.PasswordHash, user.FirstName, user.LastName).Scan(
            &user.ID, &user.CreatedAt, &user.UpdatedAt)
    })

    // Handle error and return
}
```

### Transaction Tracing

Transactions are automatically traced:

```go
err := db.WithTx(ctx, func(tx pgx.Tx) error {
    // Transaction operations are traced under "TRANSACTION" span
    return someOperation(tx)
})
```

## Viewing Traces

### Jaeger UI

When running with Jaeger tracing enabled, you can view traces at `http://localhost:16686`.

1. Start the services with tracing enabled (default in development)
2. Make API calls that trigger database operations
3. Open Jaeger UI and search for traces
4. Look for spans with names like `db.SELECT`, `db.INSERT`, etc.

### Trace Structure

A typical request trace might look like:

```
HTTP Request (from Gin middleware)
├── db.SELECT (user lookup)
├── db.INSERT (user creation)
└── db.TRANSACTION (if using transactions)
```

## Configuration

Tracing is configured in each service's `config.yaml`:

```yaml
tracing:
  enabled: true
  service_name: "user-service"
  collector_url: "http://jaeger:4318/v1/traces"
  sampling_rate: 1.0  # Sample all requests in development
```

## Performance Considerations

- Tracing adds minimal overhead but should be disabled in high-throughput production environments
- Use appropriate sampling rates (e.g., 0.1 for 10% sampling in production)
- Database spans include the full SQL query, which may contain sensitive data in logs

## Best Practices

1. **Always pass context**: Ensure `ctx context.Context` is passed through all database operations
2. **Use appropriate operation types**: Choose the correct tracing function (TraceDBQuery, TraceDBInsert, etc.)
3. **Include table names**: Use fully qualified table names (e.g., "user_service.users")
4. **Handle errors properly**: The tracing helpers automatically record errors on spans
5. **Test with tracing enabled**: Verify traces are created correctly in development

## Extending Tracing

To add tracing to new database operations:

1. Import the database tracing package
2. Wrap your database calls with the appropriate tracing function
3. Ensure context is properly passed through

Example for a custom repository:

```go
import "github.com/v-egorov/service-boilerplate/common/database"

func (r *MyRepository) CustomQuery(ctx context.Context, param string) error {
    query := "SELECT * FROM my_table WHERE param = $1"

    return database.TraceDBQuery(ctx, "my_table", query, func(ctx context.Context) error {
        rows, err := r.db.Query(ctx, query, param)
        // Process rows...
        return err
    })
}
```