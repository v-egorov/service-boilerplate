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
	"github.com/v-egorov/service-boilerplate/common/alerting"
	"github.com/v-egorov/service-boilerplate/common/config"
	"github.com/v-egorov/service-boilerplate/common/database"
	"github.com/v-egorov/service-boilerplate/common/logging"
	"github.com/v-egorov/service-boilerplate/common/tracing"
	"github.com/v-egorov/service-boilerplate/services/user-service/internal/handlers"
	"github.com/v-egorov/service-boilerplate/services/user-service/internal/repository"
	"github.com/v-egorov/service-boilerplate/services/user-service/internal/services"
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

	// Initialize repository and service only if database is available
	var userHandler *handlers.UserHandler
	var healthHandler *handlers.HealthHandler

	if db != nil {
		// Initialize repository
		userRepo := repository.NewUserRepository(db.GetPool(), logger.Logger)

		// Initialize service
		userService := services.NewUserService(userRepo, logger.Logger)

		// Initialize handlers
		userHandler = handlers.NewUserHandler(userService, logger.Logger)
		healthHandler = handlers.NewHealthHandler(db.GetPool(), logger.Logger, cfg)
	} else {
		// Initialize handlers without database
		healthHandler = handlers.NewHealthHandler(nil, logger.Logger, cfg)
		userHandler = nil // User operations won't work without database
	}

	// Initialize service request logger
	serviceLogger := logging.NewServiceRequestLogger(logger.Logger, "user-service")

	// Initialize alert manager
	alertManager := alerting.NewAlertManager(logger.Logger, "user-service", &cfg.Alerting, serviceLogger.GetMetricsCollector())

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

	// Middleware
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(serviceLogger.RequestResponseLogger())
	if cfg.Tracing.Enabled {
		router.Use(tracing.HTTPMiddleware(cfg.Tracing.ServiceName))
	}

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

		// Metrics endpoint
		v1.GET("/metrics", func(c *gin.Context) {
			metrics := serviceLogger.GetMetricsCollector().GetMetrics()
			c.JSON(http.StatusOK, metrics)
		})

		// Alerts endpoint
		v1.GET("/alerts", func(c *gin.Context) {
			alerts := alertManager.GetActiveAlerts()
			c.JSON(http.StatusOK, gin.H{"alerts": alerts})
		})

		// User endpoints (only if database is available)
		if userHandler != nil {
			users := v1.Group("/users")
			{
				users.POST("", userHandler.CreateUser)
				users.GET("/:id", userHandler.GetUser)
				users.GET("/by-email/:email", userHandler.GetUserByEmail)
				users.GET("/by-email/:email/with-password", userHandler.GetUserWithPasswordByEmail)
				users.PUT("/:id", userHandler.ReplaceUser)  // Full resource replacement
				users.PATCH("/:id", userHandler.UpdateUser) // Partial resource update
				users.DELETE("/:id", userHandler.DeleteUser)
				users.GET("", userHandler.ListUsers)
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
		logger.Info(fmt.Sprintf("Starting server on %s", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", err)
	}

	logger.Info("Server exited")
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
