package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	mcpclient "github.com/v-egorov/service-boilerplate/services/mcp-server/internal/client"
)

// ListObjectsParams defines parameters for the list_objects tool.
type ListObjectsParams struct {
	ObjectTypeID int64 `json:"object_type_id"` // required: filter by type ID
	Page         *int  `json:"page,omitempty"` // optional, default 1
	PageSize     *int  `json:"page_size,omitempty"` // optional, default 20
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

		ctx = mcpclient.WithIdentity(ctx, request.Header)

		objects, err := objClient.ListObjects(ctx, args.ObjectTypeID, page, pageSize)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Failed to list objects: %v", err))},
				IsError: true,
			}, nil
		}

		data, _ := json.Marshal(objects)
		return &mcp.CallToolResult{
			Content:         []mcp.Content{mcp.NewTextContent(string(data))},
			StructuredContent: objects,
		}, nil
	})

	// Tool: get_object
	getTool := mcp.NewTool(
		"get_object",
		mcp.WithDescription("Get a single object by its public ID (UUID). Returns full object details."),
		mcp.WithInputSchema[GetObjectParams](),
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

		data, _ := json.Marshal(obj)
		return &mcp.CallToolResult{
			Content:         []mcp.Content{mcp.NewTextContent(string(data))},
			StructuredContent: obj,
		}, nil
	})
}
