#!/bin/bash
# Test script for multiple KUBECONFIG file support

set -e

echo "=== Testing Multiple KUBECONFIG Support ==="
echo

# Create test config files
TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

cat > $TEST_DIR/config1.yaml <<EOF
apiVersion: v1
kind: Config
current-context: cluster1-context
contexts:
- name: cluster1-context
  context:
    cluster: cluster1
    user: user1
    namespace: namespace1
- name: shared-context
  context:
    cluster: cluster1
    user: user1
clusters:
- name: cluster1
  cluster:
    server: https://cluster1.example.com
users:
- name: user1
  user:
    token: token1
EOF

cat > $TEST_DIR/config2.yaml <<EOF
apiVersion: v1
kind: Config
current-context: cluster2-context
contexts:
- name: cluster2-context
  context:
    cluster: cluster2
    user: user2
    namespace: namespace2
- name: shared-context
  context:
    cluster: cluster2
    user: user2
clusters:
- name: cluster2
  cluster:
    server: https://cluster2.example.com
users:
- name: user2
  user:
    token: token2
EOF

# Test with multiple KUBECONFIG files
export KUBECONFIG="$TEST_DIR/config1.yaml:$TEST_DIR/config2.yaml"

echo "KUBECONFIG=$KUBECONFIG"
echo

echo "1. Testing context listing (should show contexts from both files):"
echo "   Expected: cluster1-context, cluster2-context, shared-context (from config1)"
echo
./kubectl-ctx 2>&1 | head -10 || echo "Interactive test - skipped"
echo

echo "2. Testing current context (should be from first file):"
CURRENT=$(./kubectl-ctx 2>&1 | head -1)
echo "   Current: $CURRENT"
echo

echo "3. Switching to context from second file:"
echo "   Switching to cluster2-context..."
./kubectl-ctx cluster2-context 2>&1 || true
echo

echo "4. Verifying switch with kubectl:"
VERIFY=$(KUBECONFIG="$TEST_DIR/config1.yaml:$TEST_DIR/config2.yaml" kubectl config current-context)
echo "   kubectl shows: $VERIFY"
echo

if [[ "$VERIFY" = "cluster2-context" ]]; then
    echo "✅ SUCCESS: Multiple KUBECONFIG support works!"
else
    echo "❌ FAILED: Expected cluster2-context but got $VERIFY"
fi

echo
echo "Test completed. Config files in: $TEST_DIR"
