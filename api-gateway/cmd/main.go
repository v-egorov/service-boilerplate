package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/api-gateway/internal/handlers"
	"github.com/v-egorov/service-boilerplate/api-gateway/internal/middleware"
	"github.com/v-egorov/service-boilerplate/api-gateway/internal/services"
	"github.com/v-egorov/service-boilerplate/common/alerting"
	"github.com/v-egorov/service-boilerplate/common/config"
	"github.com/v-egorov/service-boilerplate/common/logging"
	commonMiddleware "github.com/v-egorov/service-boilerplate/common/middleware"
	"github.com/v-egorov/service-boilerplate/common/tracing"
)

// jwtKeyCache holds the cached JWT public key with TTL
type jwtKeyCache struct {
	key       interface{}
	fetchedAt time.Time
	ttl       time.Duration
	mutex     sync.RWMutex
}

var (
	globalKeyCache = &jwtKeyCache{
		ttl: 1 * time.Hour, // Refresh key every hour
	}
)

func main() {
	// Load configuration
	cfg, err := config.Load(".")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := logging.NewLogger(logging.Config{
		Level:              cfg.Logging.Level,
		Format:             cfg.Logging.Format,
		Output:             cfg.Logging.Output,
		DualOutput:         cfg.Logging.DualOutput,
		ServiceName:        cfg.App.Name,
		StripANSIFromFiles: cfg.Logging.StripANSIFromFiles,
	})

	// Initialize tracing
	tracerProvider, err := tracing.InitTracer(cfg.Tracing)
	if err != nil {
		logger.Warn("Failed to initialize tracing", err)
	} else if tracerProvider != nil {
		defer func() {
			if err := tracing.ShutdownTracer(tracerProvider); err != nil {
				logger.Error("Failed to shutdown tracer", err)
			}
		}()
	}

	// Initialize service registry
	serviceRegistry := services.NewServiceRegistry(logger.Logger)

	// Register services with configurable URLs
	authServiceURL := cfg.GetServiceURL("auth", "http://auth-service:8083")
	userServiceURL := cfg.GetServiceURL("user", "http://user-service:8081")

	// Apply development environment overrides for localhost development
	if cfg.App.Environment == "development" && os.Getenv("DOCKER_ENV") != "true" {
		authServiceURL = strings.Replace(authServiceURL, "auth-service", "localhost", 1)
		userServiceURL = strings.Replace(userServiceURL, "user-service", "localhost", 1)
	}

	serviceRegistry.RegisterService("auth-service", authServiceURL)
	serviceRegistry.RegisterService("user-service", userServiceURL)

	// Initialize handlers
	gatewayHandler := handlers.NewGatewayHandler(serviceRegistry, logger.Logger, cfg)

	// Initialize request logger
	requestLogger := logging.NewServiceRequestLogger(logger.Logger, cfg.App.Name)

	// Initialize alert manager
	alertManager := alerting.NewAlertManager(logger.Logger, "api-gateway", &cfg.Alerting, requestLogger.GetMetricsCollector())

	// Start alert checking goroutine
	if cfg.Alerting.Enabled {
		go func() {
			ticker := time.NewTicker(1 * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				alertManager.CheckMetrics()
			}
		}()
	}

	// Start JWT key refresh routine
	startKeyRefreshRoutine(authServiceURL, logger.Logger)

	// Setup Gin router
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Initialize JWT middleware
	var jwtPublicKey interface{}
	var revocationChecker commonMiddleware.TokenRevocationChecker

	if cfg.JWT.PublicKey != "" {
		// Try to decode as base64 first, then as PEM
		jwtPublicKey = []byte(cfg.JWT.PublicKey) // Simplified for demo
		// For now, no revocation checker when using configured key
		revocationChecker = nil
	} else {
		// Try to fetch public key from auth-service (with caching)
		logger.Info("JWT public key not configured, attempting to fetch from auth-service")
		publicKey, err := getCachedKey(authServiceURL, logger.Logger)
		if err != nil {
			logger.Warn("Failed to fetch public key from auth-service, attempting JWT_PUBLIC_KEY environment fallback", err)

			// Try JWT_PUBLIC_KEY environment variable as fallback
			if envKey := os.Getenv("JWT_PUBLIC_KEY"); envKey != "" {
				logger.Info("Using JWT_PUBLIC_KEY environment variable as fallback")
				// Parse the PEM-encoded public key from environment
				block, _ := pem.Decode([]byte(envKey))
				if block == nil {
					logger.Warn("Failed to decode PEM block from JWT_PUBLIC_KEY environment variable")
					jwtPublicKey = nil
					revocationChecker = nil
				} else {
					envPublicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
					if err != nil {
						logger.WithError(err).Warn("Failed to parse public key from JWT_PUBLIC_KEY environment variable")
						jwtPublicKey = nil
						revocationChecker = nil
					} else {
						if rsaKey, ok := envPublicKey.(*rsa.PublicKey); ok {
							jwtPublicKey = rsaKey
							logger.Info("Successfully loaded JWT public key from JWT_PUBLIC_KEY environment variable")
							// Note: No revocation checker for env-based keys
							revocationChecker = nil
						} else {
							logger.Warn("JWT_PUBLIC_KEY environment variable contains non-RSA key")
							jwtPublicKey = nil
							revocationChecker = nil
						}
					}
				}
			} else {
				logger.Warn("JWT_PUBLIC_KEY environment variable not set, JWT validation disabled")
				jwtPublicKey = nil
				revocationChecker = nil
			}
		} else {
			jwtPublicKey = publicKey
			logger.Info("Successfully fetched JWT public key from auth-service")

			// Create HTTP-based revocation checker
			revocationChecker = &httpTokenRevocationChecker{
				authServiceURL: authServiceURL,
				logger:         logger.Logger,
			}
		}
	}

	// Middleware
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RequestIDMiddleware())
	if cfg.Tracing.Enabled {
		router.Use(tracing.HTTPMiddleware(cfg.Tracing.ServiceName))
	}
	router.Use(commonMiddleware.JWTMiddleware(jwtPublicKey, logger.Logger, revocationChecker))
	router.Use(requestLogger.RequestResponseLogger())

	// Health check endpoints (public, no auth required)
	router.GET("/health", gatewayHandler.LivenessHandler)
	router.GET("/ready", gatewayHandler.ReadinessHandler)
	router.GET("/live", gatewayHandler.LivenessHandler)
	router.GET("/status", gatewayHandler.StatusHandler) // Direct status endpoint
	router.GET("/ping", gatewayHandler.PingHandler)     // Direct ping endpoint

	// Public monitoring endpoints (no auth required)
	router.GET("/api/v1/status", gatewayHandler.StatusHandler)
	router.GET("/api/v1/ping", gatewayHandler.PingHandler)

	// Observability endpoints
	router.GET("/api/v1/metrics", func(c *gin.Context) {
		metrics := requestLogger.GetMetricsCollector().GetMetrics()
		c.JSON(http.StatusOK, metrics)
	})

	router.GET("/api/v1/alerts", func(c *gin.Context) {
		alerts := alertManager.GetActiveAlerts()
		c.JSON(http.StatusOK, gin.H{"alerts": alerts})
	})

	// Public auth endpoints (no auth required)
	auth := router.Group("/api/v1/auth")
	{
		auth.POST("/login", gatewayHandler.ProxyRequest("auth-service"))
		auth.POST("/register", gatewayHandler.ProxyRequest("auth-service"))
		auth.POST("/refresh", gatewayHandler.ProxyRequest("auth-service"))
		auth.POST("/logout", gatewayHandler.ProxyRequest("auth-service"))
	}

	// API routes (require authentication)
	api := router.Group("/api")
	api.Use(commonMiddleware.RequireAuth())
	{
		// Protected auth routes
		protectedAuth := api.Group("/v1/auth")
		{
			protectedAuth.GET("/me", gatewayHandler.ProxyRequest("auth-service"))

			// Admin RBAC routes (require admin role)
			admin := protectedAuth.Group("")
			admin.Use(commonMiddleware.RequireRole("admin"))
			{
				// Role management
				admin.POST("/roles", gatewayHandler.ProxyRequest("auth-service"))
				admin.GET("/roles", gatewayHandler.ProxyRequest("auth-service"))
				admin.GET("/roles/:role_id", gatewayHandler.ProxyRequest("auth-service"))
				admin.PUT("/roles/:role_id", gatewayHandler.ProxyRequest("auth-service"))
				admin.DELETE("/roles/:role_id", gatewayHandler.ProxyRequest("auth-service"))

				// Permission management
				admin.POST("/permissions", gatewayHandler.ProxyRequest("auth-service"))
				admin.GET("/permissions", gatewayHandler.ProxyRequest("auth-service"))
				admin.GET("/permissions/:permission_id", gatewayHandler.ProxyRequest("auth-service"))
				admin.PUT("/permissions/:permission_id", gatewayHandler.ProxyRequest("auth-service"))
				admin.DELETE("/permissions/:permission_id", gatewayHandler.ProxyRequest("auth-service"))

				// Role-Permission management
				admin.POST("/roles/:role_id/permissions", gatewayHandler.ProxyRequest("auth-service"))
				admin.DELETE("/roles/:role_id/permissions/:perm_id", gatewayHandler.ProxyRequest("auth-service"))
				admin.GET("/roles/:role_id/permissions", gatewayHandler.ProxyRequest("auth-service"))

				// User-Role management
				admin.POST("/users/:user_id/roles", gatewayHandler.ProxyRequest("auth-service"))
				admin.DELETE("/users/:user_id/roles/:role_id", gatewayHandler.ProxyRequest("auth-service"))
				admin.GET("/users/:user_id/roles", gatewayHandler.ProxyRequest("auth-service"))
				admin.PUT("/users/:user_id/roles", gatewayHandler.ProxyRequest("auth-service"))
			}
		}

		// User service routes
		users := api.Group("/v1/users")
		{
			users.POST("", gatewayHandler.ProxyRequest("user-service"))
			users.GET("/:id", gatewayHandler.ProxyRequest("user-service"))
			users.PUT("/:id", gatewayHandler.ProxyRequest("user-service"))
			users.PATCH("/:id", gatewayHandler.ProxyRequest("user-service"))
			users.DELETE("/:id", gatewayHandler.ProxyRequest("user-service"))
			users.GET("", gatewayHandler.ProxyRequest("user-service"))
		}
	}

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		logger.Info(fmt.Sprintf("Starting API Gateway on %s", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start API Gateway", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down API Gateway...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("API Gateway forced to shutdown", err)
	}

	logger.Info("API Gateway exited")
}

