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
	"flag"
	"fmt"
	"os"

	"github.com/reza-gholizade/k8s-mcp-server/handlers"
	"github.com/reza-gholizade/k8s-mcp-server/pkg/helm"
	"github.com/reza-gholizade/k8s-mcp-server/pkg/k8s"
	"github.com/reza-gholizade/k8s-mcp-server/tools"

	"github.com/mark3labs/mcp-go/server"
)

// main initializes the Kubernetes client, sets up the MCP server with
// Kubernetes tool handlers, and starts the server in the configured mode.
func main() {
	// Parse command line flags
	var mode string
	var port string

	flag.StringVar(&port, "port", getEnvOrDefault("SERVER_PORT", "8080"), "Server port")
	flag.StringVar(&mode, "mode", getEnvOrDefault("SERVER_MODE", "sse"), "Server mode: 'stdio' or 'sse'")
	flag.Parse()

	// Create MCP server
	s := server.NewMCPServer(
		"MCP K8S & Helm Server",
		"1.0.0",
		server.WithResourceCapabilities(true, true), // Enable resource listing and subscription capabilities
	)

	// Create a Kubernetes client
	client, err := k8s.NewClient("")
	if err != nil {
		fmt.Printf("Failed to create Kubernetes client: %v\n", err)
		return
	}

	// Create Helm client with default kubeconfig path
	helmClient, err := helm.NewClient("")
	if err != nil {
		fmt.Printf("Failed to create Helm client: %v\n", err)
		return
	}

	// Register Kubernetes tools
	s.AddTool(tools.GetAPIResourcesTool(), handlers.GetAPIResources(client))
	s.AddTool(tools.ListResourcesTool(), handlers.ListResources(client))
	s.AddTool(tools.GetResourcesTool(), handlers.GetResources(client))
	s.AddTool(tools.DescribeResourcesTool(), handlers.DescribeResources(client))
	s.AddTool(tools.GetPodsLogsTools(), handlers.GetPodsLogs(client))
	s.AddTool(tools.GetNodeMetricsTools(), handlers.GetNodeMetrics(client))
	s.AddTool(tools.GetPodMetricsTool(), handlers.GetPodMetrics(client))
	s.AddTool(tools.GetEventsTool(), handlers.GetEvents(client))
	s.AddTool(tools.CreateOrUpdateResourceTool(), handlers.CreateOrUpdateResource(client))

	// Register Helm tools
	s.AddTool(tools.HelmInstallTool(), handlers.HelmInstall(helmClient))
	s.AddTool(tools.HelmUpgradeTool(), handlers.HelmUpgrade(helmClient))
	s.AddTool(tools.HelmUninstallTool(), handlers.HelmUninstall(helmClient))
	s.AddTool(tools.HelmListTool(), handlers.HelmList(helmClient))
	s.AddTool(tools.HelmGetTool(), handlers.HelmGet(helmClient))
	s.AddTool(tools.HelmHistoryTool(), handlers.HelmHistory(helmClient))
	s.AddTool(tools.HelmRollbackTool(), handlers.HelmRollback(helmClient))

	// Start server based on mode
	switch mode {
	case "stdio":
		fmt.Println("Starting server in stdio mode...")
		if err := server.ServeStdio(s); err != nil {
			fmt.Printf("Failed to start stdio server: %v\n", err)
			return
		}
	case "sse":
		fmt.Printf("Starting server in SSE mode on port %s...\n", port)
		sse := server.NewSSEServer(s)
		if err := sse.Start(":" + port); err != nil {
			fmt.Printf("Failed to start SSE server: %v\n", err)
			return
		}
		fmt.Printf("SSE server started on port %s\n", port)
	default:
		fmt.Printf("Unknown server mode: %s. Use 'stdio' or 'sse'.\n", mode)
		return
	}
}

// getEnvOrDefault returns the value of the environment variable or the default value if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
