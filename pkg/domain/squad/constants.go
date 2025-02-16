package squad_common

import billing_entities "github.com/replay-api/replay-api/pkg/domain/billing/entities"

const (
	OperationTypeSquadAmount         billing_entities.BillableOperationKey = "SquadAmount"
	OperationTypePlayerProfileAmount billing_entities.BillableOperationKey = "PlayerProfileAmount"
)
