package billing_entities

import (
	common "github.com/replay-api/replay-api/pkg/domain"
)

type BillingPeriodType string

const (
	Monthly  BillingPeriodType = "monthly"
	Yearly   BillingPeriodType = "yearly"
	Lifetime BillingPeriodType = "lifetime"
)

type Plan struct {
	common.BaseEntity
	Name            string                        `json:"name"`
	Description     string                        `json:"description"`
	Prices          map[BillingPeriodType][]Price `json:"prices"`
	OperationLimits []BillableItem                `json:"operation_limits"`
	IsFree          bool                          `json:"is_free"`
}

type BillableItem struct {
	OperationKey BillableOperationKey `json:"operation_key"`
	Name         string               `json:"name"`
	Description  string               `json:"description"`
	Limit        uint                 `json:"limit"`
}

type Price struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}
