package main

import (
	"log/slog"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/camaeel/kubectl-ctx/internal/context"
	"github.com/camaeel/kubectl-ctx/internal/utils/logging"
	"github.com/spf13/cobra"
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
		slog.Error("Error occurred:", "error", err)
		os.Exit(1)
	}
}

func runSwitch(_ *cobra.Command, args []string) error {
	// Create context manager
	manager, err := context.NewManager()
	if err != nil {
		return err
	}

	// Get available contexts
	contexts := manager.ListContexts()
	currentContext := manager.GetCurrentContext()

	var targetContext string

	// If argument provided, use it; otherwise show interactive selection
	if len(args) > 0 {
		targetContext = args[0]

		// Validate context exists
		if err := manager.ValidateContext(targetContext); err != nil {
			return err
		}
	} else {
		// Show current context
		if currentContext == "" {
			slog.Warn("No current context set")
		} else {
			slog.Info("Current context", "context", currentContext)
		}

		// Interactive selection
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

	// Switch context
	if err := manager.SwitchContext(targetContext); err != nil {
		return err
	}

	slog.Info("Switched to context", "context", targetContext)
	return nil
}
