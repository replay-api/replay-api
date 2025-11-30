#!/bin/bash

set -e

NAMESPACE="replay-api"
NEW_VERSION=$1
CURRENT_SLOT=$(kubectl get service replay-rest-api-service -n $NAMESPACE -o jsonpath='{.spec.selector.slot}')

if [ -z "$NEW_VERSION" ]; then
    echo "Usage: ./blue-green-deploy.sh <new-version-tag>"
    exit 1
fi

echo "Current active slot: $CURRENT_SLOT"

# Determine new slot
if [ "$CURRENT_SLOT" == "blue" ]; then
    NEW_SLOT="green"
else
    NEW_SLOT="blue"
fi

echo "Deploying new version ($NEW_VERSION) to $NEW_SLOT slot..."

# Update image in new slot deployment
kubectl set image deployment/replay-rest-api-$NEW_SLOT \
    replay-rest-api=replay-api:$NEW_VERSION \
    -n $NAMESPACE

# Wait for rollout to complete
echo "Waiting for rollout to complete..."
kubectl rollout status deployment/replay-rest-api-$NEW_SLOT -n $NAMESPACE --timeout=5m

# Run smoke tests on new slot
echo "Running smoke tests on $NEW_SLOT slot..."
./scripts/smoke-test.sh $NEW_SLOT

if [ $? -eq 0 ]; then
    echo "Smoke tests passed! Switching traffic to $NEW_SLOT..."

    # Update service selector to point to new slot
    kubectl patch service replay-rest-api-service -n $NAMESPACE \
        -p "{\"spec\":{\"selector\":{\"slot\":\"$NEW_SLOT\"}}}"

    echo "Traffic successfully switched to $NEW_SLOT!"
    echo "Old $CURRENT_SLOT deployment is still running for quick rollback if needed."
    echo "To rollback: kubectl patch service replay-rest-api-service -n $NAMESPACE -p '{\"spec\":{\"selector\":{\"slot\":\"$CURRENT_SLOT\"}}}'"
else
    echo "Smoke tests failed! Keeping traffic on $CURRENT_SLOT."
    echo "New deployment in $NEW_SLOT slot is available for debugging."
    exit 1
fi
