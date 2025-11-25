# Billing Audit Report

## Audit Date
2025-11-25

## Objective
verify all monetizable operations have billing implemented

## Findings

### ✅ Usecases WITH Billing

**Squad Operations:**
- ✅ create_squad.go - `OperationTypeCreateSquadProfile`
- ✅ delete_squad.go - `OperationTypeDeleteSquadProfile`
- ✅ add_squad_member.go - `OperationTypeAddSquadMember`
- ✅ remove_squad_member.go - `OperationTypeRemoveSquadMember`
- ✅ update_squad_member_role.go - `OperationTypeUpdateSquadMemberRole`

**Player Operations:**
- ✅ create_player.go - `OperationTypeCreatePlayerProfile`
- ✅ delete_player.go - `OperationTypeDeletePlayerProfile`

### ❌ Usecases MISSING Billing

**Squad Operations:**
- ❌ update_squad.go - NO BILLING
  - modifies: name, symbol, description, logo, links
  - should use: `OperationTypeUpdateSquadProfile` (NEW)

**Player Operations:**
- ❌ update_player.go - NO BILLING
  - modifies: nickname, avatar, slug, roles, description
  - should use: `OperationTypeUpdatePlayerProfile` (NEW)

## Analysis

### Current Coverage
- 7 of 9 squad/player usecases have billing (78%)
- Missing: update operations

### Issue
update operations modify resources and should be billable per user requirement: "everything is billable"

profile updates could be rate-limited per plan:
- free plan: 1 update per month
- starter: 5 updates per month
- pro/team: unlimited updates

## Recommendations

### 1. Add Missing Operation Types

```go
// in billable_entry.go
OperationTypeUpdateSquadProfile  BillableOperationKey = "UpdateSquadProfile"
OperationTypeUpdatePlayerProfile BillableOperationKey = "UpdatePlayerProfile"
```

### 2. Add Billing to update_squad.go

```go
type UpdateSquadUseCase struct {
    billableOperationHandler billing_in.BillableOperationCommandHandler
    // ... other fields
}

func (uc *UpdateSquadUseCase) Exec(ctx, squadID, cmd) (*Squad, error) {
    // 1. auth
    // 2. fetch squad
    // 3. ownership check

    // 4. billing validation
    if err := uc.billableOperationHandler.Validate(ctx, billing_in.BillableOperationCommand{
        OperationID: billing_entities.OperationTypeUpdateSquadProfile,
        UserID:      common.GetResourceOwner(ctx).UserID,
        Amount:      1,
    }); err != nil {
        return nil, err
    }

    // 5. update squad
    squad, err := uc.SquadWriter.Update(ctx, &squad)
    if err != nil {
        return nil, err
    }

    // 6. billing execution
    uc.billableOperationHandler.Exec(ctx, billingCmd)

    // 7. history & logging
    return squad, nil
}
```

### 3. Add Billing to update_player.go

same pattern as update_squad.go using `OperationTypeUpdatePlayerProfile`

### 4. Update IoC Container

register `billableOperationHandler` in both UpdateSquadUseCase and UpdatePlayerUseCase constructors

### 5. Update BILLING.md

add new operation types to operations reference

## Priority

**HIGH** - required for monetization consistency

all create/modify operations must be billable to:
- enforce plan limits
- track usage analytics
- generate revenue

## Testing

after implementation, verify:
1. free plan users hit quota limit on updates
2. billing entries created in database
3. subscription usage tracked correctly

## Next Steps

1. add operation constants
2. refactor update usecases
3. update IoC registrations
4. add tests for quota enforcement
5. document in BILLING.md
