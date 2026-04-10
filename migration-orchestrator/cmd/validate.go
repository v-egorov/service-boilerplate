package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/v-egorov/service-boilerplate/migration-orchestrator/internal/orchestrator"
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

			// 1. Validate configuration file exists and is valid
			logger.Info("Checking configuration files...")
			migrationConfig, err := orch.LoadMigrationConfig()
			if err != nil {
				validationErrors = append(validationErrors, fmt.Sprintf("Configuration error: %v", err))
			} else {
				logger.Info("✅ Configuration file is valid")
			}

			// 2. Validate environment exists in config
			if migrationConfig != nil {
				if _, exists := migrationConfig.Environments[appConfig.Environment]; !exists {
					validationErrors = append(validationErrors, fmt.Sprintf("Environment '%s' not found in configuration", appConfig.Environment))
				} else {
					logger.Info("✅ Environment configuration exists")
				}
			}

			// 3. Validate migration state
			state, err := orch.GetMigrationState(ctx)
			if err != nil {
				validationErrors = append(validationErrors, fmt.Sprintf("Failed to get migration state: %v", err))
			} else {
				logger.Info("✅ Migration state retrieved successfully")

				// Check for failed migrations
				if state.FailedCount > 0 {
					validationErrors = append(validationErrors, fmt.Sprintf("Found %d failed migrations", state.FailedCount))
				}
			}

			// 4. Validate migration files exist
			if migrationConfig != nil {
				envConfig, exists := migrationConfig.Environments[appConfig.Environment]
				if exists {
					if err := orch.ValidateMigrationFilesExist(envConfig.Migrations); err != nil {
						validationErrors = append(validationErrors, err.Error())
					}
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
