package prompts

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	mcpclient "github.com/v-egorov/service-boilerplate/services/mcp-server/internal/client"
)

// GetObjectInfoArgs defines arguments for the get_object_info prompt.
type GetObjectInfoArgs struct {
	ObjectTypeID int64 `json:"object_type_id"` // required: type to list objects from
	Page         *int  `json:"page,omitempty"` // optional, default 1
	PageSize     *int  `json:"page_size,omitempty"` // optional, default 10
}

// RegisterGetObjectInfoPrompt registers the get_object_info prompt template.
func RegisterGetObjectInfoPrompt(mcpServer *server.MCPServer, objClient *mcpclient.ObjectsClient) {
	prompt := mcp.NewPrompt(
		"get_object_info",
		mcp.WithPromptDescription("Summarize objects of a given type and suggest next tool calls for deeper exploration."),
		mcp.WithArgument("object_type_id", mcp.RequiredArgument(), mcp.ArgumentDescription("The object type ID to query instances for")),
	)

	mcpServer.AddPrompt(prompt, func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		var args GetObjectInfoArgs
		if req.Params.Arguments == nil {
			return mcp.NewGetPromptResult("", []mcp.PromptMessage{
				{Role: mcp.RoleAssistant, Content: mcp.NewTextContent("'object_type_id' is required")},
			}), nil
		}
		if idStr, ok := req.Params.Arguments["object_type_id"]; ok {
			args.ObjectTypeID, _ = strconv.ParseInt(idStr, 10, 64)
		}

		page := 1
		if args.Page != nil && *args.Page > 0 {
			page = *args.Page
		}
		pageSize := 10
		if args.PageSize != nil && *args.PageSize > 0 {
			pageSize = *args.PageSize
		}

		ctx = mcpclient.WithIdentity(ctx, req.Header)

		objects, err := objClient.ListObjects(ctx, args.ObjectTypeID, page, pageSize)
		if err != nil {
			return mcp.NewGetPromptResult("", []mcp.PromptMessage{
				{Role: mcp.RoleAssistant, Content: mcp.NewTextContent(fmt.Sprintf("Failed to list objects: %v", err))},
			}), nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# Objects of Type ID %d\n\n", args.ObjectTypeID))
		sb.WriteString(fmt.Sprintf("Found %d object(s):\n\n", len(objects)))

		for i, obj := range objects {
			if name, ok := obj["name"].(string); ok && name != "" {
				pubID := fmt.Sprintf("%v", obj["public_id"])
				sb.WriteString(fmt.Sprintf("%d. **%s** (ID: %s)\n", i+1, name, pubID))
			} else {
				sb.WriteString(fmt.Sprintf("%d. Object #%d\n", i+1, i+1))
			}
		}

		sb.WriteString("\n## Next Steps\n\n")
		sb.WriteString("- Use `get_object` with a `public_id` to get full details of a specific object\n")
		sb.WriteString(fmt.Sprintf("- Try page %d for more results", page+1))

		return mcp.NewGetPromptResult("", []mcp.PromptMessage{
			{Role: mcp.RoleAssistant, Content: mcp.NewTextContent(sb.String())},
		}), nil
	})
}
