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
	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/client"
	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/handlers"
	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/repository"
	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/services"
	"github.com/v-egorov/service-boilerplate/services/auth-service/internal/utils"
)

// authServiceRevocationChecker implements TokenRevocationChecker for auth-service
type authServiceRevocationChecker struct {
	authService *services.AuthService
}

func (c *authServiceRevocationChecker) IsTokenRevoked(tokenString string) bool {
	_, err := c.authService.ValidateToken(context.Background(), tokenString)
	return err != nil // If validation fails, token is considered revoked or invalid
}

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
	var authHandler *handlers.AuthHandler
	var healthHandler *handlers.HealthHandler
	var revocationChecker middleware.TokenRevocationChecker
	var jwtUtils *utils.JWTUtils
	var keyRotationManager *services.KeyRotationManager

	if db != nil {
		// Initialize JWT utils with database connection
		var err error
		jwtUtils, err = utils.NewJWTUtils(db.Pool)
		if err != nil {
			logger.Fatal("Failed to initialize JWT utils", err)
		}

		// Initialize key rotation manager
		rotationConfig := services.RotationConfig{
			Enabled:        true,
			Type:           "time",
			IntervalDays:   30,
			MaxTokens:      100000,
			OverlapMinutes: 60,
			CheckInterval:  1 * time.Hour, // Check every hour
		}
		keyRotationManager = services.NewKeyRotationManager(jwtUtils, db.Pool, rotationConfig, logger.Logger)

		// Initialize repository
		authRepo := repository.NewAuthRepository(db.GetPool())

		// Initialize user service client
		// Always use Docker service name since we're running in containers
		userServiceURL := "http://user-service:8081/api/v1"
		userClient := client.NewUserClient(userServiceURL, logger.Logger)

		// Initialize service
		authService := services.NewAuthService(authRepo, userClient, jwtUtils, logger.Logger)

		// Initialize handlers
		authHandler = handlers.NewAuthHandler(authService, logger.Logger)
		healthHandler = handlers.NewHealthHandler(db.GetPool(), jwtUtils, keyRotationManager, logger.Logger, cfg)

		// Create token revocation checker for JWT middleware
		revocationChecker = &authServiceRevocationChecker{
			authService: authService,
		}
	} else {
		// Initialize handlers without database
		healthHandler = handlers.NewHealthHandler(nil, nil, nil, logger.Logger, cfg)
		authHandler = nil // Auth operations won't work without database
		revocationChecker = nil
		jwtUtils = nil
		keyRotationManager = nil
	}

	// Initialize service request logger
	serviceLogger := logging.NewServiceRequestLogger(logger.Logger, "auth-service")

	// Initialize alert manager
	alertManager := alerting.NewAlertManager(logger.Logger, "auth-service", &cfg.Alerting, serviceLogger.GetMetricsCollector())

	// Set JWT health checker if JWT utils are available
	if jwtUtils != nil {
		alertManager.SetJWTChecker(func() (bool, string) {
			keyID := jwtUtils.GetKeyID()
			if keyID == "" {
				return false, "No active JWT key found"
			}
			_, err := jwtUtils.GetPublicKeyPEM()
			if err != nil {
				return false, fmt.Sprintf("Failed to access JWT key: %v", err)
			}
			return true, ""
		})
	}

	// Start key rotation manager
	if keyRotationManager != nil {
		go func() {
			ctx := context.Background()
			keyRotationManager.Start(ctx)
		}()
	}

	// Start alert checking goroutine
	if cfg.Alerting.Enabled {
		go func() {
			ticker := time.NewTicker(1 * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				alertManager.CheckMetrics()
				alertManager.CheckJWTKeys()
			}
		}()
	}

	// Setup Gin router
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Request ID middleware to extract X-Request-ID header and store in context
	requestIDMiddleware := func() gin.HandlerFunc {
		return func(c *gin.Context) {
			requestID := c.GetHeader("X-Request-ID")
			if requestID != "" {
				ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
				c.Request = c.Request.WithContext(ctx)
			}
			c.Next()
		}
	}

	// Middleware
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(requestIDMiddleware())
	if cfg.Tracing.Enabled {
		router.Use(tracing.HTTPMiddleware(cfg.Tracing.ServiceName))
	}
	// JWT middleware for authentication
	router.Use(middleware.JWTMiddleware(jwtUtils.GetPublicKey(), logger.Logger, revocationChecker))
	router.Use(serviceLogger.RequestResponseLogger())

	// Health check endpoints (public, no auth required)
	router.GET("/health", healthHandler.LivenessHandler)
	router.GET("/ready", healthHandler.ReadinessHandler)
	router.GET("/live", healthHandler.LivenessHandler)
	router.GET("/status", healthHandler.StatusHandler) // Direct status endpoint
	router.GET("/ping", healthHandler.PingHandler)     // Direct ping endpoint

	// Public key endpoint (public, no auth required)
	if authHandler != nil {
		router.GET("/public-key", authHandler.GetPublicKey)
	}

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

		// Auth endpoints (only if database is available)
		if authHandler != nil {
			auth := v1.Group("/auth")
			{
				auth.POST("/login", authHandler.Login)
				auth.POST("/register", authHandler.Register)
				auth.POST("/refresh", authHandler.RefreshToken)
				auth.POST("/logout", authHandler.Logout)

				// Token validation endpoint (public - validates the token in the request)
				auth.POST("/validate-token", authHandler.ValidateToken)

				// Protected routes
				protected := auth.Group("")
				protected.Use(middleware.RequireAuth())
				{
					protected.GET("/me", authHandler.GetCurrentUser)
				}

				// Admin routes
				admin := auth.Group("")
				admin.Use(middleware.RequireAuth())
				admin.Use(middleware.RequireRole("admin"))
				{
					admin.POST("/rotate-keys", authHandler.RotateKeys)
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
