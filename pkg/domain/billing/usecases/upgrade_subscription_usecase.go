package billing_usecases

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	billing_out "github.com/replay-api/replay-api/pkg/domain/billing/ports/out"
)

// UpgradeSubscriptionUseCase handles subscription upgrades to higher-tier plans.
//
// This is a critical financial use case for subscription management that enables
// users to upgrade from their current plan (Free → Starter → Pro → Team → Business → Custom).
//
// Flow:
//  1. Authentication check - validates UserID is present in context
//  2. Current subscription retrieval - fetches user's active subscription
//  3. Current plan retrieval - gets the plan details for comparison
//  4. Target plan retrieval - gets the destination plan details
//  5. Upgrade validation - ensures target is higher tier than current
//  6. Availability check - confirms target plan is active and available
//  7. Subscription update - modifies plan reference and records history
//  8. Persistence - saves updated subscription
//
// Plan Hierarchy (lowest to highest):
//   - Free: Basic features, limited operations
//   - Starter: Entry-level paid tier
//   - Pro: Individual competitive player
//   - Team: Small team/squad management
//   - Business: Organization/enterprise features
//   - Custom: Tailored enterprise solutions
//
// Security:
//   - Requires authenticated context with valid UserID
//   - Validates upgrade path (no downgrades through this use case)
//
// Audit Trail:
//   - Records upgrade history with timestamp and reason
//   - Logs successful and failed upgrade attempts
//
// Dependencies:
//   - SubscriptionReader: Current subscription lookup
//   - SubscriptionWriter: Subscription updates
//   - PlanReader: Plan details retrieval
type UpgradeSubscriptionUseCase struct {
	subscriptionReader billing_out.SubscriptionReader
	subscriptionWriter billing_out.SubscriptionWriter
	planReader         billing_out.PlanReader
}

// NewUpgradeSubscriptionUseCase creates a new UpgradeSubscriptionUseCase
func NewUpgradeSubscriptionUseCase(
	subscriptionReader billing_out.SubscriptionReader,
	subscriptionWriter billing_out.SubscriptionWriter,
	planReader billing_out.PlanReader,
) billing_in.UpgradeSubscriptionCommandHandler {
	return &UpgradeSubscriptionUseCase{
		subscriptionReader: subscriptionReader,
		subscriptionWriter: subscriptionWriter,
		planReader:         planReader,
	}
}

// Exec executes the subscription upgrade
func (uc *UpgradeSubscriptionUseCase) Exec(ctx context.Context, cmd billing_in.UpgradeSubscriptionCommand) error {
	// 1. Authentication check
	resourceOwner := common.GetResourceOwner(ctx)
	if resourceOwner.UserID == uuid.Nil {
		return common.NewErrUnauthorized()
	}

	slog.InfoContext(ctx, "Processing subscription upgrade",
		"user_id", cmd.UserID,
		"target_plan_id", cmd.PlanID,
	)

	// 2. Get current subscription
	currentSub, err := uc.subscriptionReader.GetCurrentSubscription(ctx, resourceOwner)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get current subscription", "error", err)
		return fmt.Errorf("failed to get current subscription: %w", err)
	}

	if currentSub == nil {
		return fmt.Errorf("no active subscription found")
	}

	// 3. Get current plan
	currentPlan, err := uc.planReader.GetByID(ctx, currentSub.PlanID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get current plan", "error", err)
		return fmt.Errorf("failed to get current plan: %w", err)
	}

	// 4. Get target plan
	targetPlan, err := uc.planReader.GetByID(ctx, cmd.PlanID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get target plan", "error", err)
		return fmt.Errorf("failed to get target plan: %w", err)
	}

	if targetPlan == nil {
		return fmt.Errorf("target plan not found: %s", cmd.PlanID)
	}

	// 5. Validate upgrade is valid (target plan must be higher tier)
	if !isUpgrade(currentPlan.Kind, targetPlan.Kind) {
		return fmt.Errorf("invalid upgrade: %s -> %s is not an upgrade", currentPlan.Kind, targetPlan.Kind)
	}

	// 6. Check if target plan is available
	if !targetPlan.IsAvailable || !targetPlan.IsActive {
		return fmt.Errorf("target plan is not available")
	}

	now := time.Now()

	// 7. Update subscription
	currentSub.PlanID = cmd.PlanID
	currentSub.IsFree = targetPlan.IsFree
	currentSub.UpdatedAt = now
	currentSub.History = append(currentSub.History, billing_entities.SubscriptionHistory{
		Date:   now,
		Status: billing_entities.SubscriptionStatusRenewed,
		Reason: fmt.Sprintf("Upgraded from %s to %s", currentPlan.Kind, targetPlan.Kind),
	})

	// 8. Persist changes
	_, err = uc.subscriptionWriter.Update(ctx, currentSub)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update subscription", "error", err)
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	slog.InfoContext(ctx, "Subscription upgraded successfully",
		"user_id", cmd.UserID,
		"from_plan", currentPlan.Kind,
		"to_plan", targetPlan.Kind,
		"subscription_id", currentSub.ID,
	)

	return nil
}

// isUpgrade checks if the target plan is an upgrade from the current plan
func isUpgrade(current, target billing_entities.PlanKindType) bool {
	planOrder := map[billing_entities.PlanKindType]int{
		billing_entities.PlanKindTypeFree:     1,
		billing_entities.PlanKindTypeStarter:  2,
		billing_entities.PlanKindTypePro:      3,
		billing_entities.PlanKindTypeTeam:     4,
		billing_entities.PlanKindTypeBusiness: 5,
		billing_entities.PlanKindTypeCustom:   6,
	}

	return planOrder[target] > planOrder[current]
}

// Ensure UpgradeSubscriptionUseCase implements the interface
var _ billing_in.UpgradeSubscriptionCommandHandler = (*UpgradeSubscriptionUseCase)(nil)

