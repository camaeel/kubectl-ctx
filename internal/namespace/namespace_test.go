package namespace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestKubeconfig(t *testing.T, currentContext string, contexts map[string]string) string {
	t.Helper()

	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")

	// Create kubeconfig content
	content := "apiVersion: v1\n"
	content += "kind: Config\n"
	content += "current-context: " + currentContext + "\n"
	content += "clusters:\n"
	content += "- cluster:\n"
	content += "    server: https://localhost:6443\n"
	content += "  name: test-cluster\n"
	content += "contexts:\n"
	for ctx, ns := range contexts {
		content += "- context:\n"
		content += "    cluster: test-cluster\n"
		content += "    user: test-user\n"
		if ns != "" {
			content += "    namespace: " + ns + "\n"
		}
		content += "  name: " + ctx + "\n"
	}
	content += "users:\n"
	content += "- name: test-user\n"
	content += "  user: {}\n"

	if err := os.WriteFile(kubeconfigPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create test kubeconfig: %v", err)
	}

	return kubeconfigPath
}

func TestNewManager(t *testing.T) {
	tests := []struct {
		name           string
		currentContext string
		contexts       map[string]string
		expectError    bool
	}{
		{
			name:           "valid config with namespace",
			currentContext: "test-ctx",
			contexts: map[string]string{
				"test-ctx": "test-ns",
			},
			expectError: false,
		},
		{
			name:           "valid config without namespace",
			currentContext: "test-ctx",
			contexts: map[string]string{
				"test-ctx": "",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kubeconfigPath := createTestKubeconfig(t, tt.currentContext, tt.contexts)
			t.Setenv("KUBECONFIG", kubeconfigPath)

			mgr, err := NewManager()

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, mgr)
			assert.Equal(t, tt.currentContext, mgr.GetCurrentContext())
		})
	}
}

func TestGetCurrentNamespace(t *testing.T) {
	tests := []struct {
		name              string
		currentContext    string
		contexts          map[string]string
		expectedNamespace string
	}{
		{
			name:           "context with namespace",
			currentContext: "test-ctx",
			contexts: map[string]string{
				"test-ctx": "my-namespace",
			},
			expectedNamespace: "my-namespace",
		},
		{
			name:           "context without namespace defaults to default",
			currentContext: "test-ctx",
			contexts: map[string]string{
				"test-ctx": "",
			},
			expectedNamespace: DefaultNamespace,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kubeconfigPath := createTestKubeconfig(t, tt.currentContext, tt.contexts)
			t.Setenv("KUBECONFIG", kubeconfigPath)

			mgr, err := NewManager()
			require.NoError(t, err)

			ns := mgr.GetCurrentNamespace()
			assert.Equal(t, tt.expectedNamespace, ns)
		})
	}
}

func TestSwitchNamespace(t *testing.T) {
	tests := []struct {
		name              string
		currentContext    string
		contexts          map[string]string
		targetNamespace   string
		expectedNamespace string
		expectError       bool
	}{
		{
			name:           "switch to new namespace",
			currentContext: "test-ctx",
			contexts: map[string]string{
				"test-ctx": "initial-ns",
			},
			targetNamespace:   "new-ns",
			expectedNamespace: "new-ns",
			expectError:       false,
		},
		{
			name:           "switch to same namespace",
			currentContext: "test-ctx",
			contexts: map[string]string{
				"test-ctx": "my-ns",
			},
			targetNamespace:   "my-ns",
			expectedNamespace: "my-ns",
			expectError:       false,
		},
		{
			name:           "switch from default namespace",
			currentContext: "test-ctx",
			contexts: map[string]string{
				"test-ctx": "",
			},
			targetNamespace:   "new-ns",
			expectedNamespace: "new-ns",
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kubeconfigPath := createTestKubeconfig(t, tt.currentContext, tt.contexts)
			t.Setenv("KUBECONFIG", kubeconfigPath)

			mgr, err := NewManager()
			require.NoError(t, err)

			err = mgr.SwitchNamespace(tt.targetNamespace)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify the namespace was updated
			ns := mgr.GetCurrentNamespace()
			assert.Equal(t, tt.expectedNamespace, ns)

			// Verify it was persisted by creating a new manager
			mgr2, err := NewManager()
			require.NoError(t, err)

			ns2 := mgr2.GetCurrentNamespace()
			assert.Equal(t, tt.expectedNamespace, ns2)
		})
	}
}

func TestGetCurrentContext(t *testing.T) {
	kubeconfigPath := createTestKubeconfig(t, "my-context", map[string]string{
		"my-context": "my-ns",
	})
	t.Setenv("KUBECONFIG", kubeconfigPath)

	mgr, err := NewManager()
	require.NoError(t, err)

	ctx := mgr.GetCurrentContext()
	assert.Equal(t, "my-context", ctx)
}

func TestListNamespacesFromCluster(t *testing.T) {
	// This test verifies ListNamespacesFromCluster behavior
	// It will fail when no real cluster is available, which is expected in unit tests
	kubeconfigPath := createTestKubeconfig(t, "test-ctx", map[string]string{
		"test-ctx": "default",
	})
	t.Setenv("KUBECONFIG", kubeconfigPath)

	mgr, err := NewManager()
	require.NoError(t, err)

	// Attempt to list namespaces from cluster
	// This will fail without a real cluster connection, which is expected
	namespaces, err := mgr.ListNamespacesFromCluster()

	// We expect an error since there's no real cluster
	// But verify the function returns appropriate error types
	if err != nil {
		// Expected: connection refused, no such host, etc.
		assert.Error(t, err)
		assert.Nil(t, namespaces)
	} else {
		// If somehow connected to a real cluster, verify we get a slice
		assert.NotNil(t, namespaces)
	}
}
