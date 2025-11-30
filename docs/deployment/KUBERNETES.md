# Kubernetes Blue-Green Deployment Guide

## Overview

This directory contains production-grade Kubernetes manifests for deploying the Wallet API with blue-green deployment strategy for zero-downtime updates.

## Architecture

```
┌──────────────────────────────────────┐
│         Ingress / Load Balancer      │
└────────────────┬─────────────────────┘
                 │
    ┌────────────▼────────────┐
    │    wallet-api Service   │  ← Switches between blue/green
    │  (selector: version)    │
    └────────┬──────┬─────────┘
             │      │
      ┌──────▼──┐ ┌▼──────┐
      │  Blue   │ │ Green │  ← Only one receives traffic
      │ Deploy  │ │ Deploy│
      └──┬──────┘ └───┬───┘
         │            │
    ┌────▼────────────▼───┐
    │  MongoDB Replica Set │  ← 3-node StatefulSet
    │  (rs0)               │
    └──────────────────────┘
```

## Directory Structure

```
k8s/
├── base/                     # Base Kubernetes manifests
│   ├── namespace.yaml        # Namespace, RBAC, NetworkPolicy
│   ├── deployment.yaml       # Blue & Green deployments
│   ├── service.yaml          # Services (main, blue, green, headless)
│   ├── configmap.yaml        # Application configuration
│   ├── secret.yaml           # Sensitive data (template)
│   ├── mongodb-statefulset.yaml  # MongoDB 3-node replica set
│   └── hpa.yaml              # HorizontalPodAutoscaler & PDB
└── README.md                 # This file
```

## Prerequisites

1. **Kubernetes Cluster**: EKS, GKE, AKS, or self-managed (>= v1.25)
2. **kubectl**: Configured to access your cluster
3. **Storage Class**: SSD storage class named `fast-ssd`
4. **Secrets Management**: External Secrets Operator, Sealed Secrets, or Vault
5. **Ingress Controller**: nginx-ingress or similar
6. **Metrics Server**: For HPA autoscaling

## Initial Setup

### 1. Create Namespace and RBAC

```bash
kubectl apply -f k8s/base/namespace.yaml
```

This creates:
- `replay-api` namespace
- Resource quotas and limits
- Service account with RBAC permissions
- Network policies for security

### 2. Configure Secrets

**IMPORTANT**: Never commit real secrets to git!

```bash
# Create MongoDB secrets
kubectl create secret generic mongodb-secrets \
  --from-literal=root-username=admin \
  --from-literal=root-password='STRONG_RANDOM_PASSWORD' \
  --from-literal=replica-set-key='RANDOM_REPLICA_KEY' \
  -n replay-api

# Create application secrets
kubectl create secret generic wallet-api-secrets \
  --from-literal=MONGO_URI='mongodb://user:password@mongodb-0.mongodb-headless:27017,mongodb-1.mongodb-headless:27017,mongodb-2.mongodb-headless:27017/replay_prod?replicaSet=rs0&authSource=admin' \
  --from-literal=ETHEREUM_RPC_URL='https://mainnet.infura.io/v3/YOUR_API_KEY' \
  --from-literal=POLYGON_RPC_URL='https://polygon-mainnet.g.alchemy.com/v2/YOUR_API_KEY' \
  --from-literal=JWT_SECRET='32-byte-random-secret' \
  --from-literal=ENCRYPTION_KEY='32-byte-hex-encryption-key' \
  -n replay-api

# Create MongoDB keyfile secret
openssl rand -base64 756 > mongodb-keyfile
kubectl create secret generic mongodb-keyfile \
  --from-file=mongodb-keyfile=./mongodb-keyfile \
  -n replay-api
rm mongodb-keyfile
```

### 3. Deploy MongoDB

```bash
kubectl apply -f k8s/base/mongodb-statefulset.yaml
```

