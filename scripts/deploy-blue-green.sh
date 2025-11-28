#!/bin/bash
set -euo pipefail

# Blue-Green Deployment Script for Wallet API
# Usage: ./deploy-blue-green.sh [blue|green] <image-tag> [--auto-switch]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
K8S_DIR="${SCRIPT_DIR}/../k8s/base"
NAMESPACE="replay-api"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check required arguments
if [ $# -lt 2 ]; then
    log_error "Usage: $0 [blue|green] <image-tag> [--auto-switch]"
    log_error "Example: $0 green v1.2.3 --auto-switch"
    exit 1
fi

TARGET_ENV="$1"
IMAGE_TAG="$2"
AUTO_SWITCH="${3:-}"

if [[ "$TARGET_ENV" != "blue" && "$TARGET_ENV" != "green" ]]; then
    log_error "First argument must be 'blue' or 'green'"
    exit 1
fi

CURRENT_ENV=$(kubectl get service wallet-api -n "$NAMESPACE" -o jsonpath='{.spec.selector.version}' 2>/dev/null || echo "blue")
IMAGE_NAME="leetgaming/replay-api:${IMAGE_TAG}"

log_info "═══════════════════════════════════════════════"
log_info "  Wallet API Blue-Green Deployment"
log_info "═══════════════════════════════════════════════"
log_info "Current environment: $CURRENT_ENV"
log_info "Target environment: $TARGET_ENV"
log_info "Image: $IMAGE_NAME"
log_info "Namespace: $NAMESPACE"
log_info "Auto-switch: ${AUTO_SWITCH:-false}"
log_info "═══════════════════════════════════════════════"

# Step 1: Update deployment with new image
log_info "Updating $TARGET_ENV deployment with image $IMAGE_NAME..."
kubectl set image deployment/wallet-api-$TARGET_ENV \
    wallet-api=$IMAGE_NAME \
    -n "$NAMESPACE"

# Step 2: Scale up target environment if it's at 0 replicas
CURRENT_REPLICAS=$(kubectl get deployment wallet-api-$TARGET_ENV -n "$NAMESPACE" -o jsonpath='{.spec.replicas}')
if [ "$CURRENT_REPLICAS" -eq 0 ]; then
    log_info "Scaling $TARGET_ENV environment from 0 to 3 replicas..."
    kubectl scale deployment wallet-api-$TARGET_ENV --replicas=3 -n "$NAMESPACE"
fi

# Step 3: Wait for rollout to complete
log_info "Waiting for $TARGET_ENV deployment to be ready..."
kubectl rollout status deployment/wallet-api-$TARGET_ENV -n "$NAMESPACE" --timeout=5m

# Step 4: Run health checks
log_info "Running health checks on $TARGET_ENV environment..."
TARGET_POD=$(kubectl get pods -n "$NAMESPACE" -l "app=wallet-api,version=$TARGET_ENV" -o jsonpath='{.items[0].metadata.name}')

if ! kubectl exec -n "$NAMESPACE" "$TARGET_POD" -- curl -sf http://localhost:8080/health/ready > /dev/null; then
    log_error "Health check failed on $TARGET_ENV environment"
    exit 1
fi

log_info "✓ Health check passed"

# Step 5: Run smoke tests
log_info "Running smoke tests against $TARGET_ENV environment..."
TARGET_SERVICE="wallet-api-$TARGET_ENV"

# Port forward for testing (run in background)
kubectl port-forward -n "$NAMESPACE" "service/$TARGET_SERVICE" 8080:80 &
PF_PID=$!
sleep 3

# Cleanup function
cleanup() {
    log_info "Cleaning up port-forward..."
    kill $PF_PID 2>/dev/null || true
}
trap cleanup EXIT

# Run basic smoke test
if ! curl -sf http://localhost:8080/health/ready > /dev/null; then
    log_error "Smoke test failed - health endpoint not responding"
    exit 1
fi

log_info "✓ Smoke tests passed"

# Step 6: Auto-switch if requested
if [ "$AUTO_SWITCH" == "--auto-switch" ]; then
    log_warn "Auto-switching traffic to $TARGET_ENV environment..."

    # Update service selector
    kubectl patch service wallet-api -n "$NAMESPACE" -p "{\"spec\":{\"selector\":{\"app\":\"wallet-api\",\"version\":\"$TARGET_ENV\"}}}"

    log_info "✓ Traffic switched to $TARGET_ENV"

    # Wait for a bit to ensure traffic is flowing
    log_info "Monitoring for 30 seconds..."
    sleep 30

    # Check if there are any errors
    ERROR_COUNT=$(kubectl logs -n "$NAMESPACE" -l "app=wallet-api,version=$TARGET_ENV" --since=30s | grep -c "ERROR" || echo "0")

    if [ "$ERROR_COUNT" -gt 10 ]; then
        log_error "High error count detected ($ERROR_COUNT errors). Consider rolling back."
    else
        log_info "✓ No significant errors detected"
    fi

    # Scale down old environment
    log_info "Scaling down $CURRENT_ENV environment to 0 replicas..."
    kubectl scale deployment wallet-api-$CURRENT_ENV --replicas=0 -n "$NAMESPACE"

else
    log_info "═══════════════════════════════════════════════"
    log_warn "Deployment complete but traffic NOT switched."
    log_warn "To manually switch traffic, run:"
    log_warn "  kubectl patch service wallet-api -n $NAMESPACE -p '{\"spec\":{\"selector\":{\"version\":\"$TARGET_ENV\"}}}'"
    log_warn ""
    log_warn "To test the $TARGET_ENV environment:"
    log_warn "  kubectl port-forward -n $NAMESPACE service/wallet-api-$TARGET_ENV 8080:80"
    log_warn ""
    log_warn "To rollback:"
    log_warn "  kubectl patch service wallet-api -n $NAMESPACE -p '{\"spec\":{\"selector\":{\"version\":\"$CURRENT_ENV\"}}}'"
    log_info "═══════════════════════════════════════════════"
fi

log_info "═══════════════════════════════════════════════"
log_info "  Deployment Summary"
log_info "═══════════════════════════════════════════════"
log_info "Deployed: $IMAGE_NAME"
log_info "Environment: $TARGET_ENV"
log_info "Active traffic: $(kubectl get service wallet-api -n "$NAMESPACE" -o jsonpath='{.spec.selector.version}')"
log_info "Blue replicas: $(kubectl get deployment wallet-api-blue -n "$NAMESPACE" -o jsonpath='{.spec.replicas}')"
log_info "Green replicas: $(kubectl get deployment wallet-api-green -n "$NAMESPACE" -o jsonpath='{.spec.replicas}')"
log_info "═══════════════════════════════════════════════"
