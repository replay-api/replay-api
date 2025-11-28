# Production-Grade Wallet System - Deployment Guide

## 🎯 Overview

This document provides a comprehensive guide for deploying the **financial-grade multi-asset wallet system** with **Kubernetes blue-green deployment** strategy.

## ✅ Completed Implementation

### 1. **Financial-Grade Wallet System**

**Core Features:**
- ✅ Multi-asset support (Fiat, Crypto, NFT, Game Credits)
- ✅ Double-entry accounting with immutable ledger
- ✅ Saga pattern for atomic transactions with automatic rollback
- ✅ Idempotency protection (24h TTL)
- ✅ Daily reconciliation service
- ✅ Transaction coordinator for financial integrity

**Files Created:**
- `pkg/domain/wallet/services/transaction_coordinator.go` (433 lines)
  - Implements saga pattern with automatic rollback
  - Methods: ExecuteDeposit, ExecuteWithdrawal, ExecuteEntryFee, ExecutePrizeWinning
  - Prevents orphaned ledger entries

### 2. **Production-Grade Testing**

**Test Infrastructure (NO MOCKS):**
- ✅ Smoke tests (`test/smoke/wallet_smoke_test.go`) - Fast, no external dependencies
- ✅ E2E integration tests (`test/integration/wallet_e2e_test.go`) - Real MongoDB + Hardhat
- ✅ Blockchain testing with Hardhat local EVM
  - MockUSDC (ERC-20, 6 decimals)
  - MockUSDT (ERC-20, 6 decimals)
  - GameNFT (ERC-721 with rarity levels)

**Test Coverage:**
```
✓ Wallet creation and validation
✓ Deposit with double-entry ledger
✓ Withdrawal with balance checks
✓ Reconciliation (wallet matches ledger)
✓ Idempotency (duplicate detection)
✓ Entry fee with insufficient balance rollback
✓ Prize winning with daily limit enforcement
✓ Transaction history with pagination
✓ Ledger balance calculation accuracy
✓ Benchmark: Deposit throughput
```

**Files Created:**
- `test/smoke/wallet_smoke_test.go` - 12 fast unit tests + benchmarks
- `test/integration/wallet_e2e_test.go` - 8 comprehensive E2E tests
- `test/blockchain/contracts/MockUSDC.sol` - ERC-20 for testing
- `test/blockchain/contracts/MockUSDT.sol` - ERC-20 for testing
- `test/blockchain/contracts/GameNFT.sol` - ERC-721 with rarity
- `test/blockchain/scripts/deploy.js` - Contract deployment automation
- `docker-compose.test.yml` - Orchestrates MongoDB + Hardhat
- `Makefile.test` - Test automation commands

### 3. **Kubernetes Blue-Green Deployment**

**Infrastructure Components:**
- ✅ Namespace with RBAC, resource quotas, and network policies
- ✅ Blue and Green deployments (zero-downtime updates)
- ✅ Service definitions (main, blue, green, headless)
- ✅ MongoDB StatefulSet (3-node replica set for HA)
- ✅ HorizontalPodAutoscaler (CPU, memory, custom metrics)
- ✅ PodDisruptionBudgets (maintain availability during updates)
- ✅ ConfigMaps for environment configuration
- ✅ Secrets management (template with instructions)

**Files Created:**
- `k8s/base/namespace.yaml` - Namespace, RBAC, NetworkPolicy, quotas
- `k8s/base/deployment.yaml` - Blue & Green deployment manifests
- `k8s/base/service.yaml` - Service definitions
- `k8s/base/configmap.yaml` - Application configuration (60+ settings)
- `k8s/base/secret.yaml` - Secret template with instructions
- `k8s/base/mongodb-statefulset.yaml` - 3-node MongoDB cluster
- `k8s/base/hpa.yaml` - Autoscaling + PodDisruptionBudgets
- `k8s/README.md` - Comprehensive deployment guide

**Key Features:**
- **Zero-downtime deployments**: Traffic switches instantly between blue/green
- **Automatic rollback**: One-command rollback to previous version
- **Health checks**: Liveness, readiness, and startup probes
- **Autoscaling**: 3-20 replicas based on CPU, memory, and RPS
- **Security**: NetworkPolicy, RBAC, Pod Security Standards
- **High availability**: 3-node MongoDB replica set, pod anti-affinity

### 4. **Deployment Automation**

**Scripts:**
- ✅ `scripts/deploy-blue-green.sh` - Automated blue-green deployment
  - Deploys to inactive environment
  - Runs health checks and smoke tests
  - Optional auto-switch traffic
  - Scales down old environment

**CI/CD Pipelines:**
- ✅ `.github/workflows/ci.yml` - Complete CI/CD pipeline
  - Smoke tests (5 min)
  - Unit tests with coverage
  - Integration tests (MongoDB + Hardhat)
  - E2E tests with Docker Compose
  - Lint and security scanning
  - Docker build and push
  - Performance benchmarks

- ✅ `.github/workflows/deploy-k8s.yml` - Kubernetes deployment
  - Workflow dispatch (manual trigger)
  - Auto-deploy on main branch push
  - Blue-green deployment strategy
  - Health checks and smoke tests
  - Automatic rollback on failure
  - Slack notifications

