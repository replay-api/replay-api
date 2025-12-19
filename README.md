# LeetGaming Replay API

Production-grade backend API for the LeetGaming.PRO esports platform.

[![Build Status](https://github.com/leetgaming-pro/replay-api/workflows/CI/badge.svg)](https://github.com/leetgaming-pro/replay-api/actions)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Coverage](https://img.shields.io/badge/coverage-45%25-yellow)](https://codecov.io/gh/leetgaming-pro/replay-api)

---

## Overview

The Replay API powers the LeetGaming platform with:

| Feature | Description |
|---------|-------------|
| **Game Replay Analysis** | CS2, Valorant demo parsing and event extraction |
| **Skill-Based Matchmaking** | Glicko-2 rating system with queue management |
| **Tournament Management** | Bracket generation, scheduling, prize distribution |
| **Financial Operations** | Multi-asset wallets, payments, withdrawals |
| **Authentication** | Steam/Google OAuth, JWT, MFA/TOTP |

---

## Quick Start

### Prerequisites

- Go 1.23+
- Docker (for local development)
- MongoDB 7+ (or use Docker)

### Run Locally

```bash
# Start dependencies
docker compose -f docker-compose.test.yml up -d

# Run API
go run cmd/rest-api/main.go

# Verify
curl http://localhost:8080/health
```

### Run with Full Platform

```bash
# From root directory
cd ..
make local-up
```

---

## Commands

| Command | Description |
|---------|-------------|
| `go run cmd/rest-api/main.go` | Start API server |
| `go test ./...` | Run all tests |
| `go test -cover ./...` | Run tests with coverage |
| `go fmt ./...` | Format code |
| `golangci-lint run` | Run linter |
| `make build` | Build binary |

---

## Architecture

```
cmd/
├── rest-api/           # HTTP server entry point
│   ├── controllers/    # HTTP handlers (adapters)
│   ├── middlewares/    # Auth, logging, rate limiting
│   └── routing/        # Route definitions
└── event-processor/    # Kafka consumer

pkg/
├── domain/             # Business logic (core)
│   ├── auth/           # Authentication
│   ├── billing/        # Payments, subscriptions
│   ├── iam/            # Identity management
│   ├── matchmaking/    # Queue, sessions, pools
│   ├── replay/         # Replay files, events
│   ├── squad/          # Teams, members
│   ├── tournament/     # Brackets, matches
│   └── wallet/         # Balances, transactions
└── infra/              # Infrastructure (adapters)
    ├── db/mongodb/     # MongoDB repositories
    ├── kafka/          # Event streaming
    └── ioc/            # Dependency injection
```

**Pattern:** Hexagonal Architecture (Ports & Adapters)

```
HTTP Request → Controller → Use Case → Repository → MongoDB
                   ↓            ↓           ↑
               Adapter      Domain       Port
```

---

## API Reference

### Endpoints

| Domain | Endpoints | Auth |
|--------|-----------|------|
| Health | `GET /health` | No |
| Auth | `POST /auth/*`, `POST /onboarding/*` | Partial |
| Players | `GET/POST/PUT/DELETE /players/*` | Yes |
| Squads | `GET/POST/PUT/DELETE /squads/*` | Yes |
| Tournaments | `GET/POST/PUT/DELETE /tournaments/*` | Yes |
| Matchmaking | `POST /matchmaking/*` | Yes |
| Wallet | `GET /wallet/*`, `POST /payments/*` | Yes |
| Replays | `GET/POST /games/{game}/replays/*` | Partial |

### OpenAPI Documentation

- **Swagger UI:** http://localhost:8080/swagger
- **OpenAPI Spec:** [docs/swagger/openapi.yaml](./docs/swagger/openapi.yaml)

---

## Testing

```bash
# Unit tests
go test ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Integration tests (requires Docker)
go test -tags=integration ./test/integration/...

# Specific package
go test -v ./pkg/domain/matchmaking/...
```

### Coverage Targets

| Package | Current | Target |
|---------|---------|--------|
| matchmaking/value-objects | 98.7% | ✅ |
| wallet/usecases | 94.0% | ✅ |
| payment/usecases | 90.0% | ✅ |
| tournament/usecases | 88.6% | ✅ |
| matchmaking/usecases | 72.4% | 80% |
| auth/* | 0% | 80% |

---

## Configuration

### Environment Variables

```bash
# Server
PORT=8080
ENV=development

# MongoDB
MONGO_URI=mongodb://admin:password@localhost:27017/leetgaming?authSource=admin

# JWT
JWT_SECRET=your-secret-key-at-least-32-characters
JWT_EXPIRY=24h

# OAuth
STEAM_API_KEY=your-steam-api-key
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret

# Stripe
STRIPE_SECRET_KEY=sk_test_xxxxx
STRIPE_WEBHOOK_SECRET=whsec_xxxxx

# Kafka
KAFKA_BROKERS=localhost:9092
```

---

## Development Guidelines

### Code Standards

```go
// Use structured logging
slog.InfoContext(ctx, "processing request", "player_id", playerID)

// Wrap errors with context
if err := repo.Save(ctx, entity); err != nil {
    return fmt.Errorf("save player %s: %w", playerID, err)
}

// Always validate resource ownership
resourceOwner := common.GetResourceOwner(ctx)
if entity.OwnerID != resourceOwner.UserID {
    return ErrForbidden
}
```

### Controller Pattern

```go
// Controllers MUST use DI-resolved handlers
func (c *Controller) CreateHandler(ctx context.Context) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. Parse request
        var req CreateRequest
        json.NewDecoder(r.Body).Decode(&req)
        
        // 2. Build command
        cmd := domain.CreateCommand{...}
        
        // 3. Execute via use case (from DI)
        result, err := c.commandHandler.Exec(r.Context(), cmd)
        
        // 4. Return response
        json.NewEncoder(w).Encode(result)
    }
}
```

### Forbidden Patterns

```go
// ❌ NEVER access repository from controller
result, err := c.repository.FindByID(ctx, id)

// ❌ NEVER use any types
var data any

// ❌ NEVER skip ownership validation
```

---

## Deployment

### Kubernetes

```bash
# Apply manifests
kubectl apply -k k8s/overlays/local

# Check status
kubectl get pods -n leetgaming -l app=replay-api

# View logs
kubectl logs -n leetgaming -l app=replay-api
```

### Docker

```bash
# Build image
docker build -t replay-api:latest .

# Run container
docker run -p 8080:8080 \
  -e MONGO_URI="mongodb://..." \
  replay-api:latest
```

---

## Documentation

| Document | Location |
|----------|----------|
| Architecture | [docs/architecture/OVERVIEW.md](./docs/architecture/OVERVIEW.md) |
| Wallet System | [docs/architecture/WALLET_SYSTEM.md](./docs/architecture/WALLET_SYSTEM.md) |
| Onboarding | [docs/development/ONBOARDING.md](./docs/development/ONBOARDING.md) |
| Deployment | [docs/deployment/KUBERNETES.md](./docs/deployment/KUBERNETES.md) |
| AI Agent Guide | [AGENTS.md](./AGENTS.md) |

---

## Contributing

1. Read [AGENTS.md](./AGENTS.md) for architecture guidelines
2. Create feature branch from `main`
3. Write tests (required)
4. Run `golangci-lint run`
5. Submit pull request

---

## License

Proprietary - © 2025 LeetGaming Pro. All rights reserved.

---

*Maintained by the LeetGaming Platform Engineering Team*  
*Last Updated: December 19, 2025*
