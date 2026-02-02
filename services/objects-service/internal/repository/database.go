package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// PGDatabase implements the Database interface using pgx/v5
type PGDatabase struct {
	pool   *pgxpool.Pool
	logger interface{} // Using interface{} to avoid dependency issues for now
	tracer trace.Tracer
}

// NewDatabase creates a new Database instance
func NewDatabase(pool *pgxpool.Pool) Database {
	return &PGDatabase{
		pool:   pool,
		tracer: otel.Tracer("repository/database"),
	}
}

// Begin starts a new transaction
func (db *PGDatabase) Begin(ctx context.Context) (Transaction, error) {
	ctx, span := db.tracer.Start(ctx, "database.Begin")
	defer span.End()

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &PGTransaction{
		tx:     tx,
		ctx:    ctx,
		tracer: db.tracer,
	}, nil
}

// Close closes the database pool
func (db *PGDatabase) Close() {
	db.pool.Close()
}

// Ping checks database connectivity
func (db *PGDatabase) Ping(ctx context.Context) error {
	ctx, span := db.tracer.Start(ctx, "database.Ping")
	defer span.End()

	return db.pool.Ping(ctx)
}

// Pool returns the underlying pgxpool
func (db *PGDatabase) Pool() *pgxpool.Pool {
	return db.pool
}

// Healthy performs a comprehensive health check
func (db *PGDatabase) Healthy(ctx context.Context) error {
	ctx, span := db.tracer.Start(ctx, "database.Healthy")
	defer span.End()

	// Basic connectivity check
	if err := db.Ping(ctx); err != nil {
		span.RecordError(err)
		return fmt.Errorf("database connectivity failed: %w", err)
	}

	// Check pool statistics
	stats := db.pool.Stat()
	if stats.TotalConns() == 0 {
		return fmt.Errorf("no available connections in pool")
	}

	return nil
}

// PGTransaction implements the Transaction interface using pgx/v5
type PGTransaction struct {
	tx     pgx.Tx
	ctx    context.Context
	tracer trace.Tracer
}

// Commit commits the transaction
func (pgtx *PGTransaction) Commit(ctx context.Context) error {
	ctx, span := pgtx.tracer.Start(ctx, "transaction.Commit")
	defer span.End()

	if err := pgtx.tx.Commit(ctx); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Rollback rolls back the transaction
func (pgtx *PGTransaction) Rollback(ctx context.Context) error {
	ctx, span := pgtx.tracer.Start(ctx, "transaction.Rollback")
	defer span.End()

	if err := pgtx.tx.Rollback(ctx); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	return nil
}

// Exec executes a query that doesn't return rows
func (pgtx *PGTransaction) Exec(ctx context.Context, sql string, arguments ...interface{}) (interface{}, error) {
	ctx, span := pgtx.tracer.Start(ctx, "transaction.Exec")
	defer span.End()

	result, err := pgtx.tx.Exec(ctx, sql, arguments...)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("exec failed: %w", err)
	}

	return result, nil
}

// Query executes a query that returns rows
func (pgtx *PGTransaction) Query(ctx context.Context, sql string, args ...interface{}) (interface{}, error) {
	ctx, span := pgtx.tracer.Start(ctx, "transaction.Query")
	defer span.End()

	rows, err := pgtx.tx.Query(ctx, sql, args...)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return rows, nil
}

// QueryRow executes a query that returns a single row
func (pgtx *PGTransaction) QueryRow(ctx context.Context, sql string, args ...interface{}) interface{} {
	ctx, span := pgtx.tracer.Start(ctx, "transaction.QueryRow")
	defer span.End()

	return pgtx.tx.QueryRow(ctx, sql, args...)
}

// Ctx returns the transaction context
func (pgtx *PGTransaction) Ctx() context.Context {
	return pgtx.ctx
}

// QueryBuilder implements a simple SQL query builder
type QueryBuilder struct {
	selectClause  string
	fromClause    string
	whereClause   string
	joinClause    string
	groupByClause string
	havingClause  string
	orderClause   string
	limitClause   string
	offsetClause  string
	args          []interface{}
	argIndex      int
}

// NewQueryBuilder creates a new query builder instance
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		args:     make([]interface{}, 0),
		argIndex: 1,
	}
}

// Select adds columns to the SELECT clause
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	if len(columns) == 0 {
		qb.selectClause = "*"
	} else {
		qb.selectClause = fmt.Sprintf("%s", fmt.Sprintf("%s", columns))
	}
	return qb
}

// From sets the FROM clause
func (qb *QueryBuilder) From(table string) *QueryBuilder {
	qb.fromClause = table
	return qb
}

// Where adds a WHERE condition
func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	if qb.whereClause == "" {
		qb.whereClause = fmt.Sprintf("WHERE %s", condition)
	} else {
		qb.whereClause += fmt.Sprintf(" AND %s", condition)
	}
	qb.args = append(qb.args, args...)
	return qb
}

