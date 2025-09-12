package workflows

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/v-egorov/service-boilerplate/cli/internal/client"
	"github.com/v-egorov/service-boilerplate/cli/internal/config"
)

// CustomWorkflowManager manages custom workflow definitions
type CustomWorkflowManager struct {
	config    *config.Config
	apiClient *client.APIClient
	workflows map[string]*Workflow
}

// NewCustomWorkflowManager creates a new custom workflow manager
func NewCustomWorkflowManager(cfg *config.Config, apiClient *client.APIClient) *CustomWorkflowManager {
	return &CustomWorkflowManager{
		config:    cfg,
		apiClient: apiClient,
		workflows: make(map[string]*Workflow),
	}
}

// LoadWorkflowsFromFile loads custom workflows from a JSON file
func (cwm *CustomWorkflowManager) LoadWorkflowsFromFile(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// File doesn't exist, create an empty workflows file
		return cwm.SaveWorkflowsToFile(filename)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read workflow file: %w", err)
	}

	var workflows map[string]*Workflow
	if err := json.Unmarshal(data, &workflows); err != nil {
		return fmt.Errorf("failed to parse workflow file: %w", err)
	}

	cwm.workflows = workflows
	return nil
}

// SaveWorkflowsToFile saves custom workflows to a JSON file
func (cwm *CustomWorkflowManager) SaveWorkflowsToFile(filename string) error {
	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(cwm.workflows, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal workflows: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write workflow file: %w", err)
	}

	return nil
}

// AddWorkflow adds a new custom workflow
func (cwm *CustomWorkflowManager) AddWorkflow(name string, workflow *Workflow) error {
	if _, exists := cwm.workflows[name]; exists {
		return fmt.Errorf("workflow '%s' already exists", name)
	}

	cwm.workflows[name] = workflow
	return nil
}

// UpdateWorkflow updates an existing custom workflow
func (cwm *CustomWorkflowManager) UpdateWorkflow(name string, workflow *Workflow) error {
	if _, exists := cwm.workflows[name]; !exists {
		return fmt.Errorf("workflow '%s' does not exist", name)
	}

	cwm.workflows[name] = workflow
	return nil
}

// DeleteWorkflow deletes a custom workflow
func (cwm *CustomWorkflowManager) DeleteWorkflow(name string) error {
	if _, exists := cwm.workflows[name]; !exists {
		return fmt.Errorf("workflow '%s' does not exist", name)
	}

	delete(cwm.workflows, name)
	return nil
}

// GetWorkflow retrieves a custom workflow by name
func (cwm *CustomWorkflowManager) GetWorkflow(name string) (*Workflow, error) {
	workflow, exists := cwm.workflows[name]
	if !exists {
		return nil, fmt.Errorf("workflow '%s' not found", name)
	}

	return workflow, nil
}

// ListWorkflows returns all custom workflows
func (cwm *CustomWorkflowManager) ListWorkflows() map[string]*Workflow {
	result := make(map[string]*Workflow)
	for name, workflow := range cwm.workflows {
		result[name] = workflow
	}
	return result
}

// ExecuteCustomWorkflow executes a custom workflow
func (cwm *CustomWorkflowManager) ExecuteCustomWorkflow(name string) error {
	workflow, err := cwm.GetWorkflow(name)
	if err != nil {
		return err
	}

	// Execute the workflow manually since the existing executor expects predefined workflows
	return cwm.executeWorkflowObject(workflow)
}

// executeWorkflowObject executes a workflow object directly
func (cwm *CustomWorkflowManager) executeWorkflowObject(workflow *Workflow) error {
	fmt.Printf("🚀 Executing custom workflow: %s\n", workflow.Name)
	fmt.Printf("📝 Description: %s\n", workflow.Description)
	fmt.Println("=====================================")

	successCount := 0
	totalSteps := len(workflow.Steps)

	for i, step := range workflow.Steps {
		fmt.Printf("\n📍 Step %d/%d: %s\n", i+1, totalSteps, step.Name)
		fmt.Printf("📖 %s\n", step.Description)

		// Execute the step
		resp, err := cwm.apiClient.CallService(step.Service, step.Method, step.Endpoint, step.Body, step.Headers)

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
	}

	fmt.Println("=====================================")
	fmt.Printf("🎉 Custom workflow completed: %d/%d steps successful\n", successCount, totalSteps)

	return nil
}

