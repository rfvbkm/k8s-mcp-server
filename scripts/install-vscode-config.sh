#!/bin/bash
# filepath: /Users/rezagholizade/Documents/k8s-mcp-server/scripts/install-vscode-config.sh

echo "Installing Kubernetes MCP Server configuration for VS Code..."

# Check if VS Code is installed
if ! command -v code &> /dev/null; then
    echo "❌ VS Code is not installed or not in PATH"
    echo "Please install VS Code first: https://code.visualstudio.com/"
    exit 1
fi

# Install MCP extension
echo "📦 Installing MCP extension..."
code --install-extension modelcontextprotocol.mcp

# Get VS Code settings directory
if [[ "$OSTYPE" == "darwin"* ]]; then
    VSCODE_SETTINGS_DIR="$HOME/Library/Application Support/Code/User"
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    VSCODE_SETTINGS_DIR="$HOME/.config/Code/User"
elif [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
    VSCODE_SETTINGS_DIR="$APPDATA/Code/User"
else
    echo "❌ Unsupported operating system"
    exit 1
fi

# Create settings directory if it doesn't exist
mkdir -p "$VSCODE_SETTINGS_DIR"

# Check if settings.json exists
SETTINGS_FILE="$VSCODE_SETTINGS_DIR/settings.json"

if [ ! -f "$SETTINGS_FILE" ]; then
    echo "📝 Creating new settings.json..."
    cat > "$SETTINGS_FILE" << 'EOF'
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
EOF
else
    echo "⚠️  settings.json already exists. Please manually add the following configuration:"
    echo ""
    echo "Add this to your VS Code settings.json:"
    cat << 'EOF'
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
EOF
fi

echo "✅ Configuration complete!"
echo ""
echo "Next steps:"
echo "1. Make sure k8s-mcp-server is in your PATH"
echo "2. Restart VS Code"
echo "3. The Kubernetes MCP server will be available in Claude/MCP-enabled tools"