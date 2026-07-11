package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
	mcpclient "github.com/v-egorov/service-boilerplate/services/mcp-server/internal/client"
)

// RegisterTypeHierarchyResource registers the objects-types resource that provides browseable access to the type hierarchy.
func RegisterTypeHierarchyResource(mcpServer *server.MCPServer, objClient *mcpclient.ObjectsClient, logger *logrus.Logger) {
	resource := mcp.Resource{
		URI:        "objects-types://hierarchy",
		Name:       "Object Type Hierarchy",
		Title:      "Full object type hierarchy tree",
		Description: "Browseable resource containing the complete object type taxonomy tree. Use list_object_types tool for flat listing, or get_object_type for individual details.",
		MIMEType:   "application/json",
	}

	mcpServer.AddResource(resource, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		tracer := otel.Tracer("mcp-server")

		ctx, span := tracer.Start(ctx, "resource.hierarchy")
		defer span.End()

		ctx = mcpclient.WithIdentity(ctx, req.Header)

		// Fetch the root type tree from objects-service
		tree, err := objClient.GetRootTree(ctx)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			logger.WithError(err).Error("Failed to fetch type hierarchy")
			return nil, fmt.Errorf("failed to fetch type hierarchy: %w", err)
		}

		data, _ := json.Marshal(tree)
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      resource.URI,
				MIMEType: resource.MIMEType,
				Text:     string(data),
			},
		}, nil
	})
}
