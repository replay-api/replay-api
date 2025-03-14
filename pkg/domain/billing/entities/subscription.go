package billing_entities

import (
	"time"

	"github.com/google/uuid"
	common "github.com/replay-api/replay-api/pkg/domain"
)

type SubscriptionStatus string

const (
	SubscriptionStatusActive          SubscriptionStatus = "Active"
	SubscriptionStatusCanceled        SubscriptionStatus = "Canceled"
	SubscriptionStatusExpired         SubscriptionStatus = "Expired"
	SubscriptionStatusPending         SubscriptionStatus = "Pending"
	SubscriptionStatusRenewed         SubscriptionStatus = "Renewed"
	SubscriptionStatusPaymentRejected SubscriptionStatus = "PaymentRejected"
)

type Subscription struct {
	common.BaseEntity
	PlanID        uuid.UUID             `json:"plan_id" bson:"plan_id"`
	BillingPeriod BillingPeriodType     `json:"billing_period" bson:"billing_period"`
	StartAt       time.Time             `json:"start_at" bson:"start_at"`
	EndAt         *time.Time            `json:"end_at" bson:"end_at"`
	Status        SubscriptionStatus    `json:"status" bson:"status"`
	History       []SubscriptionHistory `json:"history" bson:"history"`
	IsFree        bool                  `json:"is_free" bson:"is_free"`
}

type SubscriptionHistory struct {
	Date   time.Time          `json:"date" bson:"date"`
	Status SubscriptionStatus `json:"status" bson:"status"`
	Reason string             `json:"reason" bson:"reason"`
}
