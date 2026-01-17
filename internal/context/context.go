package context

import (
	"fmt"
	"sort"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Manager handles context operations
type Manager struct {
	config       *api.Config
	loadingRules *clientcmd.ClientConfigLoadingRules
}

// NewManager creates a new context manager
func NewManager() (*Manager, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	rawConfig, err := kubeConfig.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	return &Manager{
		config:       &rawConfig,
		loadingRules: loadingRules,
	}, nil
}

// GetCurrentContext returns the current context name
func (m *Manager) GetCurrentContext() string {
	return m.config.CurrentContext
}

// ListContexts returns a sorted list of all available contexts
func (m *Manager) ListContexts() []string {
	contexts := make([]string, 0, len(m.config.Contexts))
	for name := range m.config.Contexts {
		contexts = append(contexts, name)
	}
	sort.Strings(contexts)
	return contexts
}

// ValidateContext checks if a context exists
func (m *Manager) ValidateContext(name string) error {
	if _, exists := m.config.Contexts[name]; !exists {
		return fmt.Errorf("context %q not found", name)
	}
	return nil
}

// SwitchContext switches to the specified context
func (m *Manager) SwitchContext(targetContext string) error {
	if err := m.ValidateContext(targetContext); err != nil {
		return err
	}

	if targetContext == m.config.CurrentContext {
		return nil // Already on target context
	}

	m.config.CurrentContext = targetContext

	if err := clientcmd.ModifyConfig(m.loadingRules, *m.config, false); err != nil {
		return fmt.Errorf("failed to switch context: %w", err)
	}

	return nil
}
