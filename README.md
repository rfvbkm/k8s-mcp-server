# Kubernetes MCP Server

A Kubernetes Model Control Plane (MCP) server that provides tools for interacting with Kubernetes clusters through a standardized interface.

## Features

- **API Resource Discovery**: Get all available API resources in your Kubernetes cluster
- **Resource Listing**: List resources of any type with optional namespace and label filtering
- **Resource Details**: Get detailed information about specific Kubernetes resources
- **Resource Description**: Get comprehensive descriptions of Kubernetes resources, similar to `kubectl describe`
- **Pod Logs**: Retrieve logs from specific pods
- **Node Metrics**: Get resource usage metrics for specific nodes
- **Pod Metrics**: Get CPU and Memory metrics for specific pods
- **Event Listing**: List events within a namespace or for a specific resource.
- **Resource Creation**: Create new Kubernetes resources from a manifest.
- **Standardized Interface**: Uses the MCP protocol for consistent tool interaction
- **Flexible Configuration**: Supports different Kubernetes contexts and resource scopes

## Prerequisites

- Go 1.20 or later
- Access to a Kubernetes cluster
- `kubectl` configured with appropriate cluster access

## Installation

1. Clone the repository:
```bash
git clone https://github.com/reza-gholizade/k8s-mcp-server.git
cd k8s-mcp-server
```

2. Install dependencies:
```bash
go mod download
```

3. Build the server:
```bash
go build -o k8s-mcp-server main.go
```

## Usage

### Starting the Server

Run the server:
```bash
./k8s-mcp-server
```

The server will start and listen on stdin/stdout for MCP protocol messages.

### Available Tools

#### 1. `getAPIResources`

Retrieves all available API resources in the Kubernetes cluster.

**Parameters:**
- `includeNamespaceScoped` (boolean): Whether to include namespace-scoped resources (defaults to true)
- `includeClusterScoped` (boolean): Whether to include cluster-scoped resources (defaults to true)

**Example:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "getAPIResources",
  "params": {
    "arguments": {
      "includeNamespaceScoped": true,
      "includeClusterScoped": true
    }
  }
}
```

#### 2. `listResources`

Lists all instances of a specific resource type.

**Parameters:**
- `Kind` (string, required): The kind of resource to list (e.g., "Pod", "Deployment")
- `namespace` (string): The namespace to list resources from (if omitted, lists across all namespaces for namespaced resources)
- `labelSelector` (string): Filter resources by label selector

**Example:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "listResources",
  "params": {
    "arguments": {
      "Kind": "Pod",
      "namespace": "default",
      "labelSelector": "app=nginx"
    }
  }
}
```

#### 3. `getResource`

Retrieves detailed information about a specific resource.

**Parameters:**
- `kind` (string, required): The kind of resource to get (e.g., "Pod", "Deployment")
- `name` (string, required): The name of the resource to get
- `namespace` (string): The namespace of the resource (if it's a namespaced resource)

**Example:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "getResource",
  "params": {
    "arguments": {
      "kind": "Pod",
      "name": "nginx-pod",
      "namespace": "default"
    }
  }
}
```

#### 4. `describeResource`

Describes a resource in the Kubernetes cluster based on given kind and name, similar to `kubectl describe`.

**Parameters:**
- `Kind` (string, required): The kind of resource to describe (e.g., "Pod", "Deployment")
- `name` (string, required): The name of the resource to describe
- `namespace` (string): The namespace of the resource (if it's a namespaced resource)

**Example:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "describeResource",
  "params": {
    "arguments": {
      "Kind": "Pod",
      "name": "nginx-pod",
      "namespace": "default"
    }
  }
}
```

#### 5. `getPodsLogs`

Retrieves the logs of a specific pod in the Kubernetes cluster.

**Parameters:**
- `Name` (string, required): The name of the pod to get logs from.
- `namespace` (string): The namespace of the pod (if it's a namespaced resource).

**Example:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "getPodsLogs",
  "params": {
    "arguments": {
      "Name": "my-app-pod-12345",
      "namespace": "production"
    }
  }
}
```

#### 6. `getNodeMetrics`

Retrieves resource usage metrics for a specific node in the Kubernetes cluster.

**Parameters:**
- `Name` (string, required): The name of the node to get metrics from.

**Example:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "getNodeMetrics",
  "params": {
    "arguments": {
      "Name": "worker-node-1"
    }
  }
}
```

#### 7. `getPodMetrics`

Retrieves CPU and Memory metrics for a specific pod in the Kubernetes cluster.

**Parameters:**
- `namespace` (string, required): The namespace of the pod.
- `podName` (string, required): The name of the pod.

**Example:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "getPodMetrics",
  "params": {
    "arguments": {
      "namespace": "default",
      "podName": "my-app-pod-67890"
    }
  }
}
```

#### 8. `getEvents`

Retrieves events for a specific namespace or resource in the Kubernetes cluster.

**Parameters:**
- `namespace` (string): The namespace to get events from. If omitted, events from all namespaces are considered (if permitted by RBAC).
- `resourceName` (string): The name of a specific resource (e.g., a Pod name) to filter events for.
- `resourceKind` (string): The kind of the specific resource (e.g., "Pod") if `resourceName` is provided.

**Example (Namespace Events):**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "getEvents",
  "params": {
    "arguments": {
      "namespace": "default"
    }
  }
}
```

**Example (Resource Events):**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "getEvents",
  "params": {
    "arguments": {
      "namespace": "production",
      "resourceName": "my-app-pod-12345",
      "resourceKind": "Pod"
    }
  }
}
```

#### 9. `createorUpdateResource`

Creates a new resource in the Kubernetes cluster from a YAML or JSON manifest.

**Parameters:**
- `manifest` (string, required): The YAML or JSON manifest of the resource to create.
- `namespace` (string, optional): The namespace in which to create the resource. If the manifest contains a namespace, this parameter can be omitted or used to override it (behavior might depend on server implementation).

**Example:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "createResource",
  "params": {
    "arguments": {
      "namespace": "default",
      "manifest": "apiVersion: v1\nkind: Pod\nmetadata:\n  name: my-new-pod\nspec:\n  containers:\n  - name: nginx\n    image: nginx:latest"
    }
  }
}
```

## Development

### Project Structure

```
.
├── handlers/         # Tool handlers
│   └── handlers.go   # Implementation of MCP tools
├── pkg/             # Internal packages
│   └── k8s/         # Kubernetes client implementation
├── main.go          # Server entry point
├── go.mod           # Go module definition
└── go.sum           # Go module checksums
```

### Adding New Tools

To add a new tool:

1. Create a new tool definition function (e.g., `MyNewTool() mcp.Tool`) in `handlers/handlers.go`
2. Implement the tool handler function (e.g., `MyNewHandler(client *k8s.Client) func(...)`) in `handlers/handlers.go`
3. Register the tool and its handler in `main.go` using `s.AddTool()`

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to contribute to this project.

## License
gholizade.net@gmail.com

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

