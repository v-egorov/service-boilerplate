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
		Short:        "Initialize migration tracking for service",
		Long:         `Initialize the migration tracking system for a service by creating the necessary database schema and golang-migrate tracking table.`,
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

			logger.Info("Successfully connected to PostgreSQL database")

			// Validate configuration exists
			logger.Info("Validating migration configuration...")
			_, err = orch.LoadMigrationConfig()
			if err != nil {
				return fmt.Errorf("failed to load migration configuration: %w", err)
			}

			// Create schema and golang-migrate tracking table
			logger.Info("Creating migration tracking schema...")
			ctx := context.Background()
			if err := orch.InitializeMigrationSchema(ctx); err != nil {
				return fmt.Errorf("failed to initialize migration schema: %w", err)
			}

			logger.Info("✅ Migration tracking initialized successfully for service:", serviceName)
			return nil
		},
	}

	return cmd
}
