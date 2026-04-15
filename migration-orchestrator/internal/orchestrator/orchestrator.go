package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
	for versionNumber, version := range appliedVersions {
		if !version {
			continue
		}
		migrationID := fmt.Sprintf("%06d", versionNumber)
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

	// Step 2: Initialize golang-migrate tracking table via CLI
	migrationPath := filepath.Join(o.servicePath, "migrations")
	absPath, err := filepath.Abs(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	configDB := o.db.GetPool().Config().ConnConfig
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&search_path=%s",
		configDB.User, configDB.Password, configDB.Host, configDB.Port, configDB.Database, o.schemaName)

	// Use CLI to create/verify schema_migrations table (runs "up 0" which just creates table)
	cmd := exec.Command("migrate", "-path", absPath, "-database", dsn, "up", "0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Error is expected if no migrations exist, but table should still be created
		o.logger.Infof("Init output (may contain errors): %s", string(output))
	}

	o.logger.Info("Golang-migrate tracking table initialized")
	return nil
}

// RunMigrationsUp executes pending migrations for the service using CLI
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
	absPath, err := filepath.Abs(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Get database connection
	configDB := o.db.GetPool().Config().ConnConfig
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&search_path=%s",
		configDB.User, configDB.Password, configDB.Host, configDB.Port, configDB.Database, o.schemaName)

	// Use golang-migrate CLI for migrations
	o.logger.Infof("Applying migrations from: %s", absPath)
	cmd := exec.Command("migrate", "-path", absPath, "-database", dsn, "up")
	output, err := cmd.CombinedOutput()
	if err != nil {
		o.logger.Errorf("Migration up failed: %s, output: %s", err, string(output))
		return fmt.Errorf("migration up failed: %w, output: %s", err, string(output))
	}

	// Log output for visibility (CLI prints migration progress)
	if len(output) > 0 {
		o.logger.Infof("Migration output: %s", string(output))
	}

	o.logger.Info("Migration run completed successfully")
	return nil
}

// RunMigrationsDown rolls back the last N migrations using CLI
func (o *Orchestrator) RunMigrationsDown(ctx context.Context, steps int, environment string) error {
	o.logger.Infof("Starting migration rollback (%d steps) for environment: %s", steps, environment)

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
	absPath, err := filepath.Abs(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Get database connection
	configDB := o.db.GetPool().Config().ConnConfig
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&search_path=%s",
		configDB.User, configDB.Password, configDB.Host, configDB.Port, configDB.Database, o.schemaName)

	// Use golang-migrate CLI instead of library for rollback
	cmd := exec.Command("migrate", "-path", absPath, "-database", dsn, "down", strconv.Itoa(steps))
	output, err := cmd.CombinedOutput()
	if err != nil {
		o.logger.Errorf("Migration rollback failed: %s, output: %s", err, string(output))
		return fmt.Errorf("migration rollback failed: %w, output: %s", err, string(output))
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
