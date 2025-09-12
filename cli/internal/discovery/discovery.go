package discovery

import (
	"fmt"
	"net/http"
	"time"

	"github.com/v-egorov/service-boilerplate/cli/internal/config"
)

// ServiceInfo holds information about a service
type ServiceInfo struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
}

// ServiceRegistry manages service discovery
type ServiceRegistry struct {
	config   *config.Config
	services map[string]*ServiceInfo
	client   *http.Client
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry(cfg *config.Config) *ServiceRegistry {
	return &ServiceRegistry{
		config:   cfg,
		services: make(map[string]*ServiceInfo),
		client: &http.Client{
			Timeout: time.Duration(cfg.Services.Timeout) * time.Second,
		},
	}
}

// DiscoverServices discovers all available services
func (sr *ServiceRegistry) DiscoverServices() ([]*ServiceInfo, error) {
	var services []*ServiceInfo

	// Add known services from configuration
	for name, url := range sr.config.Services.ServiceURLs {
		service := &ServiceInfo{
			Name: name,
			URL:  url,
		}

		// Check service health
		if err := sr.checkServiceHealth(service); err != nil {
			service.Status = "unhealthy"
		} else {
			service.Status = "healthy"
		}

		sr.services[name] = service
		services = append(services, service)
	}

	return services, nil
}

// GetService returns information about a specific service
func (sr *ServiceRegistry) GetService(name string) (*ServiceInfo, error) {
	// Check if service is already discovered
	if service, exists := sr.services[name]; exists {
		return service, nil
	}

	// Try to discover the service
	url := sr.config.GetServiceURL(name)
	service := &ServiceInfo{
		Name: name,
		URL:  url,
	}

	if err := sr.checkServiceHealth(service); err != nil {
		service.Status = "unhealthy"
	} else {
		service.Status = "healthy"
	}

	sr.services[name] = service
	return service, nil
}

// GetAllServices returns all discovered services
func (sr *ServiceRegistry) GetAllServices() []*ServiceInfo {
	var services []*ServiceInfo
	for _, service := range sr.services {
		services = append(services, service)
	}
	return services
}

// checkServiceHealth checks if a service is healthy
func (sr *ServiceRegistry) checkServiceHealth(service *ServiceInfo) error {
	healthURL := fmt.Sprintf("%s/health", service.URL)

	resp, err := sr.client.Get(healthURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("service returned status %d", resp.StatusCode)
	}

	return nil
}

// IsServiceHealthy checks if a service is healthy
func (sr *ServiceRegistry) IsServiceHealthy(name string) bool {
	service, err := sr.GetService(name)
	if err != nil {
		return false
	}
	return service.Status == "healthy"
}

// GetServiceURL returns the URL for a service
func (sr *ServiceRegistry) GetServiceURL(name string) string {
	service, err := sr.GetService(name)
	if err != nil {
		return ""
	}
	return service.URL
}
