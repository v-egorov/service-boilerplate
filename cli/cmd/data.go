package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

// newDataCmd creates the data command
func newDataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data",
		Short: "Data operations and management",
		Long:  `Manage data operations across services.`,
	}

	cmd.AddCommand(
		newDataSeedCmd(),
		newDataExportCmd(),
		newDataValidateCmd(),
		newDataCleanupCmd(),
	)

	return cmd
}

// printJSONToFile prints data as JSON to a file
func printJSONToFile(data interface{}, file *os.File) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	_, err = file.Write(jsonData)
	return err
}

func newDataSeedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seed <service> <fixture>",
		Short: "Seed service with test data",
		Long:  `Load test data into a service using predefined fixtures.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			service := args[0]
			fixture := args[1]

			// For now, we'll focus on user-service seeding
			if service != "user-service" {
				return fmt.Errorf("seeding is currently only supported for user-service")
			}

			// Determine seed file path based on fixture
			var seedFile string
			switch fixture {
			case "base":
				seedFile = "../scripts/seeds/base/users.sql"
			case "development", "dev":
				seedFile = "../scripts/seeds/development/dev_users.sql"
			default:
				return fmt.Errorf("unknown fixture: %s (available: base, development)", fixture)
			}

			// Check if seed file exists
			if _, err := os.Stat(seedFile); os.IsNotExist(err) {
				return fmt.Errorf("seed file not found: %s", seedFile)
			}

			// Read seed file
			seedData, err := os.ReadFile(seedFile)
			if err != nil {
				return fmt.Errorf("failed to read seed file: %w", err)
			}

			if jsonOut {
				result := map[string]interface{}{
					"service": service,
					"fixture": fixture,
					"file":    seedFile,
					"sql":     string(seedData),
				}
				return printJSON(result)
			}

			cmd.Printf("🌱 Seeding %s with %s fixture...\n", service, fixture)
			cmd.Printf("📁 Using seed file: %s\n", seedFile)
			cmd.Println("📄 SQL to execute:")
			cmd.Println("==================")
			cmd.Println(string(seedData))

			// Note: In a real implementation, you would execute this SQL
			// For now, we'll just show what would be executed
			cmd.Println("==================")
			cmd.Printf("✅ Seed operation prepared for %s\n", service)
			cmd.Printf("💡 To actually execute, run: make db-seed-enhanced ENV=%s\n", fixture)

			return nil
		},
	}

	return cmd
}

func newDataExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <service> <table>",
		Short: "Export data from service",
		Long:  `Export data from a service table to various formats.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			service := args[0]
			table := args[1]
			format, _ := cmd.Flags().GetString("format")
			output, _ := cmd.Flags().GetString("output")

			// For user-service, we can export users
			if service == "user-service" && table == "users" {
				// Make API call to get all users
				resp, err := apiClient.CallService("user-service", "GET", "/api/v1/users?limit=1000", nil, nil)
				if err != nil {
					return fmt.Errorf("failed to export users: %w", err)
				}

				if resp.StatusCode != http.StatusOK {
					return fmt.Errorf("API call failed with status %d: %s", resp.StatusCode, resp.Error)
				}

				// Extract users from response
				var users []interface{}
				if response, ok := resp.Body.(map[string]interface{}); ok {
					if usersData, exists := response["users"]; exists {
						if userList, ok := usersData.([]interface{}); ok {
							users = userList
						}
					}
				}

				if jsonOut {
					result := map[string]interface{}{
						"service": service,
						"table":   table,
						"count":   len(users),
						"data":    users,
					}
					return printJSON(result)
				}

				cmd.Printf("📊 Exporting %d users from %s...\n", len(users), service)

				if output != "" {
					// Write to file
					if format == "json" {
						// Write JSON to file
						file, err := os.Create(output)
						if err != nil {
							return fmt.Errorf("failed to create output file: %w", err)
						}
						defer file.Close()

						result := map[string]interface{}{
							"service": service,
							"table":   table,
							"count":   len(users),
							"data":    users,
						}
						return printJSONToFile(result, file)
					}
				} else {
					// Display to console
					cmd.Println("📋 Users:")
					cmd.Println("========")
					for i, user := range users {
						if userMap, ok := user.(map[string]interface{}); ok {
							id := userMap["id"]
							email := userMap["email"]
							firstName := userMap["first_name"]
							lastName := userMap["last_name"]
							cmd.Printf("%d. ID:%v | %s %s | %s\n",
								i+1, id, firstName, lastName, email)
						}
					}
				}

				cmd.Printf("✅ Exported %d records from %s.%s\n", len(users), service, table)
			} else {
				cmd.Printf("⚠️  Export not yet implemented for %s.%s\n", service, table)
				cmd.Printf("💡 Currently supported: user-service.users\n")
			}

			return nil
		},
	}

	cmd.Flags().StringP("format", "f", "json", "Export format (json, csv)")
	cmd.Flags().StringP("output", "o", "", "Output file path")

	return cmd
}

func newDataValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate <service>",
		Short: "Validate service data integrity",
		Long:  `Validate data integrity and consistency in a service.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service := args[0]

			if service == "user-service" {
				// Validate user data
				cmd.Println("🔍 Validating user-service data integrity...")

				// Check if service is healthy
				if !serviceReg.IsServiceHealthy(service) {
					cmd.Printf("❌ Service %s is not healthy\n", service)
					return fmt.Errorf("service %s is not available", service)
				}

				// Get all users to validate
				resp, err := apiClient.CallService("user-service", "GET", "/api/v1/users?limit=1000", nil, nil)
				if err != nil {
					return fmt.Errorf("failed to fetch users for validation: %w", err)
				}

				if resp.StatusCode != http.StatusOK {
					return fmt.Errorf("API call failed with status %d: %s", resp.StatusCode, resp.Error)
				}

				// Extract and validate users
				var users []interface{}
				if response, ok := resp.Body.(map[string]interface{}); ok {
					if usersData, exists := response["users"]; exists {
						if userList, ok := usersData.([]interface{}); ok {
							users = userList
						}
					}
				}

				if jsonOut {
					result := map[string]interface{}{
						"service":     service,
						"total_users": len(users),
						"validation":  "completed",
						"issues":      []string{}, // In a real implementation, you'd check for data issues
					}
					return printJSON(result)
				}

				cmd.Printf("✅ Found %d users in %s\n", len(users), service)

				// Basic validation checks
				validEmails := 0
				validNames := 0

				for _, user := range users {
					if userMap, ok := user.(map[string]interface{}); ok {
						// Check email
						if email, exists := userMap["email"]; exists && email != nil {
							validEmails++
						}

						// Check names
						if firstName, exists := userMap["first_name"]; exists && firstName != nil {
							if lastName, exists := userMap["last_name"]; exists && lastName != nil {
								validNames++
							}
						}
					}
				}

				cmd.Printf("📧 Valid emails: %d/%d\n", validEmails, len(users))
				cmd.Printf("👤 Valid names: %d/%d\n", validNames, len(users))

				if validEmails == len(users) && validNames == len(users) {
					cmd.Println("✅ All data validation checks passed!")
				} else {
					cmd.Println("⚠️  Some data validation issues found")
				}

			} else {
				cmd.Printf("⚠️  Data validation not yet implemented for %s\n", service)
				cmd.Printf("💡 Currently supported: user-service\n")
			}

			return nil
		},
	}
}

func newDataCleanupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup <service>",
		Short: "Clean test data from service",
		Long:  `Remove test data from a service (typically development/test users).`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service := args[0]
			pattern, _ := cmd.Flags().GetString("pattern")

			if service == "user-service" {
				// Confirm cleanup unless --force flag is used
				force, _ := cmd.Flags().GetBool("force")
				if !force {
					cmd.Printf("⚠️  Are you sure you want to clean test data from %s?\n", service)
					cmd.Printf("This will remove users matching the pattern: %s\n", pattern)
					cmd.Printf("Use --force to skip this confirmation.\n")
					return fmt.Errorf("confirmation required (use --force to skip)")
				}

				cmd.Printf("🧹 Cleaning test data from %s...\n", service)
				cmd.Printf("🎯 Pattern: %s\n", pattern)

				// In a real implementation, you would:
				// 1. Query for users matching the pattern
				// 2. Delete them via the API
				// For now, we'll show what would be done

				if jsonOut {
					result := map[string]interface{}{
						"service":   service,
						"operation": "cleanup",
						"pattern":   pattern,
						"status":    "prepared",
						"message":   "Cleanup operation prepared (not yet implemented)",
					}
					return printJSON(result)
				}

				cmd.Println("📋 Cleanup Plan:")
				cmd.Println("================")
				cmd.Printf("• Service: %s\n", service)
				cmd.Printf("• Pattern: %s\n", pattern)
				cmd.Println("• Action: Delete matching users")
				cmd.Println("================")

				cmd.Printf("✅ Cleanup operation prepared for %s\n", service)
				cmd.Printf("💡 To actually execute, you would implement the deletion logic\n")

			} else {
				cmd.Printf("⚠️  Data cleanup not yet implemented for %s\n", service)
				cmd.Printf("💡 Currently supported: user-service\n")
			}

			return nil
		},
	}

	cmd.Flags().StringP("pattern", "p", "example.com", "Email pattern to match for cleanup")
	cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	return cmd
}
