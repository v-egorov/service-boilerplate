package workflows

import (
	"fmt"
	"sync"
	"time"

	"github.com/v-egorov/service-boilerplate/cli/internal/client"
	"github.com/v-egorov/service-boilerplate/cli/internal/config"
)

// MultiServiceOperation represents an operation that spans multiple services
type MultiServiceOperation struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Services    []string           `json:"services"`
	Steps       []MultiServiceStep `json:"steps"`
	Parallel    bool               `json:"parallel"`      // Execute steps in parallel
	StopOnError bool               `json:"stop_on_error"` // Stop execution on first error
}

// MultiServiceStep represents a single step in a multi-service operation
type MultiServiceStep struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Service     string                 `json:"service"`
	Method      string                 `json:"method"`
	Endpoint    string                 `json:"endpoint"`
	Body        map[string]interface{} `json:"body,omitempty"`
	Headers     map[string]string      `json:"headers,omitempty"`
	Required    bool                   `json:"required"`
	DependsOn   []string               `json:"depends_on,omitempty"` // Step dependencies
}

// MultiServiceResult represents the result of a multi-service operation
type MultiServiceResult struct {
	Operation      string                 `json:"operation"`
	TotalSteps     int                    `json:"total_steps"`
	CompletedSteps int                    `json:"completed_steps"`
	FailedSteps    int                    `json:"failed_steps"`
	Duration       time.Duration          `json:"duration"`
	StepResults    map[string]*StepResult `json:"step_results"`
	Errors         []string               `json:"errors,omitempty"`
}

// StepResult represents the result of a single step
type StepResult struct {
	StepName     string        `json:"step_name"`
	Service      string        `json:"service"`
	Success      bool          `json:"success"`
	ResponseTime time.Duration `json:"response_time"`
	StatusCode   int           `json:"status_code"`
	Error        string        `json:"error,omitempty"`
	Data         interface{}   `json:"data,omitempty"`
}

// MultiServiceExecutor handles execution of multi-service operations
type MultiServiceExecutor struct {
	config    *config.Config
	apiClient *client.APIClient
}

// NewMultiServiceExecutor creates a new multi-service executor
func NewMultiServiceExecutor(cfg *config.Config, apiClient *client.APIClient) *MultiServiceExecutor {
	return &MultiServiceExecutor{
		config:    cfg,
		apiClient: apiClient,
	}
}

// ExecuteOperation executes a multi-service operation
func (mse *MultiServiceExecutor) ExecuteOperation(operation *MultiServiceOperation) (*MultiServiceResult, error) {
	start := time.Now()

	result := &MultiServiceResult{
		Operation:   operation.Name,
		TotalSteps:  len(operation.Steps),
		StepResults: make(map[string]*StepResult),
		Errors:      make([]string, 0),
	}

	fmt.Printf("🚀 Executing multi-service operation: %s\n", operation.Name)
	fmt.Printf("📝 Description: %s\n", operation.Description)
	fmt.Printf("🎯 Target Services: %v\n", operation.Services)
	fmt.Printf("📊 Total Steps: %d\n", len(operation.Steps))
	fmt.Println("=====================================")

	if operation.Parallel {
		result = mse.executeParallel(operation, result)
	} else {
		result = mse.executeSequential(operation, result)
	}

	result.Duration = time.Since(start)

	fmt.Println("=====================================")
	fmt.Printf("⏱️  Total Duration: %v\n", result.Duration.Round(time.Millisecond))
	fmt.Printf("✅ Completed Steps: %d/%d\n", result.CompletedSteps, result.TotalSteps)
	fmt.Printf("❌ Failed Steps: %d\n", result.FailedSteps)

	if result.FailedSteps == 0 {
		fmt.Println("🎉 Operation completed successfully!")
	} else if result.CompletedSteps > 0 {
		fmt.Println("⚠️  Operation completed with some failures")
	} else {
		fmt.Println("❌ Operation failed completely")
	}

	return result, nil
}

// executeSequential executes steps in sequence
func (mse *MultiServiceExecutor) executeSequential(operation *MultiServiceOperation, result *MultiServiceResult) *MultiServiceResult {
	for i, step := range operation.Steps {
		fmt.Printf("\n📍 Step %d/%d: %s\n", i+1, len(operation.Steps), step.Name)
		fmt.Printf("📖 %s\n", step.Description)
		fmt.Printf("🌐 Service: %s\n", step.Service)

		stepResult := mse.executeStep(step)

		result.StepResults[step.Name] = stepResult

		if stepResult.Success {
			result.CompletedSteps++
			fmt.Printf("✅ Step completed successfully (%v)\n", stepResult.ResponseTime.Round(time.Millisecond))
		} else {
			result.FailedSteps++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", step.Name, stepResult.Error))
			fmt.Printf("❌ Step failed: %s\n", stepResult.Error)

			if operation.StopOnError && step.Required {
				fmt.Printf("🛑 Stopping execution due to required step failure\n")
				break
			}
		}
	}

	return result
}