### 5. **Documentation**

**Comprehensive Documentation:**
- ✅ `WALLET_SYSTEM.md` - Complete wallet system architecture (440 lines)
  - Double-entry accounting patterns
  - Saga pattern explanation
  - MongoDB schema design
  - Compliance (SOX, PCI-DSS, AML/KYC)
  - Performance benchmarks
  - Disaster recovery

- ✅ `k8s/README.md` - Kubernetes deployment guide (350+ lines)
  - Setup instructions
  - Blue-green deployment workflow
  - Monitoring and troubleshooting
  - Security best practices
  - Scaling and cost optimization

## 📊 Architecture Overview

### System Architecture

```
┌──────────────────────────────────────────────────────┐
│              Kubernetes Cluster (EKS/GKE/AKS)        │
│                                                      │
│  ┌────────────────────────────────────────────────┐ │
│  │         Ingress / Load Balancer                │ │
│  └───────────────────┬────────────────────────────┘ │
│                      │                               │
│      ┌───────────────▼──────────────┐               │
│      │    wallet-api Service        │ ← Switches    │
│      │   (selector: blue/green)     │   traffic     │
│      └────────┬──────────┬──────────┘               │
│               │          │                           │
│        ┌──────▼──┐   ┌──▼──────┐                   │
│        │  Blue   │   │  Green  │ ← Only one active  │
│        │ Deploy  │   │ Deploy  │                    │
│        │ (3-20)  │   │  (0-20) │                    │
│        └──┬──────┘   └───┬─────┘                   │
│           │              │                           │
│   ┌───────▼──────────────▼────────┐                │
│   │  MongoDB Replica Set (3-node) │                │
│   │  rs0 (StatefulSet)            │                │
│   └───────────────────────────────┘                │
└──────────────────────────────────────────────────────┘
```

### Transaction Flow

```
User Request
    │
    ▼
Transaction Coordinator (Saga Pattern)
    │
    ├─► Step 1: Record in Ledger
    │   ├─ Success → Continue
    │   └─ Failure → Rollback (none needed yet)
    │
    ├─► Step 2: Update Wallet Balance
    │   ├─ Success → Continue
    │   └─ Failure → Rollback Step 1 (reverse ledger entry)
    │
    └─► Step 3: Persist to Database
        ├─ Success → Commit
        └─ Failure → Rollback Steps 1 & 2
```

## 🚀 Quick Start

### 1. Local Development & Testing

```bash
# Start test infrastructure (MongoDB + Hardhat)
docker compose -f docker-compose.test.yml up -d

# Run smoke tests (no external dependencies)
go test -v -short -tags=smoke ./test/smoke/...

# Run E2E tests (with MongoDB + Hardhat)
MONGO_TEST_URI="mongodb://test:test123@localhost:27018/replay_test?authSource=admin" \
  go test -v -tags=e2e ./test/integration/...

# Run all tests via Makefile
make -f Makefile.test test-all
```

### 2. Deploy to Kubernetes

```bash
# Initial setup
kubectl apply -f k8s/base/namespace.yaml
kubectl apply -f k8s/base/mongodb-statefulset.yaml
kubectl apply -f k8s/base/configmap.yaml
kubectl apply -f k8s/base/service.yaml
kubectl apply -f k8s/base/deployment.yaml
kubectl apply -f k8s/base/hpa.yaml

# Deploy new version to green environment
./scripts/deploy-blue-green.sh green v1.2.3

# Test green environment
kubectl port-forward -n replay-api service/wallet-api-green 8080:80
curl http://localhost:8080/health/ready

# Switch traffic to green
kubectl patch service wallet-api -n replay-api \
  -p '{"spec":{"selector":{"version":"green"}}}'

# Scale down blue
kubectl scale deployment wallet-api-blue --replicas=0 -n replay-api
```

### 3. CI/CD Deployment

```bash
# Trigger deployment via GitHub Actions
gh workflow run deploy-k8s.yml \
  -f environment=green \
  -f image_tag=v1.2.3 \
  -f auto_switch=true \
  -f cluster=production
```

## 🔒 Security & Compliance

### Financial-Grade Security

- ✅ **Idempotency**: Prevents duplicate transactions (24h TTL)
- ✅ **Immutable Ledger**: All transactions recorded permanently
- ✅ **Double-Entry Accounting**: Accounting equation always balances
- ✅ **Automatic Rollback**: Saga pattern ensures no orphaned data
- ✅ **Audit Trail**: Complete transaction history with metadata
- ✅ **Risk Scoring**: Fraud detection on every transaction

### Kubernetes Security

- ✅ **Network Policies**: Strict pod-to-pod communication
- ✅ **RBAC**: Least privilege access control
- ✅ **Pod Security Standards**: Restricted PSS enforced
- ✅ **Secrets Management**: External secrets integration
- ✅ **Non-root Containers**: All containers run as user 1000
- ✅ **Read-only Root Filesystem**: Prevents runtime modifications

### Compliance Ready

