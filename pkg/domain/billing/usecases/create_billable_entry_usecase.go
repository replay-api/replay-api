package billing_usecases

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	shared "github.com/resource-ownership/go-common/pkg/common"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	billing_out "github.com/replay-api/replay-api/pkg/domain/billing/ports/out"
	iam_entities "github.com/replay-api/replay-api/pkg/domain/iam/entities"
	iam_out "github.com/replay-api/replay-api/pkg/domain/iam/ports/out"
)

type CreateBillableEntryUseCase struct {
	BillableOperationWriter billing_out.BillableEntryWriter
	BillableOperationReader billing_out.BillableEntryReader
	SubscriptionWriter      billing_out.SubscriptionWriter
	SubscriptionReader      billing_out.SubscriptionReader
	PlanReader              billing_out.PlanReader
	GroupReader             iam_out.GroupReader
}

func NewCreateBillableEntryUseCase(
	billableOperationWriter billing_out.BillableEntryWriter,
	billableOperationReader billing_out.BillableEntryReader,
	subscriptionWriter billing_out.SubscriptionWriter,
	subscriptionReader billing_out.SubscriptionReader,
	planReader billing_out.PlanReader,
	groupReader iam_out.GroupReader,
) billing_in.BillableOperationCommandHandler {
	return &CreateBillableEntryUseCase{
		BillableOperationWriter: billableOperationWriter,
		BillableOperationReader: billableOperationReader,
		SubscriptionWriter:      subscriptionWriter,
		SubscriptionReader:      subscriptionReader,
		PlanReader:              planReader,
		GroupReader:             groupReader,
	}
}

func (useCase *CreateBillableEntryUseCase) Exec(ctx context.Context, command billing_in.BillableOperationCommand) (*billing_entities.BillableEntry, *billing_entities.Subscription, error) {
	ctx, err := useCase.prepareUserContext(ctx, command.UserID)

	if err != nil {
		slog.ErrorContext(ctx, "unable to prepare user context for creating billing entry", "err", err, "command", command)
		return nil, nil, err
	}

	rxn := shared.GetResourceOwner(ctx)

	sub, err := useCase.SubscriptionReader.GetCurrentSubscription(ctx, rxn)
	if err != nil {
		return nil, nil, err
	}

	if sub == nil {
		freePlan, err := useCase.PlanReader.GetDefaultFreePlan(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "unable to get default free plan", "err", err)
			return nil, nil, fmt.Errorf("unable to get default free plan for user %v", rxn)
		}

		be := shared.NewRestrictedEntity(rxn)

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

	// TODO: chamar (grpc) pra verificar arbitrariamente termos de uso, regras de compliance, regra de seguranca/block etc...

	entry, err := useCase.BillableOperationWriter.Create(ctx, &billing_entities.BillableEntry{
		BaseEntity:     shared.NewRestrictedEntity(rxn),
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
	ctx, err := useCase.prepareUserContext(ctx, command.UserID)

	if err != nil {
		slog.ErrorContext(ctx, "unable to prepare user context for validating billing entry", "err", err, "command", command)
		return err
	}

	rxn := shared.GetResourceOwner(ctx)

	sub, err := useCase.SubscriptionReader.GetCurrentSubscription(ctx, rxn)
	if err != nil {
		slog.ErrorContext(ctx, "unable to get current subscription", "err", err, "rxn", rxn)
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
		plan, err := useCase.PlanReader.GetByID(ctx, sub.PlanID)
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

func (useCase *CreateBillableEntryUseCase) prepareUserContext(ctx context.Context, userID uuid.UUID) (context.Context, error) {
	userContext := context.WithValue(ctx, shared.UserIDKey, userID)
	groupAccountSearch := iam_entities.NewGroupAccountSearchByUser(userContext)

	groupAccount, err := useCase.GroupReader.Search(userContext, groupAccountSearch)
	if err != nil {
		slog.ErrorContext(ctx, "unable to get group account", "err", err, "userID", userID)
		return nil, fmt.Errorf("unable to fetch group account for user %v", userID)
	}

	if len(groupAccount) > 0 {
		userContext = context.WithValue(userContext, shared.GroupIDKey, groupAccount[0].ID)
	}

	return userContext, nil
}
