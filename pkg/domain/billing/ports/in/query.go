package billing_in

import (
	shared "github.com/resource-ownership/go-common/pkg/common"
	billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"
)

type PlanReader interface {
	shared.Searchable[billing_entities.Plan]
}

type SubscriptionReader interface {
	shared.Searchable[billing_entities.Subscription]
}
