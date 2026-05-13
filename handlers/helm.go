package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/reza-gholizade/k8s-mcp-server/pkg/helm"
)

// resolveHelmClient picks the helm sub-client for the requested
// kubeconfig context, falling back to KUBERNETES_CONTEXT and finally
// to the kubeconfig's current-context.
func resolveHelmClient(client *helm.Client, args map[string]interface{}) (*helm.Client, error) {
	target, err := client.ForContext(resolveContextName(args))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve kubeconfig context: %w", err)
	}
	return target, nil
}

// HelmInstall returns a handler function for the helmInstall tool

func HelmInstall(client *helm.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments type: expected map[string]interface{}")
		}

		releaseName, err := getRequiredStringArg(args, "releaseName")
		if err != nil {
			return nil, err
		}

		chartName, err := getRequiredStringArg(args, "chartName")
		if err != nil {
			return nil, err
		}

		namespace := getStringArg(args, "namespace", "default")
		repoURL := getStringArg(args, "repoURL", "")

		values := make(map[string]interface{})
		if v, exists := args["values"]; exists {
			if valuesMap, ok := v.(map[string]interface{}); ok {
				values = valuesMap
			}
		}

		target, err := resolveHelmClient(client, args)
		if err != nil {
			return nil, err
		}

		release, err := target.InstallChart(ctx, namespace, releaseName, chartName, repoURL, values)
		if err != nil {
			return nil, fmt.Errorf("failed to install chart: %w", err)
		}

		jsonResponse, err := json.Marshal(release)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HelmUpgrade returns a handler function for the helmUpgrade tool
func HelmUpgrade(client *helm.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments type: expected map[string]interface{}")
		}

		releaseName, err := getRequiredStringArg(args, "releaseName")
		if err != nil {
			return nil, err
		}

		chartName, err := getRequiredStringArg(args, "chartName")
		if err != nil {
			return nil, err
		}

		namespace := getStringArg(args, "namespace", "default")

		values := make(map[string]interface{})
		if v, exists := args["values"]; exists {
			if valuesMap, ok := v.(map[string]interface{}); ok {
				values = valuesMap
			}
		}

		target, err := resolveHelmClient(client, args)
		if err != nil {
			return nil, err
		}

		release, err := target.UpgradeChart(ctx, namespace, releaseName, chartName, values)
		if err != nil {
			return nil, fmt.Errorf("failed to upgrade chart: %w", err)
		}

		jsonResponse, err := json.Marshal(release)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HelmUninstall returns a handler function for the helmUninstall tool
func HelmUninstall(client *helm.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments type: expected map[string]interface{}")
		}

		releaseName, err := getRequiredStringArg(args, "releaseName")
		if err != nil {
			return nil, err
		}

		namespace := getStringArg(args, "namespace", "default")

		target, err := resolveHelmClient(client, args)
		if err != nil {
			return nil, err
		}

		err = target.UninstallChart(ctx, namespace, releaseName)
		if err != nil {
			return nil, fmt.Errorf("failed to uninstall chart: %w", err)
		}

		response := map[string]string{
			"status":  "success",
			"message": fmt.Sprintf("Successfully uninstalled release '%s' from namespace '%s'", releaseName, namespace),
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HelmList returns a handler function for the helmList tool
func HelmList(client *helm.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments type: expected map[string]interface{}")
		}

		namespace := getStringArg(args, "namespace", "")

		target, err := resolveHelmClient(client, args)
		if err != nil {
			return nil, err
		}

		releases, err := target.ListReleases(ctx, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to list releases: %w", err)
		}

		jsonResponse, err := json.Marshal(releases)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HelmGet returns a handler function for the helmGet tool
func HelmGet(client *helm.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments type: expected map[string]interface{}")
		}

		releaseName, err := getRequiredStringArg(args, "releaseName")
		if err != nil {
			return nil, err
		}

		namespace := getStringArg(args, "namespace", "default")

		target, err := resolveHelmClient(client, args)
		if err != nil {
			return nil, err
		}

		release, err := target.GetRelease(ctx, namespace, releaseName)
		if err != nil {
			return nil, fmt.Errorf("failed to get release: %w", err)
		}

		jsonResponse, err := json.Marshal(release)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HelmHistory returns a handler function for the helmHistory tool
func HelmHistory(client *helm.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments type: expected map[string]interface{}")
		}

		releaseName, err := getRequiredStringArg(args, "releaseName")
		if err != nil {
			return nil, err
		}

		namespace := getStringArg(args, "namespace", "default")

		target, err := resolveHelmClient(client, args)
		if err != nil {
			return nil, err
		}

		history, err := target.GetReleaseHistory(ctx, namespace, releaseName)
		if err != nil {
			return nil, fmt.Errorf("failed to get release history: %w", err)
		}

		jsonResponse, err := json.Marshal(history)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HelmRollback returns a handler function for the helmRollback tool
func HelmRollback(client *helm.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments type: expected map[string]interface{}")
		}

		releaseName, err := getRequiredStringArg(args, "releaseName")
		if err != nil {
			return nil, err
		}

		namespace := getStringArg(args, "namespace", "default")

		revision := 0
		if revStr := getStringArg(args, "revision", "0"); revStr != "0" {
			if rev, err := strconv.Atoi(revStr); err == nil {
				revision = rev
			}
		}

		target, err := resolveHelmClient(client, args)
		if err != nil {
			return nil, err
		}

		err = target.RollbackRelease(ctx, namespace, releaseName, revision)
		if err != nil {
			return nil, fmt.Errorf("failed to rollback release: %w", err)
		}

		response := map[string]interface{}{
			"status":   "success",
			"message":  fmt.Sprintf("Successfully rolled back release '%s' in namespace '%s'", releaseName, namespace),
			"revision": revision,
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

// HelmRepoAdd returns a handler function for the helmRepoAdd tool
func HelmRepoAdd(client *helm.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments type: expected map[string]interface{}")
		}

		repoName, err := getRequiredStringArg(args, "repoName")
		if err != nil {
			return nil, err
		}

		repoURL, err := getRequiredStringArg(args, "repoURL")
		if err != nil {
			return nil, err
		}

		target, err := resolveHelmClient(client, args)
		if err != nil {
			return nil, err
		}

		err = target.HelmRepoAdd(ctx, repoName, repoURL)
		if err != nil {
			return nil, fmt.Errorf("failed to add repository: %w", err)
		}

		response := map[string]interface{}{
			"status":  "success",
			"message": fmt.Sprintf("Successfully added repository '%s' with URL '%s'", repoName, repoURL),
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}

func HelmRepoList(client *helm.Client) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			args = map[string]interface{}{}
		}

		target, err := resolveHelmClient(client, args)
		if err != nil {
			return nil, err
		}

		repos, err := target.HelmRepoList(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories: %w", err)
		}

		jsonResponse, err := json.Marshal(repos)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize response: %w", err)
		}

		return mcp.NewToolResultText(string(jsonResponse)), nil
	}
}
