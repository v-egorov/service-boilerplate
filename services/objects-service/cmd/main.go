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
	"github.com/v-egorov/service-boilerplate/common/middleware"
	"github.com/v-egorov/service-boilerplate/common/tracing"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/handlers"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/repository"
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

	// Initialize service request logger
	serviceLogger := logging.NewServiceRequestLogger(logger.Logger, cfg.App.Name)

	// Initialize standard logger for structured operation logging
	// standardLogger := logging.NewStandardLogger(logger.Logger, cfg.App.Name)

	// Initialize repository and service only if database is available
	var healthHandler *handlers.HealthHandler

	if db != nil {
		// Initialize new repository layer
		pgDatabase := repository.NewPGDatabase(db.GetPool())
		repoOptions := repository.DefaultRepositoryOptions()

		// Initialize repositories
		objectTypeRepo := repository.NewObjectTypeRepository(pgDatabase, repoOptions)
		objectRepo := repository.NewObjectRepository(pgDatabase, repoOptions)

		// Initialize health handler
		healthHandler = handlers.NewHealthHandler(db.GetPool(), logger.Logger, cfg)

		// TODO: Create and wire up new handlers and routes for ObjectType and Object repositories
		// objectTypeHandler := handlers.NewObjectTypeHandler(objectTypeRepo, logger, standardLogger)
		// objectHandler := handlers.NewObjectHandler(objectRepo, logger, standardLogger)
		//
		// Add routes like:
		// objectTypes := v1.Group("/object-types")
		// {
		//     objectTypes.POST("", objectTypeHandler.CreateObjectType)
		//     objectTypes.GET("/:id", objectTypeHandler.GetObjectType)
		//     objectTypes.GET("", objectTypeHandler.ListObjectTypes)
		//     objectTypes.GET("/:id/tree", objectTypeHandler.GetTree)
		// }
		//
		// objects := v1.Group("/objects")
		// {
		//     objects.POST("", objectHandler.CreateObject)
		//     objects.GET("/:id", objectHandler.GetObject)
		//     objects.GET("", objectHandler.ListObjects)
		//     objects.GET("/search", objectHandler.SearchObjects)
		// }
		_ = objectTypeRepo
		_ = objectRepo
	} else {
		// Initialize handlers without database
		healthHandler = handlers.NewHealthHandler(nil, logger.Logger, cfg)
	}

	// Initialize alert manager
	alertManager := alerting.NewAlertManager(logger.Logger, cfg.App.Name, &cfg.Alerting, serviceLogger.GetMetricsCollector())

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
	router.Use(requestIDMiddleware()) // Extract request_id from headers and store in context
	if cfg.Tracing.Enabled {
		router.Use(tracing.HTTPMiddleware(cfg.Tracing.ServiceName))
	}
	// JWT middleware for authentication (configure jwtSecret for token validation)
	// For development, you may need to share JWT public key with auth-service
	router.Use(middleware.JWTMiddleware(nil, logger.Logger, nil)) // nil disables JWT validation
	router.Use(serviceLogger.RequestResponseLogger())

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

		// TODO: Add new ObjectType and Object endpoints when handlers are created
		// objectTypes := v1.Group("/object-types")
		// {
		//     objectTypes.POST("", objectTypeHandler.CreateObjectType)
		//     objectTypes.GET("/:id", objectTypeHandler.GetObjectType)
		//     objectTypes.GET("", objectTypeHandler.ListObjectTypes)
		//     objectTypes.GET("/:id/tree", objectTypeHandler.GetTree)
		// }
		//
		// objects := v1.Group("/objects")
		// {
		//     objects.POST("", objectHandler.CreateObject)
		//     objects.GET("/:id", objectHandler.GetObject)
		//     objects.GET("", objectHandler.ListObjects)
		//     objects.GET("/search", objectHandler.SearchObjects)
		// }
	}

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		logger.Info(fmt.Sprintf("Starting objects-service service on %s", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start objects-service service", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down objects-service service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("objects-service service forced to shutdown", err)
	}

	logger.Info("objects-service service exited")
}

// requestIDMiddleware extracts X-Request-ID from headers and stores in context for propagation
func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID != "" {
			ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
			c.Request = c.Request.WithContext(ctx)
		}
		c.Next()
	}
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
