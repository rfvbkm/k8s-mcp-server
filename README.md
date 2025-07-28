# Kubernetes MCP Server

A Kubernetes Model Context Protocol (MCP) server that provides tools for interacting with Kubernetes clusters through a standardized interface.

## Features

- **API Resource Discovery**: Get all available API resources in your Kubernetes cluster.
- **Resource Listing**: List resources of any type with optional namespace and label filtering.
- **Resource Details**: Get detailed information about specific Kubernetes resources.
- **Resource Description**: Get comprehensive descriptions of Kubernetes resources, similar to `kubectl describe`.
- **Pod Logs**: Retrieve logs from specific pods (optionally from a specific container, or all containers if unspecified).
- **Node Metrics**: Get resource usage metrics for specific nodes.
- **Pod Metrics**: Get CPU and Memory metrics for specific pods.
- **Event Listing**: List events within a namespace or for a specific resource.
- **Resource Creation/Updating**: Create new Kubernetes resources or update existing ones from a YAML or JSON manifest.
- **Standardized Interface**: Uses the MCP protocol for consistent tool interaction.
- **Flexible Configuration**: Supports different Kubernetes contexts and resource scopes.
- **Multiple Modes**: Run in `stdio` mode for CLI tools, `sse` mode, or `streamable-http` mode for web applications.
- **Security**: Runs as non-root user in Docker containers for enhanced security.

## Prerequisites

- Go 1.23 or later
- Access to a Kubernetes cluster
- `kubectl` configured with appropriate cluster access

## Installation

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/reza-gholizade/k8s-mcp-server.git
    cd k8s-mcp-server
    ```

2.  **Install dependencies:**
    ```bash
    go mod download
    ```

3.  **Build the server:**
    ```bash
    go build -o k8s-mcp-server main.go
    ```

## Usage

### Starting the Server

The server can run in three modes, configurable via command-line flags or environment variables.

#### Stdio Mode (for CLI integrations)
This mode uses standard input/output for communication.

```bash
./k8s-mcp-server --mode stdio
```
Or using environment variables:
```bash
SERVER_MODE=stdio ./k8s-mcp-server
```

#### SSE Mode (for web applications)
This mode starts an HTTP server with Server-Sent Events support.

Default (port 8080):
```bash
./k8s-mcp-server --mode sse
```
Specify a port:
```bash
./k8s-mcp-server --mode sse --port 9090
```
Or using environment variables:
```bash
SERVER_MODE=sse SERVER_PORT=9090 ./k8s-mcp-server
```
#### Streamable-HTTP Mode (for web applications)
This mode starts an HTTP server with streamable-http transport support, following the MCP specification.

Default (port 8080):
```bash
./k8s-mcp-server --mode streamable-http
```
Specify a port:
```bash
./k8s-mcp-server --mode streamable-http --port 9090
```
Or using environment variables:
```bash
SERVER_MODE=streamable-http SERVER_PORT=9090 ./k8s-mcp-server
```

The server will be available at `http://localhost:8080/mcp` (or your specified port).

If no mode is specified, it defaults to SSE on port 8080.

#### Read-Only Mode

The server supports a read-only mode that disables all write operations, providing a safer way to explore and monitor your Kubernetes cluster without the risk of making changes.

Enable read-only mode with the `--read-only` flag:

```bash
./k8s-mcp-server --read-only
```

You can combine read-only mode with any server mode:

```bash
# Read-only with stdio mode
./k8s-mcp-server --mode stdio --read-only

# Read-only with SSE mode
./k8s-mcp-server --mode sse --read-only

# Read-only with streamable-http mode
./k8s-mcp-server --mode streamable-http --read-only
```

When read-only mode is enabled, the following tools are disabled:
- `createResource` (Kubernetes resource creation/updates)
- `helmInstall` (Helm chart installations)
- `helmUpgrade` (Helm chart upgrades)
- `helmUninstall` (Helm chart uninstallations)
- `helmRollback` (Helm release rollbacks)
- `helmRepoAdd` (Helm repository additions)

