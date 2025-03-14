package billing_usecases

import (
	"context"
	"fmt"
	"log/slog"

	common "github.com/replay-api/replay-api/pkg/domain"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	billing_out "github.com/replay-api/replay-api/pkg/domain/billing/ports/out"
)

type CreateBillableEntryUseCase struct {
	BillableOperationWriter billing_out.BillableEntryWriter
	BillableOperationReader billing_out.BillableEntryReader
	SubscriptionWriter      billing_out.SubscriptionWriter
	SubscriptionReader      billing_out.SubscriptionReader
	PlanReader              billing_out.PlanReader
}

func (useCase *CreateBillableEntryUseCase) Exec(ctx context.Context, command billing_in.BillableOperationCommand) (*billing_entities.BillableEntry, *billing_entities.Subscription, error) {
	rxn := common.GetResourceOwner(ctx)
	sub, err := useCase.SubscriptionReader.GetCurrentSubscription(rxn)
	if err != nil {
		return nil, nil, err
	}

	if sub == nil {
		freePlan, err := useCase.PlanReader.GetDefaultFreePlan(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "unable to get default free plan", "err", err)
			return nil, nil, fmt.Errorf("unable to get default free plan for user %v", rxn)
		}

		be := common.NewRestrictedEntity(rxn)

		sub, err = useCase.SubscriptionWriter.Create(ctx, &billing_entities.Subscription{
			BaseEntity:    be,
			PlanID:        freePlan.ID,
			BillingPeriod: billing_entities.BillingPeriodLifetime,
			StartAt:       be.CreatedAt,
			EndAt:         nil,
			Status:        billing_entities.SubscriptionStatusActive,
			History: []billing_entities.SubscriptionHistory{
				{
					Date:   be.CreatedAt,
					Status: billing_entities.SubscriptionStatusActive,
					Reason: fmt.Sprintf("Subscription created (%s)", command.OperationID),
				},
			},
			IsFree: true,
		})

		if err != nil {
			slog.ErrorContext(ctx, "unable to create subscription", "err", err, "rxn", rxn)
			return nil, nil, fmt.Errorf("create billing entry for %s failed: unable to create subscription for user %v", command.OperationID, rxn.UserID)
		}
	}

	if err := useCase.Validate(ctx, command); err != nil {
		return nil, nil, err
	}

	entry, err := useCase.BillableOperationWriter.Create(ctx, &billing_entities.BillableEntry{
		BaseEntity:     common.NewRestrictedEntity(rxn),
		OperationID:    command.OperationID,
		PlanID:         sub.PlanID,
		SubscriptionID: sub.ID,
		Amount:         command.Amount,
		Args:           command.Args,
	})

	if err != nil {
		return nil, nil, err
	}

	return entry, sub, nil
}

func (useCase *CreateBillableEntryUseCase) Validate(ctx context.Context, command billing_in.BillableOperationCommand) error {
	rxn := common.GetResourceOwner(ctx)
	sub, err := useCase.SubscriptionReader.GetCurrentSubscription(rxn)
	if err != nil {
		return err
	}

	if sub == nil {
		freePlan, err := useCase.PlanReader.GetDefaultFreePlan(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "unable to get default free plan", "err", err, "rxn", rxn)
			return fmt.Errorf("unable to get default free plan for user %v", rxn.UserID)
		}
		if command.Amount > freePlan.OperationLimits[command.OperationID].Limit {
			return fmt.Errorf("the amount %s exceeds the limit %s for the free plan on operation %s", formatAmount(command.Amount), formatAmount(freePlan.OperationLimits[command.OperationID].Limit), command.OperationID)
		}
	} else {
		plan, err := useCase.PlanReader.GetPlanByID(ctx, sub.PlanID)
		if err != nil {
			slog.ErrorContext(ctx, "unable to get plan id", "err", err, "planID", sub.PlanID, "subscriptionID", sub.ID, "rxn", rxn)
			return fmt.Errorf("unable to get plan id for subscription ID %s", sub.ID)
		}

		entries, err := useCase.BillableOperationReader.GetEntriesBySubscriptionID(ctx, sub.ID)
		if err != nil {
			slog.ErrorContext(ctx, "unable to get entries by subscription id", "err", err)
			return fmt.Errorf("unable to get entries by subscription ID %s", sub.ID)
		}

		usedAmount := 0.0
		for _, entry := range entries[command.OperationID] {
			usedAmount += entry.Amount
		}

		if usedAmount+command.Amount > plan.OperationLimits[command.OperationID].Limit {
			return fmt.Errorf("the amount %s exceeds the limit %s for the current plan %s on operation %s", formatAmount(usedAmount+command.Amount), formatAmount(plan.OperationLimits[command.OperationID].Limit), plan.Name, command.OperationID)
		}
	}

	return nil
}

func formatAmount(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}
