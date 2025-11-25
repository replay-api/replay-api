# JIRA Tickets - Architecture Refactoring & Fixes

## P0 - Critical (Blocks Production)

### BACKEND-50: [CRITICAL] Refactor PlayerProfileController to use existing usecases

**Type:** Bug / Technical Debt
**Priority:** P0 - Critical
**Story Points:** 5

**Description:**
PlayerProfileController violates SOLID principles by implementing business logic directly instead of delegating to existing usecases.

**Current State:**
- Controller directly manipulates repositories
- Bypasses `create_player.go` and `update_player.go` usecases
- No billing validation
- No history tracking
- No duplicate checking

**Acceptance Criteria:**
- [ ] CreatePlayerProfileHandler delegates to CreatePlayerUseCase
- [ ] UpdatePlayerProfileHandler delegates to UpdatePlayerUseCase
- [ ] GetPlayerProfileHandler remains as query (OK)
- [ ] DeletePlayerProfileHandler creates and uses DeletePlayerUseCase
- [ ] All player operations enforce billing
- [ ] All player operations record history
- [ ] Duplicate player creation prevented
- [ ] All existing tests pass
- [ ] Add unit tests for controller delegation

**Files Affected:**
- `cmd/rest-api/controllers/command/player_profile_controller.go`
- `pkg/domain/squad/usecases/create_player.go`
- `pkg/domain/squad/usecases/update_player.go`
- `pkg/domain/squad/usecases/delete_player.go` (NEW)
- `pkg/infra/ioc/container.go`

**Technical Notes:**
```go
// Controller should ONLY do this:
func (ctrl *PlayerProfileController) CreatePlayerProfileHandler() {
    var cmd squad_in.CreatePlayerProfileCommand
    json.NewDecoder(r.Body).Decode(&cmd)

    player, err := ctrl.createPlayerUseCase.Exec(r.Context(), cmd)

    json.NewEncoder(w).Encode(player)
}
```

---

### BACKEND-51: Create missing Squad member management usecases

**Type:** Story / Technical Debt
**Priority:** P0 - Critical
**Story Points:** 8

**Description:**
Squad controller operations bypass usecases for member management, violating SOLID principles and missing critical business logic.

**Missing Usecases:**
1. DeleteSquadUseCase
2. AddSquadMemberUseCase
3. RemoveSquadMemberUseCase
4. UpdateSquadMemberRoleUseCase

**Acceptance Criteria:**
- [ ] Create DeleteSquadUseCase with billing/history
- [ ] Create AddSquadMemberUseCase with validation/history
- [ ] Create RemoveSquadMemberUseCase with validation/history
- [ ] Create UpdateSquadMemberRoleUseCase with permission checking
- [ ] Update SquadController to use new usecases
- [ ] Register usecases in IoC container
- [ ] Add usecase unit tests
- [ ] Verify e2e squad flows working

**Files to Create:**
- `pkg/domain/squad/usecases/delete_squad.go`
- `pkg/domain/squad/usecases/add_member.go`
- `pkg/domain/squad/usecases/remove_member.go`
- `pkg/domain/squad/usecases/update_member_role.go`

---

### BACKEND-52: Fix IoC container registrations for all usecases

**Type:** Bug
**Priority:** P0 - Critical
**Story Points:** 3

**Description:**
IoC container resolver tests failing due to missing usecase registrations and potential circular dependencies.

**Failing Tests:**
- `TestResolveOnboardSteamUserCommand`
- `TestResolverSteamUserReader`
- `TestResolverSteamUserWriter`

**Acceptance Criteria:**
- [ ] Register CreatePlayerProfileUseCase in container
- [ ] Register UpdatePlayerProfileUseCase in container
- [ ] Register DeletePlayerProfileUseCase in container
- [ ] Register all squad usecases in container
- [ ] Resolve circular dependency issues
- [ ] All IoC container tests pass
- [ ] Verify dependency graph is acyclic

**Files Affected:**
- `pkg/infra/ioc/container.go`

---

