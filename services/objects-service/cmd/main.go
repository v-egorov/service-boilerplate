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
	authclient "github.com/v-egorov/service-boilerplate/services/objects-service/internal/client"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/handlers"
	permiddleware "github.com/v-egorov/service-boilerplate/services/objects-service/internal/permiddleware"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/repository"
	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/services"
)

type contextKey string

const requestIDKey contextKey = "request_id"

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
	var objectTypeHandler *handlers.ObjectTypeHandler
	var objectHandler *handlers.ObjectHandler
	var relationshipTypeHandler *handlers.RelationshipTypeHandler
	var relationshipHandler *handlers.RelationshipHandler
	var healthHandler *handlers.HealthHandler

	if db != nil {
		// Initialize new repository layer
		pgDatabase := repository.NewPGDatabase(db.GetPool())
		repoOptions := repository.DefaultRepositoryOptions()

		// Initialize repositories
		objectTypeRepo := repository.NewObjectTypeRepository(pgDatabase, repoOptions)
		objectRepo := repository.NewObjectRepository(pgDatabase, repoOptions)
		relationshipTypeRepo := repository.NewRelationshipTypeRepository(pgDatabase, repoOptions)

		// Initialize services
		objectTypeService := services.NewObjectTypeService(objectTypeRepo)
		objectService := services.NewObjectService(objectRepo, objectTypeRepo)
		relationshipTypeService := services.NewRelationshipTypeService(relationshipTypeRepo)
		relationshipRepo := repository.NewRelationshipRepository(pgDatabase, repoOptions, objectRepo)
		relationshipService := services.NewRelationshipService(relationshipRepo, relationshipTypeRepo, objectRepo)

		// Initialize handlers
		objectTypeHandler = handlers.NewObjectTypeHandler(objectTypeService, logger.Logger)
		objectHandler = handlers.NewObjectHandler(objectService, logger.Logger)
		relationshipTypeHandler = handlers.NewRelationshipTypeHandler(relationshipTypeService, logger.Logger)
		relationshipHandler = handlers.NewRelationshipHandler(relationshipService, logger.Logger)
		healthHandler = handlers.NewHealthHandler(db.GetPool(), logger.Logger, cfg)
	} else {
		// Initialize handlers without database
		healthHandler = handlers.NewHealthHandler(nil, logger.Logger, cfg)
	}

	// Initialize auth-client for permission checks
	authClient := authclient.NewAuthClient(authclient.AuthClientConfig{
		BaseURL: cfg.AuthService.URL,
		Timeout: time.Duration(cfg.AuthService.Timeout) * time.Second,
	}, logger.Logger)

	// Initialize permission middleware (fail-closed)
	permissionMiddleware := permiddleware.NewPermissionMiddleware(permiddleware.PermissionMiddlewareConfig{
		AuthClient: authClient,
		Logger:     logger.Logger,
	})

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

		// Object Type endpoints (only if database is available)
		if objectTypeHandler != nil {
			// Object Types - Admin only (create, update, delete)
			// Note: RequireAuth removed - gateway already validates JWT and forwards user info via X-User-* headers
			objectTypesAdmin := v1.Group("/object-types")
			objectTypesAdmin.Use(permissionMiddleware("object-types:create", "object-types:update", "object-types:delete"))
			{
				objectTypesAdmin.POST("", objectTypeHandler.Create)
				objectTypesAdmin.PUT("/:id", objectTypeHandler.Update)
				objectTypesAdmin.DELETE("/:id", objectTypeHandler.Delete)
			}

			// Object Types - Read (authenticated users)
			// Note: RequireAuth removed - gateway forwards user info via X-User-* headers
			objectTypesRead := v1.Group("/object-types")
			objectTypesRead.Use(permissionMiddleware("object-types:read"))
			{
				objectTypesRead.GET("/:id", objectTypeHandler.GetByID)
				objectTypesRead.GET("/name/:name", objectTypeHandler.GetByName)
				objectTypesRead.GET("", objectTypeHandler.List)
				objectTypesRead.GET("/search", objectTypeHandler.Search)
				objectTypesRead.GET("/:id/tree", objectTypeHandler.GetTree)
				objectTypesRead.GET("/:id/children", objectTypeHandler.GetChildren)
				objectTypesRead.GET("/:id/descendants", objectTypeHandler.GetDescendants)
				objectTypesRead.GET("/:id/ancestors", objectTypeHandler.GetAncestors)
				objectTypesRead.GET("/:id/path", objectTypeHandler.GetPath)
				objectTypesRead.GET("/:id/subtree-count", objectTypeHandler.GetSubtreeObjectCount)
				objectTypesRead.POST("/:id/validate-move", objectTypeHandler.ValidateMove)
			}
		}

		// Object endpoints (only if database is available)
		if objectHandler != nil {
			// Objects - Create
			// Note: RequireAuth removed - gateway forwards user info via X-User-* headers
			objectsCreate := v1.Group("/objects")
			objectsCreate.Use(permissionMiddleware("objects:create"))
			{
				objectsCreate.POST("", objectHandler.Create)
			}

			// Objects - Read
			// Note: RequireAuth removed - gateway forwards user info via X-User-* headers
			objectsRead := v1.Group("/objects")
			objectsRead.Use(permissionMiddleware("objects:read:all", "objects:read:own"))
			{
				objectsRead.GET("/:id", objectHandler.GetByID)
				objectsRead.GET("/public-id/:public_id", objectHandler.GetByPublicID)
				objectsRead.GET("/name/:name", objectHandler.GetByName)
				objectsRead.GET("", objectHandler.List)
				objectsRead.GET("/search", objectHandler.Search)
				objectsRead.GET("/:id/children", objectHandler.GetChildren)
				objectsRead.GET("/:id/descendants", objectHandler.GetDescendants)
				objectsRead.GET("/:id/ancestors", objectHandler.GetAncestors)
				objectsRead.GET("/:id/path", objectHandler.GetPath)
				objectsRead.GET("/stats", objectHandler.GetStats)
				// Relationships for object (using public-id for UUID-based lookup)
				objectsRead.GET("/public-id/:public_id/relationships", relationshipHandler.GetForObject)
				objectsRead.GET("/public-id/:public_id/relationships/:type_key", relationshipHandler.GetForObjectByType)
			}

			// Objects - Update
			// Note: RequireAuth removed - gateway forwards user info via X-User-* headers
			objectsUpdate := v1.Group("/objects")
			objectsUpdate.Use(permissionMiddleware("objects:update:all", "objects:update:own"))
			{
				objectsUpdate.PUT("/:id", objectHandler.Update)
				objectsUpdate.PUT("/:id/metadata", objectHandler.UpdateMetadata)
				objectsUpdate.POST("/:id/tags", objectHandler.AddTags)
				objectsUpdate.DELETE("/:id/tags", objectHandler.RemoveTags)
			}

			// Objects - Delete
			// Note: RequireAuth removed - gateway forwards user info via X-User-* headers
			objectsDelete := v1.Group("/objects")
			objectsDelete.Use(permissionMiddleware("objects:delete:all", "objects:delete:own"))
			{
				objectsDelete.DELETE("/:id", objectHandler.Delete)
			}

			// Objects - Bulk operations
			objectsBulk := v1.Group("/objects")
			objectsBulk.Use(middleware.RequireAuth())
			objectsBulk.Use(permissionMiddleware("objects:create", "objects:update:all", "objects:delete:all"))
			{
				objectsBulk.POST("/bulk", objectHandler.BulkCreate)
				objectsBulk.PUT("/bulk", objectHandler.BulkUpdate)
				objectsBulk.DELETE("/bulk", objectHandler.BulkDelete)
			}

			// Relationship Types endpoints
			if relationshipTypeHandler != nil {
				// Relationship Types - Admin only (create, update, delete)
				relationshipTypesAdmin := v1.Group("/relationship-types")
				relationshipTypesAdmin.Use(permissionMiddleware("relationship-types:create", "relationship-types:update", "relationship-types:delete"))
				{
					relationshipTypesAdmin.POST("", relationshipTypeHandler.Create)
					relationshipTypesAdmin.PUT("/:type_key", relationshipTypeHandler.Update)
					relationshipTypesAdmin.DELETE("/:type_key", relationshipTypeHandler.Delete)
				}

				// Relationship Types - Read (authenticated users)
				relationshipTypesRead := v1.Group("/relationship-types")
				relationshipTypesRead.Use(permissionMiddleware("relationship-types:read"))
				{
					relationshipTypesRead.GET("/:type_key", relationshipTypeHandler.GetByTypeKey)
					relationshipTypesRead.GET("", relationshipTypeHandler.List)
				}
			}

			// Relationships endpoints
			if relationshipHandler != nil {
				// Relationships - Create (authenticated users)
				relationshipsCreate := v1.Group("/relationships")
				relationshipsCreate.Use(permissionMiddleware("relationships:create"))
				{
					relationshipsCreate.POST("", relationshipHandler.Create)
				}

				// Relationships - Read (authenticated users)
				relationshipsRead := v1.Group("/relationships")
				relationshipsRead.Use(permissionMiddleware("relationships:read"))
				{
					relationshipsRead.GET("/:public_id", relationshipHandler.GetByPublicID)
					relationshipsRead.GET("", relationshipHandler.List)
				}

				// Relationships - Update (authenticated users)
				relationshipsUpdate := v1.Group("/relationships")
				relationshipsUpdate.Use(permissionMiddleware("relationships:update"))
				{
					relationshipsUpdate.PUT("/:public_id", relationshipHandler.Update)
				}

				// Relationships - Delete (authenticated users)
				relationshipsDelete := v1.Group("/relationships")
				relationshipsDelete.Use(permissionMiddleware("relationships:delete"))
				{
					relationshipsDelete.DELETE("/:public_id", relationshipHandler.Delete)
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
			ctx := context.WithValue(c.Request.Context(), requestIDKey, requestID)
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
