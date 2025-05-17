// Package k8s provides a client for interacting with the Kubernetes API.
package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

// Client encapsulates Kubernetes client functionality including dynamic,
// discovery, and metrics clients.
// It also caches API resource information for performance.
type Client struct {
	clientset        *kubernetes.Clientset
	dynamicClient    dynamic.Interface
	discoveryClient  *discovery.DiscoveryClient
	metricsClientset *metricsclientset.Clientset // Add metrics client
	restConfig       *rest.Config
	apiResourceCache map[string]*schema.GroupVersionResource
	cacheLock        sync.RWMutex
}

// NewClient creates a new Kubernetes client.
// It initializes the standard clientset, dynamic client, discovery client,
// and metrics client using the provided kubeconfig path or the default path.
// If kubeconfigPath is empty, it defaults to ~/.kube/config.
func NewClient(kubeconfigPath string) (*Client, error) {
	var kubeconfig string
	if kubeconfigPath != "" {
		kubeconfig = kubeconfigPath
	} else if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes configuration: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery client: %w", err)
	}

	// Initialize metrics client
	metricsClient, err := metricsclientset.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client: %w", err)
	}

	return &Client{
		clientset:        clientset,
		dynamicClient:    dynamicClient,
		discoveryClient:  discoveryClient,
		metricsClientset: metricsClient, // Assign metrics client
		restConfig:       config,
		apiResourceCache: make(map[string]*schema.GroupVersionResource),
	}, nil
}

// GetAPIResources retrieves all API resource types in the cluster.
// It uses the discovery client to fetch server-preferred resources.
// Filters resources based on includeNamespaceScoped and includeClusterScoped flags.
// Returns a slice of maps, each representing an API resource, or an error.
func (c *Client) GetAPIResources(ctx context.Context, includeNamespaceScoped, includeClusterScoped bool) ([]map[string]interface{}, error) {
	resourceLists, err := c.discoveryClient.ServerPreferredResources()
	if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
		return nil, fmt.Errorf("failed to retrieve API resources: %w", err)
	}

	var resources []map[string]interface{}
	for _, resourceList := range resourceLists {
		for _, resource := range resourceList.APIResources {
			if (resource.Namespaced && !includeNamespaceScoped) || (!resource.Namespaced && !includeClusterScoped) {
				continue
			}
			resources = append(resources, map[string]interface{}{
				"name":         resource.Name,
				"singularName": resource.SingularName,
				"namespaced":   resource.Namespaced,
				"kind":         resource.Kind,
				"group":        resource.Group,
				"version":      resource.Version,
				"verbs":        resource.Verbs,
			})
		}
	}
	return resources, nil
}

