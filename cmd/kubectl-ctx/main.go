package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/AlecAivazis/survey/v2"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
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
	if len(os.Args) > 1 {
		targetContext = os.Args[1]

		// Special case: "-" means switch to previous context
		if targetContext == "-" {
			return fmt.Errorf("previous context switching not yet implemented")
		}

		// Check if context exists
		if _, exists := rawConfig.Contexts[targetContext]; !exists {
			return fmt.Errorf("context %q not found", targetContext)
		}
	} else {
		// No argument: show current context
		if currentContext == "" {
			fmt.Println("No current context set")
		} else {
			fmt.Println(currentContext)
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
		fmt.Fprintf(os.Stderr, "Already on context %q\n", targetContext)
		return nil
	}

	// Switch context by modifying the config
	rawConfig.CurrentContext = targetContext

	// Write back the configuration
	// This uses the first file in KUBECONFIG or default location
	if err := clientcmd.ModifyConfig(loadingRules, rawConfig, false); err != nil {
		return fmt.Errorf("failed to switch context: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Switched to context %q\n", targetContext)
	return nil
}