### BACKEND-53: Fix MongoDB integration test infrastructure

**Type:** Bug / Infrastructure
**Priority:** P0 - Critical
**Story Points:** 3

**Description:**
MongoDB integration tests fail because Docker container cannot connect to host MongoDB.

**Failing Tests:**
- `TestMatchMetadataRepository_Search`
- `Test_Mongo_QueryBuilder`
- `Test_Search_Player_SuccessEmpty`
- `Test_GameEventSearch_SuccessEmpty`
- `Test_SearchSteamUserRealName_Success`

**Root Cause:**
Tests run in Docker container, try to connect to `127.0.0.1:37019` which is not accessible from container.

**Solution Options:**
1. Use `host.docker.internal` in connection string for tests
2. Use `--network host` in docker run
3. Use testcontainers-go to spin up MongoDB per test

**Acceptance Criteria:**
- [ ] All MongoDB integration tests pass
- [ ] Connection string configurable via env var
- [ ] Tests use real MongoDB (NO MOCKS principle)
- [ ] Document test setup in TESTING.md

**Files Affected:**
- `pkg/infra/db/mongodb/*_test.go`
- Test configuration/environment

---

## P1 - High Priority

### BACKEND-54: Fix Steam/Google OAuth e2e tests

**Type:** Bug / Infrastructure
**Priority:** P1 - High
**Story Points:** 5

**Description:**
Steam and Google OAuth onboarding e2e tests are failing.

**Failing Tests:**
- `Test_SteamOnboarding_Success`
- `Test_GoogleOnboarding_Success`

**Possible Causes:**
1. Missing/invalid API credentials in test environment
2. Network connectivity to OAuth providers
3. OAuth callback URLs not configured for test environment

**Acceptance Criteria:**
- [ ] Steam onboarding e2e test passes
- [ ] Google onboarding e2e test passes
- [ ] Test credentials properly configured
- [ ] Document OAuth test setup
- [ ] Consider test OAuth provider for CI/CD

**Files Affected:**
- `test/cmd/rest-api-test/http_api_test.go`
- Test environment configuration

---

### BACKEND-55: Create DeletePlayerProfileUseCase

**Type:** Story
**Priority:** P1 - High
**Story Points:** 3

**Description:**
Create missing DeletePlayerProfileUseCase to handle player deletion with proper business logic.

**Requirements:**
- Authentication check
- Ownership verification (via middleware)
- Billing operation record
- History tracking
- Cascade delete related data (or mark as deleted)
- Proper error handling and logging

**Acceptance Criteria:**
- [ ] UseCase created following squad pattern
- [ ] Billing operation recorded
- [ ] History entry created
- [ ] Related data handled (squads, matches, etc.)
- [ ] Unit tests added
- [ ] Integrated into controller
- [ ] Registered in IoC container

**Files to Create:**
- `pkg/domain/squad/usecases/delete_player.go`

---

### BACKEND-56: Validate Matchmaking Pool/Lobby functionality

**Type:** Task / Testing
**Priority:** P1 - High
**Story Points:** 5

**Description:**
Comprehensive validation of matchmaking system including pool management, lobby creation, player ready status, and match start flow.

**Test Scenarios:**
1. Join matchmaking queue
2. Leave matchmaking queue
3. Get session status
4. Get pool stats
5. Create lobby
6. Join lobby
7. Leave lobby
8. Set player ready
9. Start match
10. Cancel lobby

**Acceptance Criteria:**
- [ ] All matchmaking API endpoints tested
- [ ] WebSocket lobby updates working
- [ ] Pool stats accurate
- [ ] Match start flow complete
- [ ] Error cases handled properly
- [ ] Concurrent player scenarios tested
- [ ] Document any issues found

---

### BACKEND-57: Review and test Subscription/Pricing system

**Type:** Task / Testing
**Priority:** P1 - High
**Story Points:** 5

**Description:**
Comprehensive review and testing of subscription management and pricing/plans functionality.

**Areas to Validate:**
1. Plan definitions
2. Subscription creation
3. Subscription updates
4. Billing integration
5. Usage limits enforcement
6. Plan upgrades/downgrades
7. Subscription cancellation

