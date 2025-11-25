# Kubernetes Deployment Guide

production-grade kubernetes deployment using Kind (Kubernetes in Docker)

## Prerequisites

- Docker installed and running
- Kind installed: `brew install kind` or https://kind.sigs.k8s.io/docs/user/quick-start/
- kubectl installed: `brew install kubectl`
- Make installed

## Quick Start

### Single Command Deployment

```bash
make deploy
```

this will:
1. create a Kind cluster with 1 control-plane + 3 worker nodes
2. build and load the Docker image
3. apply all Kubernetes manifests (namespace, configmaps, secrets, mongodb, api, hpa, network policies)
4. wait for pods to be ready
5. display deployment status

## Available Commands

### Deployment

| Command | Description |
|---------|-------------|
| `make deploy` | 🚀 deploy everything (cluster + build + apply) |
| `make redeploy` | 🔄 fast redeploy with new code changes |
| `make k8s-blue-green-deploy` | 🔵🟢 deploy using blue-green strategy |
| `make k8s-rollback` | ⏪ rollback to previous deployment |

### Cluster Management

| Command | Description |
|---------|-------------|
| `make k8s-cluster-create` | create Kind cluster |
| `make k8s-cluster-delete` | delete Kind cluster |
| `make k8s-status` | check deployment status |
| `make k8s-logs` | tail logs from API pods |
| `make k8s-clean` | clean up Kubernetes resources |

### Testing

| Command | Description |
|---------|-------------|
| `make test` | 🧪 run smoke tests |
| `make k8s-test` | run smoke tests against deployment |

### Utilities

| Command | Description |
|---------|-------------|
| `make k8s-scale REPLICAS=5` | 📈 scale deployment to N replicas |
| `make k8s-shell` | 🐚 open shell in API pod |
| `make k8s-port-forward` | 🔌 port forward to localhost:8080 |

## Architecture

### Components

1. **MongoDB StatefulSet**
   - persistent storage with PVC (10Gi)
   - health/readiness probes
   - resource limits: 512Mi-2Gi RAM, 500m-2000m CPU

2. **REST API Deployment**
   - 3 replicas (min)
   - rolling update strategy
   - health/readiness probes on /health and /ready
   - resource limits: 256Mi-1Gi RAM, 250m-1000m CPU

3. **Horizontal Pod Autoscaler (HPA)**
   - min: 3 replicas
   - max: 20 replicas
   - target: 70% CPU, 80% memory
   - aggressive scale-up, conservative scale-down

4. **Network Policies**
   - API pods can only talk to MongoDB
   - MongoDB only accepts connections from API pods
   - DNS allowed for all

5. **LoadBalancer Service**
   - exposes API on port 80
   - routes to pods on port 4991

### Blue-Green Deployment

blue-green deployment allows zero-downtime deployments:

```bash
# deploy new version to inactive slot
make k8s-blue-green-deploy

# this will:
# 1. deploy to inactive slot (blue or green)
# 2. run smoke tests
# 3. switch traffic if tests pass
# 4. keep old version running for quick rollback
```

manual rollback:
```bash
# get current active slot
kubectl get svc replay-rest-api-service -n replay-api -o jsonpath='{.spec.selector.slot}'

# switch back to previous slot
kubectl patch service replay-rest-api-service -n replay-api \
  -p '{"spec":{"selector":{"slot":"blue"}}}'  # or "green"
```

## Configuration

### Environment Variables

edit `k8s/base/01-configmap.yaml` for configuration:
- `DEV_ENV`: environment name
- `MONGO_URI`: MongoDB connection string
- `HTTP_PORT`: API port
- `LOG_LEVEL`: logging level

### Secrets

edit `k8s/base/02-secrets.yaml` for sensitive data:
- `MONGODB_USERNAME`
- `MONGODB_PASSWORD`
- `JWT_SECRET`
- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`

**IMPORTANT**: in production, use proper secret management (e.g., sealed-secrets, external-secrets, vault)

## Monitoring

### Check Pod Status

```bash
kubectl get pods -n replay-api -w
```

### View Logs

```bash
# tail all API pods
make k8s-logs

# specific pod
kubectl logs -n replay-api <pod-name> -f

