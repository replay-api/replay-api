#!/bin/bash

set -e

NAMESPACE="replay-api"
SLOT=${1:-blue}

echo "running smoke tests against $SLOT slot"

# Get a pod from the target slot
POD_NAME=$(kubectl get pods -n $NAMESPACE -l slot=$SLOT -o jsonpath='{.items[0].metadata.name}')

if [ -z "$POD_NAME" ]; then
    echo "error: no pods found for slot $SLOT"
    exit 1
fi

echo "testing pod: $POD_NAME"

# Port forward to the pod
kubectl port-forward -n $NAMESPACE $POD_NAME 18991:4991 &
PF_PID=$!
sleep 3

# Cleanup function
cleanup() {
    kill $PF_PID 2>/dev/null || true
}
trap cleanup EXIT

# Test 1: Health check
echo "test 1: health endpoint"
HEALTH_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:18991/health || echo "000")
if [ "$HEALTH_RESPONSE" != "200" ]; then
    echo "FAIL: health check failed (HTTP $HEALTH_RESPONSE)"
    exit 1
fi
echo "PASS: health check"

# Test 2: Ready check
echo "test 2: ready endpoint"
READY_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:18991/ready || echo "000")
if [ "$READY_RESPONSE" != "200" ]; then
    echo "FAIL: ready check failed (HTTP $READY_RESPONSE)"
    exit 1
fi
echo "PASS: ready check"

# Test 3: API availability (adjust endpoint as needed)
echo "test 3: api availability"
API_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:18991/api/v1/health || echo "000")
if [ "$API_RESPONSE" == "000" ] || [ "$API_RESPONSE" == "502" ] || [ "$API_RESPONSE" == "503" ]; then
    echo "FAIL: api availability check failed (HTTP $API_RESPONSE)"
    exit 1
fi
echo "PASS: api availability check"

echo "all smoke tests passed"
exit 0
