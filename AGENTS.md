# AI Agent Architecture Guidelines

This document provides architecture guidelines for AI coding agents working on the LeetGaming PRO platform. Following these patterns ensures consistency, maintainability, and proper separation of concerns.

## Repository Structure

```
replay-api/
├── cmd/                          # Application entry points
│   ├── rest-api/                 # REST API application
│   │   ├── controllers/          # HTTP handlers (adapters)
│   │   │   ├── command/          # Write operations (mutations)
│   │   │   └── query/            # Read operations (queries)
│   │   └── main.go               # Application bootstrap
│   ├── event-processor/          # Kafka event processor
│   └── seed/                     # Database seeding utility
├── pkg/
│   ├── domain/                   # Domain layer (business logic)
│   │   ├── {domain}/             # Bounded context (e.g., matchmaking, billing)
│   │   │   ├── entities/         # Domain entities and aggregates
│   │   │   ├── ports/            # Ports (interfaces)
│   │   │   │   ├── in/           # Inbound ports (use case interfaces)
│   │   │   │   └── out/          # Outbound ports (repository interfaces)
│   │   │   ├── usecases/         # Use case implementations
│   │   │   ├── services/         # Domain services (if needed)
│   │   │   └── value-objects/    # Value objects
│   │   └── common.go             # Shared domain concepts
│   └── infra/                    # Infrastructure layer
│       ├── db/                   # Database adapters
│       │   └── mongodb/          # MongoDB implementations
│       ├── kafka/                # Kafka client and events
│       ├── ioc/                  # Dependency injection container
│       └── http/                 # HTTP utilities
├── k8s/                          # Kubernetes manifests
│   ├── base/                     # Base kustomize resources
│   └── overlays/                 # Environment-specific overlays
│       └── local/                # Local development (Kind)
└── test/
    └── integration/              # Integration/E2E tests
```

## Hexagonal Architecture (Ports & Adapters)

This codebase follows **Hexagonal Architecture** (also known as Ports & Adapters). Understanding this pattern is CRITICAL.

### Core Principles

1. **Domain is the center**: Business logic lives in `pkg/domain/` and has NO dependencies on infrastructure
2. **Ports define contracts**: Interfaces in `ports/in/` (inbound) and `ports/out/` (outbound)
3. **Adapters implement ports**: Controllers are inbound adapters, repositories are outbound adapters
4. **Dependencies point inward**: Infrastructure depends on domain, never the reverse

### Layer Dependencies

```
Controllers (cmd/) → Use Cases (domain/usecases/) → Repositories (domain/ports/out/)
     ↓                        ↓                              ↑
  [Adapters]              [Domain]                    [Port Interfaces]
     ↓                        ↓                              ↑
Infrastructure (infra/) ─────────────────────────────────────┘
```

## CRITICAL: Controller Implementation Rules

### DO NOT access repositories directly from controllers

**WRONG** (Architecture Violation):
```go
// controller.go - WRONG
func (c *Controller) Handler(w http.ResponseWriter, r *http.Request) {
    // VIOLATION: Controller directly using repository
    result, err := c.repository.FindByID(ctx, id)
}
```

**CORRECT**:
```go
// controller.go - CORRECT
func (c *Controller) Handler(w http.ResponseWriter, r *http.Request) {
    // Use command handler (use case) resolved from DI container
    result, err := c.commandHandler.Exec(ctx, command)
}
```

### Controllers MUST:
1. Resolve command handlers from the DI container
2. Build command/query objects
3. Delegate to use case handlers
4. Transform results to HTTP responses

### Example: Proper Controller Implementation

```go
type MatchmakingController struct {
    container         container.Container
    joinQueueHandler  matchmaking_in.JoinMatchmakingQueueCommandHandler
}

func NewMatchmakingController(container container.Container) *MatchmakingController {
    ctrl := &MatchmakingController{container: container}

    // Resolve use case handlers from DI container
    if err := container.Resolve(&ctrl.joinQueueHandler); err != nil {
        slog.Error("Failed to resolve handler", "err", err)
    }

    return ctrl
}

func (ctrl *MatchmakingController) JoinQueueHandler(ctx context.Context) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. Parse request
        var req JoinQueueRequest
        json.NewDecoder(r.Body).Decode(&req)

        // 2. Build command
        cmd := matchmaking_in.JoinMatchmakingQueueCommand{
            PlayerID: req.PlayerID,
            GameID:   req.GameID,
            // ...
        }

        // 3. Execute via use case handler
        session, err := ctrl.joinQueueHandler.Exec(r.Context(), cmd)
        if err != nil {
            // Handle error
            return
        }

        // 4. Return response
        json.NewEncoder(w).Encode(response)
    }
}
```

## Resource Ownership & Authentication

All authenticated operations MUST use the `ResourceOwner` context pattern:

```go
// Get authenticated user from context
resourceOwner := common.GetResourceOwner(ctx)
if resourceOwner.UserID == uuid.Nil {
    return errors.New("Unauthorized")
}

// Use resourceOwner.UserID for ownership checks
if entity.OwnerID != resourceOwner.UserID {
    return errors.New("Forbidden")
}
```

