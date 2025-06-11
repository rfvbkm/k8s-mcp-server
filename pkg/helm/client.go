package helm

import (
    "context"
    "fmt"
    "log"
    "os"
    "path/filepath"

    "helm.sh/helm/v3/pkg/action"
    "helm.sh/helm/v3/pkg/chart/loader"
    "helm.sh/helm/v3/pkg/cli"
    "helm.sh/helm/v3/pkg/getter"
    "helm.sh/helm/v3/pkg/repo"
    "helm.sh/helm/v3/pkg/release"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
)

// Client wraps Helm operations
type Client struct {
    settings   *cli.EnvSettings
    restConfig *rest.Config
    k8sClient  kubernetes.Interface
}

// NewClient creates a new Helm client
func NewClient(kubeconfig string) (*Client, error) {
    settings := cli.New()
    
    if kubeconfig != "" {
        settings.KubeConfig = kubeconfig
    }

    // Get Kubernetes REST config
    var restConfig *rest.Config
    var err error
    
    if settings.KubeConfig != "" {
        restConfig, err = clientcmd.BuildConfigFromFlags("", settings.KubeConfig)
    } else {
        // Try to use the default kubeconfig path first
        if home := os.Getenv("HOME"); home != "" {
            defaultKubeconfig := filepath.Join(home, ".kube", "config")
            if _, err := os.Stat(defaultKubeconfig); err == nil {
                restConfig, err = clientcmd.BuildConfigFromFlags("", defaultKubeconfig)
            } else {
                restConfig, err = rest.InClusterConfig()
            }
        } else {
            restConfig, err = rest.InClusterConfig()
        }
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get Kubernetes config: %w", err)
    }

    // Create Kubernetes client
    k8sClient, err := kubernetes.NewForConfig(restConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
    }

    return &Client{
        settings:   settings,
        restConfig: restConfig,
        k8sClient:  k8sClient,
    }, nil
}

func (c *Client) InstallChart(ctx context.Context, namespace, releaseName, chartName, repoURL string, values map[string]interface{}) (*release.Release, error) {
    actionConfig := &action.Configuration{}
    if err := actionConfig.Init(c.settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
        return nil, fmt.Errorf("failed to initialize action config: %w", err)
    }

    client := action.NewInstall(actionConfig)
    client.Namespace = namespace
    client.ReleaseName = releaseName
    client.CreateNamespace = true

    // Add repository if specified
    if repoURL != "" {
        if err := c.addRepo(chartName, repoURL); err != nil {
            return nil, fmt.Errorf("failed to add repository: %w", err)
        }
    }

    // Load the chart
    chart, err := loader.Load(chartName)
    if err != nil {
        return nil, fmt.Errorf("failed to load chart: %w", err)
    }

    // Install the chart
    release, err := client.Run(chart, values)
    if err != nil {
        return nil, fmt.Errorf("failed to install chart: %w", err)
    }

    return release, nil
}

func (c *Client) UpgradeChart(ctx context.Context, namespace, releaseName, chartName string, values map[string]interface{}) (*release.Release, error) {
    actionConfig := &action.Configuration{}
    if err := actionConfig.Init(c.settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
        return nil, fmt.Errorf("failed to initialize action config: %w", err)
    }

    client := action.NewUpgrade(actionConfig)
    client.Namespace = namespace

    // Load the chart
    chart, err := loader.Load(chartName)
    if err != nil {
        return nil, fmt.Errorf("failed to load chart: %w", err)
    }

    // Upgrade the chart
    release, err := client.Run(releaseName, chart, values)
    if err != nil {
        return nil, fmt.Errorf("failed to upgrade chart: %w", err)
    }

    return release, nil
}

// UninstallChart uninstalls a Helm release
func (c *Client) UninstallChart(ctx context.Context, namespace, releaseName string) error {
    actionConfig := &action.Configuration{}
    if err := actionConfig.Init(c.settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
        return fmt.Errorf("failed to initialize action config: %w", err)
    }

    client := action.NewUninstall(actionConfig)
    _, err := client.Run(releaseName)
    if err != nil {
        return fmt.Errorf("failed to uninstall release: %w", err)
    }

    return nil
}

func (c *Client) ListReleases(ctx context.Context, namespace string) ([]*release.Release, error) {
    actionConfig := &action.Configuration{}
    if err := actionConfig.Init(c.settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
        return nil, fmt.Errorf("failed to initialize action config: %w", err)
    }

    client := action.NewList(actionConfig)
    client.AllNamespaces = namespace == ""
    
    releases, err := client.Run()
    if err != nil {
        return nil, fmt.Errorf("failed to list releases: %w", err)
    }

    return releases, nil
}

func (c *Client) GetRelease(ctx context.Context, namespace, releaseName string) (*release.Release, error) {
    actionConfig := &action.Configuration{}
    if err := actionConfig.Init(c.settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
        return nil, fmt.Errorf("failed to initialize action config: %w", err)
    }

    client := action.NewGet(actionConfig)
    release, err := client.Run(releaseName)
    if err != nil {
        return nil, fmt.Errorf("failed to get release: %w", err)
    }

    return release, nil
}

func (c *Client) GetReleaseHistory(ctx context.Context, namespace, releaseName string) ([]*release.Release, error) {
    actionConfig := &action.Configuration{}
    if err := actionConfig.Init(c.settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
        return nil, fmt.Errorf("failed to initialize action config: %w", err)
    }

    client := action.NewHistory(actionConfig)
    releases, err := client.Run(releaseName)
    if err != nil {
        return nil, fmt.Errorf("failed to get release history: %w", err)
    }

    return releases, nil
}

// RollbackRelease rolls back a Helm release
func (c *Client) RollbackRelease(ctx context.Context, namespace, releaseName string, revision int) error {
    actionConfig := &action.Configuration{}
    if err := actionConfig.Init(c.settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
        return fmt.Errorf("failed to initialize action config: %w", err)
    }

    client := action.NewRollback(actionConfig)
    client.Version = revision
    
    if err := client.Run(releaseName); err != nil {
        return fmt.Errorf("failed to rollback release: %w", err)
    }

    return nil
}

// addRepo adds a Helm repository
func (c *Client) addRepo(name, url string) error {
    repoFile := c.settings.RepositoryConfig

    // Ensure the file directory exists
    if err := os.MkdirAll(filepath.Dir(repoFile), 0755); err != nil {
        return err
    }

    // Load existing repositories
    f, err := repo.LoadFile(repoFile)
    if err != nil && !os.IsNotExist(err) {
        return err
    }
    if f == nil {
        f = repo.NewFile()
    }

    // Check if repo already exists
    if f.Has(name) {
        return nil // Already exists
    }

    // Add the repository
    entry := &repo.Entry{
        Name: name,
        URL:  url,
    }

    r, err := repo.NewChartRepository(entry, getter.All(c.settings))
    if err != nil {
        return err
    }

    if _, err := r.DownloadIndexFile(); err != nil {
        return fmt.Errorf("failed to download repository index: %w", err)
    }

    f.Update(entry)
    return f.WriteFile(repoFile, 0644)
}