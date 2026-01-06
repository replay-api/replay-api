package usecase

import (
	"context"
	"log/slog"

	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
	billing_in "github.com/replay-api/replay-api/pkg/domain/billing/ports/in"
	shared "github.com/resource-ownership/go-common/pkg/common"
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
	isAuthenticated := ctx.Value(shared.AuthenticatedKey)
	if isAuthenticated == nil || !isAuthenticated.(bool) {
		return shared.NewErrUnauthorized()
	}
	return nil
}

func (uc *BaseUseCase) RequireOwnership(ctx context.Context, resourceOwner shared.ResourceOwner) error {
	currentUser := shared.GetResourceOwner(ctx)
	if resourceOwner.UserID != currentUser.UserID {
		return shared.NewErrUnauthorized()
	}
	return nil
}

func (uc *BaseUseCase) ValidateBilling(ctx context.Context, operationType billing_entities.BillableOperationKey, amount int) error {
	billingCmd := billing_in.BillableOperationCommand{
		OperationID: operationType,
		UserID:      shared.GetResourceOwner(ctx).UserID,
		Amount:      float64(amount),
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
		UserID:      shared.GetResourceOwner(ctx).UserID,
		Amount:      float64(amount),
	}
	_, _, err := uc.billableOperationHandler.Exec(ctx, billingCmd)
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

// ExecuteOperation is a generic function to execute use case operations with common patterns
func ExecuteOperation[T any](ctx context.Context, uc *BaseUseCase, op UseCaseOperation[T]) (T, error) {
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
