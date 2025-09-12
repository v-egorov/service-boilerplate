package workflows

import (
	"fmt"
	"time"

	"github.com/v-egorov/service-boilerplate/cli/internal/client"
	"github.com/v-egorov/service-boilerplate/cli/internal/config"
)

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Service     string                 `json:"service"`
	Method      string                 `json:"method"`
	Endpoint    string                 `json:"endpoint"`
	Body        map[string]interface{} `json:"body,omitempty"`
	Headers     map[string]string      `json:"headers,omitempty"`
	Required    bool                   `json:"required"`
}

// Workflow represents a complete workflow
type Workflow struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Steps       []WorkflowStep `json:"steps"`
}

// WorkflowExecutor handles workflow execution
type WorkflowExecutor struct {
	config    *config.Config
	apiClient *client.APIClient
}

// NewWorkflowExecutor creates a new workflow executor
func NewWorkflowExecutor(cfg *config.Config, apiClient *client.APIClient) *WorkflowExecutor {
	return &WorkflowExecutor{
		config:    cfg,
		apiClient: apiClient,
	}
}

// ExecuteWorkflow executes a predefined workflow
func (we *WorkflowExecutor) ExecuteWorkflow(name string) error {
	workflow, err := we.getWorkflow(name)
	if err != nil {
		return fmt.Errorf("failed to get workflow %s: %w", name, err)
	}

	fmt.Printf("🚀 Executing workflow: %s\n", workflow.Name)
	fmt.Printf("📝 Description: %s\n", workflow.Description)
	fmt.Println("=====================================")

	successCount := 0
	totalSteps := len(workflow.Steps)

	for i, step := range workflow.Steps {
		fmt.Printf("\n📍 Step %d/%d: %s\n", i+1, totalSteps, step.Name)
		fmt.Printf("📖 %s\n", step.Description)

		// Execute the step
		resp, err := we.apiClient.CallService(step.Service, step.Method, step.Endpoint, step.Body, step.Headers)

		if err != nil {
			if step.Required {
				return fmt.Errorf("required step '%s' failed: %w", step.Name, err)
			}
			fmt.Printf("⚠️  Optional step '%s' failed: %v\n", step.Name, err)
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			fmt.Printf("✅ Step '%s' completed successfully\n", step.Name)
			successCount++
		} else {
			if step.Required {
				return fmt.Errorf("required step '%s' failed with status %d: %s", step.Name, resp.StatusCode, resp.Error)
			}
			fmt.Printf("⚠️  Optional step '%s' failed with status %d\n", step.Name, resp.StatusCode)
		}

		// Small delay between steps
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("=====================================")
	fmt.Printf("🎉 Workflow completed: %d/%d steps successful\n", successCount, totalSteps)

	if successCount == totalSteps {
		fmt.Println("✅ All steps completed successfully!")
	} else if successCount >= totalSteps/2 {
		fmt.Println("⚠️  Workflow completed with some failures")
	} else {
		return fmt.Errorf("workflow failed: only %d/%d steps completed", successCount, totalSteps)
	}

	return nil
}

// getWorkflow returns a predefined workflow by name
func (we *WorkflowExecutor) getWorkflow(name string) (*Workflow, error) {
	workflows := we.getPredefinedWorkflows()

	workflow, exists := workflows[name]
	if !exists {
		return nil, fmt.Errorf("workflow '%s' not found", name)
	}

	return workflow, nil
}

// getPredefinedWorkflows returns all predefined workflows
func (we *WorkflowExecutor) getPredefinedWorkflows() map[string]*Workflow {
	return map[string]*Workflow{
		"user-onboarding": {
			Name:        "User Onboarding",
			Description: "Complete user onboarding workflow including user creation and verification",
			Steps: []WorkflowStep{
				{
					Name:        "Create User",
					Description: "Create a new user account",
					Service:     "user-service",
					Method:      "POST",
					Endpoint:    "/api/v1/users",
					Body: map[string]interface{}{
						"email":      "newuser@example.com",
						"first_name": "New",
						"last_name":  "User",
					},
					Required: true,
				},
				{
					Name:        "Verify User",
					Description: "Verify the user was created successfully",
					Service:     "user-service",
					Method:      "GET",
					Endpoint:    "/api/v1/users",
					Required:    false,
				},
			},
		},
		"data-initialization": {
			Name:        "Data Initialization",
			Description: "Initialize system with base data",
			Steps: []WorkflowStep{
				{
					Name:        "Health Check",
					Description: "Verify all services are healthy",
					Service:     "user-service",
					Method:      "GET",
					Endpoint:    "/health",
					Required:    true,
				},
				{
					Name:        "Seed Base Data",
					Description: "Load base user data",
					Service:     "user-service",
					Method:      "GET",
					Endpoint:    "/api/v1/users",
					Required:    false,
				},
			},
		},
		"system-health-check": {
			Name:        "System Health Check",
			Description: "Comprehensive health check of all system components",
			Steps: []WorkflowStep{
				{
					Name:        "Check User Service",
					Description: "Verify user service is responding",
					Service:     "user-service",
					Method:      "GET",
					Endpoint:    "/health",
					Required:    true,
				},
				{
					Name:        "List Users",
					Description: "Verify user data is accessible",
					Service:     "user-service",
					Method:      "GET",
					Endpoint:    "/api/v1/users?limit=1",
					Required:    false,
				},
			},
		},
	}
}

// ListWorkflows returns a list of available workflows
func (we *WorkflowExecutor) ListWorkflows() []*Workflow {
	workflows := we.getPredefinedWorkflows()
	var result []*Workflow

	for _, workflow := range workflows {
		result = append(result, workflow)
	}

	return result
}