// CreateWorkflowFromTemplate creates a workflow from a template
func (cwm *CustomWorkflowManager) CreateWorkflowFromTemplate(name, templateName string) (*Workflow, error) {
	template, err := cwm.getWorkflowTemplate(templateName)
	if err != nil {
		return nil, err
	}

	// Create a copy of the template
	workflow := &Workflow{
		Name:        name,
		Description: template.Description,
		Steps:       make([]WorkflowStep, len(template.Steps)),
	}

	copy(workflow.Steps, template.Steps)
	return workflow, nil
}

// getWorkflowTemplate returns predefined workflow templates
func (cwm *CustomWorkflowManager) getWorkflowTemplate(name string) (*Workflow, error) {
	templates := map[string]*Workflow{
		"user-management": {
			Name:        "User Management Template",
			Description: "Template for user management operations",
			Steps: []WorkflowStep{
				{
					Name:        "Create User",
					Description: "Create a new user account",
					Service:     "user-service",
					Method:      "POST",
					Endpoint:    "/api/v1/users",
					Body: map[string]interface{}{
						"email":      "{{email}}",
						"first_name": "{{first_name}}",
						"last_name":  "{{last_name}}",
					},
					Required: true,
				},
				{
					Name:        "Verify User",
					Description: "Verify the user was created successfully",
					Service:     "user-service",
					Method:      "GET",
					Endpoint:    "/api/v1/users/{{user_id}}",
					Required:    false,
				},
			},
		},
		"data-backup": {
			Name:        "Data Backup Template",
			Description: "Template for data backup operations",
			Steps: []WorkflowStep{
				{
					Name:        "Export Users",
					Description: "Export all user data",
					Service:     "user-service",
					Method:      "GET",
					Endpoint:    "/api/v1/users",
					Required:    true,
				},
				{
					Name:        "Validate Export",
					Description: "Validate the exported data",
					Service:     "user-service",
					Method:      "GET",
					Endpoint:    "/api/v1/users?limit=1",
					Required:    false,
				},
			},
		},
		"health-monitoring": {
			Name:        "Health Monitoring Template",
			Description: "Template for comprehensive health monitoring",
			Steps: []WorkflowStep{
				{
					Name:        "Check API Gateway",
					Description: "Verify API gateway health",
					Service:     "api-gateway",
					Method:      "GET",
					Endpoint:    "/health",
					Required:    true,
				},
				{
					Name:        "Check User Service",
					Description: "Verify user service health",
					Service:     "user-service",
					Method:      "GET",
					Endpoint:    "/health",
					Required:    true,
				},
				{
					Name:        "Test User API",
					Description: "Test user service API functionality",
					Service:     "user-service",
					Method:      "GET",
					Endpoint:    "/api/v1/users?limit=1",
					Required:    false,
				},
			},
		},
	}

	template, exists := templates[name]
	if !exists {
		return nil, fmt.Errorf("template '%s' not found", name)
	}

	return template, nil
}

// ValidateWorkflow validates a workflow definition
func (cwm *CustomWorkflowManager) ValidateWorkflow(workflow *Workflow) error {
	if workflow.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	if len(workflow.Steps) == 0 {
		return fmt.Errorf("workflow must have at least one step")
	}

	for i, step := range workflow.Steps {
		if step.Name == "" {
			return fmt.Errorf("step %d: name is required", i+1)
		}

		if step.Service == "" {
			return fmt.Errorf("step %d (%s): service is required", i+1, step.Name)
		}

		if step.Method == "" {
			return fmt.Errorf("step %d (%s): method is required", i+1, step.Name)
		}

		if step.Endpoint == "" {
			return fmt.Errorf("step %d (%s): endpoint is required", i+1, step.Name)
		}
	}

	return nil
}

// GetWorkflowTemplates returns available workflow templates
func (cwm *CustomWorkflowManager) GetWorkflowTemplates() map[string]*Workflow {
	templates := make(map[string]*Workflow)

	templateNames := []string{"user-management", "data-backup", "health-monitoring"}
	for _, name := range templateNames {
		if template, err := cwm.getWorkflowTemplate(name); err == nil {
			templates[name] = template
		}
	}

	return templates
}

// ListWorkflowTemplates returns available workflow template names
func (cwm *CustomWorkflowManager) ListWorkflowTemplates() []string {
	return []string{"user-management", "data-backup", "health-monitoring"}
}
