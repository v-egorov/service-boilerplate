package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Host        string
	Port        int
	User        string
	Password    string
	Database    string
	SSLMode     string
	MaxConns    int32
	MinConns    int32
	MaxConnIdle time.Duration
	MaxConnLife time.Duration
}

type PostgresDB struct {
	Pool   *pgxpool.Pool
	Logger *logrus.Logger
}

func NewPostgresDB(config Config, logger *logrus.Logger) (*PostgresDB, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		config.User, config.Password, config.Host, config.Port, config.Database, config.SSLMode)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.MaxConns = config.MaxConns
	poolConfig.MinConns = config.MinConns
	poolConfig.MaxConnIdleTime = config.MaxConnIdle
	poolConfig.MaxConnLifetime = config.MaxConnLife

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Successfully connected to PostgreSQL database")

	return &PostgresDB{
		Pool:   pool,
		Logger: logger,
	}, nil
}

func (db *PostgresDB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		db.Logger.Info("Database connection pool closed")
	}
}

func (db *PostgresDB) GetPool() *pgxpool.Pool {
	return db.Pool
}

func (db *PostgresDB) HealthCheck(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

func (db *PostgresDB) Stats() *pgxpool.Stat {
	return db.Pool.Stat()
}

// Transaction helper with tracing
func (db *PostgresDB) WithTx(ctx context.Context, fn func(pgx.Tx) error) error {
	return TraceDBOperation(ctx, "TRANSACTION", "transaction", "BEGIN/COMMIT", func(ctx context.Context) error {
		tx, err := db.Pool.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)

		if err := fn(tx); err != nil {
			return err
		}

		return tx.Commit(ctx)
	})
}
