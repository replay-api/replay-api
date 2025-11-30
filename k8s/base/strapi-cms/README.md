# Strapi CMS Kubernetes Manifests

This directory contains the Kubernetes manifests for deploying Strapi CMS as part of the LeetGaming platform.

## Components

- **PostgreSQL**: Primary database for Strapi content storage
- **Redis**: Caching layer for improved performance
- **Strapi CMS**: Headless CMS for managing blog posts, announcements, and other content

## Prerequisites

1. Kubernetes cluster (v1.25+)
2. kubectl configured with cluster access
3. kustomize (v4.0+)
4. cert-manager installed (for TLS certificates)
5. nginx-ingress controller installed

## Deployment

### 1. Create secrets

Before deploying, create the required secrets:

```bash
# Generate secure values
export ADMIN_JWT_SECRET=$(openssl rand -base64 32)
export API_TOKEN_SALT=$(openssl rand -base64 32)
export APP_KEYS=$(openssl rand -base64 32),$(openssl rand -base64 32)
export TRANSFER_TOKEN_SALT=$(openssl rand -base64 32)
export JWT_SECRET=$(openssl rand -base64 32)
export DATABASE_USERNAME=strapi_user
export DATABASE_PASSWORD=$(openssl rand -base64 24)
export REDIS_PASSWORD=$(openssl rand -base64 24)
export AWS_ACCESS_KEY_ID=your_aws_key
export AWS_SECRET_ACCESS_KEY=your_aws_secret

# Apply with substitutions
envsubst < 02-secrets.yaml | kubectl apply -f -
```

### 2. Deploy using Kustomize

```bash
# Apply all manifests
kubectl apply -k .

# Or apply individually in order
kubectl apply -f 00-namespace.yaml
kubectl apply -f 01-configmap.yaml
kubectl apply -f 02-secrets.yaml
kubectl apply -f 03-postgres-statefulset.yaml
kubectl apply -f 04-redis-deployment.yaml
kubectl apply -f 05-strapi-deployment.yaml
kubectl apply -f 06-hpa.yaml
kubectl apply -f 07-ingress.yaml
kubectl apply -f 08-network-policy.yaml
```

### 3. Verify deployment

```bash
# Check all pods are running
kubectl get pods -n strapi-cms

# Check services
kubectl get svc -n strapi-cms

# Check ingress
kubectl get ingress -n strapi-cms
```

## Configuration

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `ADMIN_JWT_SECRET` | JWT secret for admin authentication | Yes |
| `API_TOKEN_SALT` | Salt for API token generation | Yes |
| `APP_KEYS` | Application encryption keys | Yes |
| `DATABASE_USERNAME` | PostgreSQL username | Yes |
| `DATABASE_PASSWORD` | PostgreSQL password | Yes |
| `AWS_ACCESS_KEY_ID` | AWS key for S3 uploads | Yes |
| `AWS_SECRET_ACCESS_KEY` | AWS secret for S3 uploads | Yes |

### Scaling

The deployment includes HPA (Horizontal Pod Autoscaler) configured to:
- Min replicas: 2
- Max replicas: 5
- Target CPU utilization: 70%
- Target Memory utilization: 80%

### Network Policies

Network policies are configured to:
- Allow ingress only from nginx-ingress and replay-api namespace
- Allow egress to PostgreSQL, Redis, and external HTTPS
- Isolate database and cache pods

## Monitoring

Add Prometheus ServiceMonitor:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: strapi-cms-monitor
  namespace: strapi-cms
spec:
  selector:
    matchLabels:
      app: strapi-cms
  endpoints:
    - port: http
      path: /metrics
      interval: 30s
```

## Backup

PostgreSQL backups should be configured using:
- pg_dump for logical backups
- WAL archiving for point-in-time recovery
- S3 storage for backup retention

## Content Types

The CMS will be configured with:
- Blog Posts
- News/Announcements
- FAQ
- Documentation
- Landing Pages
- Game Updates

## Integration with Frontend

The CMS API will be consumed by the Next.js frontend at:
- Public API: `https://cms.leetgaming.gg/api`
- Internal API: `http://cms.internal.leetgaming.gg/api`

## Troubleshooting

### Pod not starting
```bash
kubectl logs -f deployment/strapi-cms -n strapi-cms
```

### Database connection issues
```bash
kubectl exec -it postgres-strapi-0 -n strapi-cms -- psql -U strapi_user -d strapi_cms
```

### Redis connection issues
```bash
kubectl exec -it deployment/redis-strapi -n strapi-cms -- redis-cli -a $REDIS_PASSWORD ping
```
