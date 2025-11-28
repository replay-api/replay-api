# Developer Onboarding Guide

> **Welcome to the LeetGaming Platform Engineering Team!**

This guide will help you get up and running with the Replay API codebase in under 30 minutes.

---

## 🎯 Quick Start (5 Minutes)

### Prerequisites

```bash
# Required
✓ Go 1.23+
✓ Docker Desktop
✓ kubectl (optional, for K8s work)
✓ Git

# Optional (recommended)
✓ VS Code + Go extension
✓ MongoDB Compass (database GUI)
✓ Postman / Insomnia (API testing)
```

### Clone & Setup

```bash
# Clone repository
git clone https://github.com/leetgaming-pro/replay-api.git
cd replay-api

# Install dependencies
go mod download

# Verify build
go build ./cmd/rest-api
```

**Expected output:**
```
Successfully built: rest-api
```

---

## 📁 Project Structure

```
replay-api/
├── cmd/                          # Application entrypoints
│   ├── rest-api/                 # HTTP API server
│   ├── async-api/                # Background workers
│   └── cli/                      # CLI tools
│
├── pkg/                          # Application code
│   ├── domain/                   # Domain layer (business logic)
│   │   ├── wallet/              # 💰 Wallet system (YOU'LL WORK HERE)
│   │   │   ├── entities/        # Domain entities (Wallet, Ledger)
│   │   │   ├── services/        # Business logic (TransactionCoordinator)
│   │   │   ├── ports/           # Interfaces (in/out ports)
│   │   │   └── value-objects/   # Immutable values (Amount, Currency)
│   │   ├── replay/              # Replay analysis
│   │   ├── matchmaking/         # Tournament & matching
│   │   └── iam/                 # Authentication & authorization
│   │
│   ├── infra/                    # Infrastructure layer
│   │   ├── db/mongodb/          # MongoDB repositories
│   │   ├── clients/             # External API clients
│   │   └── events/              # Event bus
│   │
│   └── app/                      # Application layer
│       ├── handlers/             # HTTP handlers
│       └── jobs/                 # Background jobs
│
├── test/                         # Tests (NO MOCKS!)
│   ├── smoke/                    # Fast unit tests
│   ├── integration/              # E2E tests with real services
│   └── blockchain/               # Hardhat EVM contracts
│
├── k8s/                          # Kubernetes manifests
│   └── base/                     # Base K8s resources
│
├── docs/                         # 📚 Documentation
│   ├── architecture/             # System design docs
│   ├── development/              # Dev guides (you are here!)
│   ├── deployment/               # Ops guides
│   └── testing/                  # Testing strategies
│
├── scripts/                      # Automation scripts
│   └── deploy-blue-green.sh     # K8s deployment script
│
└── docker-compose.test.yml       # Test infrastructure
```

---

## 🏗️ Architecture Crash Course

### Hexagonal Architecture (Ports & Adapters)

```
┌─────────────────────────────────────────────────────────────────┐
│                    Hexagonal Architecture                        │
│                                                                  │
│  ┌────────────────────────────────────────────────────────┐    │
│  │              Application Core (Domain)                  │    │
│  │                                                         │    │
│  │  ┌──────────────────────────────────────────────────┐ │    │
│  │  │  Entities (Pure Business Objects)                │ │    │
│  │  │  • UserWallet                                    │ │    │
│  │  │  • LedgerEntry                                   │ │    │
│  │  │  • IdempotentOperation                           │ │    │
│  │  └──────────────────────────────────────────────────┘ │    │
│  │                                                         │    │
│  │  ┌──────────────────────────────────────────────────┐ │    │
│  │  │  Services (Business Logic)                       │ │    │
│  │  │  • TransactionCoordinator (Saga Pattern)        │ │    │
│  │  │  • LedgerService (Double-Entry)                 │ │    │
│  │  │  • ReconciliationService                        │ │    │
│  │  └──────────────────────────────────────────────────┘ │    │
│  │                                                         │    │
│  │  ┌──────────────────────────────────────────────────┐ │    │
│  │  │  Ports (Interfaces)                              │ │    │
│  │  │  • IN Ports:  Use case interfaces (commands)    │ │    │
│  │  │  • OUT Ports: Repository interfaces             │ │    │
│  │  └──────────────────────────────────────────────────┘ │    │
│  └────────────────────────────────────────────────────────┘    │
│                                                                  │
│  ┌────────────────────────────────────────────────────────┐    │
│  │              Adapters (Infrastructure)                  │    │
│  │                                                         │    │
│  │  Inbound Adapters:         Outbound Adapters:         │    │
│  │  • HTTP REST API           • MongoDB Repository       │    │
│  │  • gRPC API                • Redis Cache              │    │
│  │  • Message Queue           • Blockchain Client        │    │
│  └────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘

KEY PRINCIPLES:
1. Domain logic NEVER depends on infrastructure
2. Dependencies point INWARD (towards domain)
3. Interfaces defined in domain, implemented in infrastructure
4. Easy to swap implementations (e.g., MongoDB → PostgreSQL)
```

