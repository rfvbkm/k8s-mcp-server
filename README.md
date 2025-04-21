# Kubernetes MCP Server

A Kubernetes Model Control Plane (MCP) server that provides tools for interacting with Kubernetes clusters through a standardized interface.

## Features

- **API Resource Discovery**: Get all available API resources in your Kubernetes cluster
- **Resource Listing**: List resources of any type with optional namespace and label filtering
- **Standardized Interface**: Uses the MCP protocol for consistent tool interaction
- **Flexible Configuration**: Supports different Kubernetes contexts and resource scopes

## Prerequisites

- Go 1.16 or later
- Access to a Kubernetes cluster
- `kubectl` configured with appropriate cluster access

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/k8s-mcp-server.git
cd k8s-mcp-server
```

2. Install dependencies:
```bash
go mod download
```

3. Build the server:
```bash
go build
```

## Usage

### Starting the Server

Run the server:
```bash
./k8s-mcp-server
```

The server will start and listen on stdin/stdout for MCP protocol messages.

### Available Tools

#### 1. Get API Resources

Get all available API resources in your Kubernetes cluster:

```json
{
  "tool": "getAPIResources",
  "params": {
    "includeNamespaceScoped": true,
    "includeClusterScoped": true
  }
}
```

Parameters:
- `includeNamespaceScoped` (boolean): Include namespace-scoped resources (default: true)
- `includeClusterScoped` (boolean): Include cluster-scoped resources (default: true)

#### 2. List Resources

List resources of a specific type:

```json
{
  "tool": "listResources",
  "params": {
    "Kind": "Pod",
    "namespace": "default",
    "labelSelector": "app=myapp"
  }
}
```

Parameters:
- `Kind` (string, required): The type of resource to list (e.g., "Pod", "Deployment")
- `namespace` (string): The namespace to list resources in (optional)
- `labelSelector` (string): A label selector to filter resources (optional)

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

1. Create a new tool definition in `handlers/handlers.go`
2. Implement the tool handler function
3. Register the tool in `main.go`

## License

gholizade.net@gmail.com

