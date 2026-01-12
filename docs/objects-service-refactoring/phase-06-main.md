# Phase 6: Main Application

**Estimated Time**: 1 hour
**Status**: ⬜ Not Started
**Dependencies**: Phase 5 (Handlers)

## Overview

Update the main application file (`cmd/main.go`) to wire together all new components: database connection, repositories, services, and handlers. Replace old entity-related initialization with new object type and object components.

## Tasks

### 6.1 Update Main Application File

**File**: `cmd/main.go`

**Steps**:
1. Remove entity-related imports
2. Add new imports for object types and objects
3. Replace repository initialization
4. Replace service initialization
5. Replace handler initialization
6. Register new routes

```go
package main

import (
    "context"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/sirupsen/logrus"
    
    "github.com/v-egorov/service-boilerplate/common/database"
    "github.com/v-egorov/service-boilerplate/services/objects-service/internal/handlers"
    "github.com/v-egorov/service-boilerplate/services/objects-service/internal/repository"
    "github.com/v-egorov/service-boilerplate/services/objects-service/internal/services"
    "github.com/v-egorov/service-boilerplate/services/objects-service/pkg/config"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    logger := logrus.New()
    logger.SetLevel(logrus.InfoLevel)
    if cfg.Environment == "development" {
        logger.SetLevel(logrus.DebugLevel)
    }
    
    db, err := database.NewPostgresDB(database.Config{
        Host:        cfg.DB.Host,
        Port:        cfg.DB.Port,
        User:        cfg.DB.User,
        Password:    cfg.DB.Password,
        Database:    cfg.DB.Database,
        SSLMode:     cfg.DB.SSLMode,
        MaxConns:    cfg.DB.MaxConns,
        MinConns:    cfg.DB.MinConns,
        MaxConnIdle: time.Hour,
        MaxConnLife: 24 * time.Hour,
    }, logger)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()
    
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    pool := db.GetPool()
    repositories := initRepositories(pool, logger)
    services := initServices(repositories, logger)
    handlers := initHandlers(services)
    
    router := setupRouter(handlers, cfg)
    
    server := &http.Server{
        Addr:         ":" + cfg.Port,
        Handler:      router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
    
    go func() {
        logger.Infof("Starting server on port %s", cfg.Port)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatalf("Failed to start server: %v", err)
        }
    }()
    
    gracefulShutdown(server, ctx)
}

func initRepositories(pool *pgxpool.Pool, logger *logrus.Logger) *Repositories {
    return &Repositories{
        ObjectType: repository.NewObjectTypeRepository(pool, logger),
        Object:     repository.NewObjectRepository(pool, logger),
    }
}

func initServices(repos *Repositories, logger *logrus.Logger) *Services {
    return &Services{
        ObjectType: services.NewObjectTypeService(repos.ObjectType, logger),
        Object:     services.NewObjectService(repos.Object, repos.ObjectType, logger),
    }
}

func initHandlers(services *Services) *Handlers {
    return &Handlers{
        ObjectType: handlers.NewObjectTypeHandler(services.ObjectType),
        Object:     handlers.NewObjectHandler(services.Object),
    }
}

func setupRouter(handlers *Handlers, cfg *config.Config, logger *logrus.Logger) *gin.Engine {
    router := gin.New()
    
    if cfg.Environment != "production" {
        gin.SetMode(gin.DebugMode)
        logger.SetLevel(logrus.DebugLevel)
    } else {
        gin.SetMode(gin.ReleaseMode)
        logger.SetLevel(logrus.InfoLevel)
    }
    
    router.Use(gin.Recovery())
    router.Use(gin.Logger())
    router.Use(corsMiddleware())
    router.Use(authMiddleware(logger))
    
    handlers.ObjectType.RegisterRoutes(router)
    handlers.Object.RegisterRoutes(router)
    
    router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": "healthy",
            "service": "objects-service",
        })
    })
    
    router.GET("/", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "service": "objects-service",
            "version": "1.0.0",
        })
    })
    
    return router
}

func corsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        
        c.Next()
    }
}

func authMiddleware(logger *logrus.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.Set("user_id", "system")
            logger.Debug("No auth header, using system user")
            c.Next()
            return
        }
        
        token := extractToken(authHeader)
        if token == "" {
            c.Set("user_id", "system")
            logger.Debug("Invalid auth header format, using system user")
            c.Next()
            return
        }
        
        userID, err := validateToken(token)
        if err != nil {
            c.Set("user_id", "system")
            logger.WithError(err).Debug("Token validation failed, using system user")
            c.Next()
            return
        }
        
        c.Set("user_id", userID)
        logger.WithField("user_id", userID).Debug("User authenticated")
        c.Next()
    }
}

func extractToken(authHeader string) string {
    if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
        return authHeader[7:]
    }
    return ""
}

func validateToken(token string) (string, error) {
    // TODO: Implement proper JWT validation
    // For now, return dummy user ID
    return "user123", nil
}

func gracefulShutdown(server *http.Server, ctx context.Context, logger *logrus.Logger) {
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    logger.Info("Shutting down server...")
    
    shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
    
    if err := server.Shutdown(shutdownCtx); err != nil {
        logger.WithError(err).Error("Server forced to shutdown")
    } else {
        logger.Info("Server shutdown gracefully")
    }
}

type Repositories struct {
    ObjectType *repository.ObjectTypeRepository
    Object     *repository.ObjectRepository
}

type Services struct {
    ObjectType *services.ObjectTypeService
    Object     *services.ObjectService
}

type Handlers struct {
    ObjectType *handlers.ObjectTypeHandler
    Object     *handlers.ObjectHandler
}
```