### Request Flow Example

```
HTTP Request: POST /api/wallet/deposit
        │
        ▼
┌───────────────────────────────────────────────────────────┐
│  HTTP Handler (cmd/rest-api/handlers/wallet_handler.go)  │
│    • Parse request body                                   │
│    • Validate input                                       │
│    • Extract JWT user ID                                  │
└────────────────────┬──────────────────────────────────────┘
                     │
                     ▼
┌───────────────────────────────────────────────────────────┐
│  Application Service (pkg/domain/wallet/services/)       │
│    • WalletService.Deposit(command)                      │
│    • Coordinate business logic                           │
└────────────────────┬──────────────────────────────────────┘
                     │
                     ▼
┌───────────────────────────────────────────────────────────┐
│  Transaction Coordinator (Saga Pattern)                   │
│    Step 1: Record in ledger (double-entry)               │
│    Step 2: Update wallet balance                         │
│    Step 3: Persist to database                           │
│    • Automatic rollback on ANY failure                   │
└────────────────────┬──────────────────────────────────────┘
                     │
                     ▼
┌───────────────────────────────────────────────────────────┐
│  MongoDB Repository (pkg/infra/db/mongodb/)              │
│    • Save wallet document                                 │
│    • Save ledger entries (immutable)                     │
└───────────────────────────────────────────────────────────┘
```

---

## 🚀 Run Locally (15 Minutes)

### Option 1: With Docker Compose (Recommended)

```bash
# Start MongoDB + Hardhat
docker compose -f docker-compose.test.yml up -d

# Wait for services to be healthy (30 seconds)
docker compose -f docker-compose.test.yml ps

# Run the API server
go run cmd/rest-api/main.go
```

**Expected output:**
```
2025/11/25 14:30:00 INFO Starting Replay API server port=8080
2025/11/25 14:30:00 INFO Connected to MongoDB
2025/11/25 14:30:00 INFO Server listening on :8080
```

### Option 2: Minimal Setup (Go only)

```bash
# Build the binary
go build -o ./bin/replay-api ./cmd/rest-api

# Run with minimal config (uses in-memory mocks)
MONGO_URI="mongodb://localhost:27017" \
./bin/replay-api
```

### Test the API

```bash
# Health check
curl http://localhost:8080/health

# Expected: {"status":"ok","timestamp":"2025-11-25T14:30:00Z"}
```

---

## 🧪 Running Tests

### Smoke Tests (Fastest - 10 seconds)

```bash
# No external dependencies
go test -v -short -tags=smoke ./test/smoke/...
```

**What it tests:**
- Wallet creation & validation
- Balance calculations
- Ledger entry creation
- Business logic (no database)

### Integration Tests (2-3 minutes)

```bash
# Requires MongoDB + Hardhat running
docker compose -f docker-compose.test.yml up -d

# Run tests
MONGO_TEST_URI="mongodb://test:test123@localhost:27018/replay_test?authSource=admin" \
  go test -v -tags=integration ./test/integration/...
```

**What it tests:**
- Real MongoDB operations
- Saga pattern rollback
- Idempotency protection
- Double-entry accounting
- Blockchain interactions (Hardhat)

### All Tests

```bash
# Run everything
make -f Makefile.test test-all
```

---

## 💡 Your First Task: Add a Feature

Let's add a simple feature to understand the codebase flow.

### Task: Add "Get Wallet Balance" Endpoint

#### Step 1: Define the Use Case (Domain Layer)

**File:** `pkg/domain/wallet/ports/in/wallet_queries.go`

```go
package wallet_in

import (
    "context"
    "github.com/google/uuid"
    wallet_vo "github.com/psavelis/team-pro/replay-api/pkg/domain/wallet/value-objects"
)

type GetBalanceQuery struct {
    UserID   uuid.UUID
    Currency wallet_vo.Currency
}

type GetBalanceResult struct {
    Balance  wallet_vo.Amount
    Currency wallet_vo.Currency
}

type WalletQueryService interface {
    GetBalance(ctx context.Context, query GetBalanceQuery) (*GetBalanceResult, error)
}
```

#### Step 2: Implement in Service (Domain Layer)

**File:** `pkg/domain/wallet/services/wallet_service.go`