All other read-only operations remain available, including listing resources, getting logs, viewing metrics, and inspecting Helm releases.

#### Tool Category Flags
You can selectively disable entire categories of tools using these flags:

**Disable Kubernetes Tools:**
```bash
./k8s-mcp-server --no-k8s
```

**Disable Helm Tools:**
```bash
./k8s-mcp-server --no-helm
```

**Combine with other flags:**
```bash
# Read-only mode with only Kubernetes tools (no Helm)
./k8s-mcp-server --read-only --no-helm

# Read-only mode with only Helm tools (no Kubernetes)
./k8s-mcp-server --read-only --no-k8s

# SSE mode with only Kubernetes tools
./k8s-mcp-server --mode sse --no-helm

```

**Note:** You cannot use both `--no-k8s` and `--no-helm` together, as this would result in no available tools. The server will exit with an error if both flags are provided.

When `--no-k8s` is enabled, all Kubernetes tools are disabled:
- `getAPIResources`, `listResources`, `getResource`, `describeResource`
- `getPodsLogs`, `getNodeMetrics`, `getPodMetrics`, `getEvents`
- `createResource` (if not in read-only mode)

When `--no-helm` is enabled, all Helm tools are disabled:
- `helmList`, `helmGet`, `helmHistory`, `helmRepoList`
- `helmInstall`, `helmUpgrade`, `helmUninstall`, `helmRollback`, `helmRepoAdd` (if not in read-only mode)

### Using the Docker Image

You can also run the server using the pre-built Docker image from Docker Hub.

1.  **Pull the image:**
    ```bash
    docker pull ginnux/k8s-mcp-server:latest
    ```
    You can replace `latest` with a specific version tag (e.g., `1.0.0`).

2.  **Run the container:**

    *   **SSE Mode (default behavior of the image):**
        ```bash
        docker run -p 8080:8080 -v ~/.kube/config:/home/appuser/.kube/config:ro ginnux/k8s-mcp-server:latest
        ```
        This maps port 8080 of the container to port 8080 on your host and mounts your Kubernetes config read-only to the non-root user's home directory. The server will be available at `http://localhost:8080`. The image defaults to `sse` mode on port `8080`.

    *   **Streamable-HTTP Mode:**
        ```bash
        docker run -p 8080:8080 -v ~/.kube/config:/home/appuser/.kube/config:ro ginnux/k8s-mcp-server:latest --mode streamable-http
        ```
        This runs the server in streamable-http mode. The server will be available at `http://localhost:8080/mcp`.

    *   **Stdio Mode:**
        ```bash
        docker run -i --rm -v ~/.kube/config:/home/appuser/.kube/config:ro ginnux/k8s-mcp-server:latest --mode stdio
        ```
        The `-i` flag is important for interactive stdio communication. `--rm` cleans up the container after exit.

    *   **Custom Port for SSE Mode:**
        ```bash
        docker run -p 9090:9090 -v ~/.kube/config:/home/appuser/.kube/config:ro ginnux/k8s-mcp-server:latest --mode sse --port 9090
        ```

    *   **Custom Port for Streamable-HTTP Mode:**
        ```bash
        docker run -p 9090:9090 -v ~/.kube/config:/home/appuser/.kube/config:ro ginnux/k8s-mcp-server:latest --mode streamable-http --port 9090
        ```

    *   **Alternative: Mount entire .kube directory:**
        ```bash
        docker run -p 8080:8080 -v ~/.kube:/home/appuser/.kube:ro ginnux/k8s-mcp-server:latest
        ```

#### Using with Docker Compose