// GetResource retrieves detailed information about a specific resource.
// It uses the dynamic client to fetch the resource by kind, name, and namespace.
// It utilizes a cached GroupVersionResource (GVR) for efficiency.
// Returns the unstructured content of the resource as a map, or an error.
func (c *Client) GetResource(ctx context.Context, kind, name, namespace string) (map[string]interface{}, error) {
	gvr, err := c.getCachedGVR(kind)
	if err != nil {
		return nil, err
	}

	var obj *unstructured.Unstructured
	if namespace != "" {
		obj, err = c.dynamicClient.Resource(*gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	} else {
		obj, err = c.dynamicClient.Resource(*gvr).Get(ctx, name, metav1.GetOptions{})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve resource: %w", err)
	}

	return obj.UnstructuredContent(), nil
}

// ListResources lists all instances of a specific resource type.
// It uses the dynamic client and supports filtering by namespace, labelSelector,
// and fieldSelector.
// It utilizes a cached GroupVersionResource (GVR) for efficiency.
// Returns a slice of maps, each representing a resource instance, or an error.
func (c *Client) ListResources(ctx context.Context, kind, namespace, labelSelector, fieldSelector string) ([]map[string]interface{}, error) {
	gvr, err := c.getCachedGVR(kind)
	if err != nil {
		return nil, err
	}

	options := metav1.ListOptions{
		LabelSelector: labelSelector,
		FieldSelector: fieldSelector,
	}

	var list *unstructured.UnstructuredList
	if namespace != "" {
		list, err = c.dynamicClient.Resource(*gvr).Namespace(namespace).List(ctx, options)
	} else {
		list, err = c.dynamicClient.Resource(*gvr).List(ctx, options)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	var resources []map[string]interface{}
	for _, item := range list.Items {
		resources = append(resources, item.UnstructuredContent())
	}
	return resources, nil
}

// CreateOrUpdateResource creates a new resource or updates an existing one.
// It parses the provided manifest string into an unstructured object.
// It uses the dynamic client to first attempt an update, and if that fails
// (e.g., resource not found), it attempts to create the resource.
// Requires the resource manifest to include a name.
// Returns the unstructured content of the created/updated resource, or an error.
func (c *Client) CreateOrUpdateResource(ctx context.Context, kind, namespace, manifest string) (map[string]interface{}, error) {
	obj := &unstructured.Unstructured{}
	if err := json.Unmarshal([]byte(manifest), &obj.Object); err != nil {
		return nil, fmt.Errorf("failed to parse resource manifest: %w", err)
	}

	gvr, err := c.getCachedGVR(kind)
	if err != nil {
		return nil, err
	}

	var result *unstructured.Unstructured
	if namespace != "" {
		obj.SetNamespace(namespace)
	}

	if obj.GetName() == "" {
		return nil, fmt.Errorf("resource name is required")
	}

	// Try to update the resource; if it doesn't exist, create it
	result, err = c.dynamicClient.Resource(*gvr).Namespace(obj.GetNamespace()).Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		result, err = c.dynamicClient.Resource(*gvr).Namespace(obj.GetNamespace()).Create(ctx, obj, metav1.CreateOptions{})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create or update resource: %w", err)
	}

	return result.UnstructuredContent(), nil
}

// DeleteResource deletes a specific resource.
// It uses the dynamic client to delete the resource by kind, name, and namespace.
// It utilizes a cached GroupVersionResource (GVR) for efficiency.
// Returns an error if the deletion fails.
func (c *Client) DeleteResource(ctx context.Context, kind, name, namespace string) error {
	gvr, err := c.getCachedGVR(kind)
	if err != nil {
		return err
	}

	var deleteErr error
	if namespace != "" {
		deleteErr = c.dynamicClient.Resource(*gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	} else {
		deleteErr = c.dynamicClient.Resource(*gvr).Delete(ctx, name, metav1.DeleteOptions{})
	}
	if deleteErr != nil {
		return fmt.Errorf("failed to delete resource: %w", deleteErr)
	}
	return nil
}

// getCachedGVR retrieves the GroupVersionResource for a given kind, using a cache for performance
func (c *Client) getCachedGVR(kind string) (*schema.GroupVersionResource, error) {
	c.cacheLock.RLock()
	if gvr, exists := c.apiResourceCache[kind]; exists {
		c.cacheLock.RUnlock()
		return gvr, nil
	}
	c.cacheLock.RUnlock()

	// Cache miss; fetch from discovery client
	resourceLists, err := c.discoveryClient.ServerPreferredResources()
	if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
		return nil, fmt.Errorf("failed to retrieve API resources: %w", err)
	}

	for _, resourceList := range resourceLists {
		gv, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			continue
		}
		for _, resource := range resourceList.APIResources {
			if resource.Kind == kind {
				gvr := &schema.GroupVersionResource{
					Group:    gv.Group,
					Version:  gv.Version,
					Resource: resource.Name,
				}
				c.cacheLock.Lock()
				c.apiResourceCache[kind] = gvr
				c.cacheLock.Unlock()
				return gvr, nil
			}
		}
	}

	return nil, fmt.Errorf("resource type %s not found", kind)
}

// DescribeResource retrieves detailed information about a specific resource, similar to GetResource.
// It uses the dynamic client to fetch the resource by kind, name, and namespace.
// It utilizes a cached GroupVersionResource (GVR) for efficiency.
// Returns the unstructured content of the resource as a map, or an error.
// Note: This function currently has the same implementation as GetResource.
func (c *Client) DescribeResource(ctx context.Context, kind, name, namespace string) (map[string]interface{}, error) {
	gvr, err := c.getCachedGVR(kind)
	if err != nil {
		return nil, err
	}

	var obj *unstructured.Unstructured
	if namespace != "" {
		obj, err = c.dynamicClient.Resource(*gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	} else {
		obj, err = c.dynamicClient.Resource(*gvr).Get(ctx, name, metav1.GetOptions{})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve resource: %w", err)
	}

	return obj.UnstructuredContent(), nil
}

// GetPodsLogs retrieves the logs for a specific pod.
// It uses the corev1 clientset to fetch logs, limiting to the last 100 lines by default.
// If containerName is provided, it gets logs for that specific container.
// If containerName is empty and the pod has multiple containers, it gets logs from all containers.
// Returns the logs as a string, or an error.
func (c *Client) GetPodsLogs(ctx context.Context, namespace, containerName, podName string) (string, error) {
	tailLines := int64(100)
	podLogOptions := &corev1.PodLogOptions{
		TailLines: &tailLines,
	}

	// If container name is provided, use it
	if containerName != "" {
		podLogOptions.Container = containerName
		req := c.clientset.CoreV1().Pods(namespace).GetLogs(podName, podLogOptions)
		logs, err := req.Stream(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to get logs for container '%s': %w", containerName, err)
		}
		defer logs.Close()

		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, logs); err != nil {
			return "", fmt.Errorf("failed to read logs: %w", err)
		}
		return buf.String(), nil
	}

	// If no container name provided, first get the pod to check its containers
	pod, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get pod details: %w", err)
	}

	// If the pod has only one container, get logs from that container
	if len(pod.Spec.Containers) == 1 {
		req := c.clientset.CoreV1().Pods(namespace).GetLogs(podName, podLogOptions)
		logs, err := req.Stream(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to get logs: %w", err)
		}
		defer logs.Close()

		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, logs); err != nil {
			return "", fmt.Errorf("failed to read logs: %w", err)
		}
		return buf.String(), nil
	}

	// If the pod has multiple containers, get logs from each container
	var allLogs strings.Builder
	for _, container := range pod.Spec.Containers {
		containerLogOptions := podLogOptions.DeepCopy()
		containerLogOptions.Container = container.Name

		req := c.clientset.CoreV1().Pods(namespace).GetLogs(podName, containerLogOptions)
		logs, err := req.Stream(ctx)
		if err != nil {
			allLogs.WriteString(fmt.Sprintf("\n--- Error getting logs for container %s: %v ---\n", container.Name, err))
			continue
		}

		allLogs.WriteString(fmt.Sprintf("\n--- Logs for container %s ---\n", container.Name))
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, logs)
		logs.Close()

		if err != nil {
			allLogs.WriteString(fmt.Sprintf("Error reading logs: %v\n", err))
		} else {
			allLogs.WriteString(buf.String())
		}
	}

	return allLogs.String(), nil
}

