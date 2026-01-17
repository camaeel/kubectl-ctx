package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// CreateKubeconfig creates a temporary kubeconfig file with the specified contexts and optional namespaces.
// contexts is a map where key is context name and value is namespace (use empty string for no namespace).
// Returns the path to the created kubeconfig file.
func CreateKubeconfig(t *testing.T, currentContext string, contexts map[string]string) string {
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

	require.NoError(t, os.WriteFile(kubeconfigPath, []byte(content), 0600))

	return kubeconfigPath
}
