package domain

import (
	"context"
	"log/slog"

	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
)

type BaseUseCase struct {
	billableOperationHandler billing_in.BillableOperationCommandHandler
}

func NewBaseUseCase(billableOperationHandler billing_in.BillableOperationCommandHandler) *BaseUseCase {
	return &BaseUseCase{
		billableOperationHandler: billableOperationHandler,
	}
}

func (uc *BaseUseCase) RequireAuthentication(ctx context.Context) error {
	isAuthenticated := ctx.Value(AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return NewErrUnauthorized()
	}
	return nil
}

func (uc *BaseUseCase) RequireOwnership(ctx context.Context, resourceOwner ResourceOwner) error {
	currentUser := GetResourceOwner(ctx)
	if resourceOwner.UserID != currentUser.UserID {
		return NewErrUnauthorized()
	}
	return nil
}

func (uc *BaseUseCase) ValidateBilling(ctx context.Context, operationType billing_entities.BillableOperationKey, amount int) error {
	billingCmd := billing_in.BillableOperationCommand{
		OperationID: operationType,
		UserID:      GetResourceOwner(ctx).UserID,
		Amount:      amount,
	}
	err := uc.billableOperationHandler.Validate(ctx, billingCmd)
	if err != nil {
		slog.ErrorContext(ctx, "Billing validation failed", "operation", operationType, "error", err)
		return err
	}
	return nil
}

func (uc *BaseUseCase) ExecuteBilling(ctx context.Context, operationType billing_entities.BillableOperationKey, amount int) {
	billingCmd := billing_in.BillableOperationCommand{
		OperationID: operationType,
		UserID:      GetResourceOwner(ctx).UserID,
		Amount:      amount,
	}
	err := uc.billableOperationHandler.Exec(ctx, billingCmd)
	if err != nil {
		slog.WarnContext(ctx, "Failed to execute billing", "operation", operationType, "error", err)
	}
}

type UseCaseOperation[T any] struct {
	OperationType   billing_entities.BillableOperationKey
	Amount          int
	RequireAuth     bool
	ValidateBilling bool
	ExecuteBilling  bool
	Execute         func(ctx context.Context) (T, error)
	LogMessage      string
	LogFields       map[string]interface{}
}

func (uc *BaseUseCase) ExecuteOperation[T any](ctx context.Context, op UseCaseOperation[T]) (T, error) {
	var zero T

	if op.RequireAuth {
		if err := uc.RequireAuthentication(ctx); err != nil {
			return zero, err
		}
	}

	if op.ValidateBilling {
		if err := uc.ValidateBilling(ctx, op.OperationType, op.Amount); err != nil {
			return zero, err
		}
	}

	result, err := op.Execute(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Operation failed", "operation", op.OperationType, "error", err)
		return zero, err
	}

	if op.ExecuteBilling {
		uc.ExecuteBilling(ctx, op.OperationType, op.Amount)
	}

	if op.LogMessage != "" {
		logArgs := []interface{}{op.LogMessage}
		for k, v := range op.LogFields {
			logArgs = append(logArgs, k, v)
		}
		slog.InfoContext(ctx, op.LogMessage, logArgs...)
	}

	return result, nil
}
