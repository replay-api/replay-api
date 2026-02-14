package billing_usecases

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"

	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	billing_out "github.com/replay-api/replay-api/pkg/domain/billing/ports/out"
	payment_out "github.com/replay-api/replay-api/pkg/domain/payment/ports/out"
)

// CheckoutSubscriptionUseCase orchestrates the payment → subscription activation flow.
// After a successful payment, it creates or upgrades a subscription to the target plan.
type CheckoutSubscriptionUseCase struct {
	subscriptionReader billing_out.SubscriptionReader
	subscriptionWriter billing_out.SubscriptionWriter
	planReader         billing_out.PlanReader
	paymentRepo        payment_out.PaymentRepository
}

// Compile-time interface verification
var _ billing_in.CheckoutSubscriptionCommandHandler = (*CheckoutSubscriptionUseCase)(nil)

// NewCheckoutSubscriptionUseCase creates a new checkout subscription use case
func NewCheckoutSubscriptionUseCase(
	subReader billing_out.SubscriptionReader,
	subWriter billing_out.SubscriptionWriter,
	planReader billing_out.PlanReader,
	paymentRepo payment_out.PaymentRepository,
) billing_in.CheckoutSubscriptionCommandHandler {
	return &CheckoutSubscriptionUseCase{
		subscriptionReader: subReader,
		subscriptionWriter: subWriter,
		planReader:         planReader,
		paymentRepo:        paymentRepo,
	}
}

// Exec processes a checkout after payment succeeds: validates payment, creates/upgrades subscription
func (uc *CheckoutSubscriptionUseCase) Exec(ctx context.Context, cmd billing_in.CheckoutSubscriptionCommand) (*billing_entities.Subscription, error) {
	// 1. Auth check
	rxn := shared.GetResourceOwner(ctx)
	if rxn.UserID == uuid.Nil {
		return nil, shared.NewErrUnauthorized()
	}

	slog.InfoContext(ctx, "Processing checkout subscription",
		"user_id", rxn.UserID,
		"plan_id", cmd.PlanID,
		"payment_id", cmd.PaymentID,
		"billing_period", cmd.BillingPeriod,
	)

	// 2. Validate payment if provided
	if cmd.PaymentID != uuid.Nil {
		payment, err := uc.paymentRepo.FindByID(ctx, cmd.PaymentID)
		if err != nil {
			slog.ErrorContext(ctx, "Payment not found", "payment_id", cmd.PaymentID, "error", err)
			return nil, fmt.Errorf("payment not found: %w", err)
		}

		if payment.Status != "succeeded" && payment.Status != "processing" {
			return nil, fmt.Errorf("payment is not in a valid state for checkout: %s", payment.Status)
		}

		slog.InfoContext(ctx, "Payment validated for checkout",
			"payment_id", payment.ID,
			"payment_status", payment.Status,
			"payment_amount", payment.Amount,
		)
	}

	// 3. Get target plan
	targetPlan, err := uc.planReader.GetByID(ctx, cmd.PlanID)
	if err != nil {
		slog.ErrorContext(ctx, "Target plan not found", "plan_id", cmd.PlanID, "error", err)
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	if !targetPlan.IsAvailable {
		return nil, fmt.Errorf("plan %s is not available", targetPlan.Name)
	}

	// 4. Determine billing period
	billingPeriod := billing_entities.BillingPeriodMonthly
	switch cmd.BillingPeriod {
	case "yearly":
		billingPeriod = billing_entities.BillingPeriodYearly
	case "quarterly":
		// quarterly maps to monthly in the domain (3-month billing cycle)
		billingPeriod = billing_entities.BillingPeriodMonthly
	}

	// 5. Check for existing subscription
	currentSub, err := uc.subscriptionReader.GetCurrentSubscription(ctx, rxn)
	if err == nil && currentSub != nil && currentSub.Status == billing_entities.SubscriptionStatusActive {
		// Upgrade existing subscription
		slog.InfoContext(ctx, "Upgrading existing subscription",
			"subscription_id", currentSub.ID,
			"current_plan_id", currentSub.PlanID,
			"target_plan_id", cmd.PlanID,
		)

		currentSub.PlanID = cmd.PlanID
		currentSub.BillingPeriod = billingPeriod
		currentSub.IsFree = targetPlan.IsFree
		currentSub.Status = billing_entities.SubscriptionStatusActive
		currentSub.History = append(currentSub.History, billing_entities.SubscriptionHistory{
			Date:   time.Now().UTC(),
			Status: billing_entities.SubscriptionStatusActive,
			Reason: fmt.Sprintf("Upgraded to %s via checkout (payment: %s)", targetPlan.Name, cmd.PaymentID),
		})

		// Set billing period end date
		now := time.Now().UTC()
		var endAt time.Time
		switch cmd.BillingPeriod {
		case "yearly":
			endAt = now.AddDate(1, 0, 0)
		case "quarterly":
			endAt = now.AddDate(0, 3, 0)
		default:
			endAt = now.AddDate(0, 1, 0)
		}
		currentSub.EndAt = &endAt

		updated, err := uc.subscriptionWriter.Update(ctx, currentSub)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to update subscription", "error", err)
			return nil, fmt.Errorf("failed to update subscription: %w", err)
		}

		slog.InfoContext(ctx, "Subscription upgraded successfully",
			"subscription_id", updated.ID,
			"plan_name", targetPlan.Name,
		)

		return updated, nil
	}

	// 6. Create new subscription
	now := time.Now().UTC()
	var endAt time.Time
	switch cmd.BillingPeriod {
	case "yearly":
		endAt = now.AddDate(1, 0, 0)
	case "quarterly":
		endAt = now.AddDate(0, 3, 0)
	default:
		endAt = now.AddDate(0, 1, 0)
	}

	subscription := &billing_entities.Subscription{
		BaseEntity:    shared.NewPrivateEntity(rxn),
		PlanID:        cmd.PlanID,
		BillingPeriod: billingPeriod,
		StartAt:       now,
		EndAt:         &endAt,
		Status:        billing_entities.SubscriptionStatusActive,
		IsFree:        targetPlan.IsFree,
		Args:          cmd.Args,
		History: []billing_entities.SubscriptionHistory{
			{
				Date:   now,
				Status: billing_entities.SubscriptionStatusActive,
				Reason: fmt.Sprintf("Created via checkout for plan %s (payment: %s)", targetPlan.Name, cmd.PaymentID),
			},
		},
	}

	// Set resource ownership is handled by NewPrivateEntity

	created, err := uc.subscriptionWriter.Create(ctx, subscription)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create subscription", "error", err)
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	slog.InfoContext(ctx, "Subscription created successfully",
		"subscription_id", created.ID,
		"plan_name", targetPlan.Name,
		"billing_period", cmd.BillingPeriod,
		"ends_at", endAt,
	)

	return created, nil
}
