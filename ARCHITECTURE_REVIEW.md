# Architecture Review - SOLID Violations and Technical Debt

**Date:** 2025-11-25
**Status:** v1.0.0 merged ✅
**Reviewer:** Claude Code

## Executive Summary

Critical SOLID violations found in controllers that bypass existing usecases, violating:
- **Single Responsibility Principle (SRP)**: Controllers contain business logic
- **Dependency Inversion Principle (DIP)**: Controllers depend on concrete implementations
- **DRY Principle**: Business logic duplicated between controllers and usecases

## Critical Issues

### 1. Player Profile Controller SOLID Violations

**File:** `cmd/rest-api/controllers/command/player_profile_controller.go`

**Problem:**
- Controller implements CRUD logic directly
- Bypasses existing usecases: `create_player.go`, `update_player.go`
- Contains business rules that should be in domain layer
- No billing validation
- No history tracking
- No duplicate checking

**Existing Usecases (NOT BEING USED):**
- `pkg/domain/squad/usecases/create_player.go` - Full player creation with billing, validation, history
- `pkg/domain/squad/usecases/update_player.go` - Full player update logic

**Impact:**
- Business logic scattered across layers
- Cannot test business rules independently
- Billing not enforced for player operations
- Audit trail (history) missing
- Duplicate players can be created

**Required Fix:**
```go
// BEFORE (Current - WRONG)
func (ctrl *PlayerProfileController) CreatePlayerProfileHandler() {
    // Controller doing business logic directly
    player, err := createPlayerCommandHandler.Exec(...)
    // BUT createPlayerCommandHandler is just calling repository directly!
}

// AFTER (Correct)
func (ctrl *PlayerProfileController) CreatePlayerProfileHandler() {
    // Controller delegates to usecase
    player, err := ctrl.createPlayerUseCase.Exec(ctx, cmd)
    // UseCase handles: auth, billing, validation, creation, history
}
```

### 2. Squad Controller - Partially Correct

**File:** `cmd/rest-api/controllers/command/squad_controller.go`

**Status:** ✅ CORRECT - Uses `createSquadCommandHandler` which delegates to usecase

**BUT:** Missing usecases for:
- Get Squad
- Update Squad (uses usecase ✅)
- Delete Squad
- Add Member
- Remove Member
- Update Member Role

**Required:** Create missing usecases for all squad operations

### 3. Dependency Injection Issues

**File:** `pkg/infra/ioc/container.go`

**Problems:**
- IoC container resolver tests failing
- Circular dependency risks
- Missing registrations for usecases

### 4. Test Infrastructure Failures

#### MongoDB Integration Tests
**Status:** ❌ FAILING
**Reason:** Tests run in Docker container, cannot connect to host MongoDB on 127.0.0.1:37019
**Fix:** Use `host.docker.internal` or `--network host`

#### Steam/Google Auth E2E Tests
**Status:** ❌ FAILING
**Tests:**
- `Test_SteamOnboarding_Success`
- `Test_GoogleOnboarding_Success`

**Likely Causes:**
1. Missing/invalid API credentials in test environment
2. Network connectivity issues
3. Mock services not set up (we use REAL services per NO MOCKS principle)

## Architecture Compliance Matrix

| Component | Uses Usecase | Auth Check | Billing | History | Validation | Status |
|-----------|--------------|------------|---------|---------|------------|--------|
| Squad Create | ✅ | ✅ | ✅ | ✅ | ✅ | CORRECT |
| Squad Update | ✅ | ✅ | ❌ | ❌ | ✅ | PARTIAL |
| Squad Delete | ❌ | MW* | ❌ | ❌ | ❌ | NEEDS USECASE |
| Squad Get | ❌ | MW* | N/A | N/A | ✅ | OK (Read-only) |
| Player Create | ❌ | MW* | ❌ | ❌ | ❌ | **CRITICAL** |
| Player Update | ❌ | MW* | ❌ | ❌ | ❌ | **CRITICAL** |
| Player Delete | ❌ | MW* | ❌ | ❌ | ❌ | **CRITICAL** |
| Player Get | ❌ | MW* | N/A | N/A | ✅ | OK (Read-only) |

