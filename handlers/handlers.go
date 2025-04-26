package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s-mcp-server/pkg/k8s"

	"github.com/mark3labs/mcp-go/mcp"
)

// and is registered with the mcp server.
func GetAPIResourcesTool() mcp.Tool {
	return mcp.NewTool(
		"getAPIResources",
		mcp.WithDescription("Get all API resources in the Kubernetes cluster\n"+
			"CreateGetAPIResourcesTool creates a tool for getting API resources\n"+
			"GetAPIResourcesHandler handles the getAPIResources tool\n"+
			"It retrieves the API resources from the Kubernetes cluster\n"+
			"and returns them as a response.\n"+
			"The context is used to specify the cluster resource context\n"+
			"e.g. 'beta' or 'prod'.\n"+
			"The function returns a mcp.CallToolResult containing the API resources\n"+
			"or an error if the operation fails.\n"+
			"The function also handles the inclusion of namespace scoped\n"+
			"and cluster scoped resources based on the provided parameters.\n"+
			"The function is designed to be used as a handler for the mcp tool"),
		mcp.WithBoolean("includeNamespaceScoped", mcp.Description("Include namespace scoped resources")),
		mcp.WithBoolean("includeClusterScoped", mcp.Description("Include cluster scoped resources")),
	)

}

func GetAPIResources(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Helper function to extract string arguments with a default value
		getStringArg := func(key string, defaultValue string) string {
			if val, ok := request.Params.Arguments[key].(string); ok {
				return val
			}
			return defaultValue
		}

		// Helper function to extract boolean arguments with a default value
		getBoolArg := func(key string, defaultValue bool) bool {
			if val, ok := request.Params.Arguments[key].(bool); ok {
				return val
			}
			return defaultValue
		}

		// Extract arguments
		clusterContext := getStringArg("context", "default")
		includeNamespaceScoped := getBoolArg("includeNamespaceScoped", true)
		includeClusterScoped := getBoolArg("includeClusterScoped", true)

		// Fetch API resources
		resources, err := client.GetAPIResources(ctx, includeNamespaceScoped, includeClusterScoped)
		if err != nil {
			return nil, fmt.Errorf("failed to get API resources for context '%s': %w", clusterContext, err)
		}

		// Serialize response to JSON
		jsonResponse, err := json.Marshal(resources)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		// Return JSON response using NewToolResultJSON
		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// List all resources in the Kubernetes cluster of a specific type
func ListResourcesTool() mcp.Tool {
	return mcp.NewTool(
		"listResources",
		mcp.WithDescription("List all resources in the Kubernetes cluster of a specific type"),
		mcp.WithString("Kind", mcp.Required(), mcp.Description("The type of resource to list")),
		mcp.WithString("namespace", mcp.Description("The namespace to list resources in")),
		mcp.WithString("labelSelector", mcp.Description("A label selector to filter resources")),
	)
}

func ListResources(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Helper function to extract string arguments with a default value
		getStringArg := func(key string, defaultValue string) string {
			if val, ok := request.Params.Arguments[key].(string); ok {
				return val
			}
			return defaultValue
		}

		// Extract arguments
		kind := getStringArg("Kind", "")
		namespace := getStringArg("namespace", "")
		labelSelector := getStringArg("labelSelector", "")

		// Fetch resources
		resources, err := client.ListResources(ctx, kind, namespace, labelSelector, "")
		if err != nil {
			return nil, fmt.Errorf("failed to list resources for kind '%s': %w", kind, err)
		}

		// Serialize response to JSON
		jsonResponse, err := json.Marshal(resources)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		// Return JSON response using NewToolResultJSON
		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// GetResourceTool is a tool for getting a specific resource in the Kubernetes cluster
func GetResourcesTool() mcp.Tool {
	return mcp.NewTool(
		"getResource",
		mcp.WithDescription("Get a specific resource in the Kubernetes cluster"),
		mcp.WithString("kind", mcp.Required(), mcp.Description("The type of resource to get")),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the resource to get")),
		mcp.WithString("namespace", mcp.Description("The namespace of the resource")),
	)
}

// GetResource is a tool for listing resources in the Kubernetes cluster
func GetResources(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		kind, ok := request.Params.Arguments["kind"].(string)
		if !ok || kind == "" {
			return nil, fmt.Errorf("missing required parameter: kind")
		}

		name, ok := request.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("missing required parameter: name")
		}

		namespace, _ := request.Params.Arguments["namespace"].(string)

		resource, err := client.GetResource(ctx, kind, name, namespace)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(resource)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}



//describe resource
func DescribeResourcesTool() mcp.Tool {
	return mcp.NewTool(
		"describeResource",
		mcp.WithDescription("Describe a resource in the Kubernetes cluster based on given kind and name"),
		mcp.WithString("Kind", mcp.Required(), mcp.Description("The type of resource to describe")),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the resource to describe")),
		mcp.WithString("namespace", mcp.Description("The namespace of the resource")),
	)
}
func DescribeResources(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Helper function to extract string arguments with a default value
		getStringArg := func(key string, defaultValue string) string {
			if val, ok := request.Params.Arguments[key].(string); ok {
				return val
			}
			return defaultValue
		}

		// Extract arguments
		kind := getStringArg("Kind", "")
		name := getStringArg("name", "")
		namespace := getStringArg("namespace", "")

		// Fetch resource description
		resourceDescription, err := client.DescribeResource(ctx, kind, name, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to describe resource '%s' of kind '%s': %w", name, kind, err)
		}

		// Serialize response to JSON
		jsonResponse, err := json.Marshal(resourceDescription)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		// Return JSON response using NewToolResultJSON
		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}
