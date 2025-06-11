package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// HelmInstallTool returns the MCP tool definition for installing Helm charts
func HelmInstallTool() mcp.Tool {
	return mcp.NewTool("helmInstall",
		mcp.WithDescription("Install a Helm chart to the Kubernetes cluster"),
		mcp.WithString("releaseName", mcp.Required(), mcp.Description("Name of the Helm release")),
		mcp.WithString("chartName", mcp.Required(), mcp.Description("Name or path of the Helm chart")),
		mcp.WithString("namespace", mcp.Description("Kubernetes namespace for the release")),
		mcp.WithString("repoURL", mcp.Description("Helm repository URL (optional)")),
		mcp.WithObject("values", mcp.Description("Values to override in the chart")),
	)
}

// HelmUpgradeTool returns the MCP tool definition for upgrading Helm releases
func HelmUpgradeTool() mcp.Tool {
	return mcp.NewTool("helmUpgrade",
		mcp.WithDescription("Upgrade an existing Helm release"),
		mcp.WithString("releaseName", mcp.Required(), mcp.Description("Name of the Helm release to upgrade")),
		mcp.WithString("chartName", mcp.Required(), mcp.Description("Name or path of the Helm chart")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Kubernetes namespace of the release")),
		mcp.WithObject("values",mcp.Required(), mcp.Description("Values to override in the chart")),
	)
}

// HelmUninstallTool returns the MCP tool definition for uninstalling Helm releases
func HelmUninstallTool() mcp.Tool {
	return mcp.NewTool("helmUninstall",
		mcp.WithDescription("Uninstall a Helm release from the Kubernetes cluster"),
		mcp.WithString("releaseName", mcp.Required(), mcp.Description("Name of the Helm release to uninstall")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Kubernetes namespace of the release")),
	)
}

// HelmListTool returns the MCP tool definition for listing Helm releases
func HelmListTool() mcp.Tool {
	return mcp.NewTool("helmList",
		mcp.WithDescription("List all Helm releases in the cluster or a specific namespace"),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Kubernetes namespace to list releases from (empty for all namespaces)")),
	)
}

// HelmGetTool returns the MCP tool definition for getting Helm release details
func HelmGetTool() mcp.Tool {
	return mcp.NewTool("helmGet",
		mcp.WithDescription("Get details of a specific Helm release"),
		mcp.WithString("releaseName", mcp.Required(), mcp.Description("Name of the Helm release")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Kubernetes namespace of the release")),
	)
}

// HelmHistoryTool returns the MCP tool definition for getting Helm release history
func HelmHistoryTool() mcp.Tool {
	return mcp.NewTool("helmHistory",
		mcp.WithDescription("Get the history of a Helm release"),
		mcp.WithString("releaseName", mcp.Required(), mcp.Description("Name of the Helm release")),
		mcp.WithString("namespace",mcp.Required(), mcp.Description("Kubernetes namespace of the release")),
	)
}

// HelmRollbackTool returns the MCP tool definition for rolling back Helm releases
func HelmRollbackTool() mcp.Tool {
	return mcp.NewTool("helmRollback",
		mcp.WithDescription("Rollback a Helm release to a previous revision"),
		mcp.WithString("releaseName", mcp.Required(), mcp.Description("Name of the Helm release to rollback")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Kubernetes namespace of the release")),
		mcp.WithNumber("revision",mcp.Required(), mcp.Description("Revision number to rollback to (0 for previous)")),
	)
}
