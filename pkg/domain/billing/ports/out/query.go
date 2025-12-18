package billing_out

import (
	"context"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
)

type SubscriptionReader interface {
	GetCurrentSubscription(ctx context.Context, rxn common.ResourceOwner) (*billing_entities.Subscription, error)
}

type PlanReader interface {
	common.Searchable[billing_entities.Plan]
	GetDefaultFreePlan(ctx context.Context) (*billing_entities.Plan, error)
	GetByID(ctx context.Context, id uuid.UUID) (*billing_entities.Plan, error)
	GetAvailablePlans(ctx context.Context) ([]*billing_entities.Plan, error)
}

type BillableEntryReader interface {
	GetEntriesBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) (map[billing_entities.BillableOperationKey][]billing_entities.BillableEntry, error)
}
