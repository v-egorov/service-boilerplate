package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	// For now, assume we're running from the migration-orchestrator directory
	// and the services are in ../services/
	servicePath := filepath.Join("..", "services", serviceName)
	if _, err := os.Stat(servicePath); os.IsNotExist(err) {
		// Try current directory if that doesn't work
		servicePath = filepath.Join("services", serviceName)
		if _, err := os.Stat(servicePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("service directory not found: %s (tried ../services/%s, cwd: %s)", serviceName, serviceName, cwd)
		}
	}

	// Convert service name to schema name (hyphens to underscores)
	schemaName := strings.ReplaceAll(serviceName, "-", "_")

	return &Orchestrator{
		db:          db,
		logger:      logger.WithService(serviceName),
		serviceName: serviceName,
		servicePath: servicePath,
		schemaName:  schemaName,
	}, nil
}

// LoadMigrationConfig loads the migration configuration files
func (o *Orchestrator) LoadMigrationConfig() (*types.MigrationConfig, *types.DependencyConfig, error) {
	configPath := filepath.Join(o.servicePath, "migrations")

	// Load environments.json
	envConfig := &types.MigrationConfig{}
	envFile := filepath.Join(configPath, "environments.json")
	if err := o.loadJSONFile(envFile, envConfig); err != nil {
		return nil, nil, fmt.Errorf("failed to load environments config: %w", err)
	}

	// Load dependencies.json
	depConfig := &types.DependencyConfig{}
	depFile := filepath.Join(configPath, "dependencies.json")
	if err := o.loadJSONFile(depFile, depConfig); err != nil {
		return nil, nil, fmt.Errorf("failed to load dependencies config: %w", err)
	}

	return envConfig, depConfig, nil
}

// RunMigrationsUp executes pending migrations for the service
func (o *Orchestrator) RunMigrationsUp(ctx context.Context, environment string) error {
	o.logger.Info("Starting migration run up for environment:", environment)
	o.logger.Info("Loading migration configuration...")

	// Load configuration
	o.logger.Info("About to load migration config...")
	migrationConfig, depConfig, err := o.LoadMigrationConfig()
	if err != nil {
		o.logger.Errorf("Failed to load migration config: %v", err)
		return fmt.Errorf("failed to load migration config: %w", err)
	}
	o.logger.Info("Migration config loaded successfully")

	// Get current migration state
	currentState, err := o.GetMigrationState(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current migration state: %w", err)
	}

	// Get environment-specific migrations
	o.logger.Infof("Looking for environment: %s", environment)
	envConfig, exists := migrationConfig.Environments[environment]
	if !exists {
		o.logger.Errorf("Environment '%s' not found in configuration", environment)
		return fmt.Errorf("environment '%s' not found in configuration", environment)
	}
	o.logger.Info("Environment found, proceeding with sync...")

	// Sync orchestrator tracking with golang-migrate
	o.logger.Info("About to sync orchestrator tracking...")
	appliedVersions := o.getAppliedVersionsFromGolangMigrate()
	o.logger.Infof("Got applied versions: %v", appliedVersions)
	o.syncOrchestratorTrackingWithGolangMigrate(appliedVersions, currentState)
	o.logger.Info("Sync completed")

	// Refresh state after sync
	currentState, err = o.GetMigrationState(ctx)
	if err != nil {
		return fmt.Errorf("failed to get updated migration state: %w", err)
	}

	// Determine which migrations to run
	pendingMigrations := o.getPendingMigrations(envConfig, depConfig, currentState)

	if len(pendingMigrations) == 0 {
		o.logger.Info("No pending migrations found")
		return nil
	}

	o.logger.Infof("Found %d pending migrations", len(pendingMigrations))

	// Perform risk assessment
	if err := o.assessMigrationRisks(pendingMigrations, depConfig); err != nil {
		return fmt.Errorf("risk assessment failed: %w", err)
	}

	// Execute migrations in dependency order
	for _, migrationID := range pendingMigrations {
		if err := o.executeMigrationUp(ctx, migrationID, environment, depConfig); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", migrationID, err)
		}
	}

	o.logger.Info("Migration run completed successfully")
	return nil
}

