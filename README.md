# kubectl-ctx and kubectl-ns

Simple context and namespace switchers for kubectl using client-go libraries.

## Features

- ✅ **Multiple KUBECONFIG support** - automatically handles `KUBECONFIG=file1:file2:file3`
- ✅ **Context switching** - switch between Kubernetes contexts
- ✅ **Namespace switching** - switch namespaces within current context
- ✅ **Interactive mode** - select from list when no argument provided
- ✅ **Uses kubectl's libraries** - same behavior as kubectl for config merging

## Installation

```bash
# Build both tools
go build -o kubectl-ctx ./cmd/kubectl-ctx
go build -o kubectl-ns ./cmd/kubectl-ns

# Install to PATH
cp kubectl-ctx kubectl-ns /usr/local/bin/
```

## Usage

### kubectl-ctx (Context Switcher)

```bash
# Show current context
kubectl-ctx

# Switch to a specific context
kubectl-ctx my-context

# Interactive mode (shows numbered list)
kubectl-ctx
# Then select from the list
```

### kubectl-ns (Namespace Switcher)

```bash
# Show current namespace
kubectl-ns

# Switch to a specific namespace
kubectl-ns my-namespace

# Interactive mode (prompts for input)
kubectl-ns
# Then enter namespace name
```

## How It Works

Both tools use Kubernetes' `client-go` libraries:
- `clientcmd.NewDefaultClientConfigLoadingRules()` - handles KUBECONFIG env var and file merging
- `clientcmd.ModifyConfig()` - writes changes back to the appropriate config file
- Follows kubectl's exact behavior for multi-file configurations

## Multiple KUBECONFIG Files

Works seamlessly with multiple config files:

```bash
export KUBECONFIG=~/.kube/config:~/.kube/config-prod:~/.kube/config-dev
kubectl-ctx  # Shows contexts from all files merged
```

The tools automatically:
- Merge contexts from all files (first occurrence wins for duplicates)
- Write current-context changes to the first file in KUBECONFIG
- Handle relative paths and LocationOfOrigin tracking

## Differences from Original kubectx/kubens

- Uses client-go instead of custom YAML parsing
- Simpler interactive mode (no fzf dependency yet)
- Larger binary size (~15MB vs ~5MB) due to client-go
- Only supports switching (no rename/delete operations)
- Guaranteed compatibility with kubectl behavior

## Future Enhancements

- [ ] Add previous context/namespace switching (`-` flag)
- [ ] Integrate fzf for better interactive experience
- [ ] Query actual namespaces from cluster for kubectl-ns
- [ ] Add shell completion
- [ ] Add context aliasing
