package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/v-egorov/service-boilerplate/migration-orchestrator/internal/database"
	"github.com/v-egorov/service-boilerplate/migration-orchestrator/pkg/types"
	"github.com/v-egorov/service-boilerplate/migration-orchestrator/pkg/utils"
)

// Orchestrator handles migration execution and tracking
type Orchestrator struct {
	db          *database.Database
	logger      *utils.Logger
	serviceName string
	servicePath string
	schemaName  string
}

// NewOrchestrator creates a new migration orchestrator
func NewOrchestrator(db *database.Database, logger *utils.Logger, serviceName string) (*Orchestrator, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	servicePath := filepath.Join("services", serviceName)
	if _, err := os.Stat(servicePath); os.IsNotExist(err) {
		servicePath = filepath.Join("..", "services", serviceName)
		if _, err := os.Stat(servicePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("service directory not found: %s (tried services/%s, cwd: %s)", serviceName, serviceName, cwd)
		}
	}

	schemaName := strings.ReplaceAll(serviceName, "-", "_")

	return &Orchestrator{
		db:          db,
		logger:      logger.WithService(serviceName),
		serviceName: serviceName,
		servicePath: servicePath,
		schemaName:  schemaName,
	}, nil
}

// LoadMigrationConfig loads ONLY the environments.json configuration
func (o *Orchestrator) LoadMigrationConfig() (*types.MigrationConfig, error) {
	configPath := filepath.Join(o.servicePath, "migrations")
	envFile := filepath.Join(configPath, "environments.json")

	config := &types.MigrationConfig{}
	if err := o.loadJSONFile(envFile, config); err != nil {
		return nil, fmt.Errorf("failed to load environments config: %w", err)
	}

	return config, nil
}

// GetMigrationState returns the current migration state
func (o *Orchestrator) GetMigrationState(ctx context.Context) (*types.ServiceMigrationState, error) {
	appliedVersions := o.getAppliedVersionsFromGolangMigrate()

	var executions []types.MigrationExecution
	for _, version := range appliedVersions {
		migrationID := fmt.Sprintf("%06d", version)
		now := time.Now()
		execution := types.MigrationExecution{
			MigrationID: migrationID,
			Status:      types.StatusCompleted,
			CompletedAt: &now,
		}
		executions = append(executions, execution)
	}

	return &types.ServiceMigrationState{
		ServiceName:  o.serviceName,
		SchemaName:   o.schemaName,
		Executions:   executions,
		AppliedCount: len(executions),
		FailedCount:  0,
	}, nil
}

// InitializeMigrationSchema creates the service schema and golang-migrate tracking table
func (o *Orchestrator) InitializeMigrationSchema(ctx context.Context) error {
	o.logger.Infof("Creating schema: %s", o.schemaName)

	// Step 1: Create schema if it does not exist
	schemaQuery := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", o.schemaName)
	_, err := o.db.GetPool().Exec(ctx, schemaQuery)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	// Step 2: Initialize golang-migrate tracking table
	migrationPath := filepath.Join(o.servicePath, "migrations")
	m, err := o.createMigrateInstance(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// Calling Version() creates the schema_migrations table if it does not exist
	_, _, err = m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to initialize golang-migrate table: %w", err)
	}

	o.logger.Info("Golang-migrate tracking table initialized")
	return nil
}

// RunMigrationsUp executes pending migrations for the service
func (o *Orchestrator) RunMigrationsUp(ctx context.Context, environment string) error {
	o.logger.Info("Starting migration run up for environment:", environment)

	// Load configuration
	config, err := o.LoadMigrationConfig()
	if err != nil {
		return fmt.Errorf("failed to load migration config: %w", err)
	}

	// Validate environment exists
	envConfig, exists := config.Environments[environment]
	if !exists {
		return fmt.Errorf("environment '%s' not found in configuration", environment)
	}

	// Get migration directory path
	migrationPath := filepath.Join(o.servicePath, "migrations", envConfig.Migrations)

	// Create golang-migrate instance
	m, err := o.createMigrateInstance(migrationPath)
	if err != nil {
		return err
	}
	defer m.Close()

	// Run all migrations up
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	o.logger.Info("Migration run completed successfully")
	return nil
}

