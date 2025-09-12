package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/v-egorov/service-boilerplate/cli/internal/monitoring"
)

// newHealthCmd creates the health command
func newHealthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Health checks and monitoring",
		Long:  `Check health and status of services and infrastructure.`,
	}

	cmd.AddCommand(
		newHealthCheckCmd(),
		newHealthServicesCmd(),
		newHealthDatabaseCmd(),
		newHealthDependenciesCmd(),
	)

	return cmd
}

func newHealthCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Comprehensive health check all services",
		Long:  `Perform a comprehensive health check of all services with detailed metrics.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			monitor := monitoring.NewMonitor(appConfig, serviceReg, apiClient)

			cmd.Println("🏥 Running comprehensive system health check...")
			cmd.Println("============================================")

			systemMetrics, err := monitor.GetSystemMetrics()
			if err != nil {
				return fmt.Errorf("failed to get system metrics: %w", err)
			}

			if jsonOut {
				return printJSON(systemMetrics)
			}

			// Display system overview
			cmd.Printf("📊 System Overview:\n")
			cmd.Printf("  • Total Services: %d\n", systemMetrics.TotalServices)
			cmd.Printf("  • Healthy Services: %d\n", systemMetrics.HealthyServices)
			cmd.Printf("  • Unhealthy Services: %d\n", systemMetrics.UnhealthyServices)
			cmd.Printf("  • Average Response Time: %v\n", systemMetrics.AverageResponseTime.Round(time.Millisecond))
			cmd.Printf("  • System Uptime: %.1f%%\n", systemMetrics.UptimePercentage)
			cmd.Printf("  • Check Time: %s\n", systemMetrics.Timestamp.Format("15:04:05"))

			cmd.Println("\n📋 Service Details:")
			cmd.Println("==================")

			for serviceName, health := range systemMetrics.ServiceMetrics {
				statusIcon := "❌"
				if health.Status == "healthy" {
					statusIcon = "✅"
				} else if health.Status == "degraded" {
					statusIcon = "⚠️"
				}

				cmd.Printf("%s %s\n", statusIcon, serviceName)
				cmd.Printf("  Status: %s\n", health.Status)
				cmd.Printf("  Response Time: %v\n", health.ResponseTime.Round(time.Millisecond))
				cmd.Printf("  Uptime: %.1f%%\n", health.Uptime)
				cmd.Printf("  Total Requests: %d\n", health.TotalRequests)
				if health.ErrorCount > 0 {
					cmd.Printf("  Errors: %d\n", health.ErrorCount)
				}
				cmd.Printf("  Last Checked: %s\n", health.LastChecked.Format("15:04:05"))
				cmd.Println()
			}

			// Overall assessment
			if systemMetrics.UptimePercentage >= 90.0 {
				cmd.Println("🎉 System Health: EXCELLENT")
			} else if systemMetrics.UptimePercentage >= 75.0 {
				cmd.Println("⚠️  System Health: GOOD")
			} else {
				cmd.Println("❌ System Health: NEEDS ATTENTION")
			}

			return nil
		},
	}
}

func newHealthServicesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "services",
		Short: "Check individual service health",
		Long:  `Check the health status of individual services with detailed metrics.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			monitor := monitoring.NewMonitor(appConfig, serviceReg, apiClient)

			services := serviceReg.GetAllServices()

			if jsonOut {
				healthStatuses := make(map[string]interface{})
				for _, service := range services {
					health, err := monitor.CheckServiceHealth(service.Name)
					if err != nil {
						healthStatuses[service.Name] = map[string]interface{}{
							"error": err.Error(),
						}
					} else {
						healthStatuses[service.Name] = health
					}
				}
				return printJSON(healthStatuses)
			}

			cmd.Println("🔍 Checking individual service health...")
			cmd.Println("=====================================")

			for _, service := range services {
				cmd.Printf("\n🌐 Service: %s\n", service.Name)
				cmd.Printf("   URL: %s\n", service.URL)

				health, err := monitor.CheckServiceHealth(service.Name)
				if err != nil {
					cmd.Printf("   ❌ Error: %v\n", err)
					continue
				}

				statusIcon := "❌"
				if health.Status == "healthy" {
					statusIcon = "✅"
				} else if health.Status == "degraded" {
					statusIcon = "⚠️"
				}

				cmd.Printf("   Status: %s %s\n", statusIcon, health.Status)
				cmd.Printf("   Response Time: %v\n", health.ResponseTime.Round(time.Millisecond))
				cmd.Printf("   Uptime: %.1f%%\n", health.Uptime)
				cmd.Printf("   Total Requests: %d\n", health.TotalRequests)
				if health.ErrorCount > 0 {
					cmd.Printf("   Errors: %d\n", health.ErrorCount)
				}

				if health.Details != nil {
					if successful, ok := health.Details["successful_checks"].(int); ok {
						if total, ok := health.Details["total_checks"].(int); ok {
							cmd.Printf("   Health Checks: %d/%d passed\n", successful, total)
						}
					}
				}
			}

			return nil
		},
	}
}

func newHealthDatabaseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "database",
		Short: "Database connectivity and health",
		Long:  `Check database connectivity, performance, and basic health metrics.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("🗄️  Database Health Check")
			cmd.Println("========================")

			// For now, we'll check if we can reach the database through the services
			// In a real implementation, you might want to add direct database connectivity checks

			monitor := monitoring.NewMonitor(appConfig, serviceReg, apiClient)

			// Check if user service can access the database by attempting a simple query
			start := time.Now()
			resp, err := apiClient.CallService("user-service", "GET", "/api/v1/users?limit=1", nil, nil)
			responseTime := time.Since(start)

			if jsonOut {
				result := map[string]interface{}{
					"database_check": map[string]interface{}{
						"service":       "user-service",
						"endpoint":      "/api/v1/users?limit=1",
						"response_time": responseTime.String(),
						"status":        "unknown",
						"can_connect":   false,
					},
				}

				if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
					result["database_check"].(map[string]interface{})["status"] = "healthy"
					result["database_check"].(map[string]interface{})["can_connect"] = true
				} else if err != nil {
					result["database_check"].(map[string]interface{})["error"] = err.Error()
				}

				return printJSON(result)
			}

			if err != nil {
				cmd.Printf("❌ Database connectivity check failed\n")
				cmd.Printf("   Error: %v\n", err)
				cmd.Printf("   Response Time: %v\n", responseTime.Round(time.Millisecond))
				return nil
			}

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				cmd.Printf("✅ Database connectivity: HEALTHY\n")
				cmd.Printf("   Response Time: %v\n", responseTime.Round(time.Millisecond))
				cmd.Printf("   Status Code: %d\n", resp.StatusCode)

				// Try to get some basic database info
				health, err := monitor.CheckServiceHealth("user-service")
				if err == nil && health.Details != nil {
					cmd.Printf("   Service Uptime: %.1f%%\n", health.Uptime)
				}
			} else {
				cmd.Printf("⚠️  Database connectivity: DEGRADED\n")
				cmd.Printf("   Status Code: %d\n", resp.StatusCode)
				cmd.Printf("   Response Time: %v\n", responseTime.Round(time.Millisecond))
				if resp.Error != "" {
					cmd.Printf("   Error: %s\n", resp.Error)
				}
			}

			cmd.Println("\n💡 Note: This check verifies database connectivity through the user service.")
			cmd.Println("   For direct database checks, consider adding database-specific endpoints.")

			return nil
		},
	}
}

func newHealthDependenciesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dependencies",
		Short: "Check service dependencies",
		Long:  `Analyze service dependencies and their health status.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("🔗 Service Dependencies Analysis")
			cmd.Println("===============================")

			services := serviceReg.GetAllServices()
			monitor := monitoring.NewMonitor(appConfig, serviceReg, apiClient)

			// Define service dependencies (this could be made configurable)
			dependencies := map[string][]string{
				"user-service": {"database"},
				"api-gateway":  {"user-service"},
			}

			if jsonOut {
				result := map[string]interface{}{
					"services":     make(map[string]interface{}),
					"dependencies": dependencies,
				}

				for _, service := range services {
					health, err := monitor.CheckServiceHealth(service.Name)
					serviceInfo := map[string]interface{}{
						"status": "unknown",
					}

					if err != nil {
						serviceInfo["error"] = err.Error()
					} else {
						serviceInfo["status"] = health.Status
						serviceInfo["response_time"] = health.ResponseTime.String()
						serviceInfo["uptime"] = health.Uptime
					}

					// Add dependency information
					if deps, exists := dependencies[service.Name]; exists {
						serviceInfo["dependencies"] = deps
					}

					result["services"].(map[string]interface{})[service.Name] = serviceInfo
				}

				return printJSON(result)
			}

			cmd.Printf("📊 Analyzing %d services and their dependencies...\n\n", len(services))

			for _, service := range services {
				cmd.Printf("🔍 Service: %s\n", service.Name)
				cmd.Printf("   URL: %s\n", service.URL)

				// Check service health
				health, err := monitor.CheckServiceHealth(service.Name)
				if err != nil {
					cmd.Printf("   ❌ Health Check: FAILED (%v)\n", err)
				} else {
					statusIcon := "❌"
					if health.Status == "healthy" {
						statusIcon = "✅"
					} else if health.Status == "degraded" {
						statusIcon = "⚠️"
					}
					cmd.Printf("   %s Health: %s\n", statusIcon, health.Status)
					cmd.Printf("   Response Time: %v\n", health.ResponseTime.Round(time.Millisecond))
				}

				// Check dependencies
				if deps, exists := dependencies[service.Name]; exists {
					cmd.Printf("   🔗 Dependencies: %v\n", deps)

					// For database dependency, we can check if the service can access it
					for _, dep := range deps {
						if dep == "database" {
							if err == nil && health != nil && health.Status == "healthy" {
								cmd.Printf("   ✅ Database: Accessible\n")
							} else {
								cmd.Printf("   ❌ Database: Not accessible\n")
							}
						}
					}
				} else {
					cmd.Printf("   ℹ️  Dependencies: None defined\n")
				}

				cmd.Println()
			}

			cmd.Println("💡 Dependency analysis helps identify service coupling and potential failure points.")
			cmd.Println("   Consider adding dependency definitions to improve monitoring accuracy.")

			return nil
		},
	}
}
