// Package main is the entry point for the Kubernetes MCP server.
// Manage Kubernetes Cluster workloads via MCP.
// It initializes the MCP server, sets up the Kubernetes client,
// and registers the necessary handlers for various Kubernetes operations.
// It also starts the server to listen for incoming requests on stdin/stdout.
// It uses the MCP Go library to create the server and handle requests.
// The server is capable of handling various Kubernetes operations
// such as listing resources, getting resource details, and retrieving logs.

package main

import (
	"fmt"

	"github.com/reza-gholizade/k8s-mcp-server/handlers"
	"github.com/reza-gholizade/k8s-mcp-server/pkg/k8s"

	"github.com/mark3labs/mcp-go/server"
)

// main initializes the Kubernetes client, sets up the MCP server with
// Kubernetes tool handlers, and starts the server listening on stdio.
func main() {
	// Create MCP server
	s := server.NewMCPServer(
		" MCP K8S Server",
		"1.0.0",
		server.WithResourceCapabilities(true, true), // Enable resource listing and subscription capabilities
	)

	// Create a Kubernetes client
	client, err := k8s.NewClient("")
	if err != nil {
		fmt.Printf("Failed to create Kubernetes client: %v\n", err)
		return
	}

	// Register the tool and its handler with the server
	s.AddTool(handlers.GetAPIResourcesTool(), handlers.GetAPIResources(client))
	s.AddTool(handlers.ListResourcesTool(), handlers.ListResources(client))
	s.AddTool(handlers.GetResourcesTool(), handlers.GetResources(client))
	s.AddTool(handlers.DescribeResourcesTool(), handlers.DescribeResources(client))
	s.AddTool(handlers.GetPodsLogsTools(), handlers.GetPodsLogs(client))
	s.AddTool(handlers.GetNodeMetricsTools(), handlers.GetNodeMetrics(client))
	s.AddTool(handlers.GetPodMetricsTool(), handlers.GetPodMetrics(client))
	s.AddTool(handlers.GetEventsTool(), handlers.GetEvents(client))
	s.AddTool(handlers.CreateOrUpdateResourceTool(), handlers.CreateOrUpdateResource(client))


	// Start SSE server
	sse := server.NewSSEServer(s)
	port := ":8080"
	if err := sse.Start(port); err != nil {
		fmt.Printf("Failed to start SSE server: %v\n", err)
		return
	}
	fmt.Printf("SSE server started on port %s\n", port)
}

