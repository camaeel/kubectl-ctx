package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/camaeel/kubectl-ctx/internal/utils/logging"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const defaultNamespace = "default"

func main() {
	logging.SetupCLILogger()

	if err := run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run() error {
	// Load kubeconfig
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	// Get raw config
	rawConfig, err := kubeConfig.RawConfig()
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Get current context
	currentContext := rawConfig.CurrentContext
	if currentContext == "" {
		return fmt.Errorf("no current context set")
	}

	context, exists := rawConfig.Contexts[currentContext]
	if !exists {
		return fmt.Errorf("current context %q not found in config", currentContext)
	}

	// Get current namespace
	currentNamespace := context.Namespace
	if currentNamespace == "" {
		currentNamespace = defaultNamespace
	}

	var targetNamespace string

	// If argument provided, use it; otherwise show interactive selection
	if len(os.Args) > 1 {
		targetNamespace = os.Args[1]

		// Special case: "-" means switch to previous namespace
		if targetNamespace == "-" {
			return fmt.Errorf("previous namespace switching not yet implemented")
		}
	} else {
		// Try to get namespaces from cluster
		namespaces, err := getNamespacesFromCluster(kubeConfig)
		if err != nil {
			// If we can't connect to cluster, allow manual input
			slog.Warn("Could not fetch namespaces from cluster, falling back to manual input", "error", err)
			prompt := &survey.Input{
				Message: "Enter namespace name:",
				Default: currentNamespace,
			}
			if err := survey.AskOne(prompt, &targetNamespace); err != nil {
				return err
			}
		} else {
			// Show interactive selection with actual namespaces
			prompt := &survey.Select{
				Message: "Select namespace:",
				Options: namespaces,
				Default: currentNamespace,
			}
			if err := survey.AskOne(prompt, &targetNamespace); err != nil {
				return err
			}
		}
	}

	// Don't switch if already on target namespace
	slog.Info("Already on namespace", "namespace", targetNamespace)
	if targetNamespace == currentNamespace {
		slog.Warn("already on namespace", "namespace", targetNamespace)
		return nil
	}

	// Update namespace in the current context
	context.Namespace = targetNamespace
	rawConfig.Contexts[currentContext] = context

	// Write back the configuration
	if err := clientcmd.ModifyConfig(loadingRules, rawConfig, false); err != nil {
		return fmt.Errorf("failed to switch namespace: %w", err)
	}

	slog.Info("Switched to namespace", "namespace", targetNamespace, "context", currentContext)
	return nil
}

// getNamespacesFromCluster attempts to fetch namespaces from the cluster
func getNamespacesFromCluster(config clientcmd.ClientConfig) ([]string, error) {
	restConfig, err := config.ClientConfig()
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