// RunMigrationsDown rolls back migrations for the service with dependency checking
func (o *Orchestrator) RunMigrationsDown(ctx context.Context, steps int, environment string) error {
	o.logger.Infof("Starting intelligent migration rollback (%d steps) for environment: %s", steps, environment)

	// Load configuration for dependency checking
	_, depConfig, err := o.LoadMigrationConfig()
	if err != nil {
		o.logger.Warnf("Could not load migration config for dependency checking: %v", err)
		o.logger.Info("Proceeding with basic rollback...")
	}

	// Get recent migration executions
	executions, err := o.getRecentExecutions(ctx, environment, steps)
	if err != nil {
		return fmt.Errorf("failed to get recent executions: %w", err)
	}

	if len(executions) == 0 {
		o.logger.Info("No migrations to rollback")
		return nil
	}

	// Check for dependent migrations that would be affected
	affectedMigrations := o.checkRollbackDependencies(executions, depConfig)
	if len(affectedMigrations) > 0 {
		o.logger.Warnf("⚠️  Rollback may affect %d dependent migration(s): %v", len(affectedMigrations), affectedMigrations)
		o.logger.Warn("   These migrations depend on the ones being rolled back")
		o.logger.Warn("   Consider rolling back more steps or using targeted rollback")
	}

	o.logger.Infof("Rolling back %d migrations", len(executions))

	// Rollback migrations in reverse order
	for i := len(executions) - 1; i >= 0; i-- {
		execution := executions[i]
		if err := o.executeMigrationDown(ctx, execution.MigrationID, environment); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", execution.MigrationID, err)
		}
	}

	o.logger.Info("Migration rollback completed successfully")
	return nil
}