Create a `docker-compose.yml` file:
```yaml
version: '3.8'
services:
  k8s-mcp-server:
    image: ginnux/k8s-mcp-server:latest # Or a specific version
    container_name: k8s-mcp-server
    ports:
      - "8080:8080" # Host:Container, adjust if using a different SERVER_PORT
    volumes:
      - ~/.kube:/home/appuser/.kube:ro # Mount kubeconfig read-only to non-root user home
    environment:
      - KUBECONFIG=/home/appuser/.kube/config
      - SERVER_MODE=sse # Can be 'stdio', 'sse', or 'streamable-http'
      - SERVER_PORT=8080 # Port for SSE/streamable-http modes
    # command: ["--read-only"] # Uncomment this line to enable read-only mode
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
    # To run in stdio mode with docker-compose, you might need to adjust 'ports',
    # add 'stdin_open: true' and 'tty: true', and potentially override the command.
    # For example, to force stdio mode:
    # command: ["--mode", "stdio"]
    # stdin_open: true
    # tty: true
    # For streamable-http mode, simply change SERVER_MODE to 'streamable-http'
```
To enable read-only mode, use the `command` override as shown above.

Then start with:
```bash
docker compose up -d
```
To see logs: `docker compose logs -f k8s-mcp-server`.

#### Security Considerations

The Docker image runs as a non-root user (`appuser` with UID 1001) for enhanced security:
- The application binary is located at `/usr/local/bin/k8s-mcp-server`
- The kubeconfig should be mounted to `/home/appuser/.kube/config`
- Health checks are enabled to monitor container status
- The container includes minimal dependencies (ca-certificates and curl only)

#### Making API Calls (SSE/Streamable-HTTP Mode)
Once the server is running in SSE or streamable-http mode, you can make JSON-RPC calls to its HTTP endpoint:
```bash
curl -X POST -H "Content-Type: application/json" -d '{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "getAPIResources",
  "params": {
    "arguments": {
      "includeNamespaceScoped": true,
      "includeClusterScoped": true
    }
  }
}' http://localhost:8080/
```

You can also check the health status:
```bash
curl -f http://localhost:8080/
```

### Available Tools

#### 1. `getAPIResources`

Retrieves all available API resources in the Kubernetes cluster.

**Parameters:**
- `includeNamespaceScoped` (boolean, optional): Whether to include namespace-scoped resources (defaults to true).
- `includeClusterScoped` (boolean, optional): Whether to include cluster-scoped resources (defaults to true).

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
- `Kind` (string, required): The kind of resource to list (e.g., "Pod", "Deployment").
- `namespace` (string, optional): The namespace to list resources from. If omitted, lists across all namespaces for namespaced resources (subject to RBAC).
- `labelSelector` (string, optional): Filter resources by label selector (e.g., "app=nginx,env=prod").

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
- `kind` (string, required): The kind of resource to get (e.g., "Pod", "Deployment").
- `name` (string, required): The name of the resource to get.
- `namespace` (string, optional): The namespace of the resource (required for namespaced resources).

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

Describes a resource in the Kubernetes cluster, similar to `kubectl describe`.

**Parameters:**
- `Kind` (string, required): The kind of resource to describe (e.g., "Pod", "Deployment").
- `name` (string, required): The name of the resource to describe.
- `namespace` (string, optional): The namespace of the resource (required for namespaced resources).

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

Retrieves the logs of a specific pod.

**Parameters:**
- `Name` (string, required): The name of the pod.
- `namespace` (string, required): The namespace of the pod.
- `containerName` (string, optional): The specific container name within the pod. If omitted:
    - If the pod has one container, its logs are fetched.
    - If the pod has multiple containers, logs from all containers are fetched and concatenated.

