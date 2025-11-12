package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/v-egorov/service-boilerplate/migration-orchestrator/internal/orchestrator"
)

func newDownCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "down <service-name> [steps]",
		Short:        "Rollback migrations",
		Long:         `Rollback the specified number of migrations for the given service.`,
		Args:         cobra.RangeArgs(1, 2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceName := args[0]
			steps := 1

			// Parse steps argument if provided
			if len(args) > 1 {
				parsedSteps, err := strconv.Atoi(args[1])
				if err != nil {
					return fmt.Errorf("invalid steps value '%s': must be a number", args[1])
				}
				if parsedSteps < 1 {
					return fmt.Errorf("steps must be at least 1")
				}
				steps = parsedSteps
			}

			logger.Info("⬇️  Rolling back", steps, "migrations for service:", serviceName)

			// Create orchestrator
			orch, err := orchestrator.NewOrchestrator(db, logger, serviceName)
			if err != nil {
				return fmt.Errorf("failed to create orchestrator: %w", err)
			}

			// Run rollback
			ctx := context.Background()
			if err := orch.RunMigrationsDown(ctx, steps, appConfig.Environment); err != nil {
				return fmt.Errorf("migration rollback failed: %w", err)
			}

			fmt.Printf("✅ Successfully rolled back %d migration(s)\n", steps)
			return nil
		},
	}

	return cmd
}