// RunMigrationsDown rolls back the last N migrations
func (o *Orchestrator) RunMigrationsDown(ctx context.Context, steps int, environment string) error {
	o.logger.Infof("Starting migration rollback (%d steps) for environment: %s", steps, environment)

	// Load configuration
	config, err := o.LoadMigrationConfig()
	if err != nil {
		return fmt.Errorf("failed to load migration config: %w", err)
	}

	// Get environment migrations
	envConfig, exists := config.Environments[environment]
	if !exists {
		return fmt.Errorf("environment '%s' not found", environment)
	}

	// Get migration directory path
	migrationPath := filepath.Join(o.servicePath, "migrations", envConfig.Migrations)

	// Create golang-migrate instance
	m, err := o.createMigrateInstance(migrationPath)
	if err != nil {
		return err
	}
	defer m.Close()

	// Use golang-migrate native rollback
	if err := m.Steps(steps); err != nil && err != migrate.ErrNoChange {
		return err
	}

	o.logger.Info("Rollback completed successfully")
	return nil
}

// ValidateMigrationFilesExist validates that the migration directory exists
func (o *Orchestrator) ValidateMigrationFilesExist(migrationDir string) error {
	migrationPath := filepath.Join(o.servicePath, "migrations", migrationDir)
	if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
		return fmt.Errorf("migration directory does not exist: %s", migrationPath)
	}
	return nil
}

// getAppliedVersionsFromGolangMigrate returns map of applied migration versions
func (o *Orchestrator) getAppliedVersionsFromGolangMigrate() map[int]bool {
	applied := make(map[int]bool)

	fullTableName := fmt.Sprintf("%s.schema_migrations", o.schemaName)
	if !o.tableExists(fullTableName) {
		return applied
	}

	query := fmt.Sprintf("SELECT version FROM %s WHERE dirty = false", fullTableName)
	rows, err := o.db.GetPool().Query(context.Background(), query)
	if err != nil {
		o.logger.Errorf("Failed to query golang-migrate tracking table: %v", err)
		return applied
	}
	defer rows.Close()

	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			continue
		}
		applied[version] = true
	}

	return applied
}

// tableExists checks if a table exists in the database
func (o *Orchestrator) tableExists(tableName string) bool {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema || '.' || table_name = $1
		)`
	var exists bool
	err := o.db.GetPool().QueryRow(context.Background(), query, tableName).Scan(&exists)
	return err == nil && exists
}

// schemaExists checks if a schema exists
func (o *Orchestrator) schemaExists() bool {
	query := fmt.Sprintf("SELECT 1 FROM information_schema.schemata WHERE schema_name = '%s'", o.schemaName)
	var exists int
	err := o.db.GetPool().QueryRow(context.Background(), query).Scan(&exists)
	return err == nil && exists == 1
}

// createMigrateInstance creates a golang-migrate instance with schema support via search_path
func (o *Orchestrator) createMigrateInstance(migrationPath string) (*migrate.Migrate, error) {
	if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("migration path does not exist: %s", migrationPath)
	}

	config := o.db.GetPool().Config().ConnConfig
	// Use search_path to point to our schema, golang-migrate will use it for schema_migrations table
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&search_path=%s",
		config.User, config.Password, config.Host, config.Port, config.Database, o.schemaName)

	m, err := migrate.New(fmt.Sprintf("file://%s", migrationPath), connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return m, nil
}

// loadJSONFile is a private helper to load JSON files
func (o *Orchestrator) loadJSONFile(filename string, v interface{}) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON from %s: %w", filename, err)
	}

	return nil
}

// ServicePath returns the service path
func (o *Orchestrator) ServicePath() string {
	return o.servicePath
}