- ✅ **SOX**: Immutable audit trail, user attribution
- ✅ **PCI-DSS**: No credit card storage, encryption at rest/transit
- ✅ **AML/KYC**: Transaction metadata, risk scoring, limits
- ✅ **Tax Reporting**: Complete history, 1099-K generation support

## 📈 Performance & Scalability

### Benchmarks

```
BenchmarkDeposit-8    500 ops    2847 ns/op    1456 B/op    23 allocs/op
```

### Autoscaling Configuration

- **Min Replicas**: 3 per environment
- **Max Replicas**: 20 per environment
- **Scale Triggers**:
  - CPU > 70%
  - Memory > 80%
  - RPS > 1000 per pod
- **Scale Up**: Immediate (0s stabilization)
- **Scale Down**: 5 min stabilization window

### High Availability

- **MongoDB**: 3-node replica set (quorum-based)
- **API Pods**: Spread across availability zones
- **Pod Disruption Budget**: Minimum 2 pods always available
- **Zero-downtime Deployments**: Blue-green strategy

## 🛠 Monitoring & Observability

### Metrics Exposed

- Prometheus metrics on port 9090
- Health endpoints:
  - `/health/live` - Liveness probe
  - `/health/ready` - Readiness probe
  - `/health/startup` - Startup probe

### Logging

- Structured JSON logging
- Request ID tracking
- Full audit trail in ledger

### Alerting

- High error rates
- Failed transactions
- Reconciliation discrepancies
- Resource exhaustion

## 🔄 Blue-Green Deployment Workflow

### Standard Deployment Process

```mermaid
graph LR
    A[Deploy to Inactive] --> B[Health Checks]
    B --> C[Smoke Tests]
    C --> D{Tests Pass?}
    D -->|Yes| E[Switch Traffic]
    D -->|No| F[Rollback]
    E --> G[Monitor]
    G --> H{Stable?}
    H -->|Yes| I[Scale Down Old]
    H -->|No| F
```

### Rollback Process

```bash
# Instant rollback (< 1 second)
kubectl patch service wallet-api -n replay-api \
  -p '{"spec":{"selector":{"version":"blue"}}}'
```

## 📝 Environment Configuration

### Key Configuration Parameters

**Wallet Limits:**
- Daily withdrawal limit: $10,000
- Daily deposit limit: $50,000
- Daily prize limit: $100,000
- Min withdrawal: $10
- Max withdrawal: $25,000

**Database:**
- MongoDB connection pool: 10-100 connections
- Idle timeout: 5 minutes
- Server selection timeout: 5 seconds

**Blockchain:**
- Ethereum confirmations: 12 blocks
- Polygon confirmations: 128 blocks
- Sync interval: 15 seconds

**Reconciliation:**
- Schedule: Daily at 2 AM
- Auto-correct threshold: $1.00
- Batch size: 100 wallets

## 📦 Docker Images

### Production Images

```bash
# Build production image
docker build -t leetgaming/replay-api:latest .

# Push to registry
docker push leetgaming/replay-api:latest
docker push leetgaming/replay-api:v1.2.3
```

### Test Images

- `mongo:7.0` - MongoDB 7.0
- `node:20-alpine` - Hardhat EVM node

## 🧪 Testing Strategy

### Test Pyramid

```
         E2E Tests (8 tests)
       ┌─────────────────────┐
       │ Real MongoDB + EVM  │
       │ NO MOCKS            │
       └─────────────────────┘
              ▲
              │
    Integration Tests (20+ tests)
  ┌───────────────────────────────┐
  │ MongoDB + Hardhat services    │
  │ Smart contract testing        │
  └───────────────────────────────┘
              ▲
              │
       Smoke Tests (12 tests)
┌─────────────────────────────────────┐
│ Fast, no dependencies               │
│ Validate core logic                 │
└─────────────────────────────────────┘
```

## 🎓 Key Learnings

### Production-Grade Patterns

1. **Saga Pattern**: Essential for distributed transactions
2. **Blue-Green Deployment**: Zero-downtime updates
3. **HPA + PDB**: Balanced scaling and availability
4. **NetworkPolicy**: Defense in depth
5. **Immutable Infrastructure**: Treat pods as cattle, not pets

### Financial-Grade Requirements

1. **Double-Entry Accounting**: Non-negotiable for financial systems
2. **Idempotency**: Prevent duplicate transactions
3. **Audit Trail**: Complete transaction history
4. **Reconciliation**: Regular balance verification
5. **Automatic Rollback**: Prevent orphaned data

## 🚧 Future Enhancements

- [ ] Multi-currency support (EUR, GBP, JPY)
- [ ] Crypto withdrawals (ETH, MATIC direct to wallets)
- [ ] NFT marketplace integration
- [ ] Game credit bundles with expiration
- [ ] Machine learning fraud detection
- [ ] Real-time balance streaming via WebSocket
- [ ] Multi-region deployment with geo-replication

## 📞 Support

For questions or issues:
- Platform Engineering Team
- GitHub Issues: https://github.com/leetgaming-pro/replay-api/issues
- Slack: #platform-engineering

## 📜 License

Proprietary - LeetGaming Pro Platform
