package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	mcpclient "github.com/v-egorov/service-boilerplate/services/mcp-server/internal/client"
)

// ListObjectsResult is the output schema wrapper for list_objects.
// Wrapped in an object with domain-specific key "objects" so structuredContent
// is always a JSON object (not a bare array), satisfying MCP spec and Python SDK 1.26+.
type ListObjectsResult struct {
	Items []map[string]interface{} `json:"objects"`
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
	Page          *int    `json:"page,omitempty"`       // optional, default 1
	PageSize      *int    `json:"page_size,omitempty"`  // optional, default 20
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

		page := 1
		if args.Page != nil && *args.Page > 0 {
			page = *args.Page
		}
		pageSize := 20
		if args.PageSize != nil && *args.PageSize > 0 {
			pageSize = *args.PageSize
		}

		var typeKeyPrefix string
		if args.TypeKeyPrefix != nil {
			typeKeyPrefix = *args.TypeKeyPrefix
		}

		ctx = mcpclient.WithIdentity(ctx, request.Header)

		objects, err := objClient.ListObjects(ctx, args.ObjectTypeID, page, pageSize, typeKeyPrefix)
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

		return &mcp.CallToolResult{
			Content:         []mcp.Content{mcp.NewTextContent(fmt.Sprintf("list_objects returned %d objects for type %s", len(objects), typeName))},
			StructuredContent: ListObjectsResult{Items: objects},
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

		ctx = mcpclient.WithIdentity(ctx, request.Header)

		obj, err := objClient.GetObjectByPublicID(ctx, args.PublicID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Failed to get object: %v", err))},
				IsError: true,
			}, nil
		}

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
