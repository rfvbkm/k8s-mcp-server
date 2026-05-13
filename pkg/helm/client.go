package helm

import (
	"context"
	"fmt"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/reza-gholizade/k8s-mcp-server/pkg/k8s"
)

// Client wraps Helm operations. A root Client (returned by NewClient)
// owns a kubeconfig source and a cache of per-context sub-clients
// produced via ForContext.
type Client struct {
	settings         *cli.EnvSettings
	restConfig       *rest.Config
	k8sClient        kubernetes.Interface
	restClientGetter genericclioptions.RESTClientGetter

	// contextName is the kubeconfig context this client is bound to.
	// Empty means current-context (or N/A for single-endpoint auth).
	contextName string

	// source holds the kubeconfig source shared between root and
	// sub-clients. Set only via NewClient.
	source *k8s.ConfigSource

	// root back-points to the owning root client; nil on the root.
	root *Client

	// subClients caches per-context Clients. Populated on the root only.
	subClients map[string]*Client
	subLock    sync.RWMutex
}

// customRESTClientGetter is a custom RESTClientGetter that uses a pre-built rest.Config
// instead of reading from kubeconfig files. This ensures Helm uses the same authentication
// method that was used to build the restConfig (KUBECONFIG_DATA, KUBERNETES_SERVER/TOKEN, etc.)
type customRESTClientGetter struct {
	restConfig *rest.Config
}

// ToRESTConfig returns the pre-built REST config
func (g *customRESTClientGetter) ToRESTConfig() (*rest.Config, error) {
	return g.restConfig, nil
}

// ToRawKubeConfigLoader returns a clientcmd.ClientConfig that uses the pre-built config
func (g *customRESTClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return &customClientConfig{restConfig: g.restConfig}
}

// ToDiscoveryClient returns a discovery client using the pre-built REST config
func (g *customRESTClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(g.restConfig)
	if err != nil {
		return nil, err
	}
	return memory.NewMemCacheClient(discoveryClient), nil
}

// ToRESTMapper returns a REST mapper using the discovery client
func (g *customRESTClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	discoveryClient, err := g.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	expander := restmapper.NewShortcutExpander(mapper, discoveryClient, nil)
	return expander, nil
}

// customClientConfig implements clientcmd.ClientConfig interface
type customClientConfig struct {
	restConfig *rest.Config
}

// RawConfig returns an empty api.Config since we're using a direct rest.Config
func (c *customClientConfig) RawConfig() (api.Config, error) {
	return api.Config{}, fmt.Errorf("raw config not available when using direct REST config")
}

// ClientConfig returns the pre-built REST config
func (c *customClientConfig) ClientConfig() (*rest.Config, error) {
	return c.restConfig, nil
}

// Namespace returns the default namespace from the config
func (c *customClientConfig) Namespace() (string, bool, error) {
	return "default", false, nil
}

// ConfigAccess returns nil as we don't use file-based config access
func (c *customClientConfig) ConfigAccess() clientcmd.ConfigAccess {
	return nil
}

// NewClient creates a new Helm client bound to the kubeconfig's
// current-context. Per-request context switching is available through
// (*Client).ForContext.
//
// It uses the same authentication methods as the Kubernetes client:
//  1. Kubeconfig content from KUBECONFIG_DATA environment variable
//  2. API server URL and token from KUBERNETES_SERVER and KUBERNETES_TOKEN environment variables
//  3. In-cluster authentication (service account token)
//  4. Kubeconfig file path (provided or default ~/.kube/config)
func NewClient(kubeconfig string) (*Client, error) {
	source, err := k8s.LoadKubeconfigSource(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig source: %w", err)
	}

	restConfig, err := source.RESTConfigFor("")
	if err != nil {
		return nil, fmt.Errorf("failed to build REST config: %w", err)
	}

	settingsKubeConfig := kubeconfig
	if settingsKubeConfig == "" {
		if envPath := os.Getenv("KUBECONFIG"); envPath != "" {
			settingsKubeConfig = envPath
		} else if srcPath := source.KubeconfigPath(); srcPath != "" {
			settingsKubeConfig = srcPath
		}
	}

	client, err := buildHelmClient(restConfig, settingsKubeConfig)
	if err != nil {
		return nil, err
	}

	client.source = source
	client.subClients = make(map[string]*Client)
	return client, nil
}

// buildHelmClient creates the low-level Helm pieces (settings,
// rest client getter, kubernetes clientset) for a given rest.Config.
// It does not wire any multi-context machinery.
func buildHelmClient(restConfig *rest.Config, settingsKubeConfig string) (*Client, error) {
	settings := cli.New()

	// Use a custom RESTClientGetter that returns our pre-built restConfig.
	// This ensures Helm uses the same authentication method
	// (KUBECONFIG_DATA, KUBERNETES_SERVER/TOKEN, in-cluster, etc.)
	// instead of trying to read from settings.KubeConfig.
	restClientGetter := &customRESTClientGetter{restConfig: restConfig}

	if settingsKubeConfig != "" {
		settings.KubeConfig = settingsKubeConfig
	}

	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return &Client{
		settings:         settings,
		restConfig:       restConfig,
		k8sClient:        k8sClient,
		restClientGetter: restClientGetter,
	}, nil
}

