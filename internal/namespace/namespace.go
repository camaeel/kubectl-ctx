package namespace

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

const DefaultNamespace = "default"

// Manager handles namespace operations
type Manager struct {
	config         *api.Config
	loadingRules   *clientcmd.ClientConfigLoadingRules
	currentContext string
}

// NewManager creates a new namespace manager
func NewManager() (*Manager, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	rawConfig, err := kubeConfig.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	currentContext := rawConfig.CurrentContext
	if currentContext == "" {
		return nil, fmt.Errorf("no current context set")
	}

	if _, exists := rawConfig.Contexts[currentContext]; !exists {
		return nil, fmt.Errorf("current context %q not found in config", currentContext)
	}

	return &Manager{
		config:         &rawConfig,
		loadingRules:   loadingRules,
		currentContext: currentContext,
	}, nil
}

// GetCurrentNamespace returns the current namespace for the current context
func (m *Manager) GetCurrentNamespace() string {
	ctx := m.config.Contexts[m.currentContext]
	if ctx.Namespace == "" {
		return DefaultNamespace
	}
	return ctx.Namespace
}

// GetCurrentContext returns the current context name
func (m *Manager) GetCurrentContext() string {
	return m.currentContext
}

// ListNamespacesFromCluster fetches namespaces from the cluster
func (m *Manager) ListNamespacesFromCluster() ([]string, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	restConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	namespaceList, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	namespaces := make([]string, 0, len(namespaceList.Items))
	for _, ns := range namespaceList.Items {
		namespaces = append(namespaces, ns.Name)
	}

	return namespaces, nil
}

// SwitchNamespace switches to the specified namespace
func (m *Manager) SwitchNamespace(targetNamespace string) error {
	if targetNamespace == m.GetCurrentNamespace() {
		return nil // Already on target namespace
	}

	// Update namespace in the current context
	ctx := m.config.Contexts[m.currentContext]
	ctx.Namespace = targetNamespace
	m.config.Contexts[m.currentContext] = ctx

	// Write back the configuration
	if err := clientcmd.ModifyConfig(m.loadingRules, *m.config, false); err != nil {
		return fmt.Errorf("failed to switch namespace: %w", err)
	}

	return nil
}
