package types

// ServiceDependencyGraph represents dependencies between services
type ServiceDependencyGraph struct {
	Services      map[string]*ServiceDependencyInfo `json:"services"`
	ResolvedOrder []string                          `json:"resolved_order"`
}

// ServiceDependencyInfo represents dependency information for a single service
type ServiceDependencyInfo struct {
	ServiceName            string                   `json:"service_name"`
	DependsOnServices      []string                 `json:"depends_on_services"`
	CrossServiceMigrations map[string][]string      `json:"cross_service_migrations"`
	MigrationDependencies  map[string]MigrationInfo `json:"migration_dependencies"`
}

// ServiceExecutionPlan represents the plan for executing services in dependency order
type ServiceExecutionPlan struct {
	ServiceOrder   []string                          `json:"service_order"`
	ServiceDetails map[string]*ServiceDependencyInfo `json:"service_details"`
	CircularDeps   []string                          `json:"circular_dependencies,omitempty"`
}
