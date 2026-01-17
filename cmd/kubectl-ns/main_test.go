package main

import (
	"path/filepath"
	"testing"

	ns "github.com/camaeel/kubectl-ctx/internal/namespace"
	"github.com/camaeel/kubectl-ctx/internal/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunSwitch_WithArgument(t *testing.T) {
	// Create test kubeconfig
	kubeconfigPath := testutil.CreateKubeconfig(t, "test-ctx", map[string]string{
		"test-ctx": "initial-ns",
	})

	// Set KUBECONFIG environment variable (automatically restored after test)
	t.Setenv("KUBECONFIG", kubeconfigPath)

	// Create a new command with args
	cmd := &cobra.Command{}
	args := []string{"target-ns"}

	// Run the switch
	err := runSwitch(cmd, args)
	require.NoError(t, err)

	// Verify the namespace was changed
	mgr, err := ns.NewManager()
	require.NoError(t, err)
	assert.Equal(t, "target-ns", mgr.GetCurrentNamespace())
}

func TestRunSwitch_AlreadyOnTargetNamespace(t *testing.T) {
	// Create test kubeconfig with namespace already set
	kubeconfigPath := testutil.CreateKubeconfig(t, "test-ctx", map[string]string{
		"test-ctx": "my-namespace",
	})

	t.Setenv("KUBECONFIG", kubeconfigPath)

	// Try to switch to the same namespace
	cmd := &cobra.Command{}
	args := []string{"my-namespace"}

	err := runSwitch(cmd, args)
	require.NoError(t, err)

	// Verify the namespace is still the same
	mgr, err := ns.NewManager()
	require.NoError(t, err)
	assert.Equal(t, "my-namespace", mgr.GetCurrentNamespace())
}

func TestRunSwitch_SwitchFromDefault(t *testing.T) {
	// Create test kubeconfig without namespace (defaults to "default")
	kubeconfigPath := testutil.CreateKubeconfig(t, "test-ctx", map[string]string{
		"test-ctx": "",
	})

	t.Setenv("KUBECONFIG", kubeconfigPath)

	cmd := &cobra.Command{}
	args := []string{"new-namespace"}

	err := runSwitch(cmd, args)
	require.NoError(t, err)

	// Verify the namespace was changed
	mgr, err := ns.NewManager()
	require.NoError(t, err)
	assert.Equal(t, "new-namespace", mgr.GetCurrentNamespace())
}

func TestRunSwitch_InvalidKubeconfig(t *testing.T) {
	// Set KUBECONFIG to non-existent file
	tmpDir := t.TempDir()
	t.Setenv("KUBECONFIG", filepath.Join(tmpDir, "nonexistent"))

	cmd := &cobra.Command{}
	args := []string{"some-namespace"}

	err := runSwitch(cmd, args)
	assert.Error(t, err)
}

func TestRunSwitch_NoCurrentContext(t *testing.T) {
	// Create kubeconfig without current-context
	kubeconfigPath := testutil.CreateKubeconfig(t, "", map[string]string{
		"test-ctx": "",
	})

	t.Setenv("KUBECONFIG", kubeconfigPath)

	cmd := &cobra.Command{}
	args := []string{"some-namespace"}

	err := runSwitch(cmd, args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no current context")
}

func TestRunSwitch_MultipleNamespaceSwitches(t *testing.T) {
	// Create test kubeconfig
	kubeconfigPath := testutil.CreateKubeconfig(t, "test-ctx", map[string]string{
		"test-ctx": "ns1",
	})

	t.Setenv("KUBECONFIG", kubeconfigPath)

	cmd := &cobra.Command{}

	// Switch to ns2
	err := runSwitch(cmd, []string{"ns2"})
	require.NoError(t, err)

	mgr, err := ns.NewManager()
	require.NoError(t, err)
	assert.Equal(t, "ns2", mgr.GetCurrentNamespace())

	// Switch to ns3
	err = runSwitch(cmd, []string{"ns3"})
	require.NoError(t, err)

	mgr, err = ns.NewManager()
	require.NoError(t, err)
	assert.Equal(t, "ns3", mgr.GetCurrentNamespace())

	// Switch back to ns1
	err = runSwitch(cmd, []string{"ns1"})
	require.NoError(t, err)

	mgr, err = ns.NewManager()
	require.NoError(t, err)
	assert.Equal(t, "ns1", mgr.GetCurrentNamespace())
}

func TestRunSwitch_MultipleKubeconfigFiles(t *testing.T) {
	// Create two kubeconfig files
	kubeconfig1 := testutil.CreateKubeconfig(t, "ctx1", map[string]string{
		"ctx1": "ns1",
	})

	kubeconfig2 := testutil.CreateKubeconfig(t, "ctx2", map[string]string{
		"ctx2": "ns2",
	})

	// Set KUBECONFIG with multiple files (colon-separated on Unix)
	t.Setenv("KUBECONFIG", kubeconfig1+":"+kubeconfig2)

	// Verify current context from first file
	mgr, err := ns.NewManager()
	require.NoError(t, err)
	assert.Equal(t, "ctx1", mgr.GetCurrentContext())
	assert.Equal(t, "ns1", mgr.GetCurrentNamespace())

	// Switch namespace in the current context
	cmd := &cobra.Command{}
	err = runSwitch(cmd, []string{"new-ns"})
	require.NoError(t, err)

	// Verify the switch
	mgr, err = ns.NewManager()
	require.NoError(t, err)
	assert.Equal(t, "new-ns", mgr.GetCurrentNamespace())
	assert.Equal(t, "ctx1", mgr.GetCurrentContext())
}
