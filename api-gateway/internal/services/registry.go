package services

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

type ServiceRegistry struct {
	services map[string]string
	mu       sync.RWMutex
	logger   *logrus.Logger
}

func NewServiceRegistry(logger *logrus.Logger) *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string]string),
		logger:   logger,
	}
}

func (r *ServiceRegistry) RegisterService(name, url string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.services[name] = url
	r.logger.WithFields(logrus.Fields{
		"service": name,
		"url":     url,
	}).Info("Service registered")
}

func (r *ServiceRegistry) GetServiceURL(name string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	url, exists := r.services[name]
	if !exists {
		return "", fmt.Errorf("service %s not found", name)
	}

	return url, nil
}

func (r *ServiceRegistry) ListServices() map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make(map[string]string)
	for name, url := range r.services {
		services[name] = url
	}

	return services
}

func (r *ServiceRegistry) UnregisterService(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.services, name)
	r.logger.WithField("service", name).Info("Service unregistered")
}
