package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the CLI configuration
type Config struct {
	Environment string         `mapstructure:"environment"`
	Services    ServiceConfig  `mapstructure:",squash"`
	API         APIConfig      `mapstructure:",squash"`
	Database    DatabaseConfig `mapstructure:",squash"`
	Logging     LoggingConfig  `mapstructure:",squash"`
}

// ServiceConfig holds service-related configuration
type ServiceConfig struct {
	GatewayURL    string                       `mapstructure:"gateway_url"`
	ServiceURLs   map[string]string            `mapstructure:"service_urls"`
	HealthChecks  map[string]HealthCheckConfig `mapstructure:"health_checks"`
	DefaultPort   int                          `mapstructure:"default_port"`
	Timeout       int                          `mapstructure:"timeout"`
	RetryAttempts int                          `mapstructure:"retry_attempts"`
}

// HealthCheckConfig defines health check endpoints for a service
type HealthCheckConfig struct {
	Health    string `mapstructure:"health"`
	Readiness string `mapstructure:"readiness"`
	Liveness  string `mapstructure:"liveness"`
}

// APIConfig holds API-related configuration
type APIConfig struct {
	BaseURL string `mapstructure:"base_url"`
	Version string `mapstructure:"version"`
	Timeout int    `mapstructure:"timeout"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// Load loads configuration from various sources
func Load(configFile string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Load from config file if specified
	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		// Look for config file in current directory or home directory
		v.SetConfigName("boilerplate-cli")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		if home, err := os.UserHomeDir(); err == nil {
			v.AddConfigPath(filepath.Join(home, ".config"))
			v.AddConfigPath(home)
		}
	}

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		// Config file is optional, so we don't return an error
		if configFile != "" {
			return nil, fmt.Errorf("failed to read config file %s: %w", configFile, err)
		}
	}

	// Bind environment variables
	v.SetEnvPrefix("BOILERPLATE_CLI")
	v.AutomaticEnv()

	// Unmarshal into config struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Service defaults
	v.SetDefault("environment", "development")
	v.SetDefault("gateway_url", "http://localhost:8080")
	v.SetDefault("service_urls", map[string]string{
		"user-service": "http://localhost:8081",
	})
	v.SetDefault("default_port", 8080)
	v.SetDefault("timeout", 30)
	v.SetDefault("retry_attempts", 3)

	// API defaults
	v.SetDefault("base_url", "http://localhost:8080")
	v.SetDefault("version", "v1")
	v.SetDefault("api_timeout", 30)

	// Database defaults
	v.SetDefault("db_host", "localhost")
	v.SetDefault("db_port", 5432)
	v.SetDefault("db_user", "postgres")
	v.SetDefault("db_password", "postgres")
	v.SetDefault("db_database", "service_db")
	v.SetDefault("db_ssl_mode", "disable")

	// Logging defaults
	v.SetDefault("log_level", "info")
	v.SetDefault("log_format", "text")
	v.SetDefault("log_output", "stdout")
}

// GetServiceURL returns the URL for a specific service
func (c *Config) GetServiceURL(serviceName string) string {
	if url, exists := c.Services.ServiceURLs[serviceName]; exists {
		return url
	}

	// Fallback to gateway URL for unknown services
	return c.Services.GatewayURL
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}
