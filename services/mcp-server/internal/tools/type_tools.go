package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
	mcpclient "github.com/v-egorov/service-boilerplate/services/mcp-server/internal/client"
)

// ListObjectTypesParams defines parameters for the list_object_types tool.
type ListObjectTypesParams struct {
	Limit int `json:"limit,omitempty"` // optional, default 100
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
	)

	mcpServer.AddTool(listTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args ListObjectTypesParams
		if err := request.BindArguments(&args); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Invalid arguments: %v", err))},
				IsError: true,
			}, nil
		}

		types, err := objClient.ListObjectTypes()
		if err != nil {
			logger.WithError(err).Error("Failed to list object types")
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Failed to list object types: %v", err))},
				IsError: true,
			}, nil
		}

		data, _ := json.Marshal(types)
		return &mcp.CallToolResult{
			Content:         []mcp.Content{mcp.NewTextContent(string(data))},
			StructuredContent: types,
		}, nil
	})

	// Tool: get_object_type
	getTool := mcp.NewTool(
		"get_object_type",
		mcp.WithDescription("Get a single object type by ID or name. Returns full type details including hierarchy info."),
		mcp.WithInputSchema[GetObjectTypeParams](),
	)

	mcpServer.AddTool(getTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

		var typ map[string]interface{}
		var err error
		if args.ID != nil {
			typ, err = objClient.GetObjectTypeByID(*args.ID)
		} else {
			typ, err = objClient.GetObjectTypeByName(*args.Name)
		}

		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("Failed to get object type: %v", err))},
				IsError: true,
			}, nil
		}

		data, _ := json.Marshal(typ)
		return &mcp.CallToolResult{
			Content:         []mcp.Content{mcp.NewTextContent(string(data))},
			StructuredContent: typ,
		}, nil
	})
}
