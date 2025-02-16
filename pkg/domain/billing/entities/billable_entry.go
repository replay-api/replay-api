package billing_entities

import common "github.com/replay-api/replay-api/pkg/domain"

type BillableEntry struct {
	common.BaseEntity
	OperationID string  `json:"operation_id"`
	PlanID      string  `json:"plan_id"`
	Amount      float64 `json:"amount"`
}

type BillableOperationKey string
