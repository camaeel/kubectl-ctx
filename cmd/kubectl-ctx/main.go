package main

import (
	"fmt"
	"log/slog"
	"os"
	"sort"

	"github.com/AlecAivazis/survey/v2"
	"github.com/camaeel/kubectl-ctx/internal/utils/logging"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

var rootCmd = &cobra.Command{
	Use:   "kubectl-ctx [CONTEXT_NAME]",
	Short: "Switch between Kubernetes contexts",
	Long: `kubectl-ctx is a tool for switching between Kubernetes contexts.

With no arguments, it shows the current context and provides an interactive
menu to select a new context. With a context name argument, it switches
directly to that context.

The tool automatically handles multiple KUBECONFIG files (e.g., KUBECONFIG=file1:file2).`,
	Example: `  # Show current context and select interactively
  kubectl-ctx

  # Switch to a specific context
  kubectl-ctx my-context`,
	Args:          cobra.MaximumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runSwitch,
}

func main() {
	logging.SetupCLILogger()

	// Ensure help flags are parsed before positional args
	rootCmd.Flags().SetInterspersed(true)

	if err := rootCmd.Execute(); err != nil {
		slog.Error("Error", "error", err)
		os.Exit(1)
	}
}

func runSwitch(cmd *cobra.Command, args []string) error {
	// Load kubeconfig using client-go (handles multiple KUBECONFIG files automatically)
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	// Get raw config (merged from all KUBECONFIG files)
	rawConfig, err := kubeConfig.RawConfig()
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Get current context
	currentContext := rawConfig.CurrentContext

	// Get list of contexts
	contexts := make([]string, 0, len(rawConfig.Contexts))
	for name := range rawConfig.Contexts {
		contexts = append(contexts, name)
	}
	sort.Strings(contexts)

	if len(contexts) == 0 {
		return fmt.Errorf("no contexts found in kubeconfig")
	}

	var targetContext string

	// If argument provided, use it; otherwise show interactive selection
	if len(args) > 0 {
		targetContext = args[0]

		// Check if context exists
		if _, exists := rawConfig.Contexts[targetContext]; !exists {
			return fmt.Errorf("context %q not found", targetContext)
		}
	} else {
		// No argument: show current context
		if currentContext == "" {
			slog.Warn("No current context set")
		} else {
			slog.Info("Current context", "context", currentContext)
		}

		// Interactive selection with survey
		prompt := &survey.Select{
			Message: "Select context:",
			Options: contexts,
			Default: currentContext,
		}
		if err := survey.AskOne(prompt, &targetContext); err != nil {
			return err
		}
	}

	// Don't switch if already on target context
	if targetContext == currentContext {
		slog.Info("Already on context", "context", targetContext)
		return nil
	}

	// Switch context by modifying the config
	rawConfig.CurrentContext = targetContext

	// Write back the configuration
	// This uses the first file in KUBECONFIG or default location
	if err := clientcmd.ModifyConfig(loadingRules, rawConfig, false); err != nil {
		return fmt.Errorf("failed to switch context: %w", err)
	}

	slog.Info("Switched to context", "context", targetContext)
	return nil
}
