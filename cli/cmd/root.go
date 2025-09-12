package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/v-egorov/service-boilerplate/cli/internal/client"
	"github.com/v-egorov/service-boilerplate/cli/internal/config"
	"github.com/v-egorov/service-boilerplate/cli/internal/discovery"
)

var (
	cfgFile    string
	env        string
	verbose    bool
	jsonOut    bool
	appConfig  *config.Config
	serviceReg *discovery.ServiceRegistry
	apiClient  *client.APIClient
)

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "boilerplate-cli",
		Short: "CLI utility for service boilerplate operations",
		Long: `A CLI utility for automating business-logic related operations
in the service boilerplate. Provides tools for service interaction,
data management, and operational workflows.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig()
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.boilerplate-cli.yaml)")
	rootCmd.PersistentFlags().StringVarP(&env, "env", "e", "development", "environment (development/production)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "output in JSON format")

	// Add subcommands
	rootCmd.AddCommand(
		newServicesCmd(),
		newDataCmd(),
		newOpsCmd(),
		newDevCmd(),
		newHealthCmd(),
	)

	return rootCmd
}

// initConfig reads in config file and ENV variables if set
func initConfig() error {
	var err error
	appConfig, err = config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override environment from flag if provided
	if env != "" {
		appConfig.Environment = env
	}

	// Initialize service registry and API client
	serviceReg = discovery.NewServiceRegistry(appConfig)
	apiClient = client.NewAPIClient(appConfig)

	if verbose {
		fmt.Fprintf(os.Stderr, "Environment: %s\n", appConfig.Environment)
		fmt.Fprintf(os.Stderr, "Gateway URL: %s\n", appConfig.Services.GatewayURL)
	}

	return nil
}
