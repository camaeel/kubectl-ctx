package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/AlecAivazis/survey/v2"
	ns "github.com/camaeel/kubectl-ctx/internal/namespace"
	"github.com/camaeel/kubectl-ctx/internal/utils/logging"
	"github.com/spf13/cobra"
)

var (
	// Version is set by build flags
	Version = "dev"
)

var rootCmd = &cobra.Command{
	Use:   "kubectl-ns [NAMESPACE]",
	Short: "Switch between Kubernetes namespaces",
	Long: `kubectl-ns is a tool for switching namespaces in the current Kubernetes context.

With no arguments, it shows the current namespace and provides an interactive
menu to select a new namespace (fetched from the cluster if accessible).
With a namespace argument, it switches directly to that namespace.

The tool automatically handles multiple KUBECONFIG files (e.g., KUBECONFIG=file1:file2).`,
	Example: `  # Show current namespace and select interactively
  kubectl-ns

  # Switch to a specific namespace
  kubectl-ns kube-system`,
	Args:          cobra.MaximumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runSwitch,
}

func init() {
	rootCmd.Version = Version
}

func main() {
	logging.SetupCLILogger()

	// Ensure help flags are parsed before positional args
	rootCmd.Flags().SetInterspersed(true)

	if err := rootCmd.Execute(); err != nil {
		slog.Error("Error occurred:", "error", err)
		os.Exit(1)
	}
}

func runSwitch(_ *cobra.Command, args []string) error {
	// Create namespace manager
	manager, err := ns.NewManager()
	if err != nil {
		return err
	}

	// Get current namespace
	currentNamespace := manager.GetCurrentNamespace()
	currentContext := manager.GetCurrentContext()

	var targetNamespace string

	// If argument provided, use it; otherwise show interactive selection
	if len(args) > 0 {
		targetNamespace = args[0]
	} else {
		// Show current namespace
		slog.Info("Current namespace", "namespace", currentNamespace)

		// Try to get namespaces from cluster
		namespaces, err := manager.ListNamespacesFromCluster()
		if err != nil {
			return fmt.Errorf("failed to fetch namespaces from cluster: %w", err)
		}

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

	// Don't switch if already on target namespace
	if targetNamespace == currentNamespace {
		slog.Info("Already on namespace", "namespace", targetNamespace)
		return nil
	}

	// Switch namespace
	if err := manager.SwitchNamespace(targetNamespace); err != nil {
		return err
	}

	slog.Info("Switched to namespace", "namespace", targetNamespace, "context", currentContext)
	return nil
}
