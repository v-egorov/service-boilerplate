package cmd

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/v-egorov/service-boilerplate/cli/internal/workflows"
)

// newOpsCmd creates the ops command
func newOpsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ops",
		Short: "Business operations and workflows",
		Long:  `Execute business operations and predefined workflows.`,
	}

	cmd.AddCommand(
		newOpsUserCmd(),
		newOpsWorkflowCmd(),
		newOpsMultiCmd(),
	)

	return cmd
}

func newOpsUserCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "User operations",
	}

	cmd.AddCommand(
		newOpsUserCreateCmd(),
		newOpsUserListCmd(),
		newOpsUserUpdateCmd(),
		newOpsUserDeleteCmd(),
	)

	return cmd
}

func newOpsMultiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multi",
		Short: "Multi-service operations",
		Long:  `Execute operations that span multiple services simultaneously.`,
	}

	cmd.AddCommand(
		newOpsMultiExecuteCmd(),
		newOpsMultiListCmd(),
	)

	return cmd
}

func newOpsMultiExecuteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "execute <operation>",
		Short: "Execute a multi-service operation",
		Long:  `Execute a predefined multi-service operation with coordinated steps across services.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			operationName := args[0]

			executor := workflows.NewMultiServiceExecutor(appConfig, apiClient)
			operations := executor.GetPredefinedOperations()

			operation, exists := operations[operationName]
			if !exists {
				return fmt.Errorf("multi-service operation '%s' not found", operationName)
			}

			result, err := executor.ExecuteOperation(operation)
			if err != nil {
				return fmt.Errorf("failed to execute multi-service operation: %w", err)
			}

			if jsonOut {
				return printJSON(result)
			}

			return nil
		},
	}
}

func newOpsMultiListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available multi-service operations",
		Long:  `Display all available multi-service operations with their descriptions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			executor := workflows.NewMultiServiceExecutor(appConfig, apiClient)
			operations := executor.GetPredefinedOperations()

			if jsonOut {
				return printJSON(operations)
			}

			cmd.Println("🔧 Available Multi-Service Operations:")
			cmd.Println("=====================================")

			for name, operation := range operations {
				cmd.Printf("• %s\n", name)
				cmd.Printf("  📝 %s\n", operation.Description)
				cmd.Printf("  🎯 Services: %v\n", operation.Services)
				cmd.Printf("  📊 Steps: %d\n", len(operation.Steps))
				if operation.Parallel {
					cmd.Printf("  ⚡ Execution: Parallel\n")
				} else {
					cmd.Printf("  🔄 Execution: Sequential\n")
				}
				cmd.Println()
			}

			return nil
		},
	}
}

func newOpsUserCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <email> <first_name> <last_name>",
		Short: "Create user via user-service API",
		Long:  `Create a new user in the user service with the specified email and name.`,
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			email := args[0]
			firstName := args[1]
			lastName := args[2]

			// Prepare request body
			userData := map[string]interface{}{
				"email":      email,
				"first_name": firstName,
				"last_name":  lastName,
			}

			// Make API call to user service
			resp, err := apiClient.CallService("user-service", "POST", "/api/v1/users", userData, nil)
			if err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}

			if jsonOut {
				return printJSON(resp.Body)
			}

			if resp.StatusCode == http.StatusCreated {
				cmd.Printf("✅ User created successfully!\n")
				if user, ok := resp.Body.(map[string]interface{}); ok {
					if id, exists := user["id"]; exists {
						cmd.Printf("User ID: %v\n", id)
					}
					if email, exists := user["email"]; exists {
						cmd.Printf("Email: %v\n", email)
					}
					if firstName, exists := user["first_name"]; exists {
						cmd.Printf("First Name: %v\n", firstName)
					}
					if lastName, exists := user["last_name"]; exists {
						cmd.Printf("Last Name: %v\n", lastName)
					}
				}
			} else {
				cmd.Printf("❌ Failed to create user (Status: %d)\n", resp.StatusCode)
				if resp.Error != "" {
					cmd.Printf("Error: %s\n", resp.Error)
				}
			}

			return nil
		},
	}

	return cmd
}

func newOpsUserListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List users with pagination",
		Long:  `List users from the user service with optional pagination.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			limit, _ := cmd.Flags().GetInt("limit")
			offset, _ := cmd.Flags().GetInt("offset")

			// Build query parameters
			endpoint := fmt.Sprintf("/api/v1/users?limit=%d&offset=%d", limit, offset)

			// Make API call to user service
			resp, err := apiClient.CallService("user-service", "GET", endpoint, nil, nil)
			if err != nil {
				return fmt.Errorf("failed to list users: %w", err)
			}

			if jsonOut {
				return printJSON(resp.Body)
			}

			if resp.StatusCode == http.StatusOK {
				if response, ok := resp.Body.(map[string]interface{}); ok {
					if users, exists := response["users"]; exists {
						if userList, ok := users.([]interface{}); ok {
							cmd.Printf("📋 Users (showing %d, offset %d):\n", limit, offset)
							cmd.Println("=====================================")

							for i, user := range userList {
								if userMap, ok := user.(map[string]interface{}); ok {
									id := userMap["id"]
									email := userMap["email"]
									firstName := userMap["first_name"]
									lastName := userMap["last_name"]

									cmd.Printf("%d. %v - %s %s (%s)\n",
										offset+i+1, id, firstName, lastName, email)
								}
							}

							// Show pagination info
							if len(userList) == limit {
								cmd.Printf("\n💡 Use --offset %d to see more users\n", offset+limit)
							}
						}
					}
				}
			} else {
				cmd.Printf("❌ Failed to list users (Status: %d)\n", resp.StatusCode)
				if resp.Error != "" {
					cmd.Printf("Error: %s\n", resp.Error)
				}
			}

			return nil
		},
	}

	cmd.Flags().IntP("limit", "l", 10, "Number of users to retrieve")
	cmd.Flags().IntP("offset", "o", 0, "Offset for pagination")

	return cmd
}

func newOpsUserUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id> <email> <first_name> <last_name>",
		Short: "Update user via API",
		Long:  `Update an existing user with new information.`,
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID := args[0]
			email := args[1]
			firstName := args[2]
			lastName := args[3]

			// Prepare request body
			userData := map[string]interface{}{
				"email":      email,
				"first_name": firstName,
				"last_name":  lastName,
			}

			// Build endpoint
			endpoint := fmt.Sprintf("/api/v1/users/%s", userID)

			// Make API call to user service
			resp, err := apiClient.CallService("user-service", "PUT", endpoint, userData, nil)
			if err != nil {
				return fmt.Errorf("failed to update user: %w", err)
			}

			if jsonOut {
				return printJSON(resp.Body)
			}

			if resp.StatusCode == http.StatusOK {
				cmd.Printf("✅ User %s updated successfully!\n", userID)
				if user, ok := resp.Body.(map[string]interface{}); ok {
					if email, exists := user["email"]; exists {
						cmd.Printf("Email: %v\n", email)
					}
					if firstName, exists := user["first_name"]; exists {
						cmd.Printf("First Name: %v\n", firstName)
					}
					if lastName, exists := user["last_name"]; exists {
						cmd.Printf("Last Name: %v\n", lastName)
					}
				}
			} else {
				cmd.Printf("❌ Failed to update user (Status: %d)\n", resp.StatusCode)
				if resp.Error != "" {
					cmd.Printf("Error: %s\n", resp.Error)
				}
			}

			return nil
		},
	}

	return cmd
}

func newOpsUserDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete user via API",
		Long:  `Delete an existing user from the system.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID := args[0]

			// Confirm deletion unless --force flag is used
			force, _ := cmd.Flags().GetBool("force")
			if !force {
				cmd.Printf("⚠️  Are you sure you want to delete user %s? (This action cannot be undone)\n", userID)
				cmd.Printf("Use --force to skip this confirmation.\n")
				return fmt.Errorf("confirmation required (use --force to skip)")
			}

			// Build endpoint
			endpoint := fmt.Sprintf("/api/v1/users/%s", userID)

			// Make API call to user service
			resp, err := apiClient.CallService("user-service", "DELETE", endpoint, nil, nil)
			if err != nil {
				return fmt.Errorf("failed to delete user: %w", err)
			}

			if jsonOut {
				return printJSON(resp)
			}

			if resp.StatusCode == http.StatusNoContent {
				cmd.Printf("✅ User %s deleted successfully!\n", userID)
			} else {
				cmd.Printf("❌ Failed to delete user (Status: %d)\n", resp.StatusCode)
				if resp.Error != "" {
					cmd.Printf("Error: %s\n", resp.Error)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	return cmd
}

func newOpsWorkflowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow <name>",
		Short: "Execute predefined business workflow",
		Long: `Execute a predefined workflow that performs multiple operations in sequence.

Use 'list' as the name to see all available workflows.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Create workflow executor
			workflowExecutor := workflows.NewWorkflowExecutor(appConfig, apiClient)

			// List available workflows if requested
			if name == "list" {
				availableWorkflows := workflowExecutor.ListWorkflows()

				if jsonOut {
					return printJSON(availableWorkflows)
				}

				cmd.Println("📋 Available Workflows:")
				cmd.Println("=======================")
				for _, workflow := range availableWorkflows {
					cmd.Printf("• %s: %s\n", workflow.Name, workflow.Description)
					cmd.Printf("  Steps: %d\n", len(workflow.Steps))
				}
				return nil
			}

			// Execute the workflow
			err := workflowExecutor.ExecuteWorkflow(name)
			if err != nil {
				return fmt.Errorf("workflow execution failed: %w", err)
			}

			return nil
		},
	}

	return cmd
}
