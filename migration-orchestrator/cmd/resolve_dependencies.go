package cmd

// DEPRECATED: Cross-service dependency resolution is no longer needed
// Each service manages its own migrations independently
// Each service manages its own migrations independently
// func newResolveDependenciesCmd() *cobra.Command {
// 	cmd := &cobra.Command{
// 		Use:          "resolve-dependencies <service1> <service2> ...",
// 		Short:        "Resolve cross-service dependencies and output execution order",
// 		Long:         `Analyze migration dependencies across services and output the correct execution order.`,
// 		Args:         cobra.MinimumNArgs(1),
// 		SilenceUsage: true,
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			serviceNames := args
//
// 			logger.Info("🔗 Resolving cross-service dependencies...")
//
// 			// Output the resolved order (same as input - no cross-service deps)
// 			fmt.Printf("SERVICES_ORDER:%s\n", strings.Join(serviceNames, " "))
// 			return nil
// 		},
// 	}
//
// 	return cmd
// }
