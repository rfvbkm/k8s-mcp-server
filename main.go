package main

import (
	"fmt"
	"github.com/reza-gholizade/k8s-mcp-server/handlers"
	"github.com/reza-gholizade/k8s-mcp-server/pkg/k8s"

	"github.com/mark3labs/mcp-go/server"
)

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

	// Start the stdio server, which listens on stdin/stdout
	fmt.Println("Starting MCP stdio server. Listening on stdin...")
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
