package orchestrator

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/v-egorov/service-boilerplate/migration-orchestrator/internal/database"
	"github.com/v-egorov/service-boilerplate/migration-orchestrator/pkg/types"
	"github.com/v-egorov/service-boilerplate/migration-orchestrator/pkg/utils"
)

// TestOrchestratorIntegration tests the full orchestrator workflow
func TestOrchestratorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/test_db?sslmode=disable"
	}

	// Connect to test database
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Skipf("Cannot connect to test database: %v", err)
	}
	defer pool.Close()

	// Create test schema
	testSchema := "test_service"
	_, err = pool.Exec(context.Background(), fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", testSchema))
	if err != nil {
		t.Fatalf("Failed to drop test schema: %v", err)
	}

	_, err = pool.Exec(context.Background(), fmt.Sprintf("CREATE SCHEMA %s", testSchema))
	if err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	// Set search path
	_, err = pool.Exec(context.Background(), fmt.Sprintf("SET search_path TO %s", testSchema))
	if err != nil {
		t.Fatalf("Failed to set search path: %v", err)
	}

	// Create test database connection
	testDB := &database.Database{Pool: pool}

	// Create test logger
	logger := utils.NewLogger(false, false)

	// Create orchestrator
	orch, err := NewOrchestrator(testDB, logger, "test-service")
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}

	// Override schema name for testing
	orch.schemaName = testSchema

	ctx := context.Background()

	// Test 1: Create migration executions table
	t.Run("CreateMigrationExecutionsTable", func(t *testing.T) {
		err := orch.CreateMigrationExecutionsTable(ctx)
		if err != nil {
			t.Fatalf("Failed to create migration executions table: %v", err)
		}

		// Verify table exists
		var exists bool
		err = pool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = $1 AND table_name = 'migration_executions')", testSchema).Scan(&exists)
		if err != nil {
			t.Fatalf("Failed to check table existence: %v", err)
		}
		if !exists {
			t.Error("Migration executions table was not created")
		}
	})

	// Test 2: Get migration state (should be empty initially)
	t.Run("GetMigrationState", func(t *testing.T) {
		state, err := orch.GetMigrationState(ctx)
		if err != nil {
			t.Fatalf("Failed to get migration state: %v", err)
		}

		if state.ServiceName != "test-service" {
			t.Errorf("Expected service name 'test-service', got '%s'", state.ServiceName)
		}

		if state.SchemaName != testSchema {
			t.Errorf("Expected schema name '%s', got '%s'", testSchema, state.SchemaName)
		}

		if len(state.Executions) != 0 {
			t.Errorf("Expected 0 executions initially, got %d", len(state.Executions))
		}
	})

	// Test 3: Test dependency resolution
	t.Run("ResolveDependencies", func(t *testing.T) {
		candidateMigrations := []string{"000001", "000002", "000003"}
		applied := map[string]bool{"000001": true}

		depConfig := &types.DependencyConfig{
			Migrations: map[string]types.MigrationInfo{
				"000001": {DependsOn: []string{}},
				"000002": {DependsOn: []string{"000001"}},
				"000003": {DependsOn: []string{"000002"}},
			},
		}

		result := orch.resolveDependencies(candidateMigrations, applied, depConfig)

		expected := []string{"000002", "000003"}
		if len(result) != len(expected) {
			t.Errorf("Expected %d migrations, got %d", len(expected), len(result))
		}

		// Check that 000001 is not in result (already applied)
		for _, mig := range result {
			if mig == "000001" {
				t.Error("000001 should not be in result (already applied)")
			}
		}
	})

	// Test 4: Test migration ID to version conversion
	t.Run("MigrationIDToVersion", func(t *testing.T) {
		version, err := orch.migrationIDToVersion("000042")
		if err != nil {
			t.Fatalf("Failed to convert migration ID: %v", err)
		}
		if version != 42 {
			t.Errorf("Expected version 42, got %d", version)
		}
	})

	// Cleanup
	t.Cleanup(func() {
		_, err := pool.Exec(context.Background(), fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", testSchema))
		if err != nil {
			t.Logf("Failed to cleanup test schema: %v", err)
		}
	})
}

// TestOrchestratorWithMockData tests orchestrator with mock migration data
func TestOrchestratorWithMockData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test would require setting up mock migration files and database state
	// For now, it's a placeholder for more comprehensive integration testing

	t.Run("MockMigrationWorkflow", func(t *testing.T) {
		// TODO: Implement full workflow test with mock data
		t.Skip("Full workflow test not yet implemented")
	})
}

// BenchmarkResolveDependencies benchmarks the dependency resolution algorithm
func BenchmarkResolveDependencies(b *testing.B) {
	o := &Orchestrator{}

	// Create a larger dependency graph for benchmarking
	candidateMigrations := make([]string, 100)
	applied := make(map[string]bool)
	depConfig := &types.DependencyConfig{
		Migrations: make(map[string]types.MigrationInfo),
	}

	for i := 1; i <= 100; i++ {
		migID := fmt.Sprintf("%06d", i)
		candidateMigrations[i-1] = migID

		dependsOn := []string{}
		if i > 1 {
			dependsOn = []string{fmt.Sprintf("%06d", i-1)}
		}

		depConfig.Migrations[migID] = types.MigrationInfo{
			DependsOn: dependsOn,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		o.resolveDependencies(candidateMigrations, applied, depConfig)
	}
}
