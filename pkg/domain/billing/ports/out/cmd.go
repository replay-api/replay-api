package billing_out

import (
	"context"

	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
)

type BillableEntryWriter interface {
	// CreateBillableOperation(operationID billing_entities.BillableOperationKey, planID uuid.UUID, payableID *uuid.UUID, subscriptionID uuid.UUID, amount float64, args map[string]interface{}, rxn common.ResourceOwner) error
	Create(ctx context.Context, billableOperation *billing_entities.BillableEntry) (*billing_entities.BillableEntry, error)
}

type SubscriptionWriter interface {
	// CreateSubscription(planID uuid.UUID, args map[string]interface{}, rxn common.ResourceOwner) (*billing_entities.Subscription, error)
	Create(ctx context.Context, subscription *billing_entities.Subscription) (*billing_entities.Subscription, error)
	Update(ctx context.Context, subscription *billing_entities.Subscription) (*billing_entities.Subscription, error)
	Cancel(ctx context.Context, subscription *billing_entities.Subscription) (*billing_entities.Subscription, error)
}
