# kubectl-ctx and kubectl-ns

Simple context and namespace switchers for kubectl using client-go libraries.

## Features

- ✅ **Multiple KUBECONFIG support** - automatically handles `KUBECONFIG=file1:file2:file3`
- ✅ **Context switching** - switch between Kubernetes contexts
- ✅ **Namespace switching** - switch namespaces within current context
- ✅ **Interactive mode** - select from list when no argument provided
- ✅ **Uses kubectl's libraries** - same behavior as kubectl for config merging

## Installation

### Using Pre-built Binaries (Recommended)

Download the latest release for your platform:

```bash
# Linux (amd64)
curl -LO https://github.com/camaeel/kubectl-ctx/releases/latest/download/kubectl-ctx-linux-amd64
curl -LO https://github.com/camaeel/kubectl-ctx/releases/latest/download/kubectl-ns-linux-amd64
chmod +x kubectl-ctx-linux-amd64 kubectl-ns-linux-amd64
sudo mv kubectl-ctx-linux-amd64 /usr/local/bin/kubectl-ctx
sudo mv kubectl-ns-linux-amd64 /usr/local/bin/kubectl-ns

# Linux (arm64)
curl -LO https://github.com/camaeel/kubectl-ctx/releases/latest/download/kubectl-ctx-linux-arm64
curl -LO https://github.com/camaeel/kubectl-ctx/releases/latest/download/kubectl-ns-linux-arm64
chmod +x kubectl-ctx-linux-arm64 kubectl-ns-linux-arm64
sudo mv kubectl-ctx-linux-arm64 /usr/local/bin/kubectl-ctx
sudo mv kubectl-ns-linux-arm64 /usr/local/bin/kubectl-ns

# macOS (Intel)
curl -LO https://github.com/camaeel/kubectl-ctx/releases/latest/download/kubectl-ctx-darwin-amd64
curl -LO https://github.com/camaeel/kubectl-ctx/releases/latest/download/kubectl-ns-darwin-amd64
chmod +x kubectl-ctx-darwin-amd64 kubectl-ns-darwin-amd64
sudo mv kubectl-ctx-darwin-amd64 /usr/local/bin/kubectl-ctx
sudo mv kubectl-ns-darwin-amd64 /usr/local/bin/kubectl-ns

# macOS (Apple Silicon)
curl -LO https://github.com/camaeel/kubectl-ctx/releases/latest/download/kubectl-ctx-darwin-arm64
curl -LO https://github.com/camaeel/kubectl-ctx/releases/latest/download/kubectl-ns-darwin-arm64
chmod +x kubectl-ctx-darwin-arm64 kubectl-ns-darwin-arm64
sudo mv kubectl-ctx-darwin-arm64 /usr/local/bin/kubectl-ctx
sudo mv kubectl-ns-darwin-arm64 /usr/local/bin/kubectl-ns

# Windows (PowerShell)
curl.exe -LO https://github.com/camaeel/kubectl-ctx/releases/latest/download/kubectl-ctx-windows-amd64.exe
curl.exe -LO https://github.com/camaeel/kubectl-ctx/releases/latest/download/kubectl-ns-windows-amd64.exe
# Move to a directory in your PATH
```

### Verify Installation

```bash
kubectl-ctx --version
kubectl-ns --version
```

### Building from Source

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
kubectl ctx

# Switch to a specific context
kubectl ctx my-context

# Interactive mode (shows numbered list)
kubectl ctx
# Then select from the list
```

### kubectl-ns (Namespace Switcher)

```bash
# Show current namespace
kubectl ns

# Switch to a specific namespace
kubectl ns my-namespace

# Interactive mode (prompts for input)
kubectl ns
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
kubectl ctx  # Shows contexts from all files merged
```

The tools automatically:
- Merge contexts from all files (first occurrence wins for duplicates)
- Write current-context changes to the first file in KUBECONFIG
- Handle relative paths and LocationOfOrigin tracking

## Differences from Original kubectx/kubens

- Uses client-go instead of custom YAML parsing
- Simpler interactive mode (no fzf dependency)
- Only supports switching (no rename/delete operations)
- Guaranteed compatibility with kubectl behavior (support for multiple KUBECONFIG files)