// httpTokenRevocationChecker implements TokenRevocationChecker using HTTP calls to auth-service
type httpTokenRevocationChecker struct {
	authServiceURL string
	logger         *logrus.Logger
}

func (c *httpTokenRevocationChecker) IsTokenRevoked(tokenString string) bool {
	// Call auth-service to validate the token
	req, err := http.NewRequest("POST", c.authServiceURL+"/api/v1/auth/validate-token", nil)
	if err != nil {
		c.logger.WithError(err).Error("Failed to create token validation request")
		return true // Consider token revoked if we can't check
	}

	req.Header.Set("Authorization", "Bearer "+tokenString)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.logger.WithError(err).Warn("Failed to validate token with auth-service")
		return true // Consider token revoked if service is unavailable
	}
	defer resp.Body.Close()

	return resp.StatusCode != http.StatusOK
}

// checkAuthServiceHealth checks if auth-service is healthy before attempting key fetch
func checkAuthServiceHealth(authServiceURL string, logger *logrus.Logger) error {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(authServiceURL + "/health")
	if err != nil {
		logger.WithError(err).Warn("Auth-service health check failed - service may not be ready")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.WithField("status", resp.StatusCode).Warn("Auth-service health check returned non-200 status")
		return fmt.Errorf("auth-service health check returned status %d", resp.StatusCode)
	}

	logger.Debug("Auth-service health check passed")
	return nil
}

