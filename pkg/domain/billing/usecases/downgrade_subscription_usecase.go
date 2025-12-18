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

// DowngradeSubscriptionUseCase handles subscription downgrades to lower-tier plans
type DowngradeSubscriptionUseCase struct {
	subscriptionReader billing_out.SubscriptionReader
	subscriptionWriter billing_out.SubscriptionWriter
	planReader         billing_out.PlanReader
	billableReader     billing_out.BillableEntryReader
}

// NewDowngradeSubscriptionUseCase creates a new DowngradeSubscriptionUseCase
func NewDowngradeSubscriptionUseCase(
	subscriptionReader billing_out.SubscriptionReader,
	subscriptionWriter billing_out.SubscriptionWriter,
	planReader billing_out.PlanReader,
	billableReader billing_out.BillableEntryReader,
) billing_in.DowngradeSubscriptionCommandHandler {
	return &DowngradeSubscriptionUseCase{
		subscriptionReader: subscriptionReader,
		subscriptionWriter: subscriptionWriter,
		planReader:         planReader,
		billableReader:     billableReader,
	}
}

// Exec executes the subscription downgrade
func (uc *DowngradeSubscriptionUseCase) Exec(ctx context.Context, cmd billing_in.DowngradeSubscriptionCommand) error {
	// 1. Authentication check
	resourceOwner := common.GetResourceOwner(ctx)
	if resourceOwner.UserID == uuid.Nil {
		return common.NewErrUnauthorized()
	}

	slog.InfoContext(ctx, "Processing subscription downgrade",
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

	// 5. Validate downgrade is valid (target plan must be lower tier)
	if !isDowngrade(currentPlan.Kind, targetPlan.Kind) {
		return fmt.Errorf("invalid downgrade: %s -> %s is not a downgrade", currentPlan.Kind, targetPlan.Kind)
	}

	// 6. Check if target plan is available
	if !targetPlan.IsAvailable || !targetPlan.IsActive {
		return fmt.Errorf("target plan is not available")
	}

	// 7. Check if current usage exceeds target plan limits
	validationErrors := uc.validateUsageLimits(ctx, currentSub, targetPlan)
	if len(validationErrors) > 0 {
		slog.WarnContext(ctx, "Downgrade blocked due to usage limits",
			"user_id", cmd.UserID,
			"errors", validationErrors,
		)
		return fmt.Errorf("downgrade blocked: current usage exceeds target plan limits: %v", validationErrors)
	}

	now := time.Now()

	// 8. Schedule downgrade for end of current billing period (or apply immediately for free plans)
	if currentPlan.IsFree {
		// Apply immediately for free plans
		currentSub.PlanID = cmd.PlanID
		currentSub.IsFree = targetPlan.IsFree
		currentSub.UpdatedAt = now
		currentSub.History = append(currentSub.History, billing_entities.SubscriptionHistory{
			Date:   now,
			Status: billing_entities.SubscriptionStatusRenewed,
			Reason: fmt.Sprintf("Downgraded from %s to %s (immediate)", currentPlan.Kind, targetPlan.Kind),
		})
	} else {
		// For paid plans, schedule downgrade at end of billing period
		// Store pending downgrade in Args
		if currentSub.Args == nil {
			currentSub.Args = make(map[string]interface{})
		}
		currentSub.Args["pending_downgrade_plan_id"] = cmd.PlanID.String()
		currentSub.Args["pending_downgrade_scheduled_at"] = now.Format(time.RFC3339)
		currentSub.UpdatedAt = now
		currentSub.History = append(currentSub.History, billing_entities.SubscriptionHistory{
			Date:   now,
			Status: billing_entities.SubscriptionStatusActive,
			Reason: fmt.Sprintf("Downgrade to %s scheduled for end of billing period", targetPlan.Kind),
		})
	}

	// 9. Persist changes
	_, err = uc.subscriptionWriter.Update(ctx, currentSub)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update subscription", "error", err)
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	slog.InfoContext(ctx, "Subscription downgrade processed",
		"user_id", cmd.UserID,
		"from_plan", currentPlan.Kind,
		"to_plan", targetPlan.Kind,
		"subscription_id", currentSub.ID,
		"immediate", currentPlan.IsFree,
	)

	return nil
}

// validateUsageLimits checks if current usage exceeds target plan limits
func (uc *DowngradeSubscriptionUseCase) validateUsageLimits(
	ctx context.Context,
	sub *billing_entities.Subscription,
	targetPlan *billing_entities.Plan,
) []string {
	var errors []string

	// Get current usage for each billable operation
	for operationKey, targetLimit := range targetPlan.OperationLimits {
		currentUsage := sub.GetUsage(operationKey)

		if currentUsage > targetLimit.Limit {
			errors = append(errors, fmt.Sprintf(
				"%s: current usage (%.2f) exceeds target limit (%.2f)",
				operationKey,
				currentUsage,
				targetLimit.Limit,
			))
		}
	}

	return errors
}

// isDowngrade checks if the target plan is a downgrade from the current plan
func isDowngrade(current, target billing_entities.PlanKindType) bool {
	planOrder := map[billing_entities.PlanKindType]int{
		billing_entities.PlanKindTypeFree:     1,
		billing_entities.PlanKindTypeStarter:  2,
		billing_entities.PlanKindTypePro:      3,
		billing_entities.PlanKindTypeTeam:     4,
		billing_entities.PlanKindTypeBusiness: 5,
		billing_entities.PlanKindTypeCustom:   6,
	}

	return planOrder[target] < planOrder[current]
}

// Ensure DowngradeSubscriptionUseCase implements the interface
var _ billing_in.DowngradeSubscriptionCommandHandler = (*DowngradeSubscriptionUseCase)(nil)

