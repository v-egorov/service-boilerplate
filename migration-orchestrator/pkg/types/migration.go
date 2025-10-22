package types

import (
	"time"
)

// MigrationStatus represents the status of a migration execution
type MigrationStatus string

const (
	StatusPending    MigrationStatus = "pending"
	StatusRunning    MigrationStatus = "running"
	StatusCompleted  MigrationStatus = "completed"
	StatusFailed     MigrationStatus = "failed"
	StatusRolledBack MigrationStatus = "rolled_back"
)

// MigrationExecution represents an enhanced migration execution record
type MigrationExecution struct {
	ID               int64           `json:"id" db:"id"`
	MigrationID      string          `json:"migration_id" db:"migration_id"`
	MigrationVersion string          `json:"migration_version" db:"migration_version"`
	Environment      string          `json:"environment" db:"environment"`
	Status           MigrationStatus `json:"status" db:"status"`
	StartedAt        *time.Time      `json:"started_at" db:"started_at"`
	CompletedAt      *time.Time      `json:"completed_at" db:"completed_at"`
	DurationMs       *int64          `json:"duration_ms" db:"duration_ms"`
	ExecutedBy       *string         `json:"executed_by" db:"executed_by"`
	Checksum         *string         `json:"checksum" db:"checksum"`
	Dependencies     interface{}     `json:"dependencies" db:"dependencies"` // JSONB
	Metadata         interface{}     `json:"metadata" db:"metadata"`         // JSONB
	ErrorMessage     *string         `json:"error_message" db:"error_message"`
	RollbackVersion  *string         `json:"rollback_version" db:"rollback_version"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at" db:"updated_at"`
}

// MigrationInfo represents information about a migration file
type MigrationInfo struct {
	ID                    string              `json:"id"`
	Version               string              `json:"version"`
	Description           string              `json:"description"`
	DependsOn             []string            `json:"depends_on"`
	CrossServiceDependsOn map[string][]string `json:"cross_service_depends_on,omitempty"`
	AffectsTables         []string            `json:"affects_tables"`
	EstimatedDuration     string              `json:"estimated_duration"`
	RiskLevel             string              `json:"risk_level"`
	RollbackSafe          bool                `json:"rollback_safe"`
	Environment           *string             `json:"environment,omitempty"`
}

// EnvironmentConfig represents the configuration for a specific environment
type EnvironmentConfig struct {
	Description string                 `json:"description"`
	Migrations  []string               `json:"migrations"`
	SeedFiles   []string               `json:"seed_files"`
	Config      map[string]interface{} `json:"config"`
}

// MigrationConfig represents the complete migration configuration
type MigrationConfig struct {
	Environments       map[string]EnvironmentConfig `json:"environments"`
	CurrentEnvironment string                       `json:"current_environment"`
	MigrationLocking   map[string]interface{}       `json:"migration_locking"`
}

// DependencyConfig represents the migration dependencies configuration
type DependencyConfig struct {
	Migrations   map[string]MigrationInfo `json:"migrations"`
	GlobalConfig map[string]interface{}   `json:"global_config"`
}

// ServiceMigrationState represents the current state of migrations for a service
type ServiceMigrationState struct {
	ServiceName    string               `json:"service_name"`
	SchemaName     string               `json:"schema_name"`
	CurrentVersion string               `json:"current_version"`
	PendingCount   int                  `json:"pending_count"`
	AppliedCount   int                  `json:"applied_count"`
	FailedCount    int                  `json:"failed_count"`
	LastMigration  *MigrationExecution  `json:"last_migration"`
	Executions     []MigrationExecution `json:"executions"`
}
