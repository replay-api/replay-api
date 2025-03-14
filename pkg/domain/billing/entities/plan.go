package billing_entities

import (
	common "github.com/replay-api/replay-api/pkg/domain"
)

type BillingPeriodType string

const (
	BillingPeriodMonthly  BillingPeriodType = "monthly"
	BillingPeriodYearly   BillingPeriodType = "yearly"
	BillingPeriodLifetime BillingPeriodType = "lifetime"
)

type Plan struct {
	common.BaseEntity
	Name            string                                `json:"name" bson:"name"`
	Description     string                                `json:"description" bson:"description"`
	Prices          map[BillingPeriodType][]Price         `json:"prices" bson:"prices"`
	OperationLimits map[BillableOperationKey]BillableItem `json:"operation_limits" bson:"operation_limits"`
	IsFree          bool                                  `json:"is_free" bson:"is_free"`
}

type BillableItem struct {
	Name        string  `json:"name" bson:"name"`
	Description string  `json:"description" bson:"description"`
	Limit       float64 `json:"limit" bson:"limit"`
}

type Price struct {
	Amount   float64 `json:"amount" bson:"amount"`
	Currency string  `json:"currency" bson:"currency"`
}
