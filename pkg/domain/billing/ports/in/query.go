package billing_in

import (
	common "github.com/replay-api/replay-api/pkg/domain"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
)

type PlanReader interface {
	common.Searchable[billing_entities.Plan]
}

type SubscriptionReader interface {
	common.Searchable[billing_entities.Subscription]
}
