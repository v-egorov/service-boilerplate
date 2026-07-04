package prompts

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
	mcpclient "github.com/v-egorov/service-boilerplate/services/mcp-server/internal/client"
)

// BrowseSchemaArgs defines arguments for the browse_schema prompt.
type BrowseSchemaArgs struct {
	TypeKeyPrefix *string `json:"type_key_prefix,omitempty"` // optional: filter types by prefix
}

// RegisterBrowseSchemaPrompt registers the browse_schema prompt template.
func RegisterBrowseSchemaPrompt(mcpServer *server.MCPServer, objClient *mcpclient.ObjectsClient, logger *logrus.Logger) {
	prompt := mcp.NewPrompt(
		"browse_schema",
		mcp.WithPromptDescription("Explore the object type schema hierarchy. Lists all types and provides guidance on which tools to use for deeper exploration."),
		mcp.WithArgument("type_key_prefix", mcp.ArgumentDescription("Optional prefix to filter type keys")),
	)

	mcpServer.AddPrompt(prompt, func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		var args BrowseSchemaArgs
		if req.Params.Arguments != nil {
			argsBytes, _ := json.Marshal(req.Params.Arguments)
			if err := json.Unmarshal(argsBytes, &args); err != nil {
				return mcp.NewGetPromptResult("", []mcp.PromptMessage{
					{Role: mcp.RoleAssistant, Content: mcp.NewTextContent(fmt.Sprintf("Invalid arguments: %v", err))},
				}), nil
			}
		}

		ctx = mcpclient.WithIdentity(ctx, req.Header)

		types, err := objClient.ListObjectTypes(ctx)
		if err != nil {
			logger.WithError(err).Error("Failed to list types for prompt")
			return mcp.NewGetPromptResult("", []mcp.PromptMessage{
				{Role: mcp.RoleAssistant, Content: mcp.NewTextContent(fmt.Sprintf("Failed to list object types: %v", err))},
			}), nil
		}

		var sb strings.Builder
		sb.WriteString("# Object Type Schema\n\n")
		sb.WriteString(fmt.Sprintf("Found %d object type(s):\n\n", len(types)))

		for _, typ := range types {
			if name, ok := typ["name"].(string); ok {
				key := ""
				if k, exists := typ["type_key"]; exists {
					key = fmt.Sprintf("%v", k)
				}
				if args.TypeKeyPrefix != nil && !strings.HasPrefix(key, *args.TypeKeyPrefix) {
					continue
				}
				sb.WriteString(fmt.Sprintf("- **%s** (`%s`): %s\n", name, key, descOrEmpty(typ)))
			}
		}

		sb.WriteString("\n## Next Steps\n\n")
		sb.WriteString("Use `get_object_type` to get details about a specific type.\n")
		sb.WriteString("Use `list_objects` to query instances of a type.\n")
		sb.WriteString("Use `browse_schema` again with `type_key_prefix` to filter results.")

		return mcp.NewGetPromptResult("", []mcp.PromptMessage{
			{Role: mcp.RoleAssistant, Content: mcp.NewTextContent(sb.String())},
		}), nil
	})
}

func descOrEmpty(typ map[string]interface{}) string {
	if d, ok := typ["description"].(string); ok && d != "" {
		return d
	}
	return "(no description)"
}