**Acceptance Criteria:**
- [ ] All subscription APIs tested
- [ ] Plan limits enforced correctly
- [ ] Billing operations tracked
- [ ] Usage metering accurate
- [ ] Upgrade/downgrade flows work
- [ ] Cancellation properly handled
- [ ] Document any issues found

---

## P2 - Medium Priority

### BACKEND-58: Add comprehensive e2e tests for Player operations

**Type:** Story / Testing
**Priority:** P2 - Medium
**Story Points:** 5

**Description:**
Add end-to-end tests covering all player profile operations after architecture refactoring.

**Test Coverage:**
- Create player profile
- Get player profile
- Update player profile
- Delete player profile
- Search for players
- Player joining squad
- Player visibility rules

**Acceptance Criteria:**
- [ ] E2E tests for all CRUD operations
- [ ] E2E tests for search functionality
- [ ] E2E tests for squad interactions
- [ ] E2E tests for visibility enforcement
- [ ] All tests passing
- [ ] Tests follow NO MOCKS principle

---

### BACKEND-59: Add comprehensive e2e tests for Squad operations

**Type:** Story / Testing
**Priority:** P2 - Medium
**Story Points:** 5

**Description:**
Add end-to-end tests covering all squad operations including member management.

**Test Coverage:**
- Create squad
- Get squad
- Update squad
- Delete squad
- Add member
- Remove member
- Update member role
- Search for squads
- Squad visibility rules

**Acceptance Criteria:**
- [ ] E2E tests for all CRUD operations
- [ ] E2E tests for member management
- [ ] E2E tests for search functionality
- [ ] E2E tests for role permissions
- [ ] E2E tests for visibility enforcement
- [ ] All tests passing

---

### BACKEND-60: Document Usecase patterns and architecture guidelines

**Type:** Documentation
**Priority:** P2 - Medium
**Story Points:** 3

**Description:**
Create comprehensive documentation of usecase patterns, SOLID principles, and architecture guidelines to prevent future violations.

**Documentation Topics:**
1. Controller responsibilities (HTTP only)
2. Usecase pattern and structure
3. Repository pattern
4. Dependency injection guidelines
5. Error handling patterns
6. Logging standards
7. Testing strategies

**Deliverables:**
- [ ] ARCHITECTURE.md with patterns and examples
- [ ] Update CONTRIBUTING.md with guidelines
- [ ] Add code examples for each pattern
- [ ] Document decision records (ADRs)

---

### BACKEND-61: Add architecture validation tests

**Type:** Story / Testing
**Priority:** P2 - Medium
**Story Points:** 3

**Description:**
Add automated tests to validate architecture rules and prevent SOLID violations.

**Validation Rules:**
1. Controllers must NOT import repositories directly
2. Controllers must only call usecases
3. Usecases must handle all business logic
4. No circular dependencies
5. Proper layer separation (cmd -> domain -> infra)

**Acceptance Criteria:**
- [ ] Architecture tests added using go-archtest or similar
- [ ] Tests fail if controllers import repositories
- [ ] Tests validate dependency direction
- [ ] Tests run in CI/CD pipeline
- [ ] Document architecture rules

---

## Summary

**Total Tickets:** 12
- P0 Critical: 4 tickets (19 story points)
- P1 High: 5 tickets (23 story points)
- P2 Medium: 3 tickets (13 story points)

**Estimated Effort:** 55 story points (~7-8 sprints for 1 developer)

**Critical Path:**
1. BACKEND-52 (IoC) → BACKEND-50 (Player refactor) → BACKEND-51 (Squad usecases)
2. BACKEND-53 (MongoDB tests) → BACKEND-56 (Matchmaking validation)
3. BACKEND-54 (Auth tests) running in parallel

**Recommended Sprint 1 (Critical):**
- BACKEND-52: Fix IoC container
- BACKEND-50: Refactor PlayerProfileController
- BACKEND-53: Fix MongoDB tests
