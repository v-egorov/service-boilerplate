package cmd

import (
	"github.com/spf13/cobra"
)

// newDevCmd creates the dev command
func newDevCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dev",
		Short: "Development utilities",
		Long:  `Development tools and utilities for the boilerplate.`,
	}

	cmd.AddCommand(
		newDevScaffoldCmd(),
		newDevTestCmd(),
		newDevMigrateCmd(),
		newDevLogsCmd(),
	)

	return cmd
}

func newDevScaffoldCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scaffold <entity> <service>",
		Short: "Generate CRUD operations for entity",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement scaffolding
			entity := args[0]
			service := args[1]
			cmd.Printf("Scaffolding CRUD operations for entity %s in service %s\n", entity, service)
			return nil
		},
	}
}

func newDevTestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test <service>",
		Short: "Run service tests with options",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement test running
			service := args[0]
			cmd.Printf("Running tests for service %s\n", service)
			return nil
		},
	}

	cmd.Flags().BoolP("watch", "w", false, "Watch for changes and re-run tests")

	return cmd
}

func newDevMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate <service> <action>",
		Short: "Database migration operations",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement migration operations
			service := args[0]
			action := args[1]
			cmd.Printf("Running migration %s for service %s\n", action, service)
			return nil
		},
	}
}

func newDevLogsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs <service>",
		Short: "Stream service logs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement log streaming
			service := args[0]
			cmd.Printf("Streaming logs for service %s\n", service)
			return nil
		},
	}

	cmd.Flags().BoolP("follow", "f", false, "Follow log output")

	return cmd
}