# MongoDB logs
kubectl logs -n replay-api mongodb-0 -f
```

### Check HPA Status

```bash
kubectl get hpa -n replay-api
kubectl describe hpa replay-rest-api-hpa -n replay-api
```

### Resource Usage

```bash
kubectl top pods -n replay-api
kubectl top nodes
```

## Troubleshooting

### Pods Not Starting

```bash
# check pod events
kubectl describe pod <pod-name> -n replay-api

# check logs
kubectl logs <pod-name> -n replay-api
```

### Image Pull Issues

```bash
# verify image is loaded into Kind
docker exec -it replay-api-cluster-control-plane crictl images | grep replay-api

# reload image
make k8s-load
```

### MongoDB Connection Issues

```bash
# test MongoDB connectivity
kubectl exec -it -n replay-api \
  $(kubectl get pod -n replay-api -l app=replay-rest-api -o jsonpath='{.items[0].metadata.name}') \
  -- sh -c 'wget -O- mongodb-service:27017'
```

### Service Not Accessible

```bash
# check service
kubectl get svc -n replay-api

# check endpoints
kubectl get endpoints -n replay-api

# port forward for testing
make k8s-port-forward
curl http://localhost:8080/health
```

## Production Considerations

### Security

- [ ] replace default secrets with actual values
- [ ] use secret management solution (vault, sealed-secrets)
- [ ] enable RBAC
- [ ] add pod security policies
- [ ] enable network encryption (TLS)
- [ ] scan images for vulnerabilities

### High Availability

- [x] multiple replicas (3+)
- [x] pod anti-affinity (spread across nodes)
- [x] HPA for auto-scaling
- [ ] pod disruption budgets
- [ ] multi-zone deployment

### Monitoring & Observability

- [ ] prometheus metrics
- [ ] grafana dashboards
- [ ] distributed tracing (jaeger/tempo)
- [ ] log aggregation (loki/elasticsearch)
- [ ] alerting rules

### Backup & Disaster Recovery

- [ ] MongoDB backup strategy
- [ ] persistent volume snapshots
- [ ] disaster recovery plan
- [ ] multi-region deployment

## Development Workflow

### Make Code Changes

```bash
# 1. make code changes
vim pkg/domain/squad/usecases/update_player.go

# 2. fast redeploy
make redeploy

# 3. check logs
make k8s-logs

# 4. test
make test
```

### Add New Dependencies

```bash
# 1. update go.mod
go get github.com/new/package

# 2. rebuild and redeploy
make redeploy
```

### Debug in Pod

```bash
# open shell in running pod
make k8s-shell

# inside pod:
ps aux
env
ls -la /app
```

## Clean Up

```bash
# delete entire cluster
make k8s-cluster-delete

# or just clean resources (keep cluster)
make k8s-clean
```

## Architecture Diagram

```
┌─────────────────────────────────────────┐
│         Kind Cluster                    │
│  ┌───────────────────────────────────┐ │
│  │  Control Plane Node               │ │
│  └───────────────────────────────────┘ │
│  ┌───────────────────────────────────┐ │
│  │  Worker Node 1                    │ │
│  │  ┌──────────┐  ┌──────────┐      │ │
│  │  │ API Pod  │  │ MongoDB  │      │ │
│  │  └──────────┘  └──────────┘      │ │
│  └───────────────────────────────────┘ │
│  ┌───────────────────────────────────┐ │
│  │  Worker Node 2                    │ │
│  │  ┌──────────┐                     │ │
│  │  │ API Pod  │                     │ │
│  │  └──────────┘                     │ │
│  └───────────────────────────────────┘ │
│  ┌───────────────────────────────────┐ │
│  │  Worker Node 3                    │ │
│  │  ┌──────────┐                     │ │
│  │  │ API Pod  │                     │ │
│  │  └──────────┘                     │ │
│  └───────────────────────────────────┘ │
└─────────────────────────────────────────┘
           │
           │ LoadBalancer (port 80)
           │
           ▼
      localhost:80
```

## Next Steps

1. customize secrets in `k8s/base/02-secrets.yaml`
2. adjust resource limits in deployments
3. configure HPA thresholds
4. add monitoring stack (prometheus + grafana)
5. implement CI/CD pipeline
6. add ingress controller for domain routing
7. enable TLS/HTTPS
8. implement backup strategy