// executeParallel executes steps in parallel
func (mse *MultiServiceExecutor) executeParallel(operation *MultiServiceOperation, result *MultiServiceResult) *MultiServiceResult {
	var wg sync.WaitGroup
	stepChan := make(chan *StepResult, len(operation.Steps))

	// Execute all steps concurrently
	for i, step := range operation.Steps {
		wg.Add(1)
		go func(stepIndex int, step MultiServiceStep) {
			defer wg.Done()

			fmt.Printf("📍 Starting parallel step %d: %s (%s)\n", stepIndex+1, step.Name, step.Service)

			stepResult := mse.executeStep(step)
			stepChan <- stepResult
		}(i, step)
	}

	// Close channel when all goroutines are done
	go func() {
		wg.Wait()
		close(stepChan)
	}()

	// Collect results
	for stepResult := range stepChan {
		result.StepResults[stepResult.StepName] = stepResult

		if stepResult.Success {
			result.CompletedSteps++
			fmt.Printf("✅ Parallel step completed: %s (%v)\n", stepResult.StepName, stepResult.ResponseTime.Round(time.Millisecond))
		} else {
			result.FailedSteps++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", stepResult.StepName, stepResult.Error))
			fmt.Printf("❌ Parallel step failed: %s (%s)\n", stepResult.StepName, stepResult.Error)
		}
	}

	return result
}

// executeStep executes a single step
func (mse *MultiServiceExecutor) executeStep(step MultiServiceStep) *StepResult {
	start := time.Now()

	result := &StepResult{
		StepName: step.Name,
		Service:  step.Service,
	}

	// Execute the API call
	resp, err := mse.apiClient.CallService(step.Service, step.Method, step.Endpoint, step.Body, step.Headers)

	result.ResponseTime = time.Since(start)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return result
	}

	result.StatusCode = resp.StatusCode
	result.Success = resp.StatusCode >= 200 && resp.StatusCode < 300

	if !result.Success {
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		if resp.Error != "" {
			result.Error += ": " + resp.Error
		}
	} else {
		result.Data = resp.Body
	}

	return result
}

// GetPredefinedOperations returns predefined multi-service operations
func (mse *MultiServiceExecutor) GetPredefinedOperations() map[string]*MultiServiceOperation {
	return map[string]*MultiServiceOperation{
		"system-bootstrap": {
			Name:        "System Bootstrap",
			Description: "Initialize all system services and verify connectivity",
			Services:    []string{"api-gateway", "user-service"},
			Parallel:    true,
			StopOnError: false,
			Steps: []MultiServiceStep{
				{
					Name:        "Check API Gateway Health",
					Description: "Verify API gateway is responding",
					Service:     "api-gateway",
					Method:      "GET",
					Endpoint:    "/health",
					Required:    true,
				},
				{
					Name:        "Check User Service Health",
					Description: "Verify user service is responding",
					Service:     "user-service",
					Method:      "GET",
					Endpoint:    "/health",
					Required:    true,
				},
				{
					Name:        "Verify User Service API",
					Description: "Test user service API endpoints",
					Service:     "user-service",
					Method:      "GET",
					Endpoint:    "/api/v1/users?limit=1",
					Required:    false,
				},
			},
		},
		"data-synchronization": {
			Name:        "Data Synchronization",
			Description: "Synchronize data across services",
			Services:    []string{"user-service"},
			Parallel:    false,
			StopOnError: true,
			Steps: []MultiServiceStep{
				{
					Name:        "Validate User Data",
					Description: "Check user data integrity",
					Service:     "user-service",
					Method:      "GET",
					Endpoint:    "/api/v1/users?limit=10",
					Required:    true,
				},
				{
					Name:        "Sync User Profiles",
					Description: "Synchronize user profile data",
					Service:     "user-service",
					Method:      "GET",
					Endpoint:    "/api/v1/users",
					Required:    false,
				},
			},
		},
		"service-migration": {
			Name:        "Service Migration",
			Description: "Migrate data and verify service compatibility",
			Services:    []string{"user-service", "api-gateway"},
			Parallel:    false,
			StopOnError: true,
			Steps: []MultiServiceStep{
				{
					Name:        "Pre-Migration Health Check",
					Description: "Verify all services are healthy before migration",
					Service:     "api-gateway",
					Method:      "GET",
					Endpoint:    "/health",
					Required:    true,
				},
				{
					Name:        "Backup User Data",
					Description: "Create backup of user data",
					Service:     "user-service",
					Method:      "GET",
					Endpoint:    "/api/v1/users",
					Required:    true,
				},
				{
					Name:        "Post-Migration Verification",
					Description: "Verify services are working after migration",
					Service:     "user-service",
					Method:      "GET",
					Endpoint:    "/health",
					Required:    true,
				},
			},
		},
	}
}