// GetPodMetrics retrieves CPU and Memory metrics for a specific pod.
// It uses the metrics clientset to fetch pod metrics.
// Returns a map containing pod metadata and container metrics, or an error.
func (c *Client) GetPodMetrics(ctx context.Context, namespace, podName string) (map[string]interface{}, error) {
	podMetrics, err := c.metricsClientset.MetricsV1beta1().PodMetricses(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics for pod '%s' in namespace '%s': %w", podName, namespace, err)
	}

	metricsResult := map[string]interface{}{
		"podName":    podName,
		"namespace":  namespace,
		"timestamp":  podMetrics.Timestamp.Time,
		"window":     podMetrics.Window.Duration.String(),
		"containers": []map[string]interface{}{},
	}

	containerMetricsList := []map[string]interface{}{}
	for _, container := range podMetrics.Containers {
		containerMetrics := map[string]interface{}{
			"name":   container.Name,
			"cpu":    container.Usage.Cpu().String(),    // Format Quantity
			"memory": container.Usage.Memory().String(), // Format Quantity
		}
		containerMetricsList = append(containerMetricsList, containerMetrics)
	}
	metricsResult["containers"] = containerMetricsList

	return metricsResult, nil
}

// GetNodeMetrics retrieves CPU and Memory metrics for a specific Node.
// It uses the metrics clientset to fetch node metrics.
// Returns a map containing node metadata and resource usage, or an error.
func (c *Client) GetNodeMetrics(ctx context.Context, nodeName string) (map[string]interface{}, error) {
	nodeMetrics, err := c.metricsClientset.MetricsV1beta1().NodeMetricses().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics for node '%s': %w", nodeName, err)
	}

	metricsResult := map[string]interface{}{
		"nodeName":  nodeName,
		"timestamp": nodeMetrics.Timestamp.Time,
		"window":    nodeMetrics.Window.Duration.String(),
		"usage": map[string]string{
			"cpu":    nodeMetrics.Usage.Cpu().String(),    // Format Quantity
			"memory": nodeMetrics.Usage.Memory().String(), // Format Quantity
		},
	}

	return metricsResult, nil
}

// GetEvents retrieves events for a specific namespace or all namespaces.
// It uses the corev1 clientset to fetch events.
// Returns a slice of maps, each representing an event, or an error.
func (c *Client) GetEvents(ctx context.Context, namespace string) ([]map[string]interface{}, error) {
	var eventList *corev1.EventList
	var err error

	if namespace != "" {
		eventList, err = c.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	} else {
		eventList, err = c.clientset.CoreV1().Events("").List(ctx, metav1.ListOptions{})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve events: %w", err)
	}

	var events []map[string]interface{}
	for _, event := range eventList.Items {
		events = append(events, map[string]interface{}{
			"name":      event.Name,
			"namespace": event.Namespace,
			"reason":    event.Reason,
			"message":   event.Message,
			"source":    event.Source.Component,
			"type":      event.Type,
			"count":     event.Count,
			"firstTime": event.FirstTimestamp.Time,
			"lastTime":  event.LastTimestamp.Time,
		})
	}
	return events, nil
}