func fetchPublicKeyFromAuthService(authServiceURL string, logger *logrus.Logger) (interface{}, error) {
	const maxRetries = 10
	const initialDelay = time.Second
	const maxDelay = 30 * time.Second

	logger.Info("Starting JWT public key fetch from auth-service with retry logic")

	// First, check if auth-service is healthy
	if err := checkAuthServiceHealth(authServiceURL, logger); err != nil {
		logger.WithError(err).Warn("Skipping JWT public key fetch due to auth-service health check failure")
		return nil, fmt.Errorf("auth-service health check failed: %w", err)
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		startTime := time.Now()

		// Calculate delay for exponential backoff (except first attempt)
		if attempt > 0 {
			delay := time.Duration(1<<uint(attempt-1)) * initialDelay // 2^(attempt-1) * initialDelay
			if delay > maxDelay {
				delay = maxDelay
			}
			logger.WithFields(logrus.Fields{
				"attempt":     attempt + 1,
				"max_retries": maxRetries,
				"delay":       delay.String(),
			}).Warn("Retrying JWT public key fetch after delay")
			time.Sleep(delay)
		}

		logger.WithFields(logrus.Fields{
			"attempt":     attempt + 1,
			"max_retries": maxRetries,
		}).Debug("Attempting to fetch JWT public key from auth-service")

		// Fetch the public key from auth-service
		resp, err := http.Get("http://auth-service:8083/public-key")
		if err != nil {
			duration := time.Since(startTime)
			logger.WithError(err).WithFields(logrus.Fields{
				"attempt":  attempt + 1,
				"duration": duration.String(),
			}).Warn("Failed to fetch public key from auth-service")
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("failed to fetch public key after %d attempts: %w", maxRetries, err)
			}
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			duration := time.Since(startTime)
			logger.WithFields(logrus.Fields{
				"attempt":  attempt + 1,
				"status":   resp.StatusCode,
				"duration": duration.String(),
			}).Warn("Auth-service returned non-200 status for public key")
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("auth-service returned status %d after %d attempts", resp.StatusCode, maxRetries)
			}
			continue
		}

		// Read the PEM-encoded public key
		pemData, err := io.ReadAll(resp.Body)
		if err != nil {
			duration := time.Since(startTime)
			logger.WithError(err).WithFields(logrus.Fields{
				"attempt":  attempt + 1,
				"duration": duration.String(),
			}).Warn("Failed to read public key response")
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("failed to read public key response after %d attempts: %w", maxRetries, err)
			}
			continue
		}

		// Parse the PEM-encoded public key
		block, _ := pem.Decode(pemData)
		if block == nil {
			duration := time.Since(startTime)
			logger.WithFields(logrus.Fields{
				"attempt":  attempt + 1,
				"duration": duration.String(),
			}).Warn("Failed to decode PEM block")
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("invalid PEM data after %d attempts", maxRetries)
			}
			continue
		}

		publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			duration := time.Since(startTime)
			logger.WithError(err).WithFields(logrus.Fields{
				"attempt":  attempt + 1,
				"duration": duration.String(),
			}).Warn("Failed to parse public key")
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("failed to parse public key after %d attempts: %w", maxRetries, err)
			}
			continue
		}

		rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
		if !ok {
			duration := time.Since(startTime)
			logger.WithFields(logrus.Fields{
				"attempt":  attempt + 1,
				"duration": duration.String(),
			}).Warn("Public key is not RSA")
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("public key is not RSA after %d attempts", maxRetries)
			}
			continue
		}

		duration := time.Since(startTime)
		logger.WithFields(logrus.Fields{
			"attempt":  attempt + 1,
			"duration": duration.String(),
		}).Info("Successfully fetched and parsed JWT public key from auth-service")
		return rsaPublicKey, nil
	}

	// This should never be reached, but just in case
	return nil, fmt.Errorf("unexpected error: exceeded max retries without proper error handling")
}