*MW = Middleware handles auth (recent addition)

## Required Actions

### Immediate (P0 - Critical)

1. **JIRA-001**: Refactor PlayerProfileController to use existing usecases
2. **JIRA-002**: Create missing squad usecases (Delete, Add/Remove/Update Member)
3. **JIRA-003**: Fix IoC container registrations for all usecases
4. **JIRA-004**: Fix MongoDB test infrastructure (host connectivity)

### High Priority (P1)

5. **JIRA-005**: Fix Steam/Google auth e2e tests
6. **JIRA-006**: Add Delete usecase for Player
7. **JIRA-007**: Validate all matchmaking pool/lobby functionality
8. **JIRA-008**: Review and test subscription/pricing system

### Medium Priority (P2)

9. **JIRA-009**: Add comprehensive e2e tests for all player operations
10. **JIRA-010**: Add comprehensive e2e tests for all squad operations
11. **JIRA-011**: Document usecase patterns and architecture guidelines
12. **JIRA-012**: Add architecture tests to prevent regression

## Affected Features

### ✅ Working (Based on architecture)
- Squad creation (uses usecase)
- Match metadata search
- Authentication middleware
- Resource ownership framework

### ❌ At Risk (Missing usecases/validation)
- Player profile CRUD operations
- Squad member management
- Billing enforcement for players
- Audit trails for players

### ⚠️ Needs Verification
- Matchmaking pool/lobby flow
- Tournament registration/management
- Subscription management
- Pricing/Plans functionality

## Testing Status

```
Domain Tests: ✅ PASSING
Integration Tests (MongoDB): ❌ FAILING (Infrastructure)
Integration Tests (IoC): ❌ FAILING (Missing registrations)
E2E Tests (Auth): ❌ FAILING (External API/Config)
E2E Tests (Player): ⚠️ UNKNOWN (Architecture issues prevent proper testing)
E2E Tests (Squad): ⚠️ PARTIAL
E2E Tests (Matchmaking): ⚠️ NEEDS REVIEW
```

## Recommendations

### 1. Controller Refactoring Pattern

All controllers MUST follow this pattern:

```go
type XController struct {
    container container.Container
    // Usecases only - NO repositories!
}

func (ctrl *XController) CreateHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. Parse request
        var cmd XCommand
        json.NewDecoder(r.Body).Decode(&cmd)

        // 2. Delegate to usecase
        result, err := ctrl.createUseCase.Exec(r.Context(), cmd)

        // 3. Handle response
        json.NewEncoder(w).Encode(result)
    }
}
```

### 2. Usecase Registration Pattern

All usecases MUST be registered in IoC container:

```go
// In pkg/infra/ioc/container.go
container.Singleton(func() squad_in.CreatePlayerProfileCommandHandler {
    return squad_usecases.NewCreatePlayerProfileUseCase(
        billableOperationHandler,
        playerWriter,
        playerReader,
        groupWriter,
        groupReader,
        historyWriter,
        mediaWriter,
    )
})
```

### 3. Test Infrastructure

- MongoDB: Use `testcontainers-go` or `--network host` for integration tests
- Auth: Use test credentials or mock OAuth providers (aligned with NO MOCKS for business logic)
- E2E: Ensure all external dependencies properly configured

## Next Steps

1. Create JIRA tickets for all identified issues
2. Implement PlayerProfileController refactoring (highest priority)
3. Fix test infrastructure
4. Verify all critical user flows end-to-end
5. Add architecture validation tests

## Notes

- ✅ v1.0.0 successfully merged
- ✅ Resource ownership middleware implemented
- ✅ Structured logging pattern established
- ❌ SOLID principles violated in player controllers
- ❌ Test infrastructure needs fixes
