package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	// networkingv1 "k8s.io/api/networking/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes"

	"k8s-mcp-server/pkg/k8s"
)

func main() {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	handleGracefulShutdown(cancel)

	// MCP server setup
	s := newMCPServer()

	// Kubernetes client setup
	clientset, err := getK8sClient()
	if err != nil {
		log.Fatalf("Failed to get Kubernetes client: %v", err)
		return
	}

	// Register tools
	registerTools(s, clientset)
	registerLogTool(s, clientset) // Register the new log tool

	// Start the server
	startServer(s)
}

func handleGracefulShutdown(cancel context.CancelFunc) {
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Println("Shutting down gracefully...")
		cancel()
	}()
}

func newMCPServer() *server.MCPServer {
	return server.NewMCPServer(
		"Kubernetes Tools",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
		server.WithRecovery(),
	)
}

func getK8sClient() (*kubernetes.Clientset, error) {
	clientset := k8s.GetClient()
	if clientset == nil {
		return nil, errors.New("failed to get Kubernetes client from k8s package")
	}
	return clientset, nil
}

func registerTools(s *server.MCPServer, clientset *kubernetes.Clientset) {
	addListTool(s, clientset)
	addDescribeTool(s, clientset)
}

func addListTool(s *server.MCPServer, clientset *kubernetes.Clientset) {
	addTool := func(name, description string, resourceType string) {
		tool := mcp.NewTool(name, mcp.WithDescription(description))
		s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			result, err := listResources(ctx, clientset, resourceType)
			if err != nil {
				return nil, logAndReturnError(err, fmt.Sprintf("Failed to handle tool: %s", name))
			}
			return mcp.NewToolResultText(result), nil
		})
	}

	addTool("list_pods", "List all pods in the cluster", "pod")
	addTool("list_nodes", "List all nodes in the cluster", "node")
	addTool("list_namespaces", "List all namespaces in the cluster", "namespace")
	addTool("list_deployments", "List all deployments in the cluster", "deployment")
	addTool("list_statefulsets", "List all statefulsets in the cluster", "statefulset")
	addTool("list_endpoints", "List all endpoints in the cluster", "endpoint")
	addTool("list_secrets", "List all secrets in the cluster", "secret")
	addTool("list_storageclasses", "List all storageclasses in the cluster", "storageclass")
}

func addDescribeTool(s *server.MCPServer, clientset *kubernetes.Clientset) {
	addTool := func(name, description string, resourceType string) {
		tool := mcp.NewTool(name,
			mcp.WithDescription(description),
			mcp.WithString("name", mcp.Required(), mcp.Description("Name of the resource")),
			mcp.WithString("namespace", mcp.Required(), mcp.Description("Namespace of the resource")),
		)
		s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			result, err := describeResource(ctx, clientset, resourceType, request)
			if err != nil {
				return nil, logAndReturnError(err, fmt.Sprintf("Failed to handle tool: %s", name))
			}
			return result, nil
		})
	}

	addTool("describe_pod", "Describe a pod", "pod")
	addTool("describe_deployment", "Describe a deployment", "deployment")
	addTool("describe_service", "Describe a service", "service")
	addTool("describe_ingress", "Describe an ingress", "ingress")
	addTool("describe_endpoint", "Describe an endpoint", "endpoint")
	addTool("describe_storageclass", "Describe a storageclass", "storageclass")
	addTool("describe_node", "Describe a node", "node")
}

