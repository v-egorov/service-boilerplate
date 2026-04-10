package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

	servicePath := filepath.Join("..", "services", serviceName)
	if _, err := os.Stat(servicePath); os.IsNotExist(err) {
		servicePath = filepath.Join("services", serviceName)
		if _, err := os.Stat(servicePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("service directory not found: %s (tried ../services/%s, cwd: %s)", serviceName, serviceName, cwd)
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

	// Validate migration files exist
	if err := o.validateMigrationFilesExist(envConfig.Migrations); err != nil {
		return err
	}

	// Get current migration state
	appliedVersions := o.getAppliedVersionsFromGolangMigrate()

	// Determine which migrations to run
	var pendingMigrations []string
	for _, migrationPath := range envConfig.Migrations {
		migrationID := o.extractMigrationID(migrationPath)
		if !o.isMigrationApplied(migrationID, appliedVersions) {
			// Check if migration is for this environment
			if o.isMigrationForEnvironment(migrationPath, environment) {
				pendingMigrations = append(pendingMigrations, migrationID)
			}
		}
	}

	if len(pendingMigrations) == 0 {
		o.logger.Info("No pending migrations found")
		return nil
	}

	o.logger.Infof("Found %d pending migrations", len(pendingMigrations))

	// Execute migrations in order
	for _, migrationID := range pendingMigrations {
		if err := o.executeMigrationUp(ctx, migrationID, environment, envConfig); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", migrationID, err)
		}
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
		o.logger.Warnf("Could not load migration config: %v", err)
		return fmt.Errorf("failed to load migration config: %w", err)
	}

	// Get environment migrations
	envConfig, exists := config.Environments[environment]
	if !exists {
		return fmt.Errorf("environment '%s' not found", environment)
	}

	// Get applied versions
	appliedVersions := o.getAppliedVersionsFromGolangMigrate()

	// Find migrations to rollback (in reverse order)
	var toRollback []string
	for i := len(envConfig.Migrations) - 1; i >= 0 && len(toRollback) < steps; i-- {
		migrationPath := envConfig.Migrations[i]
		migrationID := o.extractMigrationID(migrationPath)
		version, err := o.migrationIDToVersion(migrationID)
		if err != nil {
			continue
		}
		if appliedVersions[version] {
			// Check if migration is for this environment
			if o.isMigrationForEnvironment(migrationPath, environment) {
				toRollback = append(toRollback, migrationID)
			}
		}
	}

	if len(toRollback) == 0 {
		o.logger.Info("No migrations to rollback")
		return nil
	}

	// Rollback migrations in reverse order (most recent first)
	for i := len(toRollback) - 1; i >= 0; i-- {
		migrationID := toRollback[i]
		if err := o.executeMigrationDown(ctx, migrationID); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", migrationID, err)
		}
	}

	o.logger.Info("Migration rollback completed successfully")
	return nil
}

// validateMigrationFilesExist checks that all migration files referenced in config exist
func (o *Orchestrator) validateMigrationFilesExist(migrationPaths []string) error {
	migrationsDir := filepath.Join(o.servicePath, "migrations")
	var missingFiles []string

	for _, migrationPath := range migrationPaths {
		fullPath := filepath.Join(migrationsDir, migrationPath)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			missingFiles = append(missingFiles, migrationPath)
		}
	}

	if len(missingFiles) > 0 {
		return fmt.Errorf("missing migration files: %v", missingFiles)
	}

	return nil
}

// isMigrationForEnvironment checks if a migration applies to the given environment
// Based on -- Environment: tag in the SQL file
func (o *Orchestrator) isMigrationForEnvironment(migrationPath, environment string) bool {
	fullPath := filepath.Join(o.servicePath, "migrations", migrationPath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		o.logger.Warnf("Could not read migration file %s: %v", fullPath, err)
		return true // Don't block execution on read errors
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "-- Environment:") {
			envValue := strings.TrimSpace(strings.TrimPrefix(line, "-- Environment:"))
			if envValue == "all" || envValue == environment {
				return true
			}
		}
	}

	// Default to true if no Environment tag found (backward compatibility)
	o.logger.Warnf("No Environment tag found in %s, assuming applies to all environments", migrationPath)
	return true
}

