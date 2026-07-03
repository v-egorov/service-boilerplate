package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
	"github.com/v-egorov/service-boilerplate/common/config"
	"github.com/v-egorov/service-boilerplate/common/logging"
	mcpclient "github.com/v-egorov/service-boilerplate/services/mcp-server/internal/client"
	"github.com/v-egorov/service-boilerplate/services/mcp-server/internal/prompts"
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

	// Create HTTP client for objects-service
	objClient := mcpclient.NewObjectsClient(cfg.ObjectsService.URL, time.Duration(cfg.ObjectsService.Timeout)*time.Second)

	// Initialize MCP server with SSE transport
	mcpServer := initMCPServer(objClient, logger.Logger, cfg.App.Name, cfg.App.Version)

	// Create SSE server for HTTP handler
	sseServer := server.NewSSEServer(mcpServer)

	// Start HTTP server serving the SSE endpoint
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: sseServer.SSEHandler(),
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

	if err := sseServer.Shutdown(ctx); err != nil {
		logger.Error("MCP server forced to shutdown", err)
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
	prompts.RegisterBrowseSchemaPrompt(mcpServer, objClient, logger)
	prompts.RegisterGetObjectInfoPrompt(mcpServer, objClient)

	return mcpServer
}
