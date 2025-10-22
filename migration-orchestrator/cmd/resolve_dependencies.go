package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/v-egorov/service-boilerplate/migration-orchestrator/internal/orchestrator"
)

func newResolveDependenciesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "resolve-dependencies <service1> <service2> ...",
		Short:        "Resolve cross-service dependencies and output execution order",
		Long:         `Analyze migration dependencies across services and output the correct execution order.`,
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceNames := args

			logger.Info("🔗 Resolving cross-service dependencies...")

			// Resolve dependencies
			plan, err := orchestrator.ResolveServiceDependencies(db, logger, serviceNames)
			if err != nil {
				logger.Errorf("Dependency resolution failed: %v", err)
				// Fall back to original order
				fmt.Print(strings.Join(serviceNames, " "))
				return nil
			}

			if len(plan.CircularDeps) > 0 {
				logger.Warnf("⚠️  Circular dependencies detected: %v", plan.CircularDeps)
				logger.Warn("   Falling back to original order")
				fmt.Print(strings.Join(serviceNames, " "))
				return nil
			}

			// Output the resolved order (will be extracted by Makefile)
			fmt.Printf("SERVICES_ORDER:%s\n", strings.Join(plan.ServiceOrder, " "))
			return nil
		},
	}

	return cmd
}
