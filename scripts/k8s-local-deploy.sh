#!/bin/bash
set -e

echo "=== MailRaven Local Kubernetes Deployment ==="
echo ""

# Check prerequisites
command -v kubectl >/dev/null 2>&1 || { echo "Error: kubectl not found. Install with: brew install kubectl"; exit 1; }
command -v docker >/dev/null 2>&1 || { echo "Error: docker not found. Install Docker Desktop."; exit 1; }

# Check K8s is running
if ! kubectl cluster-info >/dev/null 2>&1; then
    echo "Error: Kubernetes cluster not running."
    echo "Enable it in Docker Desktop → Settings → Kubernetes → Enable Kubernetes"
    exit 1
fi

echo "✓ Kubernetes cluster is running"
echo ""

# Step 1: Build images
echo "Step 1/4: Building Docker images..."
docker build -t mailraven-server:local -f build/Dockerfile . --quiet
echo "  ✓ Backend image built"

docker build -t mailraven-frontend:local -f build/Dockerfile.frontend . --quiet
echo "  ✓ Frontend image built"
echo ""

# Step 2: Create namespace
echo "Step 2/4: Creating namespace..."
kubectl create namespace mailraven --dry-run=client -o yaml | kubectl apply -f -
echo "  ✓ Namespace ready"
echo ""

# Step 3: Apply manifests
echo "Step 3/4: Deploying to Kubernetes..."
kubectl apply -f deployment/kubernetes/local/secret.yaml
kubectl apply -f deployment/kubernetes/base/configmap.yaml
kubectl apply -f deployment/kubernetes/base/statefulsets.yaml
kubectl apply -f deployment/kubernetes/base/services.yaml

# Wait for postgres to be ready before deploying backend
echo "  Waiting for PostgreSQL..."
kubectl -n mailraven wait --for=condition=ready pod -l app=postgres --timeout=120s 2>/dev/null || true

kubectl apply -k deployment/kubernetes/local/
echo "  ✓ All resources deployed"
echo ""

# Step 4: Wait for pods
echo "Step 4/4: Waiting for pods to be ready..."
kubectl -n mailraven wait --for=condition=ready pod -l component=backend --timeout=120s 2>/dev/null || {
    echo "  ⚠ Backend pods taking longer than expected. Check with:"
    echo "    kubectl -n mailraven get pods"
    echo "    kubectl -n mailraven logs deployment/mailraven-backend"
}
echo ""

echo "=== Deployment Complete ==="
echo ""
echo "Access the services:"
echo "  API:      http://localhost:30080/health"
echo "  SMTP:     localhost:30025"
echo "  Metrics:  http://localhost:30080/metrics"
echo ""
echo "Or use port-forward for standard ports:"
echo "  kubectl -n mailraven port-forward svc/mailraven-http 8080:8080"
echo "  kubectl -n mailraven port-forward svc/mailraven-smtp 2525:25"
echo ""
echo "Useful commands:"
echo "  kubectl -n mailraven get pods          # See pod status"
echo "  kubectl -n mailraven logs -f deploy/mailraven-backend  # Stream logs"
echo "  kubectl -n mailraven top pods          # Resource usage"
echo "  kubectl delete namespace mailraven     # Tear down everything"
echo ""