```go
func (s *WalletService) GetBalance(ctx context.Context, query wallet_in.GetBalanceQuery) (*wallet_in.GetBalanceResult, error) {
    // Fetch wallet
    wallet, err := s.walletRepo.FindByUserID(ctx, query.UserID)
    if err != nil {
        return nil, fmt.Errorf("wallet not found: %w", err)
    }

    // Get balance
    balance := wallet.GetBalance(query.Currency)

    return &wallet_in.GetBalanceResult{
        Balance:  balance,
        Currency: query.Currency,
    }, nil
}
```

#### Step 3: Add HTTP Handler (Application Layer)

**File:** `cmd/rest-api/handlers/wallet_handler.go`

```go
func (h *WalletHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
    // Extract user ID from JWT
    userID := GetUserIDFromContext(r.Context())

    // Parse currency from query param
    currencyStr := r.URL.Query().Get("currency")
    currency, err := wallet_vo.ParseCurrency(currencyStr)
    if err != nil {
        http.Error(w, "Invalid currency", http.StatusBadRequest)
        return
    }

    // Execute query
    result, err := h.walletService.GetBalance(r.Context(), wallet_in.GetBalanceQuery{
        UserID:   userID,
        Currency: currency,
    })
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalError)
        return
    }

    // Return JSON
    json.NewEncoder(w).Encode(map[string]interface{}{
        "balance":  result.Balance.ToFloat64(),
        "currency": result.Currency.String(),
    })
}
```

#### Step 4: Register Route

**File:** `cmd/rest-api/routing/routes.go`

```go
r.HandleFunc("/api/wallet/balance", walletHandler.GetBalance).Methods("GET")
```

#### Step 5: Write Tests

**File:** `test/integration/wallet_balance_test.go`

```go
func TestE2E_GetWalletBalance(t *testing.T) {
    // Setup
    wallet := createTestWallet(t)
    depositAmount := wallet_vo.NewAmount(100.00)
    wallet.Deposit(wallet_vo.CurrencyUSD, depositAmount)
    walletRepo.Save(ctx, wallet)

    // Execute
    result, err := walletService.GetBalance(ctx, wallet_in.GetBalanceQuery{
        UserID:   wallet.ResourceOwner.UserID,
        Currency: wallet_vo.CurrencyUSD,
    })

    // Assert
    require.NoError(t, err)
    assert.Equal(t, depositAmount.ToCents(), result.Balance.ToCents())
}
```

#### Step 6: Test Locally

```bash
# Start server
go run cmd/rest-api/main.go

# Test endpoint
curl "http://localhost:8080/api/wallet/balance?currency=USD" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Expected: {"balance":100.0,"currency":"USD"}
```

---

## 🔍 Code Navigation Tips

### Find Code Quickly

```bash
# Find all wallet-related files
find . -name "*wallet*" -type f

# Search for specific function
grep -r "ExecuteDeposit" pkg/

# Find all interfaces
find pkg/domain -name "*.go" | xargs grep "type.*interface"
```

### VS Code Shortcuts

```
F12         # Go to definition
Ctrl+Click  # Go to definition
Shift+F12   # Find all references
Ctrl+P      # Quick file open
Ctrl+Shift+F # Search across files
```

### Useful Go Commands

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Check imports
goimports -w .

# View documentation
go doc github.com/psavelis/team-pro/replay-api/pkg/domain/wallet/services.TransactionCoordinator
```

---

## 📚 Key Concepts to Master

### 1. Saga Pattern (Transaction Coordinator)

**Purpose:** Ensure financial integrity with automatic rollback

**Example:**
```go
saga := coordinator.NewSaga()

saga.AddStep(TransactionStep{
    Name: "RecordInLedger",
    Execute: func(ctx context.Context) error {
        // Create ledger entries
        return ledgerService.RecordDeposit(...)
    },
    Rollback: func(ctx context.Context) error {
        // Reverse ledger entries if later step fails
        return ledgerService.RecordRefund(...)
    },
})

saga.Execute(ctx) // Automatic rollback on ANY failure
```

### 2. Double-Entry Accounting

**Every transaction creates TWO entries:**

```
Deposit $100:
  Debit:  User Asset Account    +$100  (user has more money)
  Credit: Platform Liability     +$100  (platform owes user more)

Withdrawal $50:
  Credit: User Asset Account     -$50   (user has less money)
  Debit:  Platform Liability     -$50   (platform owes user less)

Rule: Debits ALWAYS equal Credits (accounting equation balances)
```

### 3. Value Objects

**Immutable objects representing concepts:**

```go
// Amount (always in cents to avoid floating point issues)
amount := wallet_vo.NewAmount(100.00) // $100.00
cents := amount.ToCents()             // 10000 cents

