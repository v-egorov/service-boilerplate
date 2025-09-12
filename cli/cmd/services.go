package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// newServicesCmd creates the services command
func newServicesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "services",
		Short: "Service management and interaction",
		Long:  `Manage and interact with services in the boilerplate.`,
	}

	cmd.AddCommand(
		newServicesListCmd(),
		newServicesStatusCmd(),
		newServicesCallCmd(),
		newServicesProxyCmd(),
	)

	return cmd
}

// printJSON prints data as JSON
func printJSON(data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsonData))
	return nil
}

func newServicesListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all registered services",
		RunE: func(cmd *cobra.Command, args []string) error {
			services, err := serviceReg.DiscoverServices()
			if err != nil {
				return fmt.Errorf("failed to discover services: %w", err)
			}

			if jsonOut {
				return printJSON(services)
			}

			cmd.Println("Available Services:")
			cmd.Println("==================")
			for _, service := range services {
				status := "❌"
				if service.Status == "healthy" {
					status = "✅"
				}
				cmd.Printf("%s %s: %s (%s)\n", status, service.Name, service.URL, service.Status)
			}

			return nil
		},
	}
}

func newServicesStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status [service]",
		Short: "Get service status and health",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				// Check specific service
				serviceName := args[0]
				service, err := serviceReg.GetService(serviceName)
				if err != nil {
					return fmt.Errorf("failed to get service %s: %w", serviceName, err)
				}

				if jsonOut {
					return printJSON(service)
				}

				status := "❌"
				if service.Status == "healthy" {
					status = "✅"
				}
				cmd.Printf("Service: %s\n", service.Name)
				cmd.Printf("URL: %s\n", service.URL)
				cmd.Printf("Status: %s %s\n", status, service.Status)
			} else {
				// Check all services
				services, err := serviceReg.DiscoverServices()
				if err != nil {
					return fmt.Errorf("failed to discover services: %w", err)
				}

				if jsonOut {
					return printJSON(services)
				}

				cmd.Println("Service Status:")
				cmd.Println("===============")
				for _, service := range services {
					status := "❌"
					if service.Status == "healthy" {
						status = "✅"
					}
					cmd.Printf("%s %s: %s\n", status, service.Name, service.Status)
				}
			}

			return nil
		},
	}
}

func newServicesCallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "call <service> <method> <endpoint>",
		Short: "Direct API call to service",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			service := args[0]
			method := args[1]
			endpoint := args[2]

			resp, err := apiClient.CallService(service, method, endpoint, nil, nil)
			if err != nil {
				return fmt.Errorf("API call failed: %w", err)
			}

			if jsonOut {
				return printJSON(resp)
			}

			cmd.Printf("Response Status: %d\n", resp.StatusCode)
			if resp.Body != nil {
				cmd.Printf("Response Body: %v\n", resp.Body)
			}

			return nil
		},
	}

	cmd.Flags().StringP("body", "b", "", "Request body (JSON)")
	cmd.Flags().StringSliceP("header", "H", nil, "Request headers (key=value)")

	return cmd
}

func newServicesProxyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "proxy <service> <method> <path>",
		Short: "Proxy request through gateway",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement gateway proxy call
			service := args[0]
			method := args[1]
			path := args[2]
			cmd.Printf("Proxying %s request to %s at path %s\n", method, service, path)
			return nil
		},
	}
}
