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

	// Register services - use service names in Docker, localhost for local development
	userServiceURL := "http://user-service:8081"
	if cfg.App.Environment == "development" && os.Getenv("DOCKER_ENV") != "true" {
		userServiceURL = "http://localhost:8081"
	}
	serviceRegistry.RegisterService("user-service", userServiceURL)

	// Register auth-service
	serviceRegistry.RegisterService("auth-service", "http://auth-service:8083")

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
		// Try to fetch public key from auth-service
		logger.Info("JWT public key not configured, attempting to fetch from auth-service")
		publicKey, err := fetchPublicKeyFromAuthService(logger.Logger)
		if err != nil {
			logger.Warn("Failed to fetch public key from auth-service, JWT validation disabled", err)
			jwtPublicKey = nil
			revocationChecker = nil
		} else {
			jwtPublicKey = publicKey
			logger.Info("Successfully fetched JWT public key from auth-service")

			// Create HTTP-based revocation checker
			revocationChecker = &httpTokenRevocationChecker{
				authServiceURL: "http://auth-service:8083",
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

func fetchPublicKeyFromAuthService(logger *logrus.Logger) (interface{}, error) {
	// Fetch the public key from auth-service
	resp, err := http.Get("http://auth-service:8083/public-key")
	if err != nil {
		logger.WithError(err).Warn("Failed to fetch public key from auth-service")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Warn("Auth-service returned non-200 status for public key", "status", resp.StatusCode)
		return nil, fmt.Errorf("auth-service returned status %d", resp.StatusCode)
	}

	// Read the PEM-encoded public key
	pemData, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.WithError(err).Warn("Failed to read public key response")
		return nil, err
	}

	// Parse the PEM-encoded public key
	block, _ := pem.Decode(pemData)
	if block == nil {
		logger.Warn("Failed to decode PEM block")
		return nil, fmt.Errorf("invalid PEM data")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		logger.WithError(err).Warn("Failed to parse public key")
		return nil, err
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		logger.Warn("Public key is not RSA")
		return nil, fmt.Errorf("public key is not RSA")
	}

	logger.Info("Successfully fetched and parsed JWT public key from auth-service")
	return rsaPublicKey, nil
}
