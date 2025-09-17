package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/v-egorov/service-boilerplate/api-gateway/internal/handlers"
	"github.com/v-egorov/service-boilerplate/api-gateway/internal/middleware"
	"github.com/v-egorov/service-boilerplate/api-gateway/internal/services"
	"github.com/v-egorov/service-boilerplate/common/config"
	"github.com/v-egorov/service-boilerplate/common/logging"
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
		Level:       cfg.Logging.Level,
		Format:      cfg.Logging.Format,
		Output:      cfg.Logging.Output,
		ServiceName: cfg.App.Name,
	})

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

	// Setup Gin router
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RequestIDMiddleware())

	// Health check endpoints (public, no auth required)
	router.GET("/health", gatewayHandler.LivenessHandler)
	router.GET("/ready", gatewayHandler.ReadinessHandler)
	router.GET("/live", gatewayHandler.LivenessHandler)
	router.GET("/status", gatewayHandler.StatusHandler) // Direct status endpoint
	router.GET("/ping", gatewayHandler.PingHandler)     // Direct ping endpoint

	// Public monitoring endpoints (no auth required)
	router.GET("/api/v1/status", gatewayHandler.StatusHandler)
	router.GET("/api/v1/ping", gatewayHandler.PingHandler)

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
	api.Use(middleware.AuthMiddleware())
	{
		// Protected auth routes
		protectedAuth := api.Group("/v1/auth")
		{
			protectedAuth.GET("/me", gatewayHandler.ProxyRequest("auth-service"))
		}

		// User service routes
		users := api.Group("/v1/users")
		{
			users.POST("", gatewayHandler.ProxyRequest("user-service"))
			users.GET("/:id", gatewayHandler.ProxyRequest("user-service"))
			users.PUT("/:id", gatewayHandler.ProxyRequest("user-service"))
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
