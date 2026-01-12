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
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
    
    "your-project/services/objects-service/internal/handlers"
    "your-project/services/objects-service/internal/repository"
    "your-project/services/objects-service/internal/services"
    "your-project/services/objects-service/pkg/config"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()
    
    db.SetMaxOpenConns(cfg.DB.MaxOpenConns)
    db.SetMaxIdleConns(cfg.DB.MaxIdleConns)
    db.SetConnMaxLifetime(time.Hour)
    
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    repositories := initRepositories(db)
    services := initServices(repositories)
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
        log.Printf("Starting server on port %s", cfg.Port)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Failed to start server: %v", err)
        }
    }()
    
    gracefulShutdown(server, ctx)
}

func initRepositories(db *sqlx.DB) *Repositories {
    return &Repositories{
        ObjectType: repository.NewObjectTypeRepository(db),
        Object:     repository.NewObjectRepository(db),
    }
}

func initServices(repos *Repositories) *Services {
    return &Services{
        ObjectType: services.NewObjectTypeService(repos.ObjectType),
        Object:     services.NewObjectService(repos.Object, repos.ObjectType),
    }
}

func initHandlers(services *Services) *Handlers {
    return &Handlers{
        ObjectType: handlers.NewObjectTypeHandler(services.ObjectType),
        Object:     handlers.NewObjectHandler(services.Object),
    }
}

func setupRouter(handlers *Handlers, cfg *config.Config) *gin.Engine {
    router := gin.New()
    
    if cfg.Environment != "production" {
        gin.SetMode(gin.DebugMode)
    } else {
        gin.SetMode(gin.ReleaseMode)
    }
    
    router.Use(gin.Recovery())
    router.Use(gin.Logger())
    router.Use(corsMiddleware())
    router.Use(authMiddleware())
    
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

func authMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.Set("user_id", "anonymous")
            c.Next()
            return
        }
        
        token := extractToken(authHeader)
        if token == "" {
            c.Set("user_id", "anonymous")
            c.Next()
            return
        }
        
        userID, err := validateToken(token)
        if err != nil {
            c.Set("user_id", "anonymous")
            c.Next()
            return
        }
        
        c.Set("user_id", userID)
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
    return "user123", nil
}

func gracefulShutdown(server *http.Server, ctx context.Context) {
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down server...")
    
    shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
    
    if err := server.Shutdown(shutdownCtx); err != nil {
        log.Printf("Server forced to shutdown: %v", err)
    }
    
    log.Println("Server exited")
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
- [ ] Add new imports for object types and objects
- [ ] Update repository initialization
- [ ] Update service initialization
- [ ] Update handler initialization
- [ ] Register new routes
- [ ] Verify application compiles: `go build ./cmd/...`
- [ ] Test application startup: `go run cmd/main.go`
- [ ] Verify health endpoint: `curl http://localhost:8085/health`
- [ ] Verify API endpoints: `curl http://localhost:8085/api/v1/object-types`
- [ ] Update progress.md

## Testing

```bash
# Build application
cd services/objects-service
go build -o objects-service cmd/main.go

# Run application
./objects-service

# Test health endpoint
curl http://localhost:8085/health

# Test root endpoint
curl http://localhost:8085/

# Test object types endpoint
curl http://localhost:8085/api/v1/object-types

# Test objects endpoint
curl http://localhost:8085/api/v1/objects
```

## Common Issues

**Issue**: Database connection fails
**Solution**: Verify DATABASE_URL environment variable is set correctly

**Issue**: Routes not registered
**Solution**: Ensure handler.RegisterRoutes() is called after router setup

**Issue**: CORS errors
**Solution**: Check corsMiddleware() and ensure proper headers are set

**Issue**: Port already in use
**Solution**: Change PORT environment variable or kill existing process

## Next Phase

Proceed to [Phase 7: Development Test Data](phase-07-test-data.md) once all tasks in this phase are complete.
