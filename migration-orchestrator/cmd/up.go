package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/v-egorov/service-boilerplate/migration-orchestrator/internal/orchestrator"
)

func newUpCmd() *cobra.Command {
	var environment string

	cmd := &cobra.Command{
		Use:          "up <service-name>",
		Short:        "Run pending migrations up",
		Long:         `Execute all pending migrations for the specified service in the correct order.`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceName := args[0]

			logger.Info("🚀 Running migrations up for service:", serviceName)

			// Create orchestrator
			orch, err := orchestrator.NewOrchestrator(db, logger, serviceName)
			if err != nil {
				return fmt.Errorf("failed to create orchestrator: %w", err)
			}

			// Determine environment - use flag if provided, otherwise use config
			env := appConfig.Environment
			if environment != "" {
				env = environment
			}
			logger.Infof("Using environment: %s", env)

			// Run migrations
			ctx := context.Background()
			if err := orch.RunMigrationsUp(ctx, env); err != nil {
				return fmt.Errorf("migration run failed: %w", err)
			}

			fmt.Println("✅ Migrations completed successfully")
			return nil
		},
	}

	cmd.Flags().StringVarP(&environment, "env", "e", "", "environment (development/staging/production)")

	return cmd
}
