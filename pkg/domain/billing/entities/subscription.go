package billing_entities

import common "github.com/replay-api/replay-api/pkg/domain"

type SubscriptionStatus string

const (
	Active   SubscriptionStatus = "active"
	Canceled SubscriptionStatus = "canceled"
	Expired  SubscriptionStatus = "expired"
)

type Subscription struct {
	common.BaseEntity
	PlanID        string             `json:"plan_id"`
	BillingPeriod BillingPeriodType  `json:"billing_period"`
	Status        SubscriptionStatus `json:"status"`
}
