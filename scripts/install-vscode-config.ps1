Write-Host "Installing Kubernetes MCP Server configuration for VS Code..." -ForegroundColor Green

# Check if VS Code is installed
if (!(Get-Command code -ErrorAction SilentlyContinue)) {
    Write-Host "❌ VS Code is not installed or not in PATH" -ForegroundColor Red
    Write-Host "Please install VS Code first: https://code.visualstudio.com/" -ForegroundColor Yellow
    exit 1
}

# Install MCP extension
Write-Host "📦 Installing MCP extension..." -ForegroundColor Blue
& code --install-extension modelcontextprotocol.mcp

# Get VS Code settings directory
$vsCodeSettingsDir = "$env:APPDATA\Code\User"

# Create settings directory if it doesn't exist
if (!(Test-Path $vsCodeSettingsDir)) {
    New-Item -ItemType Directory -Path $vsCodeSettingsDir -Force
}

# Check if settings.json exists
$settingsFile = Join-Path $vsCodeSettingsDir "settings.json"

$mcpConfig = @'
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
'@

if (!(Test-Path $settingsFile)) {
    Write-Host "📝 Creating new settings.json..." -ForegroundColor Blue
    $mcpConfig | Out-File -FilePath $settingsFile -Encoding UTF8
} else {
    Write-Host "⚠️  settings.json already exists. Please manually add the following configuration:" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Add this to your VS Code settings.json:" -ForegroundColor Cyan
    Write-Host $mcpConfig -ForegroundColor White
}

Write-Host "✅ Configuration complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. Make sure k8s-mcp-server.exe is in your PATH" -ForegroundColor White
Write-Host "2. Restart VS Code" -ForegroundColor White
Write-Host "3. The Kubernetes MCP server will be available in Claude/MCP-enabled tools" -ForegroundColor White