package tools

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
	mcpclient "github.com/v-egorov/service-boilerplate/services/mcp-server/internal/client"
)

// ListObjectTypesResult is the output schema wrapper for list_object_types.
// Wrapped in an object with domain-specific key "types" so structuredContent
// is always a JSON object (not a bare array), satisfying MCP spec and Python SDK 1.26+.
type ListObjectTypesResult struct {
	Items []map[string]interface{} `json:"types"`
}

// GetObjectTypeResult is the output schema wrapper for get_object_type.
// Single-object tools return maps natively — this type exists only to declare
// an outputSchema so clients can validate results against it.
type GetObjectTypeResult struct {
	Item map[string]interface{} `json:"item"`
}

// ListObjectTypesParams defines parameters for the list_object_types tool.
type ListObjectTypesParams struct {
	TypeKeyPrefix  string `json:"type_key_prefix,omitempty"` // optional prefix filter on type_key
	ParentTypeID   *int64 `json:"parent_type_id,omitempty"`  // optional: only return direct children of this type
}

// GetObjectTypeParams defines parameters for the get_object_type tool.
type GetObjectTypeParams struct {
	ID   *int64  `json:"id,omitempty"`    // one of id or name must be provided
	Name *string `json:"name,omitempty"`  // one of id or name must be provided
}

// RegisterTypeTools registers the object type tools on the MCP server.
func RegisterTypeTools(mcpServer *server.MCPServer, objClient *mcpclient.ObjectsClient, logger *logrus.Logger) {
	// Tool: list_object_types
	listTool := mcp.NewTool(
		"list_object_types",
		mcp.WithDescription("List all object types in the system (schema layer). Returns type names, keys, and descriptions."),
		mcp.WithInputSchema[ListObjectTypesParams](),
		mcp.WithOutputSchema[ListObjectTypesResult](),
	)

	mcpServer.AddTool(listTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		tracer := otel.Tracer("mcp-server")

		var args ListObjectTypesParams
		if err := request.BindArguments(&args); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Invalid arguments: %v", err))},
				IsError: true,
			}, nil
		}

		ctx, span := tracer.Start(ctx, "tool.list_object_types")
		defer span.End()

		span.SetAttributes(
			attribute.String("tool.name", "list_object_types"),
		)
		if args.TypeKeyPrefix != "" {
			span.SetAttributes(attribute.String("type_key_prefix", args.TypeKeyPrefix))
		}
		if args.ParentTypeID != nil {
			span.SetAttributes(attribute.Int64("parent_type_id", *args.ParentTypeID))
		}

		ctx = mcpclient.WithIdentity(ctx, request.Header)

		types, err := objClient.ListObjectTypes(ctx, args.TypeKeyPrefix, args.ParentTypeID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			logger.WithError(err).Error("Failed to list object types")
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Failed to list object types: %v", err))},
				IsError: true,
			}, nil
		}

		span.SetStatus(codes.Ok, "")

		return &mcp.CallToolResult{
			Content:         []mcp.Content{mcp.NewTextContent(fmt.Sprintf("list_object_types returned %d types", len(types)))},
			StructuredContent: ListObjectTypesResult{Items: types},
		}, nil
})

	// Tool: get_object_type
	getTool := mcp.NewTool(
		"get_object_type",
		mcp.WithDescription("Get a single object type by ID or name. Returns full type details including hierarchy info."),
		mcp.WithInputSchema[GetObjectTypeParams](),
		mcp.WithOutputSchema[GetObjectTypeResult](),
	)

	mcpServer.AddTool(getTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		tracer := otel.Tracer("mcp-server")

		var args GetObjectTypeParams
		if err := request.BindArguments(&args); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Invalid arguments: %v", err))},
				IsError: true,
			}, nil
		}

		if args.ID == nil && args.Name == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.NewTextContent("Either 'id' or 'name' parameter is required")},
				IsError: true,
			}, nil
		}

		ctx, span := tracer.Start(ctx, "tool.get_object_type")
		defer span.End()

		span.SetAttributes(
			attribute.String("tool.name", "get_object_type"),
		)
		if args.ID != nil {
			span.SetAttributes(attribute.Int64("id", *args.ID))
		}
		if args.Name != nil {
			span.SetAttributes(attribute.String("name", *args.Name))
		}

		ctx = mcpclient.WithIdentity(ctx, request.Header)

		var typ map[string]interface{}
		var err error
		if args.ID != nil {
			typ, err = objClient.GetObjectTypeByID(ctx, *args.ID)
		} else {
			typ, err = objClient.GetObjectTypeByName(ctx, *args.Name)
		}

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Failed to get object type: %v", err))},
				IsError: true,
			}, nil
		}

		span.SetStatus(codes.Ok, "")

		// Extract name and id for summary (safe defaults if missing)
		name := "unknown"
		if n, ok := typ["name"]; ok {
			if s, ok := n.(string); ok && s != "" { name = s }
		}
		idVal := "0"
		if i, ok := typ["id"]; ok {
			switch v := i.(type) {
			case float64:
				idVal = fmt.Sprintf("%d", int(v))
			case string:
				idVal = v
			}
		}

		return &mcp.CallToolResult{
			Content:         []mcp.Content{mcp.NewTextContent(fmt.Sprintf("get_object_type: %s (id=%s)", name, idVal))},
			StructuredContent: GetObjectTypeResult{Item: typ},
		}, nil
	})
}