// Currency (type-safe enum)
currency := wallet_vo.CurrencyUSD
currency.String() // "USD"

// EVMAddress (validated Ethereum address)
addr, err := wallet_vo.NewEVMAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
```

### 4. Repository Pattern

**Abstract data access:**

```go
// Interface (domain layer)
type WalletRepository interface {
    Save(ctx context.Context, wallet *entities.UserWallet) error
    FindByID(ctx context.Context, id uuid.UUID) (*entities.UserWallet, error)
}

// Implementation (infrastructure layer)
type MongoDBWalletRepository struct { ... }

func (r *MongoDBWalletRepository) Save(ctx context.Context, wallet *entities.UserWallet) error {
    // MongoDB-specific code
}
```

---

## 🐛 Debugging Tips

### Enable Debug Logging

```bash
LOG_LEVEL=debug go run cmd/rest-api/main.go
```

### Inspect MongoDB

```bash
# Connect to MongoDB
docker exec -it replay-api-mongodb-test-1 mongosh -u test -p test123 --authenticationDatabase admin

# View wallets
use replay_test
db.wallets.find().pretty()

# View ledger entries
db.ledger_entries.find().pretty()
```

### Common Issues & Solutions

**Issue:** Tests fail with "connection refused"
**Solution:** Ensure Docker Compose is running:
```bash
docker compose -f docker-compose.test.yml ps
```

**Issue:** Build fails with "package not found"
**Solution:** Re-download dependencies:
```bash
go mod download
go mod tidy
```

**Issue:** Port 8080 already in use
**Solution:** Kill existing process:
```bash
lsof -ti:8080 | xargs kill -9
```

---

## 🎓 Learning Resources

### Internal Docs

1. [Architecture Overview](../architecture/OVERVIEW.md) - System design
2. [Wallet System Deep Dive](../architecture/WALLET_SYSTEM.md) - Financial logic
3. [Testing Strategy](../testing/TESTING_STRATEGY.md) - How we test
4. [K8s Deployment](../deployment/KUBERNETES.md) - Production deployment

### External Resources

- **Go Best Practices**: https://golang.org/doc/effective_go
- **Hexagonal Architecture**: https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html
- **Domain-Driven Design**: "Domain-Driven Design" by Eric Evans
- **Saga Pattern**: https://microservices.io/patterns/data/saga.html

---

## 🤝 Getting Help

### Slack Channels

- `#platform-engineering` - General development questions
- `#wallet-system` - Wallet-specific questions
- `#deployment-ops` - K8s and deployment
- `#help` - Stuck? Ask here!

### Code Review Process

1. Create feature branch: `git checkout -b feature/your-feature-name`
2. Make changes & commit: `git commit -m "feat: add get balance endpoint"`
3. Push & create PR: `git push origin feature/your-feature-name`
4. Request review from 2 team members
5. Address feedback
6. Merge after approval

### Coding Standards

- **Formatting**: `gofmt` (automatic in VS Code)
- **Linting**: `golangci-lint` (runs in CI)
- **Tests**: Required for all new features
- **Documentation**: Add godoc comments for public functions
- **Commit Messages**: Follow [Conventional Commits](https://www.conventionalcommits.org/)

---

## 🎯 30-Day Onboarding Checklist

### Week 1: Setup & Understanding
- [ ] Clone repository and run locally
- [ ] Read architecture docs
- [ ] Run all tests successfully
- [ ] Complete "Your First Task" exercise
- [ ] Submit first PR (docs improvement)

### Week 2: Feature Development
- [ ] Fix a "good first issue" bug
- [ ] Add a new endpoint (with tests)
- [ ] Participate in code review
- [ ] Deploy to staging environment

### Week 3: Advanced Topics
- [ ] Work on wallet system feature
- [ ] Write E2E test for new feature
- [ ] Set up local K8s (minikube/kind)
- [ ] Deploy to K8s cluster

### Week 4: Production Ready
- [ ] On-call training
- [ ] Production deployment observation
- [ ] Write runbook for your feature
- [ ] Present feature to team

---

## 🚀 You're Ready!

You now have everything you need to start contributing. Remember:

1. **Ask questions** - The team is here to help
2. **Read the docs** - They're comprehensive and up-to-date
3. **Write tests** - NO MOCKS, use real services
4. **Follow patterns** - Hexagonal architecture, DDD
5. **Have fun** - Building financial-grade systems is exciting!

**Welcome to the team! 🎉**

---

**Questions?** Post in `#platform-engineering` on Slack

**Need help?** Tag `@platform-team` for quick response

**Found a bug in this guide?** Create a PR to fix it!
