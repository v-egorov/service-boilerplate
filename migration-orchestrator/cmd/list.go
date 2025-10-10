package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list <service-name>",
		Short:        "List all migrations",
		Long:         `List all migrations for the specified service with their status.`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceName := args[0]

			logger.Info("📋 All migrations for service:", serviceName)

			// TODO: Implement migration list logic
			return fmt.Errorf("migration list command not yet implemented")
		},
	}

	return cmd
}