Wait for MongoDB pods to be ready:
```bash
kubectl wait --for=condition=ready pod -l app=mongodb -n replay-api --timeout=5m
```

Initialize replica set:
```bash
kubectl apply -f k8s/base/mongodb-statefulset.yaml  # Includes init job
```

### 4. Apply ConfigMap and Services

```bash
kubectl apply -f k8s/base/configmap.yaml
kubectl apply -f k8s/base/service.yaml
```

### 5. Deploy Blue Environment (Initial)

```bash
kubectl apply -f k8s/base/deployment.yaml
```

This creates both blue and green deployments, with blue having 3 replicas and green having 0.

### 6. Apply HPA and PDB

```bash
kubectl apply -f k8s/base/hpa.yaml
```

## Blue-Green Deployment Workflow

### Method 1: Using Deployment Script (Recommended)

```bash
# Deploy to green environment (no traffic switch)
./scripts/deploy-blue-green.sh green v1.2.3

# Deploy and auto-switch traffic
./scripts/deploy-blue-green.sh green v1.2.3 --auto-switch
```

### Method 2: Manual Deployment

#### Step 1: Deploy New Version to Inactive Environment

Determine current active environment:
```bash
ACTIVE_ENV=$(kubectl get service wallet-api -n replay-api -o jsonpath='{.spec.selector.version}')
echo "Current active: $ACTIVE_ENV"

# Set target to the opposite
if [ "$ACTIVE_ENV" == "blue" ]; then
  TARGET_ENV="green"
else
  TARGET_ENV="blue"
fi
```

Deploy new version:
```bash
# Update deployment image
kubectl set image deployment/wallet-api-$TARGET_ENV \
  wallet-api=leetgaming/replay-api:v1.2.3 \
  -n replay-api

# Scale up if needed
kubectl scale deployment wallet-api-$TARGET_ENV --replicas=3 -n replay-api

# Wait for rollout
kubectl rollout status deployment/wallet-api-$TARGET_ENV -n replay-api
```

#### Step 2: Test New Environment

```bash
# Port-forward to test
kubectl port-forward -n replay-api service/wallet-api-$TARGET_ENV 8080:80

# In another terminal, run tests
curl http://localhost:8080/health/ready
# Run your smoke tests...
```

#### Step 3: Switch Traffic

```bash
# Switch service selector to new environment
kubectl patch service wallet-api -n replay-api \
  -p "{\"spec\":{\"selector\":{\"version\":\"$TARGET_ENV\"}}}"

# Monitor for errors
kubectl logs -f -n replay-api -l "app=wallet-api,version=$TARGET_ENV"
```

#### Step 4: Scale Down Old Environment

```bash
# After monitoring for 5-10 minutes with no issues
kubectl scale deployment wallet-api-$ACTIVE_ENV --replicas=0 -n replay-api
```

### Rollback

If issues are detected:

```bash
# Instantly switch back to old environment
kubectl patch service wallet-api -n replay-api \
  -p "{\"spec\":{\"selector\":{\"version\":\"$ACTIVE_ENV\"}}}"
```

## CI/CD Integration

### GitHub Actions Workflow

```bash
# Trigger deployment via GitHub Actions
gh workflow run deploy-k8s.yml \
  -f environment=green \
  -f image_tag=v1.2.3 \
  -f auto_switch=true \
  -f cluster=production
```

### Manual Trigger via Workflow Dispatch

1. Go to GitHub Actions → Deploy to Kubernetes
2. Click "Run workflow"
3. Select:
   - Target environment: `blue` or `green`
   - Image tag: e.g., `v1.2.3` or commit SHA
   - Auto-switch: Enable to automatically switch traffic
   - Cluster: `production` or `staging`

## Monitoring

### Check Deployment Status

```bash
# View all resources
kubectl get all -n replay-api

# Check active environment
kubectl get service wallet-api -n replay-api -o jsonpath='{.spec.selector.version}'

# Check replica counts
kubectl get deployments -n replay-api

# Check pod status
kubectl get pods -n replay-api -l app=wallet-api
```

