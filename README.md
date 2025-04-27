# Kubernetes MCP Server

A Kubernetes Model Control Plane (MCP) server that provides tools for interacting with Kubernetes clusters through a standardized interface.

## Features

- **API Resource Discovery**: Get all available API resources in your Kubernetes cluster
- **Resource Listing**: List resources of any type with optional namespace and label filtering
- **Resource Details**: Get detailed information about specific Kubernetes resources
- **Resource Description**: Get comprehensive descriptions of Kubernetes resources
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

## License

gholizade.net@gmail.com

