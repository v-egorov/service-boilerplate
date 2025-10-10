package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/v-egorov/service-boilerplate/migration-orchestrator/internal/orchestrator"
	"github.com/v-egorov/service-boilerplate/migration-orchestrator/pkg/types"
)

func newValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "validate <service-name>",
		Short:        "Validate migration integrity",
		Long:         `Validate that all migrations are properly applied, dependencies are satisfied, and environment configuration is correct.`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceName := args[0]

			logger.Info("✅ Validating migrations for service:", serviceName)

			// Create orchestrator
			orch, err := orchestrator.NewOrchestrator(db, logger, serviceName)
			if err != nil {
				return fmt.Errorf("failed to create orchestrator: %w", err)
			}

			ctx := context.Background()
			var validationErrors []string

			// 1. Validate configuration files exist and are valid
			logger.Info("Checking configuration files...")
			migrationConfig, depConfig, err := orch.LoadMigrationConfig()
			if err != nil {
				validationErrors = append(validationErrors, fmt.Sprintf("Configuration error: %v", err))
			} else {
				logger.Info("✅ Configuration files are valid")
			}

			// 2. Validate environment exists in config
			if migrationConfig != nil {
				if _, exists := migrationConfig.Environments[appConfig.Environment]; !exists {
					validationErrors = append(validationErrors, fmt.Sprintf("Environment '%s' not found in configuration", appConfig.Environment))
				} else {
					logger.Info("✅ Environment configuration exists")
				}
			}

			// 3. Validate migration tracking table exists
			if !orch.MigrationExecutionsTableExists(ctx) {
				validationErrors = append(validationErrors, "Migration tracking table does not exist - run 'init' command first")
			} else {
				logger.Info("✅ Migration tracking table exists")
			}

			// 4. Validate migration state consistency
			state, err := orch.GetMigrationState(ctx)
			if err != nil {
				validationErrors = append(validationErrors, fmt.Sprintf("Failed to get migration state: %v", err))
			} else {
				logger.Info("✅ Migration state retrieved successfully")

				// Check for failed migrations
				if state.FailedCount > 0 {
					validationErrors = append(validationErrors, fmt.Sprintf("Found %d failed migrations", state.FailedCount))
				}

				// Validate dependencies if config available
				if depConfig != nil {
					depErrors := validateDependencies(state, depConfig)
					validationErrors = append(validationErrors, depErrors...)
				}
			}

			// 5. Validate migration files exist
			if migrationConfig != nil {
				envConfig, exists := migrationConfig.Environments[appConfig.Environment]
				if exists {
					fileErrors := validateMigrationFiles(orch.ServicePath(), envConfig)
					validationErrors = append(validationErrors, fileErrors...)
				}
			}

			// Report results
			if len(validationErrors) == 0 {
				logger.Info("🎉 All validations passed!")
				return nil
			} else {
				logger.Errorf("❌ Validation failed with %d error(s):", len(validationErrors))
				for _, err := range validationErrors {
					logger.Error("  • " + err)
				}
				return fmt.Errorf("validation failed with %d error(s)", len(validationErrors))
			}
		},
	}

	return cmd
}

func validateDependencies(state *types.ServiceMigrationState, depConfig *types.DependencyConfig) []string {
	var errors []string

	// Build map of applied migrations
	applied := make(map[string]bool)
	for _, exec := range state.Executions {
		if exec.Status == types.StatusCompleted {
			applied[exec.MigrationID] = true
		}
	}

	// Check dependencies for applied migrations
	for migrationID, info := range depConfig.Migrations {
		if applied[migrationID] {
			for _, dep := range info.DependsOn {
				if !applied[dep] {
					errors = append(errors, fmt.Sprintf("Migration %s is applied but dependency %s is not", migrationID, dep))
				}
			}
		}
	}

	return errors
}

func validateMigrationFiles(servicePath string, envConfig types.EnvironmentConfig) []string {
	var errors []string

	for _, migrationPath := range envConfig.Migrations {
		// For now, just check if the path pattern is valid
		// Could be enhanced to check actual file existence
		if migrationPath == "" {
			errors = append(errors, "Empty migration path found in configuration")
		}
	}

	return errors
}
