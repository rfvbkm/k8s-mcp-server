// Package handlers provides MCP tool handlers for interacting with Kubernetes.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/reza-gholizade/k8s-mcp-server/pkg/k8s"

	"github.com/mark3labs/mcp-go/mcp"
)

// GetAPIResourcesTool creates a tool for getting API resources.
// It defines the tool's name, description, and parameters for including
// namespace-scoped and cluster-scoped resources.
func GetAPIResourcesTool() mcp.Tool {
	return mcp.NewTool(
		"getAPIResources",
		mcp.WithDescription("Get all API resources in the Kubernetes cluster\n"+
			"CreateGetAPIResourcesTool creates a tool for getting API resources\n"+
			"GetAPIResourcesHandler handles the getAPIResources tool\n"+
			"It retrieves the API resources from the Kubernetes cluster\n"+
			"and returns them as a response.\n"+
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

// GetAPIResources returns a handler function for the getAPIResources tool.
// It retrieves API resources from the Kubernetes cluster based on the provided
// context and parameters (includeNamespaceScoped, includeClusterScoped).
// The result is serialized to JSON and returned.
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

// ListResourcesTool creates a tool for listing resources of a specific type.
// It defines the tool's name, description, and parameters for kind, namespace,
// and labelSelector.
func ListResourcesTool() mcp.Tool {
	return mcp.NewTool(
		"listResources",
		mcp.WithDescription("List all resources in the Kubernetes cluster of a specific type"),
		mcp.WithString("Kind", mcp.Required(), mcp.Description("The type of resource to list")),
		mcp.WithString("namespace", mcp.Description("The namespace to list resources in")),
		mcp.WithString("labelSelector", mcp.Description("A label selector to filter resources")),
	)
}

// ListResources returns a handler function for the listResources tool.
// It lists resources in the Kubernetes cluster based on the provided kind,
// namespace, and labelSelector. The result is serialized to JSON and returned.
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

// GetResourcesTool creates a tool for getting a specific resource.
// It defines the tool's name, description, and parameters for kind, name,
// and namespace.
func GetResourcesTool() mcp.Tool {
	return mcp.NewTool(
		"getResource",
		mcp.WithDescription("Get a specific resource in the Kubernetes cluster"),
		mcp.WithString("kind", mcp.Required(), mcp.Description("The type of resource to get")),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the resource to get")),
		mcp.WithString("namespace", mcp.Description("The namespace of the resource")),
	)
}

// GetResources returns a handler function for the getResource tool.
// It retrieves a specific resource from the Kubernetes cluster based on the
// provided kind, name, and namespace. The result is serialized to JSON and returned.
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

// DescribeResourcesTool creates a tool for describing a resource.
// It defines the tool's name, description, and parameters for kind, name,
// and namespace.
func DescribeResourcesTool() mcp.Tool {
	return mcp.NewTool(
		"describeResource",
		mcp.WithDescription("Describe a resource in the Kubernetes cluster based on given kind and name"),
		mcp.WithString("Kind", mcp.Required(), mcp.Description("The type of resource to describe")),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the resource to describe")),
		mcp.WithString("namespace", mcp.Description("The namespace of the resource")),
	)
}

// DescribeResources returns a handler function for the describeResource tool.
// It fetches the description (manifest) of a specific resource from the
// Kubernetes cluster based on the provided kind, name, and namespace.
// The result is serialized to JSON and returned.
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

// GetPodsLogsTools creates a tool for getting pod logs.
// It defines the tool's name, description, and parameters for the pod name
// and namespace.
func GetPodsLogsTools() mcp.Tool {
	return mcp.NewTool(
		"getPodsLogs",
		mcp.WithDescription("Get logs of a specific pod in the Kubernetes cluster"),
		mcp.WithString("Name", mcp.Required(), mcp.Description("The name of the pod to get logs from")),
		mcp.WithString("namespace", mcp.Description("The namespace of the pod")),
	)
}

// GetPodsLogs returns a handler function for the getPodsLogs tool.
// It retrieves logs for a specific pod from the Kubernetes cluster based on the
// provided name and namespace. The result is serialized to JSON and returned.
func GetPodsLogs(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, ok := request.Params.Arguments["Name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("missing required parameter: Name")
		}

		namespace, _ := request.Params.Arguments["namespace"].(string)

		logs, err := client.GetPodsLogs(ctx, namespace, name)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(logs)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// GetNodeMetricsTools creates a tool for getting node metrics.
// It defines the tool's name, description, and parameters for the node name.
func GetNodeMetricsTools() mcp.Tool {
	return mcp.NewTool(
		"getNodeMetrics",
		mcp.WithDescription("Get resource usage of a specific node in the Kubernetes cluster"),
		mcp.WithString("Name", mcp.Required(), mcp.Description("The name of the node to get resource usage from")),
	)
}

// GetNodeMetrics returns a handler function for the getNodeMetrics tool.
// It retrieves resource usage metrics for a specific node from the Kubernetes
// cluster based on the provided node name. The result is serialized to JSON
// and returned.
func GetNodeMetrics(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, ok := request.Params.Arguments["Name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("missing required parameter: Name")
		}

		resourceUsage, err := client.GetNodeMetrics(ctx, name)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(resourceUsage)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// GetPodMetricsTool creates a tool for getting pod metrics.
// It defines the tool's name, description, and parameters for the pod namespace
// and name.
func GetPodMetricsTool() mcp.Tool {
	return mcp.NewTool(
		"getPodMetrics",
		mcp.WithDescription("Get CPU and Memory metrics for a specific pod"),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace of the pod")),
		mcp.WithString("podName", mcp.Required(), mcp.Description("The name of the pod")),
	)
}

// GetPodMetrics returns a handler function for the getPodMetrics tool.
// It retrieves CPU and Memory metrics for a specific pod from the Kubernetes
// cluster based on the provided namespace and pod name. The result is
// serialized to JSON and returned.
func GetPodMetrics(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace, ok := request.Params.Arguments["namespace"].(string)
		if !ok || namespace == "" {
			return nil, fmt.Errorf("missing required parameter: namespace")
		}

		podName, ok := request.Params.Arguments["podName"].(string)
		if !ok || podName == "" {
			return nil, fmt.Errorf("missing required parameter: podName")
		}

		metrics, err := client.GetPodMetrics(ctx, namespace, podName)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(metrics)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize metrics response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// GetEventsTool creates a tool for getting events in the Kubernetes cluster.
// It defines the tool's name, description, and parameters for the namespace
// and labelSelector.
func GetEventsTool() mcp.Tool {
	return mcp.NewTool(
		"getEvents",
		mcp.WithDescription("Get events in the Kubernetes cluster"),
		mcp.WithString("namespace", mcp.Description("The namespace to get events from")),
		mcp.WithString("labelSelector", mcp.Description("A label selector to filter events")),
	)
}
// GetEvents returns a handler function for the getEvents tool.
// It retrieves events from the Kubernetes cluster based on the provided
// namespace and labelSelector. The result is serialized to JSON and returned.
func GetEvents(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace, ok := request.Params.Arguments["namespace"].(string)
		if !ok || namespace == "" {
			return nil, fmt.Errorf("missing required parameter: namespace")
		}

		events, err := client.GetEvents(ctx, namespace)
		if err != nil {
			return nil, err
		}

		jsonResponse, err := json.Marshal(events)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize events response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}


// create resource of any type or kind
func CreateResourceTool() mcp.Tool {
	return mcp.NewTool(
		"createResource",
		mcp.WithDescription("Create a resource in the Kubernetes cluster"),
		mcp.WithString("kind", mcp.Required(), mcp.Description("The type of resource to create")),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the resource to create")),
		mcp.WithString("namespace", mcp.Description("The namespace of the resource")),
		mcp.WithString("manifest", mcp.Required(), mcp.Description("The manifest of the resource to create")),
	)
}

// Create or Update Resource returns a handler function for the createResource tool.
// It creates or Updates  a resource in the Kubernetes cluster based on the provided kind,
// name, namespace, and manifest. The result is serialized to JSON and returned.
func CreateResource(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

		manifest, ok := request.Params.Arguments["manifest"].(string)
		if !ok || manifest == "" {
			return nil, fmt.Errorf("missing required parameter: manifest")
		}

		resource, err := client.CreateOrUpdateResource(ctx, kind, name, namespace, manifest)
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