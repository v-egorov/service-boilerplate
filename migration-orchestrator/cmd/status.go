package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/v-egorov/service-boilerplate/migration-orchestrator/internal/orchestrator"
	"github.com/v-egorov/service-boilerplate/migration-orchestrator/pkg/types"
)

func newStatusCmd() *cobra.Command {
	var showHistory bool
	var showDependencies bool

	cmd := &cobra.Command{
		Use:   "status <service-name>",
		Short: "Show migration status",
		Long:  `Display the current migration status for the specified service, including applied migrations, pending migrations, and execution history.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceName := args[0]

			logger.Info("📊 Migration status for service:", serviceName)

			// Create orchestrator
			orch, err := orchestrator.NewOrchestrator(db, logger, serviceName)
			if err != nil {
				return fmt.Errorf("failed to create orchestrator: %w", err)
			}

			// Get migration state
			ctx := context.Background()
			state, err := orch.GetMigrationState(ctx)
			if err != nil {
				return fmt.Errorf("failed to get migration state: %w", err)
			}

			// Load configuration for additional details
			migrationConfig, depConfig, err := orch.LoadMigrationConfig()
			if err != nil {
				logger.Warnf("Could not load migration config: %v", err)
			}

			// Display status
			if jsonOutput {
				return displayStatusJSON(state, migrationConfig, depConfig)
			} else {
				return displayStatusTable(state, migrationConfig, depConfig, showHistory, showDependencies)
			}
		},
	}

	cmd.Flags().BoolVar(&showHistory, "history", false, "Show detailed execution history")
	cmd.Flags().BoolVar(&showDependencies, "deps", false, "Show migration dependencies")

	return cmd
}

func displayStatusJSON(state *types.ServiceMigrationState, migrationConfig *types.MigrationConfig, depConfig *types.DependencyConfig) error {
	status := map[string]interface{}{
		"service_name":    state.ServiceName,
		"schema_name":     state.SchemaName,
		"current_version": state.CurrentVersion,
		"applied_count":   state.AppliedCount,
		"failed_count":    state.FailedCount,
		"pending_count":   state.PendingCount,
		"executions":      state.Executions,
	}

	if migrationConfig != nil {
		status["environments"] = migrationConfig.Environments
	}

	if depConfig != nil {
		status["dependencies"] = depConfig
	}

	return json.NewEncoder(os.Stdout).Encode(status)
}

func displayStatusTable(state *types.ServiceMigrationState, migrationConfig *types.MigrationConfig, depConfig *types.DependencyConfig, showHistory, showDependencies bool) error {
	fmt.Printf("📊 Migration Status for %s\n", state.ServiceName)
	fmt.Printf("Schema: %s\n", state.SchemaName)
	fmt.Println(strings.Repeat("=", 50))

	// Summary
	fmt.Printf("Current Version: %s\n", state.CurrentVersion)
	fmt.Printf("Applied: %d | Failed: %d | Pending: %d\n", state.AppliedCount, state.FailedCount, state.PendingCount)

	if state.LastMigration != nil {
		fmt.Printf("Last Migration: %s (%s)\n", state.LastMigration.MigrationID, state.LastMigration.Status)
		if state.LastMigration.CompletedAt != nil {
			fmt.Printf("Completed At: %s\n", state.LastMigration.CompletedAt.Format("2006-01-02 15:04:05"))
		}
	}

	fmt.Println()

	// Recent executions
	if showHistory && len(state.Executions) > 0 {
		fmt.Println("📝 Recent Migration History:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "Migration ID\tStatus\tApplied At\tDuration\tEnvironment")
		fmt.Fprintln(w, "------------\t------\t----------\t--------\t-----------")

		// Sort executions by applied_at desc
		executions := make([]types.MigrationExecution, len(state.Executions))
		copy(executions, state.Executions)
		sort.Slice(executions, func(i, j int) bool {
			if executions[i].CreatedAt.After(executions[j].CreatedAt) {
				return true
			}
			return false
		})

		for _, exec := range executions[:min(10, len(executions))] {
			appliedAt := "N/A"
			if exec.CompletedAt != nil {
				appliedAt = exec.CompletedAt.Format("2006-01-02 15:04:05")
			}

			duration := "N/A"
			if exec.DurationMs != nil {
				duration = fmt.Sprintf("%dms", *exec.DurationMs)
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", exec.MigrationID, exec.Status, appliedAt, duration, exec.Environment)
		}
		w.Flush()
		fmt.Println()
	}

	// Dependencies
	if showDependencies && depConfig != nil {
		fmt.Println("🔗 Migration Dependencies:")
		for migrationID, info := range depConfig.Migrations {
			fmt.Printf("  %s: %s\n", migrationID, info.Description)
			if len(info.DependsOn) > 0 {
				fmt.Printf("    Depends on: %v\n", info.DependsOn)
			}
			if len(info.AffectsTables) > 0 {
				fmt.Printf("    Affects: %v\n", info.AffectsTables)
			}
			fmt.Printf("    Risk: %s | Duration: %s\n", info.RiskLevel, info.EstimatedDuration)
			fmt.Println()
		}
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
