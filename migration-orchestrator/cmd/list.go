package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/v-egorov/service-boilerplate/migration-orchestrator/internal/orchestrator"
	"github.com/v-egorov/service-boilerplate/migration-orchestrator/pkg/types"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list <service-name>",
		Short:        "List all migrations",
		Long:         `List all migrations for the specified service with their status.`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceName := args[0]

			logger.Info("📋 All migrations for service:", serviceName)

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

			// Load environments configuration
			migrationConfig, err := orch.LoadMigrationConfig()
			if err != nil {
				logger.Warnf("Could not load migration config: %v", err)
			}

			// Display list
			if jsonOutput {
				return displayListJSON(state, migrationConfig, orch.ServicePath())
			} else {
				return displayListTable(state, migrationConfig, orch.ServicePath())
			}
		},
	}

	return cmd
}

func displayListJSON(state *types.ServiceMigrationState, migrationConfig *types.MigrationConfig, servicePath string) error {
	type MigrationInfo struct {
		ID        string `json:"id"`
		Status    string `json:"status"`
		Type      string `json:"type"`
		AppliedAt string `json:"applied_at,omitempty"`
	}

	var migrations []MigrationInfo

	// Add applied migrations
	for _, exec := range state.Executions {
		info := MigrationInfo{
			ID:     exec.MigrationID,
			Status: string(exec.Status),
			Type:   "applied",
		}
		if exec.CompletedAt != nil {
			info.AppliedAt = exec.CompletedAt.Format("2006-01-02 15:04:05")
		}
		migrations = append(migrations, info)
	}

	// Add pending migrations from environments.json
	if migrationConfig != nil && migrationConfig.Environments != nil {
		currentEnv := appConfig.Environment
		envConfig, exists := migrationConfig.Environments[currentEnv]
		if exists {
			// envConfig.Migrations is now a directory name (string)
			migrationDir := envConfig.Migrations

			// List files in the migration directory
			migrationPath := filepath.Join(servicePath, "migrations", migrationDir)
			files, err := os.ReadDir(migrationPath)
			if err == nil {
				for _, file := range files {
					if !file.IsDir() && strings.HasPrefix(file.Name(), "000") && strings.HasSuffix(file.Name(), ".up.sql") {
						filename := file.Name()
						migrationID := filename[:6]

						alreadyApplied := false
						for _, exec := range state.Executions {
							if exec.MigrationID == migrationID && exec.Status == "completed" {
								alreadyApplied = true
								break
							}
						}
						if !alreadyApplied {
							migrations = append(migrations, MigrationInfo{
								ID:     migrationID,
								Status: "pending",
								Type:   fmt.Sprintf("env-%s", currentEnv),
							})
						}
					}
				}
			}
		}
	}

	return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
		"service_name": state.ServiceName,
		"migrations":   migrations,
	})
}

func displayListTable(state *types.ServiceMigrationState, migrationConfig *types.MigrationConfig, servicePath string) error {
	fmt.Printf("📋 Migration List for %s\n", state.ServiceName)
	fmt.Printf("Schema: %s\n", state.SchemaName)
	fmt.Println(strings.Repeat("=", 60))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Migration ID\tStatus\tType")
	fmt.Fprintln(w, "------------\t------\t----")

	type MigrationEntry struct {
		ID     string
		Status string
		Type   string
	}

	var allMigrations []MigrationEntry

	// Add applied migrations
	for _, exec := range state.Executions {
		entry := MigrationEntry{
			ID:     exec.MigrationID,
			Status: string(exec.Status),
			Type:   "applied",
		}
		allMigrations = append(allMigrations, entry)
	}

	// Add pending migrations from environments.json
	if migrationConfig != nil && migrationConfig.Environments != nil {
		currentEnv := appConfig.Environment
		envConfig, exists := migrationConfig.Environments[currentEnv]
		if exists {
			// envConfig.Migrations is now a directory name (string)
			migrationDir := envConfig.Migrations

			// List files in the migration directory
			migrationPath := filepath.Join(servicePath, "migrations", migrationDir)
			files, err := os.ReadDir(migrationPath)
			if err == nil {
				for _, file := range files {
					if !file.IsDir() && strings.HasPrefix(file.Name(), "000") && strings.HasSuffix(file.Name(), ".up.sql") {
						filename := file.Name()
						migrationID := filename[:6]

						alreadyApplied := false
						for _, exec := range state.Executions {
							if exec.MigrationID == migrationID && exec.Status == "completed" {
								alreadyApplied = true
								break
							}
						}
						if !alreadyApplied {
							allMigrations = append(allMigrations, MigrationEntry{
								ID:     migrationID,
								Status: "pending",
								Type:   fmt.Sprintf("env-%s", currentEnv),
							})
						}
					}
				}
			}
		}
	}

	// Sort by migration ID
	sort.Slice(allMigrations, func(i, j int) bool {
		return allMigrations[i].ID < allMigrations[j].ID
	})

	// Display migrations
	for _, migration := range allMigrations {
		fmt.Fprintf(w, "%s\t%s\t%s\n", migration.ID, migration.Status, migration.Type)
	}

	w.Flush()
	fmt.Printf("\nTotal: %d migrations (%d applied, %d pending)\n",
		len(allMigrations),
		len(state.Executions),
		len(allMigrations)-len(state.Executions))

	return nil
}