**Example:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "getPodsLogs",
  "params": {
    "arguments": {
      "Name": "my-app-pod-12345",
      "namespace": "production",
      "containerName": "main-container"
    }
  }
}
```

#### 6. `getNodeMetrics`

Retrieves resource usage metrics for a specific node.

**Parameters:**
- `Name` (string, required): The name of the node.

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

Retrieves CPU and Memory metrics for a specific pod.

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

Retrieves events for a specific namespace or resource.

**Parameters:**
- `namespace` (string, optional): The namespace to get events from. If omitted, events from all namespaces are considered (subject to RBAC).
- `resourceName` (string, optional): The name of a specific resource (e.g., a Pod name) to filter events for.
- `resourceKind` (string, optional): The kind of the specific resource (e.g., "Pod") if `resourceName` is provided.

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

#### 9. `createOrUpdateResource`

Creates a new resource or updates an existing one from a YAML or JSON manifest.

**Parameters:**
- `manifest` (string, required): The YAML or JSON manifest of the resource.
- `namespace` (string, optional): The namespace in which to create/update the resource. If the manifest contains a namespace, this parameter can be used to override it. If not provided and the manifest doesn't specify one, "default" might be assumed or it might be an error depending on the resource type.

**Example:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "createOrUpdateResource",
  "params": {
    "arguments": {
      "namespace": "default",
      "manifest": "apiVersion: v1\nkind: Pod\nmetadata:\n  name: my-new-pod\nspec:\n  containers:\n  - name: nginx\n    image: nginx:latest"
    }
  }
}
```

### Helm Operations

#### 10. `helmInstall`

Install a Helm chart to the Kubernetes cluster.

**Parameters:**
- `releaseName` (string, required): Name of the Helm release
- `chartName` (string, required): Name or path of the Helm chart
- `namespace` (string, optional): Kubernetes namespace for the release (defaults to "default")
- `repoURL` (string, optional): Helm repository URL
- `values` (object, optional): Values to override in the chart

