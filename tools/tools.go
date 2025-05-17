
// Package tools provides MCP tool handlers for interacting with Kubernetes.
package tools

import (
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

// GetPodsLogsTools creates a tool for getting pod logs.
// It defines the tool's name, description, and parameters for the pod name
// and namespace.
func GetPodsLogsTools() mcp.Tool {
	return mcp.NewTool(
		"getPodsLogs",
		mcp.WithDescription("Get logs of a specific pod in the Kubernetes cluster"),
		mcp.WithString("Name", mcp.Required(), mcp.Description("The name of the pod to get logs from")),
		mcp.WithString("containerName", mcp.Description("The name of the container to get logs from")),
		mcp.WithString("namespace", mcp.Description("The namespace of the pod")),
	)
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

// create or update resource of any type or kind
func CreateOrUpdateResourceTool() mcp.Tool {
	return mcp.NewTool(
		"createResource",
		mcp.WithDescription("Create a resource in the Kubernetes cluster"),
		mcp.WithString("kind", mcp.Required(), mcp.Description("The type of resource to create")),
		mcp.WithString("namespace", mcp.Description("The namespace of the resource")),
		mcp.WithString("manifest", mcp.Required(), mcp.Description("The manifest of the resource to create")),
	)
}