// GetMigrationState returns the current migration state for the service
func (o *Orchestrator) GetMigrationState(ctx context.Context) (*types.ServiceMigrationState, error) {
	state := &types.ServiceMigrationState{
		ServiceName: o.serviceName,
		SchemaName:  o.schemaName,
		Executions:  []types.MigrationExecution{},
	}

	// Query migration executions
	query := fmt.Sprintf(`
		SELECT id, migration_id, migration_version, environment, status,
		       started_at, completed_at, duration_ms, executed_by, checksum,
		       dependencies, metadata, error_message, rollback_version,
		       created_at, updated_at
		FROM %s.migration_executions
		ORDER BY created_at DESC`, o.schemaName)

	rows, err := o.db.GetPool().Query(ctx, query)
	if err != nil {
		// Check if the table doesn't exist yet (first run)
		if strings.Contains(err.Error(), "does not exist") {
			o.logger.Info("Migration executions table does not exist yet, assuming first run")
			return state, nil
		}
		return nil, fmt.Errorf("failed to query migration executions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var exec types.MigrationExecution
		err := rows.Scan(
			&exec.ID, &exec.MigrationID, &exec.MigrationVersion, &exec.Environment, &exec.Status,
			&exec.StartedAt, &exec.CompletedAt, &exec.DurationMs, &exec.ExecutedBy, &exec.Checksum,
			&exec.Dependencies, &exec.Metadata, &exec.ErrorMessage, &exec.RollbackVersion,
			&exec.CreatedAt, &exec.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan migration execution: %w", err)
		}

		state.Executions = append(state.Executions, exec)

		// Update counters
		switch exec.Status {
		case types.StatusCompleted:
			state.AppliedCount++
		case types.StatusFailed:
			state.FailedCount++
		}

		// Set last migration
		if state.LastMigration == nil {
			state.LastMigration = &exec
		}
	}

	if len(state.Executions) > 0 {
		state.CurrentVersion = state.Executions[0].MigrationVersion
	}

	return state, nil
}

// Private methods

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

func (o *Orchestrator) getPendingMigrations(envConfig types.EnvironmentConfig, depConfig *types.DependencyConfig, state *types.ServiceMigrationState) []string {
	// Check golang-migrate tracking table for applied migrations
	appliedVersions := o.getAppliedVersionsFromGolangMigrate()

	// Convert applied versions to migration IDs
	applied := make(map[string]bool)
	for version := range appliedVersions {
		migrationID := fmt.Sprintf("%06d", version) // Convert to 000001 format
		applied[migrationID] = true

		// Since golang-migrate applies migrations sequentially, mark all lower versions as applied
		for i := 1; i < version; i++ {
			lowerID := fmt.Sprintf("%06d", i)
			applied[lowerID] = true
		}
	}

	// Also check orchestrator tracking for any additional applied migrations
	for _, exec := range state.Executions {
		if exec.Status == types.StatusCompleted {
			applied[exec.MigrationID] = true
		}
	}

	// Sync orchestrator tracking table with golang-migrate state
	o.syncOrchestratorTrackingWithGolangMigrate(appliedVersions, state)

	// Get all migrations that should be considered for this environment
	var candidateMigrations []string

	// Add base migrations that golang-migrate manages (find all .up.sql files in root migrations dir)
	baseMigrations := o.getBaseMigrations()
	candidateMigrations = append(candidateMigrations, baseMigrations...)

	// Add environment-specific migrations
	for _, migrationPath := range envConfig.Migrations {
		// Extract migration ID from path (e.g., "development/000003_dev_test_data.up.sql" -> "000003")
		parts := strings.Split(migrationPath, "/")
		if len(parts) >= 2 {
			filename := parts[len(parts)-1]
			if strings.HasPrefix(filename, "000") && strings.HasSuffix(filename, ".up.sql") {
				migrationID := filename[:6] // Extract "000003" from "000003_dev_test_data.up.sql"
				candidateMigrations = append(candidateMigrations, migrationID)
			}
		}
	}

	// Resolve dependencies and get pending migrations in correct order
	return o.resolveDependencies(candidateMigrations, applied, depConfig)
}

// resolveDependencies resolves migration dependencies and returns pending migrations in execution order
func (o *Orchestrator) resolveDependencies(candidateMigrations []string, applied map[string]bool, depConfig *types.DependencyConfig) []string {
	if depConfig == nil {
		// No dependency config, return all unapplied migrations in order
		var pending []string
		for _, migrationID := range candidateMigrations {
			if !applied[migrationID] {
				pending = append(pending, migrationID)
			}
		}
		return pending
	}

	// Build dependency graph
	graph := make(map[string][]string) // migration -> dependencies
	inDegree := make(map[string]int)   // migration -> number of unresolved dependencies

	// Initialize graph with all candidate migrations
	for _, migrationID := range candidateMigrations {
		graph[migrationID] = []string{}
		inDegree[migrationID] = 0
	}

	// Build dependencies from config
	for migrationID, info := range depConfig.Migrations {
		// Only consider migrations that are candidates for this environment
		if _, exists := graph[migrationID]; !exists {
			continue
		}

		dependencies := info.DependsOn
		graph[migrationID] = dependencies
		inDegree[migrationID] = len(dependencies)
	}

	// Reduce in-degree for migrations whose dependencies are already applied
	for migrationID, deps := range graph {
		for _, dep := range deps {
			if applied[dep] {
				inDegree[migrationID]--
			}
		}
	}

	// Perform topological sort using Kahn's algorithm
	var queue []string
	var result []string

	// Add migrations with no unresolved dependencies to queue
	for migrationID, degree := range inDegree {
		if degree == 0 && !applied[migrationID] {
			queue = append(queue, migrationID)
		}
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// For each migration that depends on current, reduce in-degree
		for migrationID, deps := range graph {
			for _, dep := range deps {
				if dep == current {
					inDegree[migrationID]--
					if inDegree[migrationID] == 0 && !applied[migrationID] {
						queue = append(queue, migrationID)
					}
				}
			}
		}
	}

	// Check for cycles (if not all migrations were processed)
	totalUnapplied := 0
	for _, applied := range applied {
		if !applied {
			totalUnapplied++
		}
	}
	if len(result) < totalUnapplied {
		o.logger.Warn("Dependency cycle detected or missing dependencies in migration graph")
		// Fall back to simple ordering
		var fallback []string
		for _, migrationID := range candidateMigrations {
			if !applied[migrationID] {
				fallback = append(fallback, migrationID)
			}
		}
		return fallback
	}

	return result
}

// assessMigrationRisks checks for high-risk migrations and provides warnings
func (o *Orchestrator) assessMigrationRisks(migrations []string, depConfig *types.DependencyConfig) error {
	if depConfig == nil {
		return nil
	}

	hasHighRisk := false
	for _, migrationID := range migrations {
		if info, exists := depConfig.Migrations[migrationID]; exists {
			if info.RiskLevel == "high" {
				hasHighRisk = true
				o.logger.Warnf("⚠️  HIGH RISK migration detected: %s - %s", migrationID, info.Description)
				o.logger.Warnf("   Affects tables: %v", info.AffectsTables)
				o.logger.Warnf("   Estimated duration: %s", info.EstimatedDuration)
			}
		}
	}

	if hasHighRisk {
		o.logger.Warn("⚠️  High-risk migrations detected. Consider running in a maintenance window.")
		o.logger.Warn("   Use --dry-run flag to preview without executing.")
	}

	return nil
}

func (o *Orchestrator) syncOrchestratorTrackingWithGolangMigrate(appliedVersions map[int]bool, state *types.ServiceMigrationState) {
	ctx := context.Background()

	o.logger.Info("🔄 Syncing orchestrator tracking with golang-migrate...")
	o.logger.Infof("Applied versions from golang-migrate: %v", appliedVersions)
	o.logger.Infof("Current orchestrator executions: %d", len(state.Executions))

	// Skip sync if migration executions table doesn't exist yet
	if len(state.Executions) == 0 && !o.migrationExecutionsTableExists(ctx) {
		o.logger.Info("Migration executions table doesn't exist yet, skipping sync until after initial migration")
		return
	}

	for version, applied := range appliedVersions {
		if !applied {
			continue
		}

		migrationID := fmt.Sprintf("%06d", version)
		o.logger.Infof("Processing migration %s (version %d)", migrationID, version)

		// Check if orchestrator has a record for this migration
		hasRecord := false
		for _, exec := range state.Executions {
			o.logger.Infof("Checking execution: %s/%s status=%s", exec.MigrationID, exec.Environment, exec.Status)
			if exec.MigrationID == migrationID && exec.Environment == "development" {
				hasRecord = true
				o.logger.Infof("Found existing record for %s with status: %s", migrationID, exec.Status)
				// If golang-migrate shows it as applied but orchestrator shows failed, update to completed
				if exec.Status == types.StatusFailed {
					o.logger.Infof("Updating failed migration %s to completed", migrationID)
					o.updateMigrationStatus(ctx, exec.ID, types.StatusCompleted, "Synchronized with golang-migrate")
				}
				break
			}
		}

		// If no orchestrator record but golang-migrate shows applied, create a record
		if !hasRecord {
			o.logger.Infof("Creating new record for migration %s", migrationID)
			o.createMigrationRecord(ctx, migrationID, "development", types.StatusCompleted, "Synchronized with golang-migrate")
		}
	}
}

func (o *Orchestrator) updateMigrationStatus(ctx context.Context, executionID int64, status types.MigrationStatus, message string) {
	query := fmt.Sprintf(`
		UPDATE %s.migration_executions
		SET status = $1, error_message = $2, completed_at = $3, updated_at = $4
		WHERE id = $5`, o.schemaName)

	now := time.Now()
	_, err := o.db.GetPool().Exec(ctx, query, status, message, now, now, executionID)
	if err != nil {
		o.logger.Errorf("Failed to update migration status: %v", err)
	}
}

func (o *Orchestrator) createMigrationRecord(ctx context.Context, migrationID, environment string, status types.MigrationStatus, message string) {
	query := fmt.Sprintf(`
		INSERT INTO %s.migration_executions
		(migration_id, migration_version, environment, status, error_message, completed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`, o.schemaName)

	now := time.Now()
	_, err := o.db.GetPool().Exec(ctx, query, migrationID, migrationID, environment, status, message, now, now, now)
	if err != nil {
		o.logger.Errorf("Failed to create migration record: %v", err)
	}
}

func (o *Orchestrator) getBaseMigrations() []string {
	var baseMigrations []string

	// Find all .up.sql files in the root migrations directory
	files, err := os.ReadDir(o.servicePath + "/migrations")
	if err != nil {
		o.logger.Errorf("Failed to read migrations directory: %v", err)
		return baseMigrations
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".up.sql") && strings.HasPrefix(file.Name(), "000") {
			// Extract migration ID (e.g., "000001_initial.up.sql" -> "000001")
			migrationID := file.Name()[:6]
			baseMigrations = append(baseMigrations, migrationID)
		}
	}

	// Sort to ensure consistent ordering
	sort.Strings(baseMigrations)
	return baseMigrations
}

