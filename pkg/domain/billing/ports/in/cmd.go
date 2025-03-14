package billing_in

import (
	"context"

	"github.com/google/uuid"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
)

type BillableOperationCommand struct {
	OperationID billing_entities.BillableOperationKey
	UserID      uuid.UUID
	Amount      float64
	Args        map[string]interface{}
}

type BillableOperationCommandHandler interface {
	Exec(ctx context.Context, command BillableOperationCommand) error
	Validate(ctx context.Context, command BillableOperationCommand) error
}

type CreateSubscriptionCommand struct {
	PlanID uuid.UUID
	Args   map[string]interface{}
}

type CreateSubscriptionCommandHandler interface {
	Exec(ctx context.Context, command CreateSubscriptionCommand) error
}

type GetFreeSubscriptionCommandHandler interface {
	Exec(ctx context.Context) (uuid.UUID, error)
}

type UpgradeSubscriptionCommand struct {
	GroupAccountID uuid.UUID
	UserID         uuid.UUID
	PlanID         uuid.UUID
	Args           map[string]interface{}
}

type UpgradeSubscriptionCommandHandler interface {
	Exec(ctx context.Context, command UpgradeSubscriptionCommand) error
}

type DowngradeSubscriptionCommand struct {
	GroupAccountID uuid.UUID
	UserID         uuid.UUID
	PlanID         uuid.UUID
	Args           map[string]interface{}
}

type DowngradeSubscriptionCommandHandler interface {
	Exec(ctx context.Context, command DowngradeSubscriptionCommand) error
}