// extractMigrationID extracts the migration ID from a file path
func (o *Orchestrator) extractMigrationID(migrationPath string) string {
	// Extract filename from path (handle both "000001.up.sql" and "subdir/000001.up.sql")
	filename := filepath.Base(migrationPath)
	if len(filename) >= 6 && strings.HasPrefix(filename, "000") {
		return filename[:6]
	}
	return ""
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

// isMigrationApplied checks if a migration has been applied
func (o *Orchestrator) isMigrationApplied(migrationID string, appliedVersions map[int]bool) bool {
	version, err := o.migrationIDToVersion(migrationID)
	if err != nil {
		return false
	}
	return appliedVersions[version]
}

// migrationIDToVersion converts migration ID string to integer version
func (o *Orchestrator) migrationIDToVersion(migrationID string) (int, error) {
	if len(migrationID) != 6 || !strings.HasPrefix(migrationID, "000") {
		return 0, fmt.Errorf("invalid migration ID format: %s", migrationID)
	}
	return strconv.Atoi(strings.TrimPrefix(migrationID, "000"))
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

// executeMigrationUp executes a single migration up
func (o *Orchestrator) executeMigrationUp(ctx context.Context, migrationID, environment string, envConfig types.EnvironmentConfig) error {
	o.logger.Infof("Executing migration %s for environment %s", migrationID, environment)

	migrationPath, err := o.findMigrationPath(migrationID, envConfig)
	if err != nil {
		return err
	}

	// Read and execute SQL
	sqlContent, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	startTime := time.Now()
	_, err = o.db.GetPool().Exec(ctx, string(sqlContent))
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		o.logger.Errorf("Migration %s failed: %v", migrationID, err)
		return err
	}

	o.logger.Infof("Migration %s completed in %dms", migrationID, duration)
	return nil
}

// findMigrationPath finds the migration file path for a given migration ID
func (o *Orchestrator) findMigrationPath(migrationID string, envConfig types.EnvironmentConfig) (string, error) {
	for _, migrationPath := range envConfig.Migrations {
		if o.extractMigrationID(migrationPath) == migrationID {
			fullPath := filepath.Join(o.servicePath, "migrations", migrationPath)
			return fullPath, nil
		}
	}
	return "", fmt.Errorf("migration %s not found in environment config", migrationID)
}

// executeMigrationDown rolls back the latest migration
func (o *Orchestrator) executeMigrationDown(ctx context.Context, migrationID string) error {
	o.logger.Infof("Rolling back migration %s", migrationID)

	migrationPath := filepath.Join(o.servicePath, "migrations")
	m, err := o.createMigrateInstance(migrationPath)
	if err != nil {
		return err
	}
	defer m.Close()

	// Handle dirty migrations first
	if err := o.handleDirtyMigrations(m); err != nil {
		return err
	}

	// Run migration down
	if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration rollback failed: %w", err)
	}

	o.logger.Infof("Migration %s rolled back successfully", migrationID)
	return nil
}

// handleDirtyMigrations handles dirty state from interrupted migrations
func (o *Orchestrator) handleDirtyMigrations(m *migrate.Migrate) error {
	version, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			return nil
		}
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	if dirty {
		o.logger.Warnf("Detected dirty migration at version %d, checking schema...", version)
		if o.schemaExists() {
			o.logger.Info("Schema exists, forcing version to clean state")
			if err := m.Force(int(version)); err != nil {
				return fmt.Errorf("failed to force clean version: %w", err)
			}
			o.logger.Info("Successfully forced migration to clean state")
		}
	}

	return nil
}

// schemaExists checks if the service schema exists
func (o *Orchestrator) schemaExists() bool {
	query := fmt.Sprintf("SELECT 1 FROM information_schema.schemata WHERE schema_name = '%s'", o.schemaName)
	var exists int
	err := o.db.GetPool().QueryRow(context.Background(), query).Scan(&exists)
	return err == nil && exists == 1
}

// createMigrateInstance creates a golang-migrate instance
func (o *Orchestrator) createMigrateInstance(migrationPath string) (*migrate.Migrate, error) {
	if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("migration path does not exist: %s", migrationPath)
	}

	config := o.db.GetPool().Config().ConnConfig
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
