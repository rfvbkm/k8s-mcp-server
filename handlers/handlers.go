// Package handlers provides MCP tool handlers for interacting with Kubernetes.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/reza-gholizade/k8s-mcp-server/pkg/k8s"

	"github.com/mark3labs/mcp-go/mcp"
)

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
		containerName, _ := request.Params.Arguments["containerName"].(string)


		logs, err := client.GetPodsLogs(ctx, namespace, containerName, name)
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

// Create or Update Resource returns a handler function for the createResource tool.
// It creates or Updates  a resource in the Kubernetes cluster based on the provided kind,
// name, namespace, and manifest. The result is serialized to JSON and returned.
func CreateOrUpdateResource(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		kind, ok := request.Params.Arguments["kind"].(string)
		if !ok || kind == "" {
			return nil, fmt.Errorf("missing required parameter: kind")
		}

		namespace, _ := request.Params.Arguments["namespace"].(string)

		manifest, ok := request.Params.Arguments["manifest"].(string)
		if !ok || manifest == "" {
			return nil, fmt.Errorf("missing required parameter: manifest")
		}

		resource, err := client.CreateOrUpdateResource(ctx, kind, namespace, manifest)
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