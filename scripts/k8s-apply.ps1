$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$K8sDir = Join-Path $Root "deploy\kubernetes"
$Namespace = if ($env:K8S_NAMESPACE) { $env:K8S_NAMESPACE } else { "raven-webmarket" }
$EnvFile = Join-Path $Root ".env"

Write-Host "=== Raven Webmarket — Kubernetes Deploy ==="

if (-not (Get-Command kubectl -ErrorAction SilentlyContinue)) {
    Write-Host "kubectl not found. Install from https://kubernetes.io/docs/tasks/tools/"
    exit 1
}

kubectl cluster-info 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "Cannot reach Kubernetes cluster. Check kubeconfig."
    exit 1
}

$metrics = kubectl get deployment metrics-server -n kube-system 2>$null
if (-not $metrics) {
    Write-Host "WARNING: metrics-server not found. HPA needs metrics-server to scale pods."
}

kubectl apply -k $K8sDir

if (Test-Path $EnvFile) {
    Write-Host "Creating/updating secret raven-env from .env..."
    kubectl create secret generic raven-env --from-env-file=$EnvFile -n $Namespace --dry-run=client -o yaml | kubectl apply -f -
} else {
    Write-Host "WARNING: .env not found. Create secret manually."
}

Write-Host ""
kubectl get deploy -n $Namespace
Write-Host ""
kubectl get hpa -n $Namespace
Write-Host ""
Write-Host "Tune HPA in Admin -> Pod Autoscale, then:"
Write-Host "  kubectl apply -f deploy\kubernetes\hpa-api.yaml"
Write-Host "  kubectl apply -f deploy\kubernetes\hpa-frontend.yaml"
