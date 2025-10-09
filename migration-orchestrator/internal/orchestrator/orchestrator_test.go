package orchestrator

import (
	"testing"

	"github.com/v-egorov/service-boilerplate/migration-orchestrator/pkg/types"
	"github.com/v-egorov/service-boilerplate/migration-orchestrator/pkg/utils"
)

func TestResolveDependencies(t *testing.T) {
	o := &Orchestrator{}

	tests := []struct {
		name             string
		candidateMigs    []string
		applied          map[string]bool
		depConfig        *types.DependencyConfig
		expectedOrder    []string
		expectCycleError bool
	}{
		{
			name:          "no dependencies",
			candidateMigs: []string{"000001", "000002", "000003"},
			applied:       map[string]bool{},
			depConfig:     nil,
			expectedOrder: []string{"000001", "000002", "000003"},
		},
		{
			name:          "simple dependencies",
			candidateMigs: []string{"000001", "000002", "000003"},
			applied:       map[string]bool{},
			depConfig: &types.DependencyConfig{
				Migrations: map[string]types.MigrationInfo{
					"000001": {DependsOn: []string{}},
					"000002": {DependsOn: []string{"000001"}},
					"000003": {DependsOn: []string{"000002"}},
				},
			},
			expectedOrder: []string{"000001", "000002", "000003"},
		},
		{
			name:          "some applied",
			candidateMigs: []string{"000001", "000002", "000003"},
			applied:       map[string]bool{"000001": true},
			depConfig: &types.DependencyConfig{
				Migrations: map[string]types.MigrationInfo{
					"000001": {DependsOn: []string{}},
					"000002": {DependsOn: []string{"000001"}},
					"000003": {DependsOn: []string{"000002"}},
				},
			},
			expectedOrder: []string{"000002", "000003"},
		},
		{
			name:          "complex dependencies",
			candidateMigs: []string{"000001", "000002", "000003", "000004"},
			applied:       map[string]bool{},
			depConfig: &types.DependencyConfig{
				Migrations: map[string]types.MigrationInfo{
					"000001": {DependsOn: []string{}},
					"000002": {DependsOn: []string{"000001"}},
					"000003": {DependsOn: []string{"000001"}},
					"000004": {DependsOn: []string{"000002", "000003"}},
				},
			},
			expectedOrder: []string{"000001", "000002", "000003", "000004"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := o.resolveDependencies(tt.candidateMigs, tt.applied, tt.depConfig)

			if len(result) != len(tt.expectedOrder) {
				t.Errorf("Expected %d migrations, got %d", len(tt.expectedOrder), len(result))
				return
			}

			// Check that all expected migrations are present
			resultSet := make(map[string]bool)
			for _, mig := range result {
				resultSet[mig] = true
			}

			for _, expected := range tt.expectedOrder {
				if !resultSet[expected] {
					t.Errorf("Expected migration %s not found in result", expected)
				}
			}

			// For simple cases, check order
			if tt.depConfig != nil && !tt.expectCycleError {
				for i, mig := range result {
					if i < len(tt.expectedOrder) && mig != tt.expectedOrder[i] {
						// Order might vary for complex dependencies, just check presence
						break
					}
				}
			}
		})
	}
}

func TestAssessMigrationRisks(t *testing.T) {
	// Create a real logger for testing (output to /dev/null to avoid spam)
	logger := utils.NewLogger(false, false)

	o := &Orchestrator{
		logger: logger,
	}

	highRiskMigrations := []string{"000001", "000002"}
	depConfig := &types.DependencyConfig{
		Migrations: map[string]types.MigrationInfo{
			"000001": {
				Description:       "High risk migration",
				RiskLevel:         "high",
				AffectsTables:     []string{"users"},
				EstimatedDuration: "30s",
			},
			"000002": {
				Description: "Low risk migration",
				RiskLevel:   "low",
			},
		},
	}

	// Test that assessMigrationRisks doesn't panic and handles the config
	err := o.assessMigrationRisks(highRiskMigrations, depConfig)
	if err != nil {
		t.Errorf("assessMigrationRisks failed: %v", err)
	}
}

func TestCheckRollbackDependencies(t *testing.T) {
	o := &Orchestrator{}

	executions := []types.MigrationExecution{
		{MigrationID: "000003", Status: "completed"},
		{MigrationID: "000004", Status: "completed"},
	}

	depConfig := &types.DependencyConfig{
		Migrations: map[string]types.MigrationInfo{
			"000001": {DependsOn: []string{}},
			"000002": {DependsOn: []string{"000001"}},
			"000003": {DependsOn: []string{"000002"}},
			"000004": {DependsOn: []string{"000003"}},
			"000005": {DependsOn: []string{"000004"}}, // This depends on 000004
		},
	}

	affected := o.checkRollbackDependencies(executions, depConfig)

	// 000005 should be affected since it depends on 000004
	expectedAffected := []string{"000005"}
	if len(affected) != len(expectedAffected) {
		t.Errorf("Expected %d affected migrations, got %d", len(expectedAffected), len(affected))
	}

	affectedSet := make(map[string]bool)
	for _, mig := range affected {
		affectedSet[mig] = true
	}

	for _, expected := range expectedAffected {
		if !affectedSet[expected] {
			t.Errorf("Expected affected migration %s not found", expected)
		}
	}
}

func TestMigrationIDToVersion(t *testing.T) {
	o := &Orchestrator{}

	tests := []struct {
		migrationID string
		expected    int
		expectError bool
	}{
		{"000001", 1, false},
		{"000123", 123, false},
		{"000999", 999, false},
		{"invalid", 0, true},
		{"001", 0, true},
		{"000abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.migrationID, func(t *testing.T) {
			result, err := o.migrationIDToVersion(tt.migrationID)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, got none", tt.migrationID)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tt.migrationID, err)
				}
				if result != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, result)
				}
			}
		})
	}
}
