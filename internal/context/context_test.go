package context

import (
	"os"
	"path/filepath"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

func createTestKubeconfig(t *testing.T, contexts []string, currentContext string) string {
	t.Helper()

	config := api.NewConfig()

	// Add contexts
	for _, name := range contexts {
		config.Contexts[name] = &api.Context{
			Cluster:  "test-cluster",
			AuthInfo: "test-user",
		}
	}

	// Add required cluster and user
	config.Clusters["test-cluster"] = &api.Cluster{
		Server: "https://test-server:6443",
	}
	config.AuthInfos["test-user"] = &api.AuthInfo{
		Token: "test-token",
	}

	config.CurrentContext = currentContext

	// Create temp file
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")

	if err := clientcmd.WriteToFile(*config, kubeconfigPath); err != nil {
		t.Fatalf("Failed to write test kubeconfig: %v", err)
	}

	// Set KUBECONFIG env var
	t.Setenv("KUBECONFIG", kubeconfigPath)

	return kubeconfigPath
}

func TestNewManager(t *testing.T) {
	createTestKubeconfig(t, []string{"dev", "prod"}, "dev")

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	if manager == nil {
		t.Fatal("NewManager() returned nil manager")
	}

	if manager.config == nil {
		t.Fatal("Manager config is nil")
	}
}

func TestGetCurrentContext(t *testing.T) {
	tests := []struct {
		name           string
		contexts       []string
		currentContext string
		want           string
	}{
		{
			name:           "with current context",
			contexts:       []string{"dev", "staging", "prod"},
			currentContext: "staging",
			want:           "staging",
		},
		{
			name:           "no current context",
			contexts:       []string{"dev", "prod"},
			currentContext: "",
			want:           "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createTestKubeconfig(t, tt.contexts, tt.currentContext)

			manager, err := NewManager()
			if err != nil {
				t.Fatalf("NewManager() failed: %v", err)
			}

			got := manager.GetCurrentContext()
			if got != tt.want {
				t.Errorf("GetCurrentContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListContexts(t *testing.T) {
	tests := []struct {
		name     string
		contexts []string
		want     []string
	}{
		{
			name:     "multiple contexts",
			contexts: []string{"prod", "dev", "staging"},
			want:     []string{"dev", "prod", "staging"}, // Should be sorted
		},
		{
			name:     "single context",
			contexts: []string{"minikube"},
			want:     []string{"minikube"},
		},
		{
			name:     "alphabetically sorted",
			contexts: []string{"z-context", "a-context", "m-context"},
			want:     []string{"a-context", "m-context", "z-context"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createTestKubeconfig(t, tt.contexts, tt.contexts[0])

			manager, err := NewManager()
			if err != nil {
				t.Fatalf("NewManager() failed: %v", err)
			}

			got := manager.ListContexts()
			if len(got) != len(tt.want) {
				t.Errorf("ListContexts() length = %v, want %v", len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ListContexts()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestValidateContext(t *testing.T) {
	createTestKubeconfig(t, []string{"dev", "staging", "prod"}, "dev")

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	tests := []struct {
		name        string
		contextName string
		wantErr     bool
	}{
		{
			name:        "valid context",
			contextName: "dev",
			wantErr:     false,
		},
		{
			name:        "another valid context",
			contextName: "prod",
			wantErr:     false,
		},
		{
			name:        "invalid context",
			contextName: "nonexistent",
			wantErr:     true,
		},
		{
			name:        "empty context name",
			contextName: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.ValidateContext(tt.contextName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateContext() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSwitchContext(t *testing.T) {
	tests := []struct {
		name         string
		initialCtx   string
		targetCtx    string
		wantErr      bool
		expectSwitch bool
	}{
		{
			name:         "switch to different context",
			initialCtx:   "dev",
			targetCtx:    "prod",
			wantErr:      false,
			expectSwitch: true,
		},
		{
			name:         "switch to same context (no-op)",
			initialCtx:   "dev",
			targetCtx:    "dev",
			wantErr:      false,
			expectSwitch: false,
		},
		{
			name:         "switch to nonexistent context",
			initialCtx:   "dev",
			targetCtx:    "nonexistent",
			wantErr:      true,
			expectSwitch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createTestKubeconfig(t, []string{"dev", "staging", "prod"}, tt.initialCtx)

			manager, err := NewManager()
			if err != nil {
				t.Fatalf("NewManager() failed: %v", err)
			}

			err = manager.SwitchContext(tt.targetCtx)
			if (err != nil) != tt.wantErr {
				t.Errorf("SwitchContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			// Verify the switch by creating a new manager
			newManager, err := NewManager()
			if err != nil {
				t.Fatalf("NewManager() after switch failed: %v", err)
			}

			if tt.expectSwitch {
				if newManager.GetCurrentContext() != tt.targetCtx {
					t.Errorf("Context not switched: got %v, want %v", newManager.GetCurrentContext(), tt.targetCtx)
				}
			} else {
				if newManager.GetCurrentContext() != tt.initialCtx {
					t.Errorf("Context changed unexpectedly: got %v, want %v", newManager.GetCurrentContext(), tt.initialCtx)
				}
			}
		})
	}
}

func TestMultipleKubeconfigFiles(t *testing.T) {
	// Create first kubeconfig
	config1 := api.NewConfig()
	config1.Contexts["ctx1"] = &api.Context{
		Cluster:  "cluster1",
		AuthInfo: "user1",
	}
	config1.Clusters["cluster1"] = &api.Cluster{Server: "https://server1:6443"}
	config1.AuthInfos["user1"] = &api.AuthInfo{Token: "token1"}
	config1.CurrentContext = "ctx1"

	tmpDir := t.TempDir()
	path1 := filepath.Join(tmpDir, "config1")
	if err := clientcmd.WriteToFile(*config1, path1); err != nil {
		t.Fatalf("Failed to write config1: %v", err)
	}

	// Create second kubeconfig
	config2 := api.NewConfig()
	config2.Contexts["ctx2"] = &api.Context{
		Cluster:  "cluster2",
		AuthInfo: "user2",
	}
	config2.Clusters["cluster2"] = &api.Cluster{Server: "https://server2:6443"}
	config2.AuthInfos["user2"] = &api.AuthInfo{Token: "token2"}

	path2 := filepath.Join(tmpDir, "config2")
	if err := clientcmd.WriteToFile(*config2, path2); err != nil {
		t.Fatalf("Failed to write config2: %v", err)
	}

	// Set KUBECONFIG with multiple files
	t.Setenv("KUBECONFIG", path1+string(os.PathListSeparator)+path2)

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() with multiple configs failed: %v", err)
	}

	// Should have both contexts
	contexts := manager.ListContexts()
	if len(contexts) != 2 {
		t.Errorf("Expected 2 contexts, got %d", len(contexts))
	}

	// Verify both contexts exist
	if err := manager.ValidateContext("ctx1"); err != nil {
		t.Errorf("ctx1 not found: %v", err)
	}
	if err := manager.ValidateContext("ctx2"); err != nil {
		t.Errorf("ctx2 not found: %v", err)
	}

	if manager.GetCurrentContext() != "ctx1" {
		t.Errorf("CurrentContext = %v, want ctx1", manager.GetCurrentContext())
	}
}
