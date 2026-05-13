// Package handlers provides MCP tool handlers for interacting with Kubernetes.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

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

// resolveContextName returns the kubeconfig context to target for the
// current request. Precedence:
//  1. `context` argument on the MCP tool call
//  2. KUBERNETES_CONTEXT environment variable
//  3. empty string, meaning the kubeconfig's current-context
func resolveContextName(args map[string]interface{}) string {
	if name := getStringArg(args, "context", ""); name != "" {
		return name
	}
	return os.Getenv("KUBERNETES_CONTEXT")
}

// resolveK8sClient picks the kubernetes sub-client for the requested
// kubeconfig context, falling back to KUBERNETES_CONTEXT and finally
// to the kubeconfig's current-context.
func resolveK8sClient(client *k8s.Client, args map[string]interface{}) (*k8s.Client, error) {
	target, err := client.ForContext(resolveContextName(args))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve kubeconfig context: %w", err)
	}
	return target, nil
}

// ListContexts returns a handler function for the listContexts tool.
// It enumerates the kubeconfig contexts the server can target and
// reports which one is the current default.
func ListContexts(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		names, current, err := client.ListContexts()
		if err != nil {
			return nil, fmt.Errorf("failed to list contexts: %w", err)
		}

		response := map[string]interface{}{
			"contexts":       names,
			"currentContext": current,
		}
		if src := client.Source(); src != nil {
			response["authMode"] = src.Mode()
			response["supportsContexts"] = src.SupportsContexts()
			if path := src.KubeconfigPath(); path != "" {
				response["kubeconfigPath"] = path
			}
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}
		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
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

		target, err := resolveK8sClient(client, args)
		if err != nil {
			return nil, err
		}

		// Fetch API resources
		resources, err := target.GetAPIResources(ctx, includeNamespaceScoped, includeClusterScoped)
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
		fieldSelector := getStringArg(args, "fieldSelector", "")

		target, err := resolveK8sClient(client, args)
		if err != nil {
			return nil, err
		}

		// Fetch resources
		resources, err := target.ListResources(ctx, kind, namespace, labelSelector, fieldSelector)
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

		target, err := resolveK8sClient(client, args)
		if err != nil {
			return nil, err
		}

		resource, err := target.GetResource(ctx, kind, name, namespace)
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

		target, err := resolveK8sClient(client, args)
		if err != nil {
			return nil, err
		}

		// Fetch resource description
		resourceDescription, err := target.DescribeResource(ctx, kind, name, namespace)
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

		target, err := resolveK8sClient(client, args)
		if err != nil {
			return nil, err
		}

		logs, err := target.GetPodsLogs(ctx, namespace, containerName, name)
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

		target, err := resolveK8sClient(client, args)
		if err != nil {
			return nil, err
		}

		resourceUsage, err := target.GetNodeMetrics(ctx, name)
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

		target, err := resolveK8sClient(client, args)
		if err != nil {
			return nil, err
		}

		metrics, err := target.GetPodMetrics(ctx, namespace, podName)
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

		target, err := resolveK8sClient(client, args)
		if err != nil {
			return nil, err
		}

		events, err := target.GetEvents(ctx, namespace)
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
func CreateOrUpdateResourceJSON(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
		kind := getStringArg(args, "kind", "")

		target, err := resolveK8sClient(client, args)
		if err != nil {
			return nil, err
		}

		resource, err := target.CreateOrUpdateResourceJSON(ctx, namespace, manifest, kind)
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

// CreateOrUpdateResourceYAML returns a handler function for the createOrUpdateResourceYAML tool.
// It creates or updates a resource in the Kubernetes cluster based on the provided
// namespace and YAML manifest. This function is specifically optimized for YAML input.
// The result is serialized to JSON and returned.
func CreateOrUpdateResourceYAML(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments type: expected map[string]interface{}")
		}

		yamlManifest, err := getRequiredStringArg(args, "yamlManifest")
		if err != nil {
			// Backward compatibility: older clients sent `manifest`.
			yamlManifest, err = getRequiredStringArg(args, "manifest")
			if err != nil {
				return nil, err
			}
		}

		namespace := getStringArg(args, "namespace", "")
		kind := getStringArg(args, "kind", "")

		target, err := resolveK8sClient(client, args)
		if err != nil {
			return nil, err
		}

		resource, err := target.CreateOrUpdateResourceYAML(ctx, namespace, yamlManifest, kind)
		if err != nil {
			return nil, fmt.Errorf("failed to create or update resource from YAML: %w", err)
		}

		jsonResponse, err := json.Marshal(resource)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// DeleteResource returns a handler function for the deleteResource tool.
// It deletes a resource in the Kubernetes cluster based on the provided
// namespace and kind. The result is serialized to JSON and returned.
func DeleteResource(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

		target, err := resolveK8sClient(client, args)
		if err != nil {
			return nil, err
		}

		err = target.DeleteResource(ctx, kind, name, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to delete resource: %w", err)
		}

		return mcp.NewToolResultText("Resource deleted successfully"), nil
	}
}

// getIngresses returns a handler function for the getIngresses tool.
// It retrieves ingress resources from the Kubernetes cluster based on the provided
// Host and Path. The result is serialized to JSON and returned.
func GetIngresses(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments type: expected map[string]interface{}")
		}

		host := getStringArg(args, "host", "")

		target, err := resolveK8sClient(client, args)
		if err != nil {
			return nil, err
		}

		ingresses, err := target.GetIngresses(ctx, host)
		if err != nil {
			return nil, fmt.Errorf("failed to get ingress resources: %w", err)
		}

		jsonResponse, err := json.Marshal(ingresses)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// RolloutRestartHandler returns a handler function for the rolloutRestart tool.
// It calls the Client.RolloutRestart method and serializes the result to JSON.
func RolloutRestart(client *k8s.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments type: expected map[string]interface{}")
		}

		kind := getStringArg(args, "kind", "")
		name := getStringArg(args, "name", "")
		namespace := getStringArg(args, "namespace", "")

		if kind == "" || name == "" || namespace == "" {
			return nil, fmt.Errorf("kind, name, and namespace are required")
		}

		target, err := resolveK8sClient(client, args)
		if err != nil {
			return nil, err
		}

		result, err := target.RolloutRestart(ctx, kind, name, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to rollout restart resource: %w", err)
		}

		jsonResponse, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}
