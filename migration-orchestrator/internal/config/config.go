package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Environment string                   `yaml:"environment" mapstructure:"environment"`
	Database    DatabaseConfig           `yaml:"database" mapstructure:"database"`
	Services    map[string]ServiceConfig `yaml:"services" mapstructure:"services"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host" mapstructure:"host"`
	Port     int    `yaml:"port" mapstructure:"port"`
	User     string `yaml:"user" mapstructure:"user"`
	Password string `yaml:"password" mapstructure:"password"`
	Database string `yaml:"database" mapstructure:"database"`
	SSLMode  string `yaml:"ssl_mode" mapstructure:"ssl_mode"`
}

type ServiceConfig struct {
	Path        string `yaml:"path" mapstructure:"path"`
	Schema      string `yaml:"schema" mapstructure:"schema"`
	Description string `yaml:"description" mapstructure:"description"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configFile string) (*Config, error) {
	if configFile == "" {
		// Try default locations
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}

		possiblePaths := []string{
			filepath.Join(home, ".migration-orchestrator.yaml"),
			filepath.Join(home, ".migration-orchestrator.yml"),
			"migration-orchestrator.yaml",
			"migration-orchestrator.yml",
		}

		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				configFile = path
				break
			}
		}
	}

	config := &Config{
		Environment: "development",
		Database: DatabaseConfig{
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     5432,
			User:     getEnvOrDefault("DB_USER", "postgres"),
			Password: getEnvOrDefault("DB_PASSWORD", "postgres"),
			Database: getEnvOrDefault("DB_NAME", "service_db"),
			SSLMode:  getEnvOrDefault("DB_SSL_MODE", "disable"),
		},
		Services: make(map[string]ServiceConfig),
	}

	if configFile != "" {
		if err := loadFromFile(configFile, config); err != nil {
			return nil, fmt.Errorf("failed to load config file %s: %w", configFile, err)
		}
	}

	// Override with environment variables
	if env := os.Getenv("MIGRATION_ENV"); env != "" {
		config.Environment = env
	}

	return config, nil
}

func loadFromFile(configFile string, config *Config) error {
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return viper.Unmarshal(config)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SaveConfig saves the configuration to a file
func (c *Config) SaveConfig(filename string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}