func (o *Orchestrator) getAppliedVersionsFromGolangMigrate() map[int]bool {
	applied := make(map[int]bool)

	// golang-migrate uses tables in public schema with service-specific names
	tableName := fmt.Sprintf("%s_schema_migrations", o.schemaName)
	query := fmt.Sprintf("SELECT version FROM public.%s WHERE dirty = false", tableName)
	rows, err := o.db.GetPool().Query(context.Background(), query)
	if err != nil {
		// Check if the table doesn't exist (fresh start scenario)
		if strings.Contains(err.Error(), "does not exist") {
			o.logger.Info("Golang-migrate tracking table does not exist yet, assuming fresh start")
			return applied
		}
		o.logger.Errorf("Failed to query golang-migrate tracking table: %v", err)
		return applied
	}
	defer rows.Close()

	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			o.logger.Errorf("Failed to scan version: %v", err)
			continue
		}
		applied[version] = true
	}

	return applied
}

// CreateMigrationExecutionsTable creates the migration_executions table in the service schema if it doesn't exist
func (o *Orchestrator) CreateMigrationExecutionsTable(ctx context.Context) error {
	o.logger.Info("Ensuring migration_executions table exists in schema:", o.schemaName)

	// Check if table already exists
	if o.migrationExecutionsTableExists(ctx) {
		o.logger.Info("Migration tracking table already exists, skipping creation")
		return nil
	}

	// First, ensure the schema exists
	schemaQuery := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", o.schemaName)
	if _, err := o.db.GetPool().Exec(ctx, schemaQuery); err != nil {
		return fmt.Errorf("failed to create schema %s: %w", o.schemaName, err)
	}

	// Create the migration_executions table
	tableQuery := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.migration_executions (
			id BIGSERIAL PRIMARY KEY,
			migration_id VARCHAR(255) NOT NULL,
			migration_version VARCHAR(255) NOT NULL,
			environment VARCHAR(50) NOT NULL,
			status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'running', 'completed', 'failed', 'rolled_back')),
			started_at TIMESTAMP WITH TIME ZONE,
			completed_at TIMESTAMP WITH TIME ZONE,
			duration_ms BIGINT,
			executed_by VARCHAR(255),
			checksum VARCHAR(255),
			dependencies JSONB,
			metadata JSONB,
			error_message TEXT,
			rollback_version VARCHAR(255),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

			UNIQUE(migration_id, environment)
		)`, o.schemaName)

	if _, err := o.db.GetPool().Exec(ctx, tableQuery); err != nil {
		return fmt.Errorf("failed to create migration_executions table: %w", err)
	}

	// Create indexes (matching what's in the initial migration)
	indexes := []string{
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_migration_executions_migration_id ON %s.migration_executions(migration_id)", o.schemaName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_migration_executions_environment ON %s.migration_executions(environment)", o.schemaName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_migration_executions_status ON %s.migration_executions(status)", o.schemaName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_migration_executions_created_at ON %s.migration_executions(created_at)", o.schemaName),
	}

	for _, indexQuery := range indexes {
		if _, err := o.db.GetPool().Exec(ctx, indexQuery); err != nil {
			o.logger.Warnf("Failed to create index: %v", err)
			// Don't fail the whole operation for index creation errors
		}
	}

	o.logger.Info("Migration tracking table created successfully")
	return nil
}

// MigrationExecutionsTableExists checks if the migration executions table exists
func (o *Orchestrator) MigrationExecutionsTableExists(ctx context.Context) bool {
	return o.migrationExecutionsTableExists(ctx)
}

func (o *Orchestrator) migrationExecutionsTableExists(ctx context.Context) bool {
	query := fmt.Sprintf("SELECT 1 FROM %s.migration_executions LIMIT 1", o.schemaName)
	_, err := o.db.GetPool().Exec(ctx, query)
	return err == nil
}

// ServicePath returns the service path
func (o *Orchestrator) ServicePath() string {
	return o.servicePath
}

func (o *Orchestrator) handleDirtyMigrations(m *migrate.Migrate) error {
	// Check if database is in dirty state
	version, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			// No migrations applied yet, nothing to do
			return nil
		}
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	if dirty {
		o.logger.Warnf("Detected dirty migration at version %d, attempting to fix...", version)

		// For dirty migrations, check if the schema exists (indicating migration partially succeeded)
		if o.schemaExists() {
			o.logger.Info("Schema exists, migration likely succeeded - forcing version to clean state")
			if err := m.Force(int(version)); err != nil {
				return fmt.Errorf("failed to force clean version: %w", err)
			}
			o.logger.Info("Successfully forced migration to clean state")
		} else {
			o.logger.Warn("Schema doesn't exist, attempting to rollback dirty migration")
			// For dirty migrations where rollback is blocked, we need to force to version 0
			if err := m.Force(0); err != nil {
				return fmt.Errorf("failed to force version to 0: %w", err)
			}
			o.logger.Info("Successfully reset migration version to 0")
		}
	}

	return nil
}

// checkRollbackDependencies identifies migrations that depend on the ones being rolled back
func (o *Orchestrator) checkRollbackDependencies(executions []types.MigrationExecution, depConfig *types.DependencyConfig) []string {
	if depConfig == nil {
		return nil
	}

	// Build reverse dependency map (what depends on what)
	reverseDeps := make(map[string][]string) // migration -> migrations that depend on it

	for migrationID, info := range depConfig.Migrations {
		for _, dep := range info.DependsOn {
			reverseDeps[dep] = append(reverseDeps[dep], migrationID)
		}
	}

	// Find all migrations being rolled back
	rollbackSet := make(map[string]bool)
	for _, exec := range executions {
		rollbackSet[exec.MigrationID] = true
	}

	// Find migrations that depend on any of the rollback migrations
	var affected []string
	for rollbackID := range rollbackSet {
		if dependents, exists := reverseDeps[rollbackID]; exists {
			for _, dependent := range dependents {
				// Only include if the dependent migration is applied and not being rolled back
				if !rollbackSet[dependent] {
					affected = append(affected, dependent)
				}
			}
		}
	}

	// Remove duplicates
	seen := make(map[string]bool)
	var unique []string
	for _, migrationID := range affected {
		if !seen[migrationID] {
			seen[migrationID] = true
			unique = append(unique, migrationID)
		}
	}

	return unique
}

func (o *Orchestrator) schemaExists() bool {
	// Check if the service schema exists
	query := fmt.Sprintf("SELECT 1 FROM information_schema.schemata WHERE schema_name = '%s'", o.schemaName)
	var exists int
	err := o.db.GetPool().QueryRow(context.Background(), query).Scan(&exists)
	return err == nil && exists == 1
}

func (o *Orchestrator) executeMigrationUp(ctx context.Context, migrationID, environment string, depConfig *types.DependencyConfig) error {
	o.logger.WithMigration(migrationID).Info("Executing migration up")

	// Check if this is a base migration (managed by golang-migrate)
	isBaseMigration := o.isBaseMigration(migrationID)

	if isBaseMigration {
		return o.executeBaseMigrationUp(ctx, migrationID, environment)
	} else {
		return o.executeEnvironmentMigrationUp(ctx, migrationID, environment)
	}
}

func (o *Orchestrator) isBaseMigration(migrationID string) bool {
	baseMigrations := o.getBaseMigrations()
	for _, base := range baseMigrations {
		if migrationID == base {
			return true
		}
	}
	return false
}

func (o *Orchestrator) executeBaseMigrationUp(ctx context.Context, migrationID, environment string) error {
	// For base migrations, check if golang-migrate already applied them
	appliedVersions := o.getAppliedVersionsFromGolangMigrate()
	version, err := o.migrationIDToVersion(migrationID)
	if err != nil {
		return fmt.Errorf("invalid migration ID: %w", err)
	}

	if appliedVersions[version] {
		o.logger.WithMigration(migrationID).Info("Migration already applied by golang-migrate, recording in orchestrator")

		// Just record it in orchestrator tracking
		executionID, err := o.recordMigrationStart(ctx, migrationID, environment)
		if err != nil {
			o.logger.Error("Failed to record migration:", err)
		} else {
			if err := o.recordMigrationSuccess(ctx, executionID, 0); err != nil {
				o.logger.Error("Failed to record migration success:", err)
			}
		}
		return nil
	}

	// If not applied, run it using golang-migrate (fresh start scenario)
	o.logger.WithMigration(migrationID).Info("Base migration not applied yet, executing with golang-migrate")

	// Record migration start in orchestrator
	executionID, err := o.recordMigrationStart(ctx, migrationID, environment)
	if err != nil {
		return fmt.Errorf("failed to record migration start: %w", err)
	}

	startTime := time.Now()

	// Create golang-migrate instance and run the migration
	migrationPath := filepath.Join(o.servicePath, "migrations")
	m, err := o.createMigrateInstance(migrationPath)
	if err != nil {
		o.recordMigrationFailure(ctx, executionID, err.Error())
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// Handle dirty migrations first
	if err := o.handleDirtyMigrations(m); err != nil {
		o.recordMigrationFailure(ctx, executionID, err.Error())
		return fmt.Errorf("failed to handle dirty migrations: %w", err)
	}

	// Run the specific migration up to this version
	if err := m.Migrate(uint(version)); err != nil {
		o.recordMigrationFailure(ctx, executionID, err.Error())
		return fmt.Errorf("failed to run migration %s: %w", migrationID, err)
	}

	// Record migration success
	duration := time.Since(startTime).Milliseconds()
	if err := o.recordMigrationSuccess(ctx, executionID, duration); err != nil {
		o.logger.Error("Failed to record migration success:", err)
	}

	o.logger.WithMigration(migrationID).Info("Base migration completed successfully")
	return nil
}

func (o *Orchestrator) executeEnvironmentMigrationUp(ctx context.Context, migrationID, environment string) error {
	o.logger.WithMigration(migrationID).Info("Executing environment-specific migration")

	// Record migration start
	executionID, err := o.recordMigrationStart(ctx, migrationID, environment)
	if err != nil {
		return fmt.Errorf("failed to record migration start: %w", err)
	}

	startTime := time.Now()

	// Load migration config to find the correct file path
	migrationConfig, _, err := o.LoadMigrationConfig()
	if err != nil {
		o.recordMigrationFailure(ctx, executionID, err.Error())
		return fmt.Errorf("failed to load migration config: %w", err)
	}

	// Find the migration file path for this environment
	var migrationPath string
	if envConfig, exists := migrationConfig.Environments[environment]; exists {
		for _, migrationFile := range envConfig.Migrations {
			// Extract migration ID from path (e.g., "development/000003_dev_test_data.up.sql" -> "000003")
			parts := strings.Split(migrationFile, "/")
			if len(parts) >= 2 {
				filename := parts[len(parts)-1]
				if strings.HasPrefix(filename, migrationID+"_") && strings.HasSuffix(filename, ".up.sql") {
					migrationPath = filepath.Join(o.servicePath, "migrations", migrationFile)
					break
				}
			}
		}
	}

	if migrationPath == "" {
		o.recordMigrationFailure(ctx, executionID, "migration file not found in environment config")
		return fmt.Errorf("migration file for %s not found in environment %s config", migrationID, environment)
	}

	// Read and execute the SQL file
	sqlContent, err := os.ReadFile(migrationPath)
	if err != nil {
		o.recordMigrationFailure(ctx, executionID, err.Error())
		return fmt.Errorf("failed to read migration file %s: %w", migrationPath, err)
	}

	// Execute the SQL
	_, err = o.db.GetPool().Exec(ctx, string(sqlContent))
	if err != nil {
		o.recordMigrationFailure(ctx, executionID, err.Error())
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration success
	duration := time.Since(startTime).Milliseconds()
	if err := o.recordMigrationSuccess(ctx, executionID, duration); err != nil {
		o.logger.Error("Failed to record migration success:", err)
	}

	o.logger.WithMigration(migrationID).Info("Environment migration completed successfully")
	return nil
}

func (o *Orchestrator) migrationIDToVersion(migrationID string) (int, error) {
	if len(migrationID) != 6 || !strings.HasPrefix(migrationID, "000") {
		return 0, fmt.Errorf("invalid migration ID format: %s", migrationID)
	}
	return strconv.Atoi(strings.TrimPrefix(migrationID, "000"))
}

func (o *Orchestrator) executeMigrationDown(ctx context.Context, migrationID, environment string) error {
	o.logger.WithMigration(migrationID).Info("Executing migration down")

	// Create golang-migrate instance
	migrationPath := filepath.Join(o.servicePath, "migrations")
	m, err := o.createMigrateInstance(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// Execute rollback
	if err := m.Steps(-1); err != nil {
		return fmt.Errorf("migration rollback failed: %w", err)
	}

	// Update execution record
	if err := o.recordMigrationRollback(ctx, migrationID, environment); err != nil {
		o.logger.Error("Failed to record migration rollback:", err)
	}

	o.logger.WithMigration(migrationID).Info("Migration rollback completed successfully")
	return nil
}

func (o *Orchestrator) createMigrateInstance(migrationPath string) (*migrate.Migrate, error) {
	o.logger.Infof("Creating migrate instance with path: %s", migrationPath)

	// Verify the path exists
	if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("migration path does not exist: %s", migrationPath)
	}

	// Create database driver
	config := o.db.GetPool().Config().ConnConfig
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&search_path=%s",
		config.User, config.Password, config.Host, config.Port, config.Database, "disable", o.schemaName)

	// Create migrate instance
	m, err := migrate.New(fmt.Sprintf("file://%s", migrationPath), connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return m, nil
}

func (o *Orchestrator) recordMigrationStart(ctx context.Context, migrationID, environment string) (int64, error) {
	query := fmt.Sprintf(`
		INSERT INTO %s.migration_executions
		(migration_id, migration_version, environment, status, started_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (migration_id, environment)
		DO UPDATE SET
			status = EXCLUDED.status,
			started_at = EXCLUDED.started_at,
			error_message = NULL,
			updated_at = EXCLUDED.updated_at
		RETURNING id`, o.schemaName)

	now := time.Now()
	var id int64
	err := o.db.GetPool().QueryRow(ctx, query,
		migrationID, migrationID, environment, types.StatusRunning, now, now, now).Scan(&id)

	return id, err
}

func (o *Orchestrator) recordMigrationSuccess(ctx context.Context, executionID int64, durationMs int64) error {
	query := fmt.Sprintf(`
		UPDATE %s.migration_executions
		SET status = $1, completed_at = $2, duration_ms = $3, updated_at = $4
		WHERE id = $5`, o.schemaName)

	now := time.Now()
	_, err := o.db.GetPool().Exec(ctx, query, types.StatusCompleted, now, durationMs, now, executionID)
	return err
}

func (o *Orchestrator) recordMigrationFailure(ctx context.Context, executionID int64, errorMsg string) error {
	query := fmt.Sprintf(`
		UPDATE %s.migration_executions
		SET status = $1, error_message = $2, updated_at = $3
		WHERE id = $4`, o.schemaName)

	now := time.Now()
	_, err := o.db.GetPool().Exec(ctx, query, types.StatusFailed, errorMsg, now, executionID)
	return err
}

func (o *Orchestrator) recordMigrationRollback(ctx context.Context, migrationID, environment string) error {
	query := fmt.Sprintf(`
		UPDATE %s.migration_executions
		SET status = $1, updated_at = $2
		WHERE migration_id = $3 AND environment = $4`, o.schemaName)

	now := time.Now()
	_, err := o.db.GetPool().Exec(ctx, query, types.StatusRolledBack, now, migrationID, environment)
	return err
}

func (o *Orchestrator) getRecentExecutions(ctx context.Context, environment string, limit int) ([]types.MigrationExecution, error) {
	query := fmt.Sprintf(`
		SELECT id, migration_id, migration_version, environment, status,
		       started_at, completed_at, duration_ms, executed_by, checksum,
		       dependencies, metadata, error_message, rollback_version,
		       created_at, updated_at
		FROM %s.migration_executions
		WHERE environment = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT $3`, o.schemaName)

	rows, err := o.db.GetPool().Query(ctx, query, environment, types.StatusCompleted, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var executions []types.MigrationExecution
	for rows.Next() {
		var exec types.MigrationExecution
		err := rows.Scan(
			&exec.ID, &exec.MigrationID, &exec.MigrationVersion, &exec.Environment, &exec.Status,
			&exec.StartedAt, &exec.CompletedAt, &exec.DurationMs, &exec.ExecutedBy, &exec.Checksum,
			&exec.Dependencies, &exec.Metadata, &exec.ErrorMessage, &exec.RollbackVersion,
			&exec.CreatedAt, &exec.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		executions = append(executions, exec)
	}

	return executions, nil
}
