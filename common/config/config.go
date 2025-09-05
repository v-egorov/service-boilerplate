package config

import (
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Server   ServerConfig   `mapstructure:"server"`
}

type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

type DatabaseConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	Database    string `mapstructure:"database"`
	SSLMode     string `mapstructure:"ssl_mode"`
	MaxConns    int32  `mapstructure:"max_conns"`
	MinConns    int32  `mapstructure:"min_conns"`
	MaxConnIdle int    `mapstructure:"max_conn_idle"`
	MaxConnLife int    `mapstructure:"max_conn_life"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	viper.AddConfigPath(".")

	// Set defaults
	setDefaults()

	// Enable environment variable override
	viper.SetEnvPrefix("APP")
	viper.AutomaticEnv()

	// Bind specific environment variables with higher priority
	viper.BindEnv("database.host", "DATABASE_HOST")
	viper.BindEnv("database.port", "DATABASE_PORT")
	viper.BindEnv("database.user", "DATABASE_USER")
	viper.BindEnv("database.password", "DATABASE_PASSWORD")
	viper.BindEnv("database.database", "DATABASE_NAME")
	viper.BindEnv("database.ssl_mode", "DATABASE_SSL_MODE")
	viper.BindEnv("logging.level", "LOGGING_LEVEL")
	viper.BindEnv("logging.format", "LOGGING_FORMAT")
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("app.environment", "APP_ENV")

	// Set environment variable defaults for Docker
	if os.Getenv("DOCKER_ENV") == "true" {
		if os.Getenv("DATABASE_HOST") == "" {
			os.Setenv("DATABASE_HOST", "postgres")
		}
	}

	// Try to read config file, but don't fail if it doesn't exist
	if err := viper.ReadInConfig(); err != nil {
		// If config file doesn't exist, continue with environment variables and defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func setDefaults() {
	// App defaults
	viper.SetDefault("app.name", "service-boilerplate")
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("app.environment", "development")

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.database", "service_db")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_conns", 10)
	viper.SetDefault("database.min_conns", 2)
	viper.SetDefault("database.max_conn_idle", 300)
	viper.SetDefault("database.max_conn_life", 3600)

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")

	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
}
