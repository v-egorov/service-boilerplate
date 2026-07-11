package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/common/config"
	"github.com/v-egorov/service-boilerplate/common/logging"
	"github.com/v-egorov/service-boilerplate/common/tracing"
	mcpclient "github.com/v-egorov/service-boilerplate/services/mcp-server/internal/client"
	"github.com/v-egorov/service-boilerplate/services/mcp-server/internal/handlers"
	promptsPkg "github.com/v-egorov/service-boilerplate/services/mcp-server/internal/prompts"
	mcpresources "github.com/v-egorov/service-boilerplate/services/mcp-server/internal/resources"
	mcptools "github.com/v-egorov/service-boilerplate/services/mcp-server/internal/tools"
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

	// Create HTTP client for objects-service
	objClient := mcpclient.NewObjectsClient(cfg.ObjectsService.URL, time.Duration(cfg.ObjectsService.Timeout)*time.Second)

	// Initialize health handler
	healthHandler := handlers.NewHealthHandler(objClient, logger.Logger, cfg.App.Name, cfg.App.Version)

	// Initialize MCP server with StreamableHTTP transport + stateful sessions
	mcpServer := initMCPServer(objClient, logger.Logger, cfg.App.Name, cfg.App.Version)

	// Create StreamableHTTP server for HTTP handler (single endpoint at /mcp)
	streamableHTTPServer := server.NewStreamableHTTPServer(mcpServer,
		server.WithEndpointPath("/mcp"),
	)

	// Build multi-route mux: health endpoints + MCP StreamableHTTP endpoint
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler.LivenessHandler)
	mux.HandleFunc("/live", healthHandler.LivenessHandler)
	mux.HandleFunc("/ready", healthHandler.ReadinessHandler)
	mux.HandleFunc("/ping", healthHandler.PingHandler)
	mux.HandleFunc("/status", healthHandler.StatusHandler)

	// MCP StreamableHTTP endpoint — single /mcp path handles all JSON-RPC communication.
	// Session management is stateful: first POST establishes session, subsequent calls echo back MCP-Session-ID header.
	mux.Handle("/mcp", otelhttp.NewHandler(streamableHTTPServer, "mcp-server"))

	// Start HTTP server serving the mux
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: mux,
	}

	go func() {
		logger.Info(fmt.Sprintf("Starting MCP Server on %s", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start MCP server", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down MCP Server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if tracerProvider != nil {
		if err := tracing.ShutdownTracer(tracerProvider); err != nil {
			logger.Error("Failed to shutdown tracer during graceful shutdown", err)
		}
	}
	if err := streamableHTTPServer.Shutdown(ctx); err != nil {
		logger.Error("StreamableHTTP server forced to shutdown", err)
	}
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("HTTP server forced to shutdown", err)
	}

	logger.Info("MCP Server exited")
}

func initMCPServer(objClient *mcpclient.ObjectsClient, logger *logrus.Logger, name, version string) *server.MCPServer {
	mcpServer := server.NewMCPServer(name, version)

	// Register tools for object types (schema layer)
	mcptools.RegisterTypeTools(mcpServer, objClient, logger)

	// Register tools for objects (instance layer)
	mcptools.RegisterObjectTools(mcpServer, objClient)

	// Register resources (browseable data at URI)
	mcpresources.RegisterTypeHierarchyResource(mcpServer, objClient, logger)

	// Register prompt templates
	promptsPkg.RegisterBrowseSchemaPrompt(mcpServer, objClient, logger)
	promptsPkg.RegisterGetObjectInfoPrompt(mcpServer, objClient)

	return mcpServer
}