### View Logs

```bash
# Stream logs from active environment
ACTIVE=$(kubectl get service wallet-api -n replay-api -o jsonpath='{.spec.selector.version}')
kubectl logs -f -n replay-api -l "app=wallet-api,version=$ACTIVE"

# View MongoDB logs
kubectl logs -f -n replay-api mongodb-0
```

### Metrics and Autoscaling

```bash
# Check HPA status
kubectl get hpa -n replay-api

# Check resource usage
kubectl top pods -n replay-api

# Check PDB status
kubectl get pdb -n replay-api
```

## Troubleshooting

### Pods Not Starting

```bash
# Describe pod to see events
kubectl describe pod <pod-name> -n replay-api

# Check init container logs
kubectl logs <pod-name> -n replay-api -c db-migration

# Check resource constraints
kubectl get resourcequotas -n replay-api
```

### MongoDB Connection Issues

```bash
# Check MongoDB replica set status
kubectl exec -it mongodb-0 -n replay-api -- mongosh -u admin -p <password> --authenticationDatabase admin

rs.status()

# Check MongoDB logs
kubectl logs mongodb-0 -n replay-api
```

### High Memory/CPU Usage

```bash
# Check current usage
kubectl top pods -n replay-api

# Check HPA decisions
kubectl describe hpa wallet-api-blue-hpa -n replay-api

# Temporarily increase limits (not recommended for production)
kubectl set resources deployment wallet-api-blue \
  --limits=cpu=2000m,memory=2Gi \
  -n replay-api
```

### Service Not Routing Traffic

```bash
# Check service endpoints
kubectl get endpoints wallet-api -n replay-api

# Verify selector matches
kubectl get service wallet-api -n replay-api -o yaml | grep -A5 selector
kubectl get pods -n replay-api -l app=wallet-api --show-labels
```

## Security Best Practices

1. **Secrets Management**: Use external secrets management (Vault, AWS Secrets Manager, GCP Secret Manager)
2. **Network Policies**: Enforce strict pod-to-pod communication
3. **RBAC**: Principle of least privilege
4. **Pod Security Standards**: Enforce restricted PSS
5. **Image Scanning**: Scan images for vulnerabilities before deployment
6. **Audit Logging**: Enable Kubernetes audit logs
7. **TLS**: Use TLS for all external communication

## Scaling

### Manual Scaling

```bash
# Scale up
kubectl scale deployment wallet-api-blue --replicas=10 -n replay-api

# Scale down
kubectl scale deployment wallet-api-blue --replicas=3 -n replay-api
```

### Adjust HPA

```bash
# Edit HPA to change thresholds
kubectl edit hpa wallet-api-blue-hpa -n replay-api

# Or apply updated manifest
kubectl apply -f k8s/base/hpa.yaml
```

## Backup and Disaster Recovery

### MongoDB Backup

```bash
# Create backup job
kubectl apply -f k8s/backup/mongodb-backup-cronjob.yaml

# Manual backup
kubectl exec -it mongodb-0 -n replay-api -- \
  mongodump --uri="mongodb://user:pass@localhost:27017/replay_prod" \
  --out=/backup/$(date +%Y%m%d)
```

### Restore from Backup

```bash
kubectl exec -it mongodb-0 -n replay-api -- \
  mongorestore --uri="mongodb://user:pass@localhost:27017/replay_prod" \
  /backup/20231201
```

## Cost Optimization

1. **Right-size Pods**: Monitor actual usage and adjust requests/limits
2. **HPA**: Let autoscaler handle traffic spikes
3. **PDB**: Ensure minimum availability during scale-down
4. **Spot Instances**: Use for non-critical workloads
5. **Resource Quotas**: Prevent runaway resource usage

## Contact

For questions or issues, contact the Platform Engineering team.
