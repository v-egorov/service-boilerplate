package tools

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	mcpclient "github.com/v-egorov/service-boilerplate/services/mcp-server/internal/client"
)

// PaginationMeta contains pagination metadata relayed from objects-service.
type PaginationMeta struct {
	Total  int64 `json:"total"`  // total number of matching results
	Limit  int   `json:"limit"`  // max results per page (actual limit used)
	Offset int   `json:"offset"` // number of results skipped
}

// ListObjectsResult is the output schema wrapper for list_objects.
// Wrapped in an object with domain-specific key "objects" so structuredContent
// is always a JSON object (not a bare array), satisfying MCP spec and Python SDK 1.26+.
type ListObjectsResult struct {
	Items      []map[string]interface{} `json:"objects"`
	Pagination *PaginationMeta           `json:"pagination,omitempty"`
}

// GetObjectResult is the output schema wrapper for get_object.
// Single-object tools return maps natively — this type exists to declare an
// outputSchema so clients can validate results against it. StructuredContent
// wraps the object in {"object": {...}} for consistency with the schema.
type GetObjectResult struct {
	Object map[string]interface{} `json:"object"`
}

// ListObjectsParams defines parameters for the list_objects tool.
type ListObjectsParams struct {
	ObjectTypeID  int64   `json:"object_type_id"` // required: filter by type ID
	Limit         *int    `json:"limit,omitempty"`        // optional, default 50 (objects-service max)
	Offset        *int    `json:"offset,omitempty"`       // optional, default 0
	TypeKeyPrefix *string `json:"type_key_prefix,omitempty"` // optional: filter objects whose type_key starts with this prefix (cross-type query)
}

// GetObjectParams defines parameters for the get_object tool.
type GetObjectParams struct {
	PublicID string `json:"public_id"` // required: UUID of the object
}

// RegisterObjectTools registers the object query tools on the MCP server.
func RegisterObjectTools(mcpServer *server.MCPServer, objClient *mcpclient.ObjectsClient) {
	// Tool: list_objects
	listTool := mcp.NewTool(
		"list_objects",
		mcp.WithDescription("List objects of a specific type. Returns paginated results with object details."),
		mcp.WithInputSchema[ListObjectsParams](),
		mcp.WithOutputSchema[ListObjectsResult](),
	)

	mcpServer.AddTool(listTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		tracer := otel.Tracer("mcp-server")

		var args ListObjectsParams
		if err := request.BindArguments(&args); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Invalid arguments: %v", err))},
				IsError: true,
			}, nil
		}

		if args.ObjectTypeID == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.NewTextContent("'object_type_id' is required")},
				IsError: true,
			}, nil
		}

		ctx, span := tracer.Start(ctx, "tool.list_objects")
		defer span.End()

		span.SetAttributes(
			attribute.String("tool.name", "list_objects"),
			attribute.Int64("object_type_id", args.ObjectTypeID),
		)

		limit := 50 // default matches objects-service default
		if args.Limit != nil && *args.Limit > 0 {
			limit = *args.Limit
		}
		offset := 0
		if args.Offset != nil && *args.Offset >= 0 {
			offset = *args.Offset
		}

		var typeKeyPrefix string
		if args.TypeKeyPrefix != nil {
			typeKeyPrefix = *args.TypeKeyPrefix
		}

		ctx = mcpclient.WithIdentity(ctx, request.Header)

		objects, paginationMeta, err := objClient.ListObjects(ctx, args.ObjectTypeID, limit, offset, typeKeyPrefix)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Failed to list objects: %v", err))},
				IsError: true,
			}, nil
		}

		typeName := fmt.Sprintf("type_%d", args.ObjectTypeID)
		if len(objects) > 0 {
			if n, ok := objects[0]["name"]; ok {
				if s, ok := n.(string); ok { typeName = s }
			}
		}

		// Build pagination metadata from objects-service response
		var pagination *PaginationMeta
		if paginationMeta != nil {
			pagination = &PaginationMeta{
				Total:  func() int64 { v, _ := paginationMeta["total"].(float64); return int64(v) }(),
				Limit:  func() int   { v, _ := paginationMeta["limit"].(float64); return int(v) }(),
				Offset: func() int   { v, _ := paginationMeta["offset"].(float64); return int(v) }(),
			}
		}

		return &mcp.CallToolResult{
			Content:         []mcp.Content{mcp.NewTextContent(fmt.Sprintf("list_objects returned %d objects for type %s", len(objects), typeName))},
			StructuredContent: ListObjectsResult{Items: objects, Pagination: pagination},
		}, nil
	})

	// Tool: get_object
	getTool := mcp.NewTool(
		"get_object",
		mcp.WithDescription("Get a single object by its public ID (UUID). Returns full object details."),
		mcp.WithInputSchema[GetObjectParams](),
		mcp.WithOutputSchema[GetObjectResult](),
	)

	mcpServer.AddTool(getTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		tracer := otel.Tracer("mcp-server")

		var args GetObjectParams
		if err := request.BindArguments(&args); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Invalid arguments: %v", err))},
				IsError: true,
			}, nil
		}

		if args.PublicID == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.NewTextContent("'public_id' is required")},
				IsError: true,
			}, nil
		}

		ctx, span := tracer.Start(ctx, "tool.get_object")
		defer span.End()

		span.SetAttributes(
			attribute.String("tool.name", "get_object"),
			attribute.String("public_id", args.PublicID),
		)

		ctx = mcpclient.WithIdentity(ctx, request.Header)

		obj, err := objClient.GetObjectByPublicID(ctx, args.PublicID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Failed to get object: %v", err))},
				IsError: true,
			}, nil
		}

		span.SetStatus(codes.Ok, "")

		// Extract name and public_id for summary (safe defaults if missing)
		name := "unknown"
		if n, ok := obj["name"]; ok {
			if s, ok := n.(string); ok && s != "" { name = s }
		}

		return &mcp.CallToolResult{
			Content:         []mcp.Content{mcp.NewTextContent(fmt.Sprintf("get_object: %s (id=%s)", name, args.PublicID))},
			StructuredContent: GetObjectResult{Object: obj},
		}, nil
	})
}
