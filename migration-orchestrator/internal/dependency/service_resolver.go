package dependency

import (
	"fmt"
	"sort"

	"github.com/v-egorov/service-boilerplate/migration-orchestrator/pkg/types"
)

// ServiceDependencyResolver handles cross-service dependency resolution
type ServiceDependencyResolver struct {
	services map[string]*types.ServiceDependencyInfo
}

// NewServiceDependencyResolver creates a new service dependency resolver
func NewServiceDependencyResolver() *ServiceDependencyResolver {
	return &ServiceDependencyResolver{
		services: make(map[string]*types.ServiceDependencyInfo),
	}
}

// AddService adds a service with its dependency information
func (r *ServiceDependencyResolver) AddService(serviceName string, depConfig *types.DependencyConfig) error {
	info := &types.ServiceDependencyInfo{
		ServiceName:            serviceName,
		DependsOnServices:      []string{},
		CrossServiceMigrations: make(map[string][]string),
		MigrationDependencies:  depConfig.Migrations,
	}

	// Extract cross-service dependencies from migration configs
	dependsOnServices := make(map[string]bool)

	for _, migrationInfo := range depConfig.Migrations {
		for service, migrations := range migrationInfo.CrossServiceDependsOn {
			if len(migrations) > 0 {
				info.CrossServiceMigrations[service] = migrations
				dependsOnServices[service] = true
			}
		}
	}

	// Convert map to slice
	for service := range dependsOnServices {
		info.DependsOnServices = append(info.DependsOnServices, service)
	}
	sort.Strings(info.DependsOnServices)

	r.services[serviceName] = info
	return nil
}

// ResolveDependencies performs topological sorting to determine service execution order
func (r *ServiceDependencyResolver) ResolveDependencies() (*types.ServiceExecutionPlan, error) {
	plan := &types.ServiceExecutionPlan{
		ServiceOrder:   []string{},
		ServiceDetails: r.services,
	}

	// Kahn's algorithm for topological sorting
	inDegree := make(map[string]int)
	queue := []string{}

	// Calculate in-degrees
	for serviceName, info := range r.services {
		inDegree[serviceName] = len(info.DependsOnServices)
		if inDegree[serviceName] == 0 {
			queue = append(queue, serviceName)
		}
	}

	// Process queue
	processed := 0
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		plan.ServiceOrder = append(plan.ServiceOrder, current)
		processed++

		// Find services that depend on current service
		for serviceName, info := range r.services {
			if contains(info.DependsOnServices, current) {
				inDegree[serviceName]--
				if inDegree[serviceName] == 0 {
					queue = append(queue, serviceName)
				}
			}
		}
	}

	// Check for circular dependencies
	if processed != len(r.services) {
		plan.CircularDeps = r.detectCircularDependencies()
		return plan, fmt.Errorf("circular dependency detected among services: %v", plan.CircularDeps)
	}

	return plan, nil
}

// detectCircularDependencies finds services involved in circular dependencies
func (r *ServiceDependencyResolver) detectCircularDependencies() []string {
	// Simple cycle detection - services that are still referenced by others
	var circular []string
	serviceNames := make([]string, 0, len(r.services))
	for name := range r.services {
		serviceNames = append(serviceNames, name)
	}

	for serviceName := range r.services {
		// Check if this service is still referenced by others
		for _, info := range r.services {
			if contains(info.DependsOnServices, serviceName) {
				circular = append(circular, serviceName)
				break
			}
		}
	}
	return circular
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetServiceOrder returns the resolved service execution order
func (r *ServiceDependencyResolver) GetServiceOrder() ([]string, error) {
	plan, err := r.ResolveDependencies()
	if err != nil {
		return nil, err
	}
	return plan.ServiceOrder, nil
}
