package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/mark3labs/mcp-go/mcp"
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

	// Build multi-route mux: health endpoints + MCP StreamableHTTP endpoint
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler.LivenessHandler)
	mux.HandleFunc("/live", healthHandler.LivenessHandler)
	mux.HandleFunc("/ready", healthHandler.ReadinessHandler)
	mux.HandleFunc("/ping", healthHandler.PingHandler)
	mux.HandleFunc("/status", healthHandler.StatusHandler)

	// MCP StreamableHTTP endpoint — single /mcp path handles all JSON-RPC communication.
	// Session management is stateful: first POST establishes session, subsequent calls echo back MCP-Session-ID header.
	// Register slog bridge for mcp-go transport logging → logrus.
	transportLogger := logger.Logger.WithField("component", "mcp-transport")
	streamableHTTPServer := server.NewStreamableHTTPServer(mcpServer,
		server.WithEndpointPath("/mcp"),
		server.WithStreamableHTTPLogger(slog.New(&logrusSlogHandler{logger: transportLogger})),
	)

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

// logrusSlogHandler bridges mcp-go's internal slog transport logging to logrus.
type logrusSlogHandler struct {
	logger *logrus.Entry
	group  string
	attrs  []slog.Attr
}

func (h *logrusSlogHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *logrusSlogHandler) Handle(_ context.Context, r slog.Record) error {
	fields := logrus.Fields{}
	for _, a := range h.attrs {
		fields[a.Key] = a.Value.Any()
	}
	if h.group != "" {
		fields["component"] = h.group
	}
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()
		return true
	})
	if r.Message != "" {
		fields["msg"] = r.Message
	}

	switch r.Level {
	case slog.LevelDebug:
		h.logger.WithFields(fields).Debug(r.Message)
	case slog.LevelInfo:
		h.logger.WithFields(fields).Info(r.Message)
	case slog.LevelWarn:
		h.logger.WithFields(fields).Warn(r.Message)
	case slog.LevelError:
		h.logger.WithFields(fields).Error(r.Message)
	default:
		h.logger.WithFields(fields).Info(r.Message)
	}
	return nil
}

func (h *logrusSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	merged := make([]slog.Attr, len(h.attrs), len(h.attrs)+len(attrs))
	copy(merged, h.attrs)
	merged = append(merged, attrs...)
	return &logrusSlogHandler{logger: h.logger, group: h.group, attrs: merged}
}

func (h *logrusSlogHandler) WithGroup(name string) slog.Handler {
	group := name
	if h.group != "" {
		group = h.group + "." + name
	}
	return &logrusSlogHandler{logger: h.logger, group: group, attrs: h.attrs}
}

// opTimeStore stores per-request start times keyed by request ID string.
type opTimeStore struct {
	mu  sync.Mutex
	tms map[string]time.Time
}

var (
	opTimes     = &opTimeStore{tms: make(map[string]time.Time)}
	hooksLogger *logrus.Entry
)

func onBeforeAny(ctx context.Context, id any, method mcp.MCPMethod, message any) {
	rid := fmt.Sprintf("%v", id)
	opTimes.mu.Lock()
	opTimes.tms[rid] = time.Now()
	opTimes.mu.Unlock()

	fields := logrus.Fields{
		"op":         string(method),
		"request_id": rid,
		"service":    "mcp-server",
	}

	switch method {
	case mcp.MethodToolsCall:
		if p, ok := message.(*mcp.CallToolParams); ok && p != nil {
			fields["tool"] = p.Name
			if args, ok := p.Arguments.(map[string]any); ok {
				fields["params"] = args
			}
		}
	case mcp.MethodResourcesRead:
		if p, ok := message.(*mcp.ReadResourceParams); ok && p != nil {
			fields["resource_uri"] = p.URI
			if len(p.Arguments) > 0 {
				fields["arguments"] = p.Arguments
			}
		}
	case mcp.MethodPromptsGet:
		if p, ok := message.(*mcp.GetPromptParams); ok && p != nil {
			fields["prompt_name"] = p.Name
			if len(p.Arguments) > 0 {
				fields["arguments"] = p.Arguments
			}
		}
	case mcp.MethodPing:
		hooksLogger.WithFields(fields).Debug("mcp_operation_start")
		return
	}

	hooksLogger.WithFields(fields).Info("mcp_operation_start")
}

func onSuccess(ctx context.Context, id any, method mcp.MCPMethod, message any, result any) {
	rid := fmt.Sprintf("%v", id)

	opTimes.mu.Lock()
	start := opTimes.tms[rid]
	delete(opTimes.tms, rid)
	opTimes.mu.Unlock()

	duration := time.Since(start).Milliseconds()

	fields := logrus.Fields{
		"op":          string(method),
		"request_id":  rid,
		"service":     "mcp-server",
		"status":      "success",
		"duration_ms": duration,
	}

	switch method {
	case mcp.MethodToolsCall:
		if p, ok := message.(*mcp.CallToolParams); ok && p != nil {
			fields["tool"] = p.Name
		}
	case mcp.MethodResourcesRead:
		if p, ok := message.(*mcp.ReadResourceParams); ok && p != nil {
			fields["resource_uri"] = p.URI
		}
	case mcp.MethodPromptsGet:
		if p, ok := message.(*mcp.GetPromptParams); ok && p != nil {
			fields["prompt_name"] = p.Name
		}
	}

	hooksLogger.WithFields(fields).Info("mcp_operation_end")
}

func onError(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
	rid := fmt.Sprintf("%v", id)

	opTimes.mu.Lock()
	start := opTimes.tms[rid]
	delete(opTimes.tms, rid)
	opTimes.mu.Unlock()

	duration := time.Since(start).Milliseconds()

	fields := logrus.Fields{
		"op":          string(method),
		"request_id":  rid,
		"service":     "mcp-server",
		"status":      "error",
		"duration_ms": duration,
		"error":       err.Error(),
	}

	switch method {
	case mcp.MethodToolsCall:
		if p, ok := message.(*mcp.CallToolParams); ok && p != nil {
			fields["tool"] = p.Name
		}
	case mcp.MethodResourcesRead:
		if p, ok := message.(*mcp.ReadResourceParams); ok && p != nil {
			fields["resource_uri"] = p.URI
		}
	case mcp.MethodPromptsGet:
		if p, ok := message.(*mcp.GetPromptParams); ok && p != nil {
			fields["prompt_name"] = p.Name
		}
	}

	hooksLogger.WithFields(fields).Error("mcp_operation_error")
}

func initMCPServer(objClient *mcpclient.ObjectsClient, logger *logrus.Logger, name, version string) *server.MCPServer {
	hooksLogger = logger.WithField("component", "mcp-hooks")

	// Initialize the MCPServer with a Hooks container so GetHooks() returns non-nil.
	mcpServer := server.NewMCPServer(name, version,
		server.WithHooks(&server.Hooks{}),
	)

	// Register operation-level hooks for per-request visibility.
	hooks := mcpServer.GetHooks()
	hooks.AddBeforeAny(onBeforeAny)
	hooks.AddOnSuccess(onSuccess)
	hooks.AddOnError(onError)

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

