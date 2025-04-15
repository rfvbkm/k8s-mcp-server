package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"k8s-mcp-server/pkg/k8s"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		cancel()
	}()

	// MCP server
	s := server.NewMCPServer(
		"Kubernetes Tools",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
		server.WithRecovery(),
	)

	// Get Kubernetes client from the k8s package
	clientset := k8s.GetClient()
	if clientset == nil {
		log.Fatalf("Failed to get Kubernetes client from k8s package")
		return
	}

	// Utility function to add tools
	addTool := func(name, description string, handler func(ctx context.Context) (string, error)) {
		tool := mcp.NewTool(name, mcp.WithDescription(description))
		s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			result, err := handler(ctx)
			if err != nil {
				return nil, logAndReturnError(err, fmt.Sprintf("Failed to handle tool: %s", name))
			}
			return mcp.NewToolResultText(result), nil
		})
	}

	// Add tools
	addTool("list_pods", "List all pods in the cluster", func(ctx context.Context) (string, error) {
		pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return "", err
		}
		return formatResourceList(pods.Items, func(pod metav1.Object) string {
			return fmt.Sprintf("- %s/%s", pod.GetNamespace(), pod.GetName())
		}), nil
	})

	addTool("list_nodes", "List all nodes in the cluster", func(ctx context.Context) (string, error) {
		nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			return "", err
		}
		return formatResourceList(nodes.Items, func(node metav1.Object) string {
			return fmt.Sprintf("- %s", node.GetName())
		}), nil
	})

	addTool("list_namespaces", "List all namespaces in the cluster", func(ctx context.Context) (string, error) {
		namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			return "", err
		}
		return formatResourceList(namespaces.Items, func(ns metav1.Object) string {
			return fmt.Sprintf("- %s", ns.GetName())
		}), nil
	})

	// Add more tools as needed...

	// Start the server
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// Utility functions
func logAndReturnError(err error, message string) error {
	log.Printf("%s: %v", message, err)
	return errors.New(message)
}

func formatResourceList(items []metav1.Object, formatFunc func(metav1.Object) string) string {
	var builder strings.Builder
	for _, item := range items {
		builder.WriteString(formatFunc(item) + "\n")
	}
	return builder.String()
}
