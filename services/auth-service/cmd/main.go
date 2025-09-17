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
	"github.com/v-egorov/service-boilerplate/common/config"
	"github.com/v-egorov/service-boilerplate/common/database"
	"github.com/v-egorov/service-boilerplate/common/logging"
	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/handlers"
	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/repository"
	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/services"
	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/utils"
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

	// Initialize database
	dbConfig := database.Config{
		Host:        cfg.Database.Host,
		Port:        cfg.Database.Port,
		User:        cfg.Database.User,
		Password:    cfg.Database.Password,
		Database:    cfg.Database.Database,
		SSLMode:     cfg.Database.SSLMode,
		MaxConns:    cfg.Database.MaxConns,
		MinConns:    cfg.Database.MinConns,
		MaxConnIdle: time.Duration(cfg.Database.MaxConnIdle) * time.Second,
		MaxConnLife: time.Duration(cfg.Database.MaxConnLife) * time.Second,
	}

	db, err := database.NewPostgresDB(dbConfig, logger.Logger)
	if err != nil {
		logger.Warn("Failed to connect to database, running without database", err)
		db = nil // Set to nil to indicate no database connection
	} else {
		defer db.Close()
	}

	// Initialize JWT utils
	jwtUtils, err := utils.NewJWTUtils()
	if err != nil {
		logger.Fatal("Failed to initialize JWT utils", err)
	}

	// Initialize repository and service only if database is available
	var authHandler *handlers.AuthHandler
	var healthHandler *handlers.HealthHandler

	if db != nil {
		// Initialize repository
		authRepo := repository.NewAuthRepository(db.GetPool())

		// Initialize service
		authService := services.NewAuthService(authRepo, jwtUtils, logger.Logger)

		// Initialize handlers
		authHandler = handlers.NewAuthHandler(authService, logger.Logger)
		healthHandler = handlers.NewHealthHandler(db.GetPool(), logger.Logger, cfg)
	} else {
		// Initialize handlers without database
		healthHandler = handlers.NewHealthHandler(nil, logger.Logger, cfg)
		authHandler = nil // Auth operations won't work without database
	}

	// Setup Gin router
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// Health check endpoints (public, no auth required)
	router.GET("/health", healthHandler.LivenessHandler)
	router.GET("/ready", healthHandler.ReadinessHandler)
	router.GET("/live", healthHandler.LivenessHandler)
	router.GET("/status", healthHandler.StatusHandler) // Direct status endpoint
	router.GET("/ping", healthHandler.PingHandler)     // Direct ping endpoint

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Health endpoints
		v1.GET("/status", healthHandler.StatusHandler)
		v1.GET("/ping", healthHandler.PingHandler)

		// Auth endpoints (only if database is available)
		if authHandler != nil {
			auth := v1.Group("/auth")
			{
				auth.POST("/login", authHandler.Login)
				auth.POST("/register", authHandler.Register)
				auth.POST("/refresh", authHandler.RefreshToken)
				auth.POST("/logout", authHandler.Logout)

				// Protected routes
				protected := auth.Group("")
				protected.Use(authHandler.AuthMiddleware())
				{
					protected.GET("/me", authHandler.GetCurrentUser)
				}
			}
		}
	}

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		logger.Info(fmt.Sprintf("Starting auth-service service on %s", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start auth-service service", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down auth-service service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("auth-service service forced to shutdown", err)
	}

	logger.Info("auth-service service exited")
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
