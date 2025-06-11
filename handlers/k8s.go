// Package handlers provides MCP tool handlers for interacting with Kubernetes.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/reza-gholizade/k8s-mcp-server/pkg/k8s"

	"github.com/mark3labs/mcp-go/mcp"
)

// Helper functions for consistent parameter extraction
func getStringArg(args map[string]interface{}, key string, defaultValue string) string {
	if val, ok := args[key].(string); ok {
		return val
	}
	return defaultValue
}

func getBoolArg(args map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := args[key].(bool); ok {
		return val
	}
	return defaultValue
}

func getRequiredStringArg(args map[string]interface{}, key string) (string, error) {
	val, ok := args[key].(string)
	if !ok || val == "" {
		return "", fmt.Errorf("missing required parameter: %s", key)
	}
	return val, nil
}

// GetAPIResources returns a handler function for the getAPIResources tool.
// It retrieves API resources from the Kubernetes cluster based on the provided
// context and parameters (includeNamespaceScoped, includeClusterScoped).
// The result is serialized to JSON and returned.
func GetAPIResources(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract arguments
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments type: expected map[string]interface{}")
		}

		includeNamespaceScoped := getBoolArg(args, "includeNamespaceScoped", true)
		includeClusterScoped := getBoolArg(args, "includeClusterScoped", true)

		// Fetch API resources
		resources, err := client.GetAPIResources(ctx, includeNamespaceScoped, includeClusterScoped)
		if err != nil {
			return nil, fmt.Errorf("failed to get API resources: %w", err)
		}

		// Serialize response to JSON
		jsonResponse, err := json.Marshal(resources)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		// Return JSON response using NewToolResultText
		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// ListResources returns a handler function for the listResources tool.
// It lists resources in the Kubernetes cluster based on the provided kind,
// namespace, and labelSelector. The result is serialized to JSON and returned.
func ListResources(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract arguments - using capital K to match your tools definition
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments type: expected map[string]interface{}")
		}

		kind, err := getRequiredStringArg(args, "Kind")
		if err != nil {
			return nil, err
		}

		namespace := getStringArg(args, "namespace", "")
		labelSelector := getStringArg(args, "labelSelector", "")

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

		// Return JSON response using NewToolResultText
		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// GetResources returns a handler function for the getResource tool.
// It retrieves a specific resource from the Kubernetes cluster based on the
// provided kind, name, and namespace. The result is serialized to JSON and returned.
func GetResources(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments type: expected map[string]interface{}")
		}

		kind, err := getRequiredStringArg(args, "kind")
		if err != nil {
			return nil, err
		}

		name, err := getRequiredStringArg(args, "name")
		if err != nil {
			return nil, err
		}

		namespace := getStringArg(args, "namespace", "")

		resource, err := client.GetResource(ctx, kind, name, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to get resource '%s' of kind '%s': %w", name, kind, err)
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
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments type: expected map[string]interface{}")
		}

		// Extract arguments - using capital K to match your tools definition
		kind, err := getRequiredStringArg(args, "Kind")
		if err != nil {
			return nil, err
		}

		name, err := getRequiredStringArg(args, "name")
		if err != nil {
			return nil, err
		}

		namespace := getStringArg(args, "namespace", "")

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

		// Return JSON response using NewToolResultText
		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// GetPodsLogs returns a handler function for the getPodsLogs tool.
// It retrieves logs for a specific pod from the Kubernetes cluster based on the
// provided name and namespace. The result is serialized to JSON and returned.
func GetPodsLogs(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments type: expected map[string]interface{}")
		}

		// Using capital N to match your tools definition
		name, err := getRequiredStringArg(args, "Name")
		if err != nil {
			return nil, err
		}

		namespace, err := getRequiredStringArg(args, "namespace")
		if err != nil {
			return nil, err
		}

		containerName := getStringArg(args, "containerName", "")

		logs, err := client.GetPodsLogs(ctx, namespace, containerName, name)
		if err != nil {
			return nil, fmt.Errorf("failed to get logs for pod '%s': %w", name, err)
		}

		// Return logs as plain text instead of JSON for better readability
		return mcp.NewToolResultText(logs), nil
	}
}

// GetNodeMetrics returns a handler function for the getNodeMetrics tool.
// It retrieves resource usage metrics for a specific node from the Kubernetes
// cluster based on the provided node name. The result is serialized to JSON
// and returned.
func GetNodeMetrics(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments type: expected map[string]interface{}")
		}

		// Using capital N to match your tools definition
		name, err := getRequiredStringArg(args, "Name")
		if err != nil {
			return nil, err
		}

		resourceUsage, err := client.GetNodeMetrics(ctx, name)
		if err != nil {
			return nil, fmt.Errorf("failed to get metrics for node '%s': %w", name, err)
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
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments type: expected map[string]interface{}")
		}

		namespace, err := getRequiredStringArg(args, "namespace")
		if err != nil {
			return nil, err
		}

		podName, err := getRequiredStringArg(args, "podName")
		if err != nil {
			return nil, err
		}

		metrics, err := client.GetPodMetrics(ctx, namespace, podName)
		if err != nil {
			return nil, fmt.Errorf("failed to get metrics for pod '%s' in namespace '%s': %w", podName, namespace, err)
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
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments type: expected map[string]interface{}")
		}

		namespace := getStringArg(args, "namespace", "")

		events, err := client.GetEvents(ctx, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to get events: %w", err)
		}

		jsonResponse, err := json.Marshal(events)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize events response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// CreateOrUpdateResource returns a handler function for the createOrUpdateResource tool.
// It creates or updates a resource in the Kubernetes cluster based on the provided
// namespace and manifest. The result is serialized to JSON and returned.
func CreateOrUpdateResource(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments type: expected map[string]interface{}")
		}

		manifest, err := getRequiredStringArg(args, "manifest")
		if err != nil {
			return nil, err
		}

		namespace := getStringArg(args, "namespace", "")
		resourceName := getStringArg(args, "resourceName", "")

		resource, err := client.CreateOrUpdateResource(ctx, namespace, manifest, resourceName)
		if err != nil {
			return nil, fmt.Errorf("failed to create or update resource: %w", err)
		}

		jsonResponse, err := json.Marshal(resource)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}
