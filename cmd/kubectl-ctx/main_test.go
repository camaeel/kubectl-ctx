package main

import (
	"os"
	"path/filepath"
	"testing"

	ctx "github.com/camaeel/kubectl-ctx/internal/context"
	"github.com/camaeel/kubectl-ctx/internal/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunSwitch_WithArgument(t *testing.T) {
	// Create test kubeconfig
	kubeconfigPath := testutil.CreateKubeconfig(t, "ctx1", map[string]string{
		"ctx1": "",
		"ctx2": "",
		"ctx3": "",
	})

	// Set KUBECONFIG environment variable (automatically restored after test)
	t.Setenv("KUBECONFIG", kubeconfigPath)

	// Create a new command with args
	cmd := &cobra.Command{}
	args := []string{"ctx2"}

	// Run the switch
	err := runSwitch(cmd, args)
	require.NoError(t, err)

	// Verify the context was changed
	mgr, err := ctx.NewManager()
	require.NoError(t, err)
	assert.Equal(t, "ctx2", mgr.GetCurrentContext())
}

func TestRunSwitch_AlreadyOnTargetContext(t *testing.T) {
	// Create test kubeconfig with context already set
	kubeconfigPath := testutil.CreateKubeconfig(t, "ctx1", map[string]string{
		"ctx1": "",
		"ctx2": "",
	})

	t.Setenv("KUBECONFIG", kubeconfigPath)

	// Try to switch to the same context
	cmd := &cobra.Command{}
	args := []string{"ctx1"}

	err := runSwitch(cmd, args)
	require.NoError(t, err)

	// Verify the context is still the same
	mgr, err := ctx.NewManager()
	require.NoError(t, err)
	assert.Equal(t, "ctx1", mgr.GetCurrentContext())
}

func TestRunSwitch_InvalidContext(t *testing.T) {
	// Create test kubeconfig
	kubeconfigPath := testutil.CreateKubeconfig(t, "ctx1", map[string]string{
		"ctx1": "",
		"ctx2": "",
	})

	t.Setenv("KUBECONFIG", kubeconfigPath)

	cmd := &cobra.Command{}
	args := []string{"nonexistent-context"}

	err := runSwitch(cmd, args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRunSwitch_InvalidKubeconfig(t *testing.T) {
	// Set KUBECONFIG to non-existent file
	tmpDir := t.TempDir()
	t.Setenv("KUBECONFIG", filepath.Join(tmpDir, "nonexistent"))

	cmd := &cobra.Command{}
	args := []string{"some-context"}

	err := runSwitch(cmd, args)
	assert.Error(t, err)
}

func TestRunSwitch_NoCurrentContext(t *testing.T) {
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")

	// Create kubeconfig without current-context
	content := `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://localhost:6443
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: ctx1
- context:
    cluster: test-cluster
    user: test-user
  name: ctx2
users:
- name: test-user
  user: {}
`
	require.NoError(t, os.WriteFile(kubeconfigPath, []byte(content), 0600))

	t.Setenv("KUBECONFIG", kubeconfigPath)

	cmd := &cobra.Command{}
	args := []string{"ctx1"}

	// Should succeed even without current context
	err := runSwitch(cmd, args)
	require.NoError(t, err)

	// Verify the context was set
	mgr, err := ctx.NewManager()
	require.NoError(t, err)
	assert.Equal(t, "ctx1", mgr.GetCurrentContext())
}

func TestRunSwitch_MultipleContextSwitches(t *testing.T) {
	// Create test kubeconfig
	kubeconfigPath := testutil.CreateKubeconfig(t, "ctx1", map[string]string{
		"ctx1": "",
		"ctx2": "",
		"ctx3": "",
	})

	t.Setenv("KUBECONFIG", kubeconfigPath)

	cmd := &cobra.Command{}

	// Switch to ctx2
	err := runSwitch(cmd, []string{"ctx2"})
	require.NoError(t, err)

	mgr, err := ctx.NewManager()
	require.NoError(t, err)
	assert.Equal(t, "ctx2", mgr.GetCurrentContext())

	// Switch to ctx3
	err = runSwitch(cmd, []string{"ctx3"})
	require.NoError(t, err)

	mgr, err = ctx.NewManager()
	require.NoError(t, err)
	assert.Equal(t, "ctx3", mgr.GetCurrentContext())

	// Switch back to ctx1
	err = runSwitch(cmd, []string{"ctx1"})
	require.NoError(t, err)

	mgr, err = ctx.NewManager()
	require.NoError(t, err)
	assert.Equal(t, "ctx1", mgr.GetCurrentContext())
}

func TestRunSwitch_ListContexts(t *testing.T) {
	// Create test kubeconfig
	kubeconfigPath := testutil.CreateKubeconfig(t, "ctx1", map[string]string{
		"ctx1": "",
		"ctx2": "",
		"ctx3": "",
	})

	t.Setenv("KUBECONFIG", kubeconfigPath)

	// Verify we can list all contexts
	mgr, err := ctx.NewManager()
	require.NoError(t, err)

	contexts := mgr.ListContexts()
	assert.ElementsMatch(t, []string{"ctx1", "ctx2", "ctx3"}, contexts)
}

func TestRunSwitch_MultipleKubeconfigFiles(t *testing.T) {
	// Create two kubeconfig files
	kubeconfig1 := testutil.CreateKubeconfig(t, "ctx1", map[string]string{
		"ctx1": "",
		"ctx2": "",
	})

	kubeconfig2 := testutil.CreateKubeconfig(t, "ctx3", map[string]string{
		"ctx3": "",
		"ctx4": "",
	})

	// Set KUBECONFIG with multiple files (colon-separated on Unix)
	t.Setenv("KUBECONFIG", kubeconfig1+":"+kubeconfig2)

	// Verify we can list contexts from both files
	mgr, err := ctx.NewManager()
	require.NoError(t, err)

	contexts := mgr.ListContexts()
	assert.ElementsMatch(t, []string{"ctx1", "ctx2", "ctx3", "ctx4"}, contexts)

	// Verify current context from first file
	assert.Equal(t, "ctx1", mgr.GetCurrentContext())

	// Switch to context from second file
	cmd := &cobra.Command{}
	err = runSwitch(cmd, []string{"ctx3"})
	require.NoError(t, err)

	// Verify the switch
	mgr, err = ctx.NewManager()
	require.NoError(t, err)
	assert.Equal(t, "ctx3", mgr.GetCurrentContext())
}