// getCachedKey returns the cached key if it's still valid, otherwise refreshes it
func getCachedKey(authServiceURL string, logger *logrus.Logger) (interface{}, error) {
	globalKeyCache.mutex.RLock()
	if globalKeyCache.key != nil && time.Since(globalKeyCache.fetchedAt) < globalKeyCache.ttl {
		key := globalKeyCache.key
		globalKeyCache.mutex.RUnlock()
		return key, nil
	}
	globalKeyCache.mutex.RUnlock()

	// Key is expired or missing, refresh it
	globalKeyCache.mutex.Lock()
	defer globalKeyCache.mutex.Unlock()

	// Double-check after acquiring write lock
	if globalKeyCache.key != nil && time.Since(globalKeyCache.fetchedAt) < globalKeyCache.ttl {
		return globalKeyCache.key, nil
	}

	logger.Info("Refreshing JWT public key cache")
	key, err := fetchPublicKeyFromAuthService(authServiceURL, logger)
	if err != nil {
		// If fetch fails, try to use existing key if available (better than nothing)
		if globalKeyCache.key != nil {
			logger.WithError(err).Warn("Failed to refresh JWT key, using expired cached key")
			return globalKeyCache.key, nil
		}
		return nil, err
	}

	globalKeyCache.key = key
	globalKeyCache.fetchedAt = time.Now()
	logger.Info("Successfully refreshed JWT public key cache")
	return key, nil
}

// startKeyRefreshRoutine starts a background goroutine to periodically refresh the key
func startKeyRefreshRoutine(authServiceURL string, logger *logrus.Logger) {
	go func() {
		ticker := time.NewTicker(30 * time.Minute) // Check every 30 minutes
		defer ticker.Stop()

		for range ticker.C {
			logger.Debug("Periodic JWT key refresh check")
			_, err := getCachedKey(authServiceURL, logger)
			if err != nil {
				logger.WithError(err).Warn("Periodic JWT key refresh failed")
			}
		}
	}()
}
