#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
K8S_DIR="$ROOT/deploy/kubernetes"
NAMESPACE="${K8S_NAMESPACE:-raven-webmarket}"
ENV_FILE="${ROOT}/.env"

echo "=== Raven Webmarket — Kubernetes Deploy ==="

if ! command -v kubectl &>/dev/null; then
  echo "kubectl not found. Install: https://kubernetes.io/docs/tasks/tools/"
  exit 1
fi

if ! kubectl cluster-info &>/dev/null; then
  echo "Cannot reach Kubernetes cluster. Check kubeconfig."
  exit 1
fi

echo "Checking metrics-server (required for HPA)..."
if ! kubectl get deployment metrics-server -n kube-system &>/dev/null; then
  echo "WARNING: metrics-server not found in kube-system."
  echo "HPA will not scale until metrics-server is installed:"
  echo "  kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml"
fi

kubectl apply -k "$K8S_DIR"

if [ -f "$ENV_FILE" ]; then
  echo "Creating/updating secret raven-env from .env..."
  kubectl create secret generic raven-env \
    --from-env-file="$ENV_FILE" \
    -n "$NAMESPACE" \
    --dry-run=client -o yaml | kubectl apply -f -
else
  echo "WARNING: $ENV_FILE not found. Create secret manually:"
  echo "  kubectl create secret generic raven-env --from-env-file=.env -n $NAMESPACE"
fi

echo ""
echo "Deployments:"
kubectl get deploy -n "$NAMESPACE"
echo ""
echo "HPA (autoscale):"
kubectl get hpa -n "$NAMESPACE"
echo ""
echo "Done. Tune HPA in Admin → Pod Autoscale, then re-apply:"
echo "  kubectl apply -f $K8S_DIR/hpa-api.yaml"
echo "  kubectl apply -f $K8S_DIR/hpa-frontend.yaml"
echo ""
echo "Watch scaling:"
echo "  kubectl get hpa -n $NAMESPACE -w"
