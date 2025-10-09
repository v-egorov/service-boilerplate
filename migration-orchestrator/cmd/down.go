package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDownCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down <service-name> [steps]",
		Short: "Rollback migrations",
		Long:  `Rollback the specified number of migrations for the given service.`,
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceName := args[0]
			steps := 1

			logger.Info("⬇️  Rolling back", steps, "migrations for service:", serviceName)

			// TODO: Implement migration down logic
			return fmt.Errorf("migration down command not yet implemented")
		},
	}

	return cmd
}