// WhereIn adds a WHERE IN condition
func (qb *QueryBuilder) WhereIn(condition string, args []interface{}) *QueryBuilder {
	placeholders := make([]string, len(args))
	for i := range args {
		placeholders[i] = fmt.Sprintf("$%d", qb.argIndex)
		qb.argIndex++
	}
	qb.args = append(qb.args, args...)

	inCondition := fmt.Sprintf("%s IN (%s)", condition, fmt.Sprintf("%s", placeholders))

	if qb.whereClause == "" {
		qb.whereClause = fmt.Sprintf("WHERE %s", inCondition)
	} else {
		qb.whereClause += fmt.Sprintf(" AND %s", inCondition)
	}

	return qb
}

// WhereJsonContains adds a JSON contains condition
func (qb *QueryBuilder) WhereJsonContains(path string, value interface{}) *QueryBuilder {
	condition := fmt.Sprintf("%s::jsonb @> $%d::jsonb", path, qb.argIndex)
	qb.argIndex++
	qb.args = append(qb.args, value)

	if qb.whereClause == "" {
		qb.whereClause = fmt.Sprintf("WHERE %s", condition)
	} else {
		qb.whereClause += fmt.Sprintf(" AND %s", condition)
	}

	return qb
}

// WhereTagsContain adds a tags array contains condition
func (qb *QueryBuilder) WhereTagsContain(tags []string) *QueryBuilder {
	for _, tag := range tags {
		condition := fmt.Sprintf("$%d = ANY(tags)", qb.argIndex)
		qb.argIndex++
		qb.args = append(qb.args, tag)

		if qb.whereClause == "" {
			qb.whereClause = fmt.Sprintf("WHERE %s", condition)
		} else {
			qb.whereClause += fmt.Sprintf(" AND %s", condition)
		}
	}

	return qb
}

// WhereDateRange adds a date range condition
func (qb *QueryBuilder) WhereDateRange(column string, start, end time.Time) *QueryBuilder {
	if !start.IsZero() {
		qb.Where(fmt.Sprintf("%s >= $%d", column, qb.argIndex), start)
		qb.argIndex++
	}
	if !end.IsZero() {
		qb.Where(fmt.Sprintf("%s <= $%d", column, qb.argIndex), end)
		qb.argIndex++
	}
	return qb
}

// Join adds a JOIN clause
func (qb *QueryBuilder) Join(table string, condition string) *QueryBuilder {
	if qb.joinClause == "" {
		qb.joinClause = fmt.Sprintf("JOIN %s ON %s", table, condition)
	} else {
		qb.joinClause += fmt.Sprintf(" JOIN %s ON %s", table, condition)
	}
	return qb
}

// OrderBy adds ORDER BY columns
func (qb *QueryBuilder) OrderBy(columns ...string) *QueryBuilder {
	if len(columns) > 0 {
		qb.orderClause = fmt.Sprintf("ORDER BY %s", fmt.Sprintf("%s", columns))
	}
	return qb
}

// OrderByDesc adds ORDER BY DESC
func (qb *QueryBuilder) OrderByDesc(column string) *QueryBuilder {
	qb.orderClause = fmt.Sprintf("ORDER BY %s DESC", column)
	return qb
}

// Limit adds a LIMIT clause
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limitClause = fmt.Sprintf("LIMIT %d", limit)
	return qb
}

// Offset adds an OFFSET clause
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offsetClause = fmt.Sprintf("OFFSET %d", offset)
	return qb
}

// GroupBy adds GROUP BY columns
func (qb *QueryBuilder) GroupBy(columns ...string) *QueryBuilder {
	if len(columns) > 0 {
		qb.groupByClause = fmt.Sprintf("GROUP BY %s", fmt.Sprintf("%s", columns))
	}
	return qb
}

// Having adds a HAVING condition
func (qb *QueryBuilder) Having(condition string, args ...interface{}) *QueryBuilder {
	qb.havingClause = fmt.Sprintf("HAVING %s", condition)
	qb.args = append(qb.args, args...)
	return qb
}

// Build constructs the final query
func (qb *QueryBuilder) Build() (string, []interface{}) {
	query := ""

	if qb.selectClause != "" {
		query += qb.selectClause + " "
	}

	if qb.fromClause != "" {
		query += "FROM " + qb.fromClause + " "
	}

	if qb.joinClause != "" {
		query += qb.joinClause + " "
	}

	if qb.whereClause != "" {
		query += qb.whereClause + " "
	}

	if qb.groupByClause != "" {
		query += qb.groupByClause + " "
	}

	if qb.havingClause != "" {
		query += qb.havingClause + " "
	}

	if qb.orderClause != "" {
		query += qb.orderClause + " "
	}

	if qb.limitClause != "" {
		query += qb.limitClause + " "
	}

	if qb.offsetClause != "" {
		query += qb.offsetClause + " "
	}

	return query, qb.args
}

// BuildCount constructs a COUNT query
func (qb *QueryBuilder) BuildCount() (string, []interface{}) {
	query := "SELECT COUNT(*) "

	if qb.fromClause != "" {
		query += "FROM " + qb.fromClause + " "
	}

	if qb.joinClause != "" {
		query += qb.joinClause + " "
	}

	if qb.whereClause != "" {
		query += qb.whereClause + " "
	}

	return query, qb.args
}

// ToStdDB converts pgx pool to standard SQL DB for compatibility
func ToStdDB(pool *pgxpool.Pool) *sql.DB {
	return stdlib.OpenDBFromPool(pool)
}
