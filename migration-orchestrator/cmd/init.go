package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/v-egorov/service-boilerplate/migration-orchestrator/internal/orchestrator"
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "init <service-name>",
		Short:        "Initialize migration tracking for a service",
		Long:         `Initialize the migration tracking system for a newly created service by creating the necessary database schema and tables.`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceName := args[0]

			logger.Info("🔧 Initializing migration tracking for service:", serviceName)

			// Create orchestrator
			orch, err := orchestrator.NewOrchestrator(db, logger, serviceName)
			if err != nil {
				return fmt.Errorf("failed to create orchestrator: %w", err)
			}

			// Validate configuration exists
			logger.Info("Validating migration configuration...")
			_, _, err = orch.LoadMigrationConfig()
			if err != nil {
				return fmt.Errorf("failed to load migration configuration: %w", err)
			}

			// Create migration tracking table
			logger.Info("Creating migration tracking table...")
			ctx := context.Background()
			if err := orch.CreateMigrationExecutionsTable(ctx); err != nil {
				return fmt.Errorf("failed to create migration tracking table: %w", err)
			}

			// Initialize golang-migrate tracking table
			logger.Info("Initializing golang-migrate tracking table...")
			if err := orch.InitializeGolangMigrateTable(ctx); err != nil {
				return fmt.Errorf("failed to initialize golang-migrate tracking table: %w", err)
			}

			logger.Info("✅ Migration tracking initialized successfully for service:", serviceName)
			return nil
		},
	}

	return cmd
}