---

## Checklist

- [ ] Remove entity imports from `cmd/main.go`
- [ ] Add pgx/v5 imports
- [ ] Add common/database imports
- [ ] Add logrus logger initialization
- [ ] Update repository initialization with pgxpool.Pool
- [ ] Update service initialization with logger
- [ ] Update handler initialization
- [ ] Register new routes
- [ ] Verify application compiles: `go build ./cmd/...`
- [ ] Test application startup: `go run cmd/main.go`
- [ ] Verify database connection: check logs for "Successfully connected to PostgreSQL database"
- [ ] Verify health endpoint: `curl http://localhost:8085/health`
- [ ] Verify API endpoints: `curl http://localhost:8085/api/v1/object-types`
- [ ] Verify database.Pool() is used, not sqlx.DB
- [ ] Verify tracing is enabled (check logs or OpenTelemetry UI)
- [ ] Update progress.md

## Testing

```bash
# Build application
cd services/objects-service
go build -o objects-service cmd/main.go

# Run application with environment variables
export DATABASE_URL="postgresql://postgres:password@localhost:5432/objects_service?sslmode=disable"
export PORT="8085"
export ENVIRONMENT="development"
./objects-service

# Or run directly
DATABASE_URL="postgresql://postgres:password@localhost:5432/objects_service?sslmode=disable" \
  PORT="8085" \
  ENVIRONMENT="development" \
  go run cmd/main.go

# Test health endpoint
curl http://localhost:8085/health

# Test root endpoint
curl http://localhost:8085/

# Test object types endpoint
curl http://localhost:8085/api/v1/object-types

# Test objects endpoint
curl http://localhost:8085/api/v1/objects

# Test database connection
export DATABASE_URL="postgresql://postgres:password@localhost:5432/objects_service?sslmode=disable"
psql "$DATABASE_URL" -c "SELECT COUNT(*) FROM object_types;"
```

## Common Issues

**Issue**: Database connection fails
**Solution**: Verify DATABASE_URL environment variable is set correctly and pgxpool.Pool is configured

**Issue**: Routes not registered
**Solution**: Ensure handler.RegisterRoutes() is called after service initialization

**Issue**: CORS errors
**Solution**: Check corsMiddleware() and ensure proper headers are set

**Issue**: Port already in use
**Solution**: Change PORT environment variable or kill existing process

**Issue**: pgxpool configuration errors
**Solution**: Ensure MinConns <= MaxConns and use proper Config struct from common/database

**Issue**: Logging not appearing
**Solution**: Ensure logrus logger is passed to all constructors and set appropriate level

**Issue**: WithTx not available
**Solution**: Use database.WithTx() helper which wraps transaction logic with tracing

## Next Phase

Proceed to [Phase 7: Development Test Data](phase-07-test-data.md) once all tasks in this phase are complete.
