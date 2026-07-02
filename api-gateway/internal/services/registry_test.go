package services

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func newTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetOutput(&discardWriter{})
	return logger
}

type discardWriter struct{}

func (d *discardWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func TestRegisterService(t *testing.T) {
	t.Run("stores service mapping correctly", func(t *testing.T) {
		logger := newTestLogger()
		reg := NewServiceRegistry(logger)

		reg.RegisterService("test-service", "http://localhost:8080")

		url, err := reg.GetServiceURL("test-service")
		assert.NoError(t, err)
		assert.Equal(t, "http://localhost:8080", url)
	})

	t.Run("overwrites existing service URL", func(t *testing.T) {
		logger := newTestLogger()
		reg := NewServiceRegistry(logger)

		reg.RegisterService("test-service", "http://old-url:8080")
		reg.RegisterService("test-service", "http://new-url:9090")

		url, err := reg.GetServiceURL("test-service")
		assert.NoError(t, err)
		assert.Equal(t, "http://new-url:9090", url)
	})
}

func TestGetServiceURL(t *testing.T) {
	t.Run("returns registered URL on success path", func(t *testing.T) {
		logger := newTestLogger()
		reg := NewServiceRegistry(logger)

		reg.RegisterService("my-service", "http://example.com:3000")

		url, err := reg.GetServiceURL("my-service")
		assert.NoError(t, err)
		assert.Equal(t, "http://example.com:3000", url)
	})

	t.Run("returns error for unknown service names", func(t *testing.T) {
		logger := newTestLogger()
		reg := NewServiceRegistry(logger)

		reg.RegisterService("existing-service", "http://localhost:8080")

		url, err := reg.GetServiceURL("nonexistent-service")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.Empty(t, url)
	})

	t.Run("returns error when registry is empty", func(t *testing.T) {
		logger := newTestLogger()
		reg := NewServiceRegistry(logger)

		url, err := reg.GetServiceURL("any-service")
		assert.Error(t, err)
		assert.Empty(t, url)
	})
}

func TestListServices(t *testing.T) {
	t.Run("returns all registered services in a map", func(t *testing.T) {
		logger := newTestLogger()
		reg := NewServiceRegistry(logger)

		reg.RegisterService("svc-a", "http://a:8080")
		reg.RegisterService("svc-b", "http://b:9090")
		reg.RegisterService("svc-c", "http://c:7070")

		services := reg.ListServices()

		assert.Len(t, services, 3)
		assert.Equal(t, "http://a:8080", services["svc-a"])
		assert.Equal(t, "http://b:9090", services["svc-b"])
		assert.Equal(t, "http://c:7070", services["svc-c"])
	})

	t.Run("returns empty map when no services registered", func(t *testing.T) {
		logger := newTestLogger()
		reg := NewServiceRegistry(logger)

		services := reg.ListServices()
		assert.Empty(t, services)
	})

	t.Run("does not leak internal reference", func(t *testing.T) {
		logger := newTestLogger()
		reg := NewServiceRegistry(logger)

		reg.RegisterService("svc-x", "http://x:1234")

		services1 := reg.ListServices()
		reg.RegisterService("svc-y", "http://y:5678")

		// Modifying returned map should not affect registry
		services1["fake"] = "http://fake"
		assert.Len(t, services1, 2) // has fake + svc-x

		services2 := reg.ListServices()
		assert.Len(t, services2, 2) // only svc-x and svc-y
	})
}

func TestUnregisterService(t *testing.T) {
	t.Run("removes service from registry", func(t *testing.T) {
		logger := newTestLogger()
		reg := NewServiceRegistry(logger)

		reg.RegisterService("to-remove", "http://remove:8080")
		assert.Len(t, reg.ListServices(), 1)

		reg.UnregisterService("to-remove")
		assert.Empty(t, reg.ListServices())
	})

	t.Run("subsequent GetServiceURL fails after unregister", func(t *testing.T) {
		logger := newTestLogger()
		reg := NewServiceRegistry(logger)

		reg.RegisterService("remove-me", "http://localhost:9000")
		reg.UnregisterService("remove-me")

		url, err := reg.GetServiceURL("remove-me")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.Empty(t, url)
	})

	t.Run("unregistering non-existent service does not panic", func(t *testing.T) {
		logger := newTestLogger()
		reg := NewServiceRegistry(logger)

		// Should not panic or error — delete from nil/empty map is safe in Go
		assert.NotPanics(t, func() {
			reg.UnregisterService("does-not-exist")
		})

		services := reg.ListServices()
		assert.Empty(t, services)
	})
}