**Example:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "helmInstall",
  "params": {
    "arguments": {
      "releaseName": "my-nginx",
      "chartName": "bitnami/nginx",
      "namespace": "web",
      "repoURL": "https://charts.bitnami.com/bitnami",
      "values": {
        "replicaCount": 3,
        "service": {
          "type": "LoadBalancer"
        }
      }
    }
  }
}
```

#### 11. `helmUpgrade`

Upgrade an existing Helm release.

#### 12. `helmUninstall`

Uninstall a Helm release from the Kubernetes cluster.

#### 13. `helmList`

List all Helm releases in the cluster or a specific namespace.

#### 14. `helmGet`

Get details of a specific Helm release.

#### 15. `helmHistory`

Get the history of a Helm release.

#### 16. `helmRollback`

Rollback a Helm release to a previous revision.

### Adding New Tools

1.  **Define the Tool**: In `tools/tools.go`, define a function that returns an `mcp.Tool` structure. This includes the tool's name, description, and input/output schemas.
2.  **Implement the Handler**: In `handlers/handlers.go`, create a handler function. This function takes `*k8s.Client` as an argument and returns a function with the signature `func(context.Context, mcp.ToolInput) (mcp.ToolOutput, error)`. This inner function will contain the logic for your tool.
3.  **Register the Tool**: In `main.go`, add your new tool to the MCP server instance using `s.AddTool(tools.YourToolDefinitionFunction(), handlers.YourToolHandlerFunction(client))`.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to contribute to this project.

## License
gholizade.net@gmail.com

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

<a href="https://glama.ai/mcp/servers/@reza-gholizade/k8s-mcp-server">
  <img width="380" height="200" src="https://glama.ai/mcp/servers/@reza-gholizade/k8s-mcp-server/badge" />
</a>

## VS Code Integration

### Quick Setup

#### Automatic Installation (Recommended)

**macOS/Linux:**
```bash
curl -sSL https://raw.githubusercontent.com/reza-gholizade/k8s-mcp-server/main/scripts/install-vscode-config.sh | bash
```

**Windows (PowerShell):**
```powershell
iex ((New-Object System.Net.WebClient).DownloadString('https://raw.githubusercontent.com/reza-gholizade/k8s-mcp-server/main/scripts/install-vscode-config.ps1'))
```

#### Manual Installation

1. **Install the MCP extension in VS Code:**
   ```bash
   code --install-extension modelcontextprotocol.mcp
   ```

2. **Add to your VS Code settings.json:**

   Open VS Code settings (Cmd/Ctrl + ,) → Open Settings JSON → Add:

   **macOS/Linux:**
   ```json
   {
     "mcp.mcpServers": {
       "k8s-mcp-server": {
         "command": "k8s-mcp-server",
         "args": ["--mode", "stdio"],
         "env": {
           "KUBECONFIG": "${env:HOME}/.kube/config"
         }
       }
     }
   }
   ```

   **Read-Only Mode (recommended for safety):**
   ```json
   {
     "mcp.mcpServers": {
       "k8s-mcp-server": {
         "command": "k8s-mcp-server",
         "args": ["--mode", "stdio", "--read-only"],
         "env": {
           "KUBECONFIG": "${env:HOME}/.kube/config"
         }
       }
     }
   }
   ```

   **Kubernetes Tools Only:**
   ```json
   {
     "mcp.mcpServers": {
       "k8s-mcp-server": {
         "command": "k8s-mcp-server",
         "args": ["--mode", "stdio", "--no-helm"],
         "env": {
           "KUBECONFIG": "${env:HOME}/.kube/config"
         }
       }
     }
   }
   ```

   **Helm Tools Only:**
   ```json
   {
     "mcp.mcpServers": {
       "k8s-mcp-server": {
         "command": "k8s-mcp-server",
         "args": ["--mode", "stdio", "--no-k8s"],
         "env": {
           "KUBECONFIG": "${env:HOME}/.kube/config"
         }
       }
     }
   }
   ```

   **Read-Only with Kubernetes Tools Only:**
   ```json
   {
     "mcp.mcpServers": {
       "k8s-mcp-server": {
         "command": "k8s-mcp-server",
         "args": ["--mode", "stdio", "--read-only", "--no-helm"],
         "env": {
           "KUBECONFIG": "${env:HOME}/.kube/config"
         }
       }
     }
   }
   ```

   **Windows:**
   ```json
   {
     "mcp.mcpServers": {
       "k8s-mcp-server": {
         "command": "k8s-mcp-server.exe",
         "args": ["--mode", "stdio"],
         "env": {
           "KUBECONFIG": "${env:USERPROFILE}/.kube/config"
         }
       }
     }
   }
   ```

3. **Ensure the binary is in your PATH:**

   Download the appropriate binary from the [releases page](https://github.com/reza-gholizade/k8s-mcp-server/releases) and add it to your system PATH.

4. **Restart VS Code**

### Usage in VS Code

Once configured, you can use the Kubernetes MCP server in VS Code with Claude or other MCP-compatible tools:

1. Open VS Code
2. Access Claude (or other MCP-enabled AI assistant)
3. Use natural language to interact with your Kubernetes cluster:
   - "List all pods in the default namespace"
   - "Show me the logs for pod nginx-123"
   - "Get the CPU usage for worker-node-1"
   - "Describe the deployment called my-app"

### Configuration Options

You can customize the configuration by modifying the settings:

```json
{
  "mcp.mcpServers": {
    "k8s-mcp-server": {
      "command": "k8s-mcp-server",
      "args": ["--mode", "stdio"],
      "env": {
        "KUBECONFIG": "/path/to/your/kubeconfig",
        "KUBERNETES_CONTEXT": "your-context-name"
      }
    }
  }
}
```

### Troubleshooting

- **Binary not found**: Ensure `k8s-mcp-server` is in your PATH
- **Kubernetes connection issues**: Verify your `KUBECONFIG` path is correct
- **Permission errors**: Ensure your kubeconfig has the necessary RBAC permissions
- **Extension not loading**: Restart VS Code after configuration changes
