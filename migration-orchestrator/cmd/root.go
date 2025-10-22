package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/v-egorov/service-boilerplate/migration-orchestrator/internal/config"
	"github.com/v-egorov/service-boilerplate/migration-orchestrator/internal/database"
	"github.com/v-egorov/service-boilerplate/migration-orchestrator/pkg/utils"
)

var (
	cfgFile     string
	environment string
	verbose     bool
	jsonOutput  bool
	appConfig   *config.Config
	db          *database.Database
	logger      *utils.Logger
)

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "migration-orchestrator",
		Short: "Database migration orchestrator for service boilerplate",
		Long: `A powerful migration orchestrator that provides enhanced tracking,
dependency management, and rollback capabilities for database migrations
across multiple services in a schema-per-service architecture.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig()
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if db != nil {
				db.Close()
			}
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.migration-orchestrator.yaml)")
	rootCmd.PersistentFlags().StringVarP(&environment, "env", "e", "development", "environment (development/staging/production)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")

	// Add subcommands
	rootCmd.AddCommand(
		newUpCmd(),
		newDownCmd(),
		newStatusCmd(),
		newListCmd(),
		newValidateCmd(),
		newInitCmd(),
		newResolveDependenciesCmd(),
	)

	return rootCmd
}

// initConfig reads in config file and ENV variables if set
func initConfig() error {
	var err error

	// Initialize logger first
	logger = utils.NewLogger(verbose, jsonOutput)

	logger.Info("Initializing migration orchestrator...")

	// Load configuration
	appConfig, err = config.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override environment from flag if provided
	if environment != "" {
		appConfig.Environment = environment
	}

	// Initialize database connection
	db, err = database.NewDatabase(appConfig.Database, logger)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	logger.Info("Configuration loaded successfully")
	logger.Info(fmt.Sprintf("Environment: %s", appConfig.Environment))
	logger.Info(fmt.Sprintf("Database: %s@%s:%d/%s", appConfig.Database.User, appConfig.Database.Host, appConfig.Database.Port, appConfig.Database.Database))

	return nil
}
