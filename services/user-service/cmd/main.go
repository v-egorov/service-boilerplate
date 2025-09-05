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
	"github.com/vegorov/service-boilerplate/common/config"
	"github.com/vegorov/service-boilerplate/common/database"
	"github.com/vegorov/service-boilerplate/common/logging"
	"github.com/vegorov/service-boilerplate/services/user-service/internal/handlers"
	"github.com/vegorov/service-boilerplate/services/user-service/internal/repository"
	"github.com/vegorov/service-boilerplate/services/user-service/internal/services"
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
		logger.Fatal("Failed to connect to database", err)
	}
	defer db.Close()

	// Initialize repository
	userRepo := repository.NewUserRepository(db.GetPool(), logger.Logger)

	// Initialize service
	userService := services.NewUserService(userRepo, logger.Logger)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService, logger.Logger)

	// Setup Gin router
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("", userHandler.CreateUser)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
			users.GET("", userHandler.ListUsers)
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