### Context Keys
- `common.AuthenticatedKey` - boolean indicating authentication status
- `common.UserIDKey` - authenticated user's UUID
- `common.TenantIDKey` - tenant/organization UUID
- `common.ClientIDKey` - API client UUID

## Use Case Implementation

Use cases implement business logic and coordinate domain operations:

```go
// pkg/domain/matchmaking/usecases/join_matchmaking_queue.go

type JoinMatchmakingQueueUseCase struct {
    billableHandler billing_in.BillableOperationHandler
    sessionRepo     matchmaking_out.MatchmakingSessionRepository
    poolRepo        matchmaking_out.MatchmakingPoolRepository
}

func (uc *JoinMatchmakingQueueUseCase) Exec(
    ctx context.Context,
    cmd matchmaking_in.JoinMatchmakingQueueCommand,
) (*matchmaking_entities.MatchmakingSession, error) {
    // 1. Validate authentication
    if !common.IsAuthenticated(ctx) {
        return nil, errors.New("Unauthorized")
    }

    // 2. Business validation
    existingSessions, _ := uc.sessionRepo.GetByPlayerID(ctx, cmd.PlayerID)
    if hasActiveSession(existingSessions) {
        return nil, errors.New("already in queue")
    }

    // 3. Validate billing (if applicable)
    if err := uc.billableHandler.Validate(ctx, billingCmd); err != nil {
        return nil, err
    }

    // 4. Execute domain logic
    session := matchmaking_entities.NewMatchmakingSession(cmd)

    // 5. Persist via repository
    if err := uc.sessionRepo.Save(ctx, session); err != nil {
        return nil, err
    }

    // 6. Execute billing
    uc.billableHandler.Exec(ctx, billingCmd)

    return session, nil
}
```

## Dependency Injection

Dependencies are registered in `pkg/infra/ioc/container.go`:

```go
// Register use case as singleton
container.Singleton(func() matchmaking_in.JoinMatchmakingQueueCommandHandler {
    var billableHandler billing_in.BillableOperationHandler
    var sessionRepo matchmaking_out.MatchmakingSessionRepository
    var poolRepo matchmaking_out.MatchmakingPoolRepository

    container.Resolve(&billableHandler)
    container.Resolve(&sessionRepo)
    container.Resolve(&poolRepo)

    return matchmaking_usecases.NewJoinMatchmakingQueueUseCase(
        billableHandler,
        sessionRepo,
        poolRepo,
    )
})
```

## Kafka Event Publishing

For event-driven operations, use the Kafka EventPublisher:

```go
// Event types are defined in pkg/infra/kafka/events.go
const (
    EventTypeQueueJoined = "QUEUE_JOINED"
    EventTypeQueueLeft   = "QUEUE_LEFT"
    EventTypeLobbyCreated = "LOBBY_CREATED"
    // ...
)

// Publish events via EventPublisher
publisher := kafka.NewEventPublisher(client)
publisher.PublishQueueEvent(ctx, &kafka.QueueEvent{
    PlayerID:  playerID,
    GameType:  "cs2",
    EventType: kafka.EventTypeQueueJoined,
})
```

## Testing Guidelines

### Unit Tests
- Test use cases with mocked repositories
- Use `testify/mock` for mocking
- Place in `*_test.go` files next to source

### Integration Tests
- Use build tags: `//go:build integration || e2e`
- Test against real MongoDB/Kafka when available
- Skip gracefully when infrastructure unavailable
- Place in `test/integration/`

```go
//go:build integration || e2e

func TestMatchmaking_FullLifecycle(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    // Test implementation
}
```

## Code Quality Rules

### NO `any` types
Use proper type definitions, interfaces, or generics instead.

### NO hardcoded URLs
Use environment variables or configuration.

### NO mocks in production code
Only use mocks in tests. Production code should use real implementations.

### Avoid over-engineering
- Don't add features beyond what's requested
- Don't add unnecessary abstractions
- Three similar lines of code is better than a premature abstraction

## Frontend Integration (leetgaming-pro-web)

The frontend MUST use the SDK client (`lib/api/client.ts`), not hardcoded routes:

```typescript
// CORRECT
import { sdk } from '@/lib/api/client';
const squads = await sdk.squads.list();

// WRONG - Never hardcode API routes
fetch('/api/v1/squads');
```

## Git Commit Guidelines

- NO "Co-authored-by: Claude" in commits
- Use conventional commits: `feat:`, `fix:`, `test:`, `docs:`
- Keep commits focused and atomic

## Questions to Ask Before Implementing

1. Does this belong in domain or infrastructure?
2. Am I accessing repositories from a controller? (If yes, refactor to use case)
3. Is the user authenticated for this operation?
4. Does this operation require resource ownership validation?
5. Should this emit a Kafka event?
6. Am I using the SDK client in the frontend?
