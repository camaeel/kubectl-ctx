# kubectl-ctx and kubectl-ns

Simple context and namespace switchers for kubectl using client-go libraries.

## Features

- ✅ **Multiple KUBECONFIG support** - automatically handles `KUBECONFIG=file1:file2:file3`
- ✅ **Context switching** - switch between Kubernetes contexts
- ✅ **Namespace switching** - switch namespaces within current context
- ✅ **Interactive mode** - select from list when no argument provided
- ✅ **Uses kubectl's libraries** - same behavior as kubectl for config merging

## Installation

### Quick Install (MacOS and Linux)

#### Install with Homebrew

```bash
brew install camaeel/tap/kubectl-ctx
```

#### Manual download
```bash
# Automatically detect platform and install latest version
VERSION=$(curl -s https://api.github.com/repos/camaeel/kubectl-ctx/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
curl -L "https://github.com/camaeel/kubectl-ctx/releases/download/${VERSION}/kubectl-ctx-${OS}-${ARCH}" -o kubectl-ctx
curl -L "https://github.com/camaeel/kubectl-ctx/releases/download/${VERSION}/kubectl-ns-${OS}-${ARCH}" -o kubectl-ns
chmod +x kubectl-ctx kubectl-ns
sudo mv kubectl-ctx kubectl-ns /usr/local/bin/
```

### Windows (PowerShell)
$VERSION = (Invoke-RestMethod -Uri "https://api.github.com/repos/camaeel/kubectl-ctx/releases/latest").tag_name
Invoke-WebRequest -Uri "https://github.com/camaeel/kubectl-ctx/releases/download/$VERSION/kubectl-ctx-windows-amd64.exe" -OutFile "kubectl-ctx.exe"
Invoke-WebRequest -Uri "https://github.com/camaeel/kubectl-ctx/releases/download/$VERSION/kubectl-ns-windows-amd64.exe" -OutFile "kubectl-ns.exe"
# Move to a directory in your PATH
```

### Verify Installation

```bash
kubectl ctx --version
kubectl ns --version
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