// ForContext returns a Helm Client bound to the requested kubeconfig
// context. An empty name returns the root client. Sub-clients are
// cached on the root for reuse.
func (c *Client) ForContext(name string) (*Client, error) {
	root := c.rootClient()
	if name == "" {
		return root, nil
	}

	root.subLock.RLock()
	if sub, ok := root.subClients[name]; ok {
		root.subLock.RUnlock()
		return sub, nil
	}
	root.subLock.RUnlock()

	if root.source == nil {
		return nil, fmt.Errorf("kubeconfig source is not initialized; cannot switch context")
	}

	cfg, err := root.source.RESTConfigFor(name)
	if err != nil {
		return nil, err
	}

	settingsKubeConfig := ""
	if root.settings != nil {
		settingsKubeConfig = root.settings.KubeConfig
	}

	sub, err := buildHelmClient(cfg, settingsKubeConfig)
	if err != nil {
		return nil, err
	}
	sub.contextName = name
	sub.source = root.source
	sub.root = root

	root.subLock.Lock()
	defer root.subLock.Unlock()
	if existing, ok := root.subClients[name]; ok {
		return existing, nil
	}
	root.subClients[name] = sub
	return sub, nil
}

// ListContexts returns the sorted list of available kubeconfig contexts
// and the name of the current-context.
func (c *Client) ListContexts() ([]string, string, error) {
	root := c.rootClient()
	if root.source == nil {
		return nil, "", fmt.Errorf("kubeconfig source is not initialized")
	}
	return root.source.Contexts()
}

// ContextName returns the kubeconfig context this client is bound to.
func (c *Client) ContextName() string { return c.contextName }

// rootClient returns the root client that owns the sub-client cache.
func (c *Client) rootClient() *Client {
	if c.root != nil {
		return c.root
	}
	return c
}

func (c *Client) InstallChart(ctx context.Context, namespace, releaseName, chartName, repoURL string, values map[string]interface{}) (*release.Release, error) {
	actionConfig := &action.Configuration{}
	if err := actionConfig.Init(c.restClientGetter, namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		return nil, fmt.Errorf("failed to initialize action config: %w", err)
	}

	client := action.NewInstall(actionConfig)
	client.Namespace = namespace
	client.ReleaseName = releaseName
	client.CreateNamespace = true
	cln, err := registry.NewClient(
		registry.ClientOptDebug(true),
		registry.ClientOptCredentialsFile(""),
		registry.ClientOptEnableCache(false))

	if err != nil {
		return nil, fmt.Errorf("failed to initialize registry: %w", err)
	}
	fmt.Println("Registry client created successfully:", cln)

	if values == nil {
		values = make(map[string]interface{})
	}

	// If repoURL is provided, add it to settings or append to chartName accordingly
	if repoURL != "" {
		client.RepoURL = repoURL
	}

	// Locate the chart (resolves repo/chart or OCI)
	chartPath, err := client.LocateChart(chartName, c.settings)
	if err != nil {
		return nil, fmt.Errorf("failed to locate chart: %w", err)
	}

	// Load the chart from the resolved path (can be a URL or OCI reference)
	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	// Run the install action
	release, err := client.Run(chart, values)
	if err != nil {
		return nil, fmt.Errorf("failed to install chart: %w", err)
	}

	return release, nil
}

func (c *Client) UpgradeChart(ctx context.Context, namespace, releaseName, chartName string, values map[string]interface{}) (*release.Release, error) {
	actionConfig := &action.Configuration{}
	if err := actionConfig.Init(c.restClientGetter, namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		return nil, fmt.Errorf("failed to initialize action config: %w", err)
	}

	// Create and assign registry client
	regClient, err := registry.NewClient(
		registry.ClientOptDebug(true),
		registry.ClientOptEnableCache(false),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize registry client: %w", err)
	}
	fmt.Println("Registry client created successfully:", regClient)

	client := action.NewUpgrade(actionConfig)
	client.Namespace = namespace

	if values == nil {
		values = make(map[string]interface{})
	}

	// Locate the chart (for both OCI and regular charts)
	chartPath, err := client.LocateChart(chartName, c.settings)
	if err != nil {
		return nil, fmt.Errorf("failed to locate chart: %w", err)
	}

	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	release, err := client.Run(releaseName, chart, values)
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade chart: %w", err)
	}

	return release, nil
}

// UninstallChart uninstalls a Helm release
func (c *Client) UninstallChart(ctx context.Context, namespace, releaseName string) error {
	actionConfig := &action.Configuration{}
	if err := actionConfig.Init(c.restClientGetter, namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
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
	if err := actionConfig.Init(c.restClientGetter, namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		return nil, fmt.Errorf("failed to initialize action config: %w", err)
	}

	client := action.NewList(actionConfig)
	client.AllNamespaces = namespace == ""

	releases, err := client.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to list releases: %w", err)
	}
	// remove useless fields from releases
	for _, release := range releases {
		release.Chart.Templates = nil
		release.Chart.Files = nil
		release.Chart.Values = nil
		release.Chart.Schema = nil
		release.Config = nil
		release.Manifest = ""
		release.Chart.Lock = nil
		release.Hooks = nil
	}

	return releases, nil
}

func (c *Client) GetRelease(ctx context.Context, namespace, releaseName string) (*release.Release, error) {
	actionConfig := &action.Configuration{}
	if err := actionConfig.Init(c.restClientGetter, namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
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
	if err := actionConfig.Init(c.restClientGetter, namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
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
	if err := actionConfig.Init(c.restClientGetter, namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
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
func (c *Client) HelmRepoAdd(ctx context.Context, name, url string) error {
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

func (c *Client) HelmRepoList(ctx context.Context) ([]*repo.Entry, error) {
	repoFile := c.settings.RepositoryConfig
	f, err := repo.LoadFile(repoFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load repository file: %w", err)
	}
	return f.Repositories, nil
}
