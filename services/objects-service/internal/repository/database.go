package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PGDatabase implements DBInterface using pgx/v5
type PGDatabase struct {
	pool *pgxpool.Pool
}

// NewPGDatabase creates a new PGDatabase instance
func NewPGDatabase(pool *pgxpool.Pool) *PGDatabase {
	return &PGDatabase{pool: pool}
}

// Query implements DBInterface
func (db *PGDatabase) Query(ctx context.Context, sql string, args ...any) (Rows, error) {
	return db.pool.Query(ctx, sql, args...)
}

// QueryRow implements DBInterface
func (db *PGDatabase) QueryRow(ctx context.Context, sql string, args ...any) Row {
	return db.pool.QueryRow(ctx, sql, args...)
}

// Exec implements DBInterface
func (db *PGDatabase) Exec(ctx context.Context, sql string, args ...any) (CommandTag, error) {
	return db.pool.Exec(ctx, sql, args...)
}

// BeginTx starts a new transaction (not part of DBInterface, but available for advanced use)
func (db *PGDatabase) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return db.pool.Begin(ctx)
}

// Health checks database connectivity
func (db *PGDatabase) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return db.pool.Ping(ctx)
}

// QueryBuilder provides a simple interface for building SQL queries
type QueryBuilder struct {
	query    string
	args     []interface{}
	argIndex int
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		args:     make([]interface{}, 0),
		argIndex: 1,
	}
}

func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.query += "SELECT " + joinStrings(columns, ", ") + " "
	return qb
}

func (qb *QueryBuilder) From(table string) *QueryBuilder {
	qb.query += "FROM " + table + " "
	return qb
}

func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	if qb.query == "" || !contains(qb.query, "WHERE") {
		qb.query += "WHERE " + condition + " "
	} else {
		qb.query += "AND " + condition + " "
	}
	qb.args = append(qb.args, args...)
	qb.argIndex += len(args)
	return qb
}

func (qb *QueryBuilder) WhereIn(condition string, args []interface{}) *QueryBuilder {
	placeholders := make([]string, len(args))
	for i := range args {
		placeholders[i] = "$" + itoa(qb.argIndex+i)
	}
	qb.argIndex += len(args)
	qb.query += "WHERE " + condition + " IN (" + joinStrings(placeholders, ", ") + ") "
	qb.args = append(qb.args, args...)
	return qb
}

func (qb *QueryBuilder) WhereTagsContain(tags []string) *QueryBuilder {
	for i, tag := range tags {
		if i == 0 {
			qb.query += "AND ($" + itoa(qb.argIndex+i) + " = ANY(tags)) "
		} else {
			qb.query += "AND ($" + itoa(qb.argIndex+i) + " = ANY(tags)) "
		}
		qb.args = append(qb.args, tag)
		qb.argIndex += len(tags)
	}
	return qb
}

func (qb *QueryBuilder) WhereJsonContains(path string, value interface{}) *QueryBuilder {
	qb.query += "AND " + path + "::jsonb @> $" + itoa(qb.argIndex) + "::jsonb "
	qb.args = append(qb.args, value)
	qb.argIndex++
	return qb
}

func (qb *QueryBuilder) WhereDateRange(column string, start, end time.Time) *QueryBuilder {
	if !start.IsZero() {
		qb.query += "AND " + column + " >= $" + itoa(qb.argIndex) + " "
		qb.args = append(qb.args, start)
		qb.argIndex++
	}
	if !end.IsZero() {
		qb.query += "AND " + column + " <= $" + itoa(qb.argIndex) + " "
		qb.args = append(qb.args, end)
		qb.argIndex++
	}
	return qb
}

func (qb *QueryBuilder) OrderBy(columns ...string) *QueryBuilder {
	qb.query += "ORDER BY " + joinStrings(columns, ", ") + " "
	return qb
}

func (qb *QueryBuilder) OrderByDesc(column string) *QueryBuilder {
	qb.query += "ORDER BY " + column + " DESC "
	return qb
}

func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.query += "LIMIT " + itoa(limit) + " "
	return qb
}

func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.query += "OFFSET " + itoa(offset) + " "
	return qb
}

func (qb *QueryBuilder) Build() (string, []interface{}) {
	return qb.query, qb.args
}

func (qb *QueryBuilder) BuildCount() (string, []interface{}) {
	return "SELECT COUNT(*) FROM " + extractTable(qb.query), qb.args
}

// Helper functions
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	negative := n < 0
	if negative {
		n = -n
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	if negative {
		result = "-" + result
	}
	return result
}

func extractTable(query string) string {
	// Simple extraction - looks for "FROM table"
	for i := 0; i < len(query)-5; i++ {
		if query[i:i+5] == "FROM " {
			// Find the end of the table name
			end := i + 5
			for end < len(query) && query[end] != ' ' && query[end] != 'W' && query[end] != 'G' && query[end] != 'O' && query[end] != 'L' {
				end++
			}
			return query[i+5 : end]
		}
	}
	return ""
}
