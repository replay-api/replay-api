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
	PlanID        uuid.UUID              `json:"plan_id" bson:"plan_id"`
	BillingPeriod BillingPeriodType      `json:"billing_period" bson:"billing_period"`
	StartAt       time.Time              `json:"start_at" bson:"start_at"`
	EndAt         *time.Time             `json:"end_at" bson:"end_at"`
	Status        SubscriptionStatus     `json:"status" bson:"status"`
	History       []SubscriptionHistory  `json:"history" bson:"history"`
	IsFree        bool                   `json:"is_free" bson:"is_free"`
	VoucherCode   string                 `json:"voucher_code" bson:"voucher_code"`
	Args          map[string]interface{} `json:"args" bson:"args"`

	usage float64 `json:"-" bson:"-"`
}

type SubscriptionHistory struct {
	Date   time.Time          `json:"date" bson:"date"`
	Status SubscriptionStatus `json:"status" bson:"status"`
	Reason string             `json:"reason" bson:"reason"`
}

func (s Subscription) GetID() uuid.UUID {
	return s.ID
}

func (s *Subscription) GetUsage(operationID BillableOperationKey) float64 {
	entries := s.BaseEntity.Includes["BillableEntry"].([]BillableEntry)

	s.usage = 0.0
	for _, entry := range entries {
		if entry.OperationID != operationID {
			continue
		}
		s.usage += entry.Amount
	}

	return s.usage
}

func (s *Subscription) GetUsageAndLimits(operationID BillableOperationKey) (float64, float64) {
	plan := s.BaseEntity.Includes["Plan"].(*Plan)

	usage := s.GetUsage(operationID)

	return usage, plan.OperationLimits[operationID].Limit
}

func (s *Subscription) Available(operationID BillableOperationKey) bool {
	plan := s.BaseEntity.Includes["Plan"].(*Plan)

	usage := s.usage

	if s.usage == 0 {
		usage = s.GetUsage(operationID)
	}

	return plan.OperationLimits[operationID].Limit > usage
}

func (s *Subscription) GetFeatures() []BillableOperationKey {
	plan := s.BaseEntity.Includes["Plan"].(*Plan)

	operations := make([]BillableOperationKey, len(plan.OperationLimits))
	count := 0
	for key := range plan.OperationLimits {
		operations[count] = key
		count++
	}

	return operations
}
