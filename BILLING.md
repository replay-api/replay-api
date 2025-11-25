# Billing Architecture

## Overview

Every monetizable operation in the API must be billed. The billing system follows SOLID principles and is integrated into all usecases through the `BaseUseCase`.

## Billable Operations

### Always Billable (Creation/Modification)
- Create squad, player profile, tournament
- Upload replay files
- Add/remove squad members
- Create match-making queue
- Mint/list/swap assets
- Create store items
- Tournament operations
- Feature flags (API access, webhooks, analytics)

### Never Billable (Read-Only)
- GET requests (search, list, view)
- Authentication operations
- Reading public data

### Rule of Thumb
If an operation:
- **Creates** a resource → Billable
- **Modifies** important data → Billable
- **Deletes** a resource → May be billable
- **Reads** data → Not billable

## Architecture

### Interfaces

```go
type BillableOperationCommandHandler interface {
    Validate(ctx context.Context, cmd BillableOperationCommand) error
    Exec(ctx context.Context, cmd BillableOperationCommand) error
}
```

**Two-Phase Billing:**
1. `Validate()` - Before operation (checks quota, subscription limits)
2. `Exec()` - After operation (records usage)

### BaseUseCase Integration

```go
type BaseUseCase struct {
    billableOperationHandler billing_in.BillableOperationCommandHandler
}

func (uc *BaseUseCase) ValidateBilling(ctx, operationType, amount) error
func (uc *BaseUseCase) ExecuteBilling(ctx, operationType, amount)
```

## Implementation Examples

### Example 1: Squad Creation

```go
func (uc *CreateSquadUseCase) Exec(ctx context.Context, cmd CreateSquadCommand) (*Squad, error) {
    // 1. Authentication
    if err := uc.BaseUseCase.RequireAuthentication(ctx); err != nil {
        return nil, err
    }

    // 2. Billing Validation (BEFORE operation)
    if err := uc.ValidateBilling(ctx, billing_entities.OperationTypeCreateSquadProfile, 1); err != nil {
        return nil, err // User quota exceeded or no subscription
    }

    // 3. Business Logic
    squad, err := uc.SquadWriter.Create(ctx, newSquad)
    if err != nil {
        return nil, err
    }

    // 4. Billing Execution (AFTER successful operation)
    uc.ExecuteBilling(ctx, billing_entities.OperationTypeCreateSquadProfile, 1)

    // 5. History & Logging
    uc.SquadHistoryWriter.Create(ctx, history)
    slog.InfoContext(ctx, "squad created", "squad_id", squad.ID)

    return squad, nil
}
```

### Example 2: Using BaseUseCase.ExecuteOperation

```go
func (uc *DeletePlayerUseCase) Exec(ctx context.Context, playerID uuid.UUID) error {
    return uc.ExecuteOperation(ctx, UseCaseOperation[*struct{}]{
        OperationType:   billing_entities.OperationTypeDeletePlayerProfile,
        Amount:          1,
        RequireAuth:     true,
        ValidateBilling: true,
        ExecuteBilling:  true,
        Execute: func(ctx context.Context) (*struct{}, error) {
            // Only business logic here
            return nil, uc.PlayerWriter.Delete(ctx, playerID)
        },
        LogMessage: "player profile deleted",
        LogFields:  map[string]interface{}{"player_id": playerID},
    })
}
```

## Billing Operations Reference

### Squad Operations
- `OperationTypeCreateSquadProfile` - Creating a squad
- `OperationTypeDeleteSquadProfile` - Deleting a squad
- `OperationTypeAddSquadMember` - Adding member
- `OperationTypeRemoveSquadMember` - Removing member
- `OperationTypeUpdateSquadMemberRole` - Updating roles
- `OperationTypeCreatePlayerProfile` - Creating player profile
- `OperationTypeDeletePlayerProfile` - Deleting player profile
- `OperationTypePlayersPerSquadAmount` - Squad size quota
- `OperationTypeSquadProfileBoostAmount` - Featured squad boost

### Replay Operations
- `OperationTypeUploadedReplayFiles` - Per file uploaded
- `OperationTypeStorageLimit` - Storage quota enforcement
- `OperationTypeMediaAlbum` - Media/clip creation

### Match-Making Operations
- `OperationTypeMatchMakingQueueAmount` - Queue slots
- `OperationTypeMatchMakingCreateAmount` - Creating lobbies
- `OperationTypeMatchMakingPriorityQueue` - Priority matchmaking
- `OperationTypeMatchMakingCreateFeaturedBoosts` - Featured lobby

### Tournament Operations (Plus+)
- `OperationTypeTournamentProfileAmount` - Creating tournaments
- `OperationTypeMatchMakingTournamentAmount` - Tournament matches
- `OperationTypeFeatureTournament` - Tournament feature access

### Web3 Operations
- `OperationTypeMintAssetAmount` - Minting NFTs
- `OperationTypeListAssetAmount` - Listing on marketplace
- `OperationTypeSwapAssetAmount` - Asset swaps
- `OperationTypeStakeAssetAmount` - Staking operations

### Wallet Operations (Plus+)
- `OperationTypeFeatureGlobalWallet` - Wallet feature access
- `OperationTypeGlobalWalletLimit` - Wallet quota
- `OperationTypeGlobalPaymentProcessingAmount` - Payment processing fees
- `OperationTypeSquadPayout` - Squad payouts

### Store Operations
- `OperationTypeStoreAmount` - Store items
- `OperationTypeStoreBoostAmount` - Featured items
- `OperationTypeStoreVoucherAmount` - Vouchers
- `OperationTypeStorePromotionAmount` - Promotions
- `OperationTypeFeatureStoreCatalogIntegration` - Store integration

