package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/v-egorov/service-boilerplate/migration-orchestrator/internal/config"
	"github.com/v-egorov/service-boilerplate/migration-orchestrator/pkg/utils"
)

type Database struct {
	Pool   *pgxpool.Pool
	Logger *utils.Logger
	Config *config.DatabaseConfig
}

func NewDatabase(dbConfig config.DatabaseConfig, logger *utils.Logger) (*Database, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database, dbConfig.SSLMode)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configure connection pool
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 1
	poolConfig.MaxConnIdleTime = 5 * time.Minute
	poolConfig.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Successfully connected to PostgreSQL database")

	return &Database{
		Pool:   pool,
		Logger: logger,
		Config: &dbConfig,
	}, nil
}

func (db *Database) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		db.Logger.Info("Database connection pool closed")
	}
}

func (db *Database) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

func (db *Database) GetPool() *pgxpool.Pool {
	return db.Pool
}