func listResources(ctx context.Context, clientset *kubernetes.Clientset, resourceType string) (string, error) {
	listOptions := metav1.ListOptions{}
	var builder strings.Builder

	resourceListFunc := func(resourceType string) (runtime.Object, error) {
		switch resourceType {
		case "pod":
			return clientset.CoreV1().Pods("").List(ctx, listOptions)
		case "node":
			return clientset.CoreV1().Nodes().List(ctx, listOptions)
		case "namespace":
			return clientset.CoreV1().Namespaces().List(ctx, listOptions)
		case "deployment":
			return clientset.AppsV1().Deployments("").List(ctx, listOptions)
		case "statefulset":
			return clientset.AppsV1().StatefulSets("").List(ctx, listOptions)
		case "endpoint":
			return clientset.CoreV1().Endpoints("").List(ctx, listOptions)
		case "secret":
			return clientset.CoreV1().Secrets("").List(ctx, listOptions)
		case "storageclass":
			return clientset.StorageV1().StorageClasses().List(ctx, listOptions)
		default:
			return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
		}
	}

	formatResource := func(resourceType string, obj runtime.Object) string {
		switch resourceType {
		case "pod":
			pod := obj.(*corev1.Pod)
			return fmt.Sprintf("- %s/%s\n", pod.Namespace, pod.Name)
		case "node":
			node := obj.(*corev1.Node)
			return fmt.Sprintf("- %s\n", node.Name)
		case "namespace":
			ns := obj.(*corev1.Namespace)
			return fmt.Sprintf("- %s\n", ns.Name)
		case "deployment":
			deployment := obj.(*appsv1.Deployment)
			return fmt.Sprintf("- %s/%s\n", deployment.Namespace, deployment.Name)
		case "statefulset":
			statefulSet := obj.(*appsv1.StatefulSet)
			return fmt.Sprintf("- %s/%s\n", statefulSet.Namespace, statefulSet.Name)
		case "endpoint":
			endpoint := obj.(*corev1.Endpoints)
			return fmt.Sprintf("- %s/%s\n", endpoint.Namespace, endpoint.Name)
		case "secret":
			secret := obj.(*corev1.Secret)
			return fmt.Sprintf("- %s/%s\n", secret.Namespace, secret.Name)
		case "storageclass":
			sc := obj.(*storagev1.StorageClass)
			return fmt.Sprintf("- %s\n", sc.Name)
		default:
			return ""
		}
	}

	resourceList, err := resourceListFunc(resourceType)
	if err != nil {
		return "", err
	}

	itemsValue := reflect.ValueOf(resourceList).Elem().FieldByName("Items")
	if !itemsValue.IsValid() {
		return "", fmt.Errorf("Items field not found in %T", resourceList)
	}

	for i := 0; i < itemsValue.Len(); i++ {
		item := itemsValue.Index(i).Interface().(runtime.Object)
		builder.WriteString(formatResource(resourceType, item))
	}

	return builder.String(), nil
}

func describeResource(ctx context.Context, clientset *kubernetes.Clientset, resourceType string, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, okName := request.Params.Arguments["name"].(string)
	namespace, okNamespace := request.Params.Arguments["namespace"].(string)

	if !okName || !okNamespace {
		return nil, errors.New("name and namespace are required parameters")
	}

	var (
		resource runtime.Object
		err      error
	)

	switch resourceType {
	case "pod":
		resource, err = clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	case "deployment":
		resource, err = clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	case "service":
		resource, err = clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	case "ingress":
		resource, err = clientset.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
	case "endpoint":
		resource, err = clientset.CoreV1().Endpoints(namespace).Get(ctx, name, metav1.GetOptions{})
	case "storageclass":
		resource, err = clientset.StorageV1().StorageClasses().Get(ctx, name, metav1.GetOptions{})
	case "node":
		node, err := clientset.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		resource = node
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	if err != nil {
		return nil, err
	}

	// Serialize the resource to YAML
	s := json.NewSerializerWithOptions(json.DefaultMetaFactory, nil, nil, json.SerializerOptions{
		Yaml:   true,
		Pretty: false,
		Strict: false,
	})
	var builder strings.Builder
	err = s.Encode(resource, &builder)
	if err != nil {
		return nil, fmt.Errorf("error encoding resource to YAML: %w", err)
	}

	return mcp.NewToolResultText(builder.String()), nil
}

func startServer(s *server.MCPServer) {
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func logAndReturnError(err error, message string) error {
	log.Printf("%s: %v", message, err)
	return errors.New(message)
}

func registerLogTool(s *server.MCPServer, clientset *kubernetes.Clientset) {
	tool := mcp.NewTool(
		"get_pod_logs",
		mcp.WithDescription("Retrieve logs from a specific pod"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Name of the pod")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Namespace of the pod")),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, okName := request.Params.Arguments["name"].(string)
		namespace, okNamespace := request.Params.Arguments["namespace"].(string)

		if !okName || !okNamespace {
			return nil, errors.New("name and namespace are required parameters")
		}

		result, err := getPodLogs(ctx, clientset, name, namespace)
		if err != nil {
			return nil, logAndReturnError(err, fmt.Sprintf("Failed to get logs for pod %s/%s", namespace, name))
		}

		return mcp.NewToolResultText(result), nil
	})
}

func getPodLogs(ctx context.Context, clientset *kubernetes.Clientset, podName, podNamespace string) (string, error) {
	req := clientset.CoreV1().Pods(podNamespace).GetLogs(podName, &corev1.PodLogOptions{})
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("error in opening stream: %w", err)
	}
	defer podLogs.Close()

	buf := new(strings.Builder)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", fmt.Errorf("error in copy information from podLogs to buf: %w", err)
	}
	str := buf.String()
	return str, nil
}