### Betting Operations (Plus+)
- `OperationTypeFeatureBetting` - Betting feature access
- `OperationTypeBettingCreateAmount` - Creating bets
- `OperationTypeBettingStakeAmount` - Placing bets
- `OperationTypeBettingPayoutAmount` - Payouts

### Engagement Features
- `OperationTypeFeatureAnalytics` - Analytics access
- `OperationTypeFeatureChatAndComments` - Chat/comments
- `OperationTypeFeatureForum` - Forum access
- `OperationTypeFeaturedBlogPostAmount` - Featured blog posts
- `OperationTypeFeatureWiki` - Wiki access
- `OperationTypeFeaturePoll` - Polls
- `OperationTypeFeatureCalendarIntegration` - Calendar integration
- `OperationTypeFeatureAchievements` - Achievements system

### Platform Features (Plus+)
- `OperationTypeNotifications` - Notification system
- `OperationTypeCustomEmails` - Custom email templates
- `OperationTypePremiumSupport` - Premium support access
- `OperationTypeDiscordIntegration` - Discord integration
- `OperationTypeFeatureAPI` - API access
- `OperationTypeFeatureWebhooks` - Webhook integrations
- `OperationTypeExclusiveContent` - Exclusive content access
- `OperationTypeAdFreeExperience` - Ad-free experience
- `OperationTypeEarlyAccessFeatures` - Early access features

## Best Practices

### 1. Always Validate Before Execute

```go
// ✅ CORRECT
if err := uc.ValidateBilling(ctx, operationType, 1); err != nil {
    return nil, err
}
result, err := uc.DoOperation(ctx)
if err != nil {
    return nil, err
}
uc.ExecuteBilling(ctx, operationType, 1)

// ❌ WRONG - Validate after operation
result, err := uc.DoOperation(ctx)
if err := uc.ValidateBilling(ctx, operationType, 1); err != nil {
    // Too late! Operation already executed
}
```

### 2. Only Bill on Success

```go
// ✅ CORRECT - Bill after successful operation
result, err := uc.DoOperation(ctx)
if err != nil {
    return nil, err // Don't bill on failure
}
uc.ExecuteBilling(ctx, operationType, 1)

// ❌ WRONG - Bill even on failure
result, err := uc.DoOperation(ctx)
uc.ExecuteBilling(ctx, operationType, 1) // Bills even if operation failed!
```

### 3. Use Correct Amount

```go
// ✅ CORRECT - Amount matches resource count
for _, file := range files {
    uc.ValidateBilling(ctx, OperationTypeUploadedReplayFiles, 1)
}

// ❌ WRONG - Bill once for multiple resources
uc.ValidateBilling(ctx, OperationTypeUploadedReplayFiles, 1) // Should be len(files)
```

### 4. Don't Bill Read Operations

```go
// ✅ CORRECT - No billing for reads
func (uc *SearchSquadsUseCase) Exec(ctx, search) ([]Squad, error) {
    return uc.SquadReader.Search(ctx, search)
    // No billing!
}

// ❌ WRONG - Billing reads
func (uc *GetSquadUseCase) Exec(ctx, id) (*Squad, error) {
    uc.ValidateBilling(ctx, OperationTypeReadSquad, 1) // Don't do this!
    return uc.SquadReader.GetByID(ctx, id)
}
```

## Security

### Billing Handler Implementation

The billing handler MUST:
1. Verify subscription exists and is active
2. Check quota limits
3. Enforce plan restrictions
4. Record usage atomically
5. Handle race conditions (multiple concurrent requests)

### Context Security

Always use `GetResourceOwner(ctx).UserID` from authenticated context:

```go
billingCmd := billing_in.BillableOperationCommand{
    OperationID: operationType,
    UserID:      common.GetResourceOwner(ctx).UserID, // ✅ From authenticated context
    Amount:      amount,
}
```

Never accept UserID from client:

```go
// ❌ WRONG - UserID from request (security vulnerability!)
billingCmd := billing_in.BillableOperationCommand{
    UserID: cmd.UserID, // Attacker could bill another user!
}
```

## Testing

Billing tests must use real MongoDB (no mocks):

```go
func TestCreateSquad_BillingValidation(t *testing.T) {
    client := setupTestMongoDB(t)
    defer cleanupTestMongoDB(t, client)

    // Setup: User with free plan (1 squad limit)
    setupUserSubscription(t, ctx, userID, freePlanID)

    // Test: First squad succeeds
    squad1, err := usecase.Exec(ctx, CreateSquadCommand{})
    assert.NoError(t, err)

    // Test: Second squad fails (quota exceeded)
    squad2, err := usecase.Exec(ctx, CreateSquadCommand{})
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "quota exceeded")
}
```

## Adding New Billable Operations

1. **Add operation constant** to `billable_entry.go`:
```go
OperationTypeNewFeature BillableOperationKey = "NewFeature"
```

2. **Add billing to usecase**:
```go
func (uc *NewFeatureUseCase) Exec(ctx, cmd) (*Result, error) {
    if err := uc.ValidateBilling(ctx, billing_entities.OperationTypeNewFeature, 1); err != nil {
        return nil, err
    }

    result, err := uc.doOperation(ctx, cmd)
    if err != nil {
        return nil, err
    }

    uc.ExecuteBilling(ctx, billing_entities.OperationTypeNewFeature, 1)
    return result, nil
}
```

3. **Configure plan limits** in subscription service

4. **Add tests** for quota enforcement
