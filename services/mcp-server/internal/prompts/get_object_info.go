package prompts

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	mcpclient "github.com/v-egorov/service-boilerplate/services/mcp-server/internal/client"
)

// GetObjectInfoArgs defines arguments for the get_object_info prompt.
type GetObjectInfoArgs struct {
	ObjectTypeID int64 `json:"object_type_id"` // required: type to list objects from
	Limit        *int  `json:"limit,omitempty"`       // optional, default 50 (max results)
	Offset       *int  `json:"offset,omitempty"`      // optional, default 0
}

// RegisterGetObjectInfoPrompt registers the get_object_info prompt template.
func RegisterGetObjectInfoPrompt(mcpServer *server.MCPServer, objClient *mcpclient.ObjectsClient) {
	prompt := mcp.NewPrompt(
		"get_object_info",
		mcp.WithPromptDescription("Summarize objects of a given type and suggest next tool calls for deeper exploration."),
		mcp.WithArgument("object_type_id", mcp.RequiredArgument(), mcp.ArgumentDescription("The object type ID to query instances for")),
	)

	mcpServer.AddPrompt(prompt, func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		tracer := otel.Tracer("mcp-server")

		var args GetObjectInfoArgs
		if req.Params.Arguments == nil {
			return mcp.NewGetPromptResult("", []mcp.PromptMessage{
				{Role: mcp.RoleAssistant, Content: mcp.NewTextContent("'object_type_id' is required")},
			}), nil
		}
		if idStr, ok := req.Params.Arguments["object_type_id"]; ok {
			args.ObjectTypeID, _ = strconv.ParseInt(idStr, 10, 64)
		}

		limit := 50 // default matches objects-service default
		if args.Limit != nil && *args.Limit > 0 {
			limit = *args.Limit
		}
		offset := 0
		if args.Offset != nil && *args.Offset >= 0 {
			offset = *args.Offset
		}

		_, span := tracer.Start(ctx, "prompt.get_object_info")
		defer span.End()

		ctx = mcpclient.WithIdentity(ctx, req.Header)

		objects, _, err := objClient.ListObjects(ctx, args.ObjectTypeID, limit, offset, "")
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
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
		sb.WriteString(fmt.Sprintf("- Use offset=%d to see the next batch of objects", offset+limit))

		return mcp.NewGetPromptResult("", []mcp.PromptMessage{
			{Role: mcp.RoleAssistant, Content: mcp.NewTextContent(sb.String())},
		}), nil
	})
}